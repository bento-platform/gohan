package main

import (
	"api/models"
	assemblyId "api/models/constants/assembly-id"
	"api/models/ingest/structs"
	"api/services"
	"api/utils"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/kelseyhightower/envconfig"
)

const (
	// Setup
	localDataDir = "data"
	baseUrl      = "https://en.wikipedia.org"
)

func main() {
	// Gather environment variables
	var cfg models.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	// TEMP: SECURITY RISK
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	//

	// Service Connections:
	// -- Elasticsearch
	es := utils.CreateEsConnection(&cfg)
	iz := services.NewIngestionService(es)
	iz.Init()

	// - create local data dirs
	if _, err := os.Stat(localDataDir); os.IsNotExist(err) {
		err := os.Mkdir(localDataDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	startTime := time.Now()

	// Start here on chromosome 1
	res, err := http.Get(fmt.Sprintf("%s/wiki/Category:Genes_on_human_chromosome_1", baseUrl))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Pluck out the navigation bar with all the chromosomes
	doc.Find("#mw-content-text > div.mw-parser-output > div.navbox > table > tbody > tr > td").Each(func(index int, item *goquery.Selection) {

		// Gather links for all chromosomes
		//var chromosomeWg sync.WaitGroup
		item.Find("div ul li").Each(func(index int, item *goquery.Selection) {

			// chromosomeWg.Add(1)
			// go func(_cwg *sync.WaitGroup) {
			// 	defer _cwg.Done()

			// link data
			chromTitle := item.Text()
			chromLinkTag := item.Find("a")
			chromLink, _ := chromLinkTag.Attr("href")

			//fmt.Printf("Chromosome #%d: %s - %s\n", index, chromTitle, chromLink)

			// process:
			// begin on an initial page, and verify the end of each page if
			// a "next page" link exists. if so, query that and continue processing
			// in the same manner
			for {
				chromPageRes, err := http.Get(fmt.Sprintf("%s%s", baseUrl, chromLink))
				if err != nil {
					fmt.Println(err)
					continue
				}
				defer chromPageRes.Body.Close()
				if res.StatusCode != 200 {
					fmt.Printf("status code error: %d %s\n", res.StatusCode, res.Status)
					continue
				}

				chromDoc, err := goquery.NewDocumentFromReader(chromPageRes.Body)
				if err != nil {
					fmt.Println(err)
					continue
				}

				processChromDoc(iz, chromTitle, chromDoc)

				hasNextPage := false
				chromDoc.Find("#mw-pages > a").EachWithBreak(func(index int, linkItem *goquery.Selection) bool {

					if strings.Contains(strings.ToLower(linkItem.Text()), "next page") {
						chromLink, _ = linkItem.Attr("href")
						hasNextPage = true

						// break
						return false
					}

					// continue loop
					return true
				})

				if !hasNextPage {
					break
				}
			}
			// gather "next page" link if available

			// }(&chromosomeWg)
		})
		// chromosomeWg.Wait()
	})

	// Done - display time lapse
	fmt.Printf("Process duration %s\n", time.Since(startTime))
}

func processChromDoc(iz *services.IngestionService, chromTitle string, chromDoc *goquery.Document) {
	// Pluck out sections with the links to all the genes on this page alphabetically
	var geneWg sync.WaitGroup
	chromDoc.Find(".mw-category-group").Each(func(index int, categorySectionItem *goquery.Selection) {
		go func(_gwg *sync.WaitGroup) {
			//defer _gwg.Done()

			// Skip this category if it's a wildcard
			isWildcard := false
			categorySectionItem.Find("h3").Each(func(index int, h3Item *goquery.Selection) {
				if h3Item.Text() == "*" {
					isWildcard = true
				}
			})
			if isWildcard {
				return
			}

			// Gather links for all chromosomes
			categorySectionItem.Find("ul li").Each(func(index int, item *goquery.Selection) {
				_gwg.Add(1)

				// link data
				geneTitle := item.Text()
				geneLinkTag := item.Find("a")
				geneLink, _ := geneLinkTag.Attr("href")

				// discover gene wiki page
				geneRes, err := http.Get(fmt.Sprintf("%s%s", baseUrl, geneLink))
				if err != nil {
					fmt.Println(err)
				}
				defer geneRes.Body.Close()
				if geneRes.StatusCode != 200 {
					fmt.Printf("status code error: %d %s\n", geneRes.StatusCode, geneRes.Status)
				}
				geneDoc, err := goquery.NewDocumentFromReader(geneRes.Body)
				if err != nil {
					fmt.Println(err)
				}

				// find assembly
				// TODO

				// find start and end positions
				var (
					aliasesRowElement *goquery.Selection
					aliasesValue      []string

					humanGeneLocationTableElement *goquery.Selection
					startHeaderElement            *goquery.Selection
					startValue                    int
					endHeaderElement              *goquery.Selection
					endValue                      int

					assemblyIdValue string
				)

				if geneDoc == nil {
					return
				}

				// Find nomenclature
				// - aliases
				// - symbol(s)
				geneDoc.Find("tr").Each(func(index int, rowElement *goquery.Selection) {
					if strings.Contains(rowElement.Text(), "Aliases") {
						aliasesRowElement = rowElement

						aliasesElement := aliasesRowElement.Find("td span").First()
						if aliasesElement != nil {
							aliasesValue = strings.Split(aliasesElement.Text(), ",")
						}
					}
				})
				// TODO: symbol(s)

				// Find gene location
				// - from start/end table
				// - "map position"
				geneDoc.Find("table").Each(func(index int, table *goquery.Selection) {
					if strings.Contains(table.Text(), "Gene location (Human)") {
						humanGeneLocationTableElement = table
						return
					}
				})

				if humanGeneLocationTableElement != nil {
					humanGeneLocationTableElement.Find("th").Each(func(index int, rowItemHeader *goquery.Selection) {
						if strings.Contains(rowItemHeader.Text(), "Start") {
							startHeaderElement = rowItemHeader
							return
						} else if strings.Contains(rowItemHeader.Text(), "End") {
							endHeaderElement = rowItemHeader
							return
						}
					})

					if startHeaderElement != nil {
						valueELement := startHeaderElement.SiblingsFiltered("td").Last()
						startClean := strings.ReplaceAll(strings.ReplaceAll(strings.Split(valueELement.Text(), "bp")[0], ",", ""), " ", "")
						startValue, _ = strconv.Atoi(startClean)
					}
					if endHeaderElement != nil {
						endValueELement := endHeaderElement.SiblingsFiltered("td").Last()
						endClean := strings.ReplaceAll(strings.ReplaceAll(strings.Split(endValueELement.Text(), "bp")[0], ",", ""), " ", "")
						endValue, _ = strconv.Atoi(endClean)
					}

				}
				// TODO: "map position"

				// Find Assembly
				// Assume the references provided, if any, containing an assembly id is
				// the assembly corresponding to the gene and its position
				geneDoc.Find("span.reference-text").EachWithBreak(func(index int, referenceListItem *goquery.Selection) bool {
					if strings.Contains(strings.ToLower(referenceListItem.Text()), "grch38") ||
						strings.Contains(strings.ToLower(referenceListItem.Text()), "grch37") ||
						strings.Contains(strings.ToLower(referenceListItem.Text()), "ncbi36") {

						// pluck out the link containing the text containing the assembly id
						// (usually the first one)
						refText := referenceListItem.Find("a").First()

						// split by colon to retrieve the assembly id
						assemblyIdValue = strings.Split(refText.Text(), ":")[0]

						// break
						return false
					}

					// keep looping
					return true
				})

				// store data
				// (temp : store to disk)
				chromosomeClean := strings.Replace(strings.Replace(chromTitle, ")", "", -1), "(", "", -1)
				chromosome, _ := strconv.Atoi(chromosomeClean)

				discoveredGene := &models.Gene{
					Name:         geneTitle,
					Nomenclature: aliasesValue,
					Chrom:        chromosome,
					AssemblyId:   assemblyId.CastToAssemblyId(assemblyIdValue),
					Start:        startValue,
					End:          endValue,
					SourceUrl:    fmt.Sprintf("%s%s", baseUrl, geneLink),
				}

				fmt.Println(discoveredGene)

				iz.GeneIngestionBulkIndexingQueue <- &structs.GeneIngestionQueueStructure{
					Gene:      discoveredGene,
					WaitGroup: _gwg,
				}

			})
		}(&geneWg)
	})
	geneWg.Wait()
}
