package main

import (
	//"fmt"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	// Setup
	localDataDir := "data"

	// - create local data dirs
	if _, err := os.Stat(localDataDir); os.IsNotExist(err) {
		err := os.Mkdir(localDataDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	baseUrl := "https://en.wikipedia.org"

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

			fmt.Printf("Chromosome #%d: %s - %s\n", index, chromTitle, chromLink)

			chromPageRes, err := http.Get(fmt.Sprintf("%s%s", baseUrl, chromLink))
			if err != nil {
				log.Fatal(err)
			}
			defer chromPageRes.Body.Close()
			if res.StatusCode != 200 {
				log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
			}
			chromDoc, err := goquery.NewDocumentFromReader(chromPageRes.Body)
			if err != nil {
				log.Fatal(err)
			}

			// Pluck out sections with the links to all the genes on this page alphabetically
			var geneWg sync.WaitGroup

			chromDoc.Find(".mw-category-group").Each(func(index int, categorySectionItem *goquery.Selection) {
				geneWg.Add(1)
				go func(_gwg *sync.WaitGroup) {
					defer _gwg.Done()

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

						// link data
						geneTitle := item.Text()
						geneLinkTag := item.Find("a")
						geneLink, _ := geneLinkTag.Attr("href")

						// discover gene wiki page
						geneRes, err := http.Get(fmt.Sprintf("%s%s", baseUrl, geneLink))
						if err != nil {
							log.Fatal(err)
						}
						defer geneRes.Body.Close()
						if res.StatusCode != 200 {
							log.Fatalf("status code error: %d %s", geneRes.StatusCode, res.Status)
						}
						geneDoc, err := goquery.NewDocumentFromReader(geneRes.Body)
						if err != nil {
							log.Fatal(err)
						}

						// find assembly
						// TODO

						// find start and end positions
						var humanGeneLocationTableElement *goquery.Selection
						var startHeaderElement *goquery.Selection
						var startValue string
						var endHeaderElement *goquery.Selection
						var endValue string

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
								startValue = valueELement.Text()
							}
							if endHeaderElement != nil {
								endValueELement := endHeaderElement.SiblingsFiltered("td").Last()
								endValue = endValueELement.Text()
							}

						}

						// store data
						// (temp : store to disk)
						chromosome := strings.Replace(strings.Replace(chromTitle, ")", "", -1), "(", "", -1)

						fmt.Printf("Chromosome #%s: Gene #%d: %s - %s\n", chromosome, index, geneTitle, geneLink)
						fmt.Printf("Start: %s\n", startValue)
						fmt.Printf("End: %s\n\n", endValue)

						var file *os.File
						thisGenePath := fmt.Sprintf("%s/%s.txt", localDataDir, geneTitle)
						if _, err := os.Stat(thisGenePath); os.IsNotExist(err) {
							file, err = os.Create(thisGenePath)
							if err != nil {
								fmt.Println(err)
								return
							}
						} else {
							file, err = os.OpenFile(thisGenePath, os.O_RDWR, 0755)
							if err != nil {
								log.Fatal(err)
							}
						}
						defer file.Close()

						writeText := fmt.Sprintf("Chromosome: %s\nStart: %s\nEnd: %s\nPath: %s", chromosome, startValue, endValue, geneLink)
						_, err = file.WriteString(writeText)
						if err != nil {
							fmt.Println(err)
							return
						}
					})
				}(&geneWg)
			})
			geneWg.Wait()

			// }(&chromosomeWg)
		})
		// chromosomeWg.Wait()
	})

}
