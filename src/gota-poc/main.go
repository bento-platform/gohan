package main

import (
	"api/models"
	"api/models/constants"
	assemblyId "api/models/constants/assembly-id"
	"api/models/constants/chromosome"
	"api/models/ingest/structs"
	"api/services"
	"api/utils"
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/kelseyhightower/envconfig"
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

	assemblyIdMap := map[constants.AssemblyId]string{
		assemblyId.GRCh38: "gencode.v38.annotation.gtf",
		assemblyId.GRCh37: "gencode.v19.annotation.gtf",
		// SKIP
		// assemblyId.NCBI36: "hg18",
		// assemblyId.NCBI35: "hg17",
		// assemblyId.NCBI34: "hg16",
	}
	assemblyIdGTFUrlMap := map[constants.AssemblyId]string{
		assemblyId.GRCh38: "http://ftp.ebi.ac.uk/pub/databases/gencode/Gencode_human/release_38/gencode.v38.annotation.gtf.gz",
		assemblyId.GRCh37: "http://ftp.ebi.ac.uk/pub/databases/gencode/Gencode_human/release_19/gencode.v19.annotation.gtf.gz",
		// SKIP
		// assemblyId.NCBI36: "",
		// assemblyId.NCBI35: "",
		// assemblyId.NCBI34: "",
	}

	var geneWg sync.WaitGroup

	for assId, fileName := range assemblyIdMap {
		// Read one file at a time

		gtfFile, err := os.Open(fileName)
		if err != nil {
			// log.Fatalf("failed to open file: %s", err)
			// Download the file
			fullURLFile := assemblyIdGTFUrlMap[assId]

			// Build fileName from fullPath
			fileURL, err := url.Parse(fullURLFile)
			if err != nil {
				log.Fatal(err)
			}
			path := fileURL.Path
			segments := strings.Split(path, "/")
			fileName = segments[len(segments)-1]

			// Create blank file
			file, err := os.Create(fileName)
			if err != nil {
				log.Fatal(err)
			}
			client := http.Client{
				CheckRedirect: func(r *http.Request, via []*http.Request) error {
					r.URL.Opaque = r.URL.Path
					return nil
				},
			}
			fmt.Printf("Downloading file %s ...\n", fileName)

			// Put content on file
			resp, err := client.Get(fullURLFile)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			size, err := io.Copy(file, resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			fmt.Printf("Downloaded a file %s with size %d\n", fileName, size)

			fmt.Printf("Unzipping %s...\n", fileName)
			gzipfile, err := os.Open(fileName)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			reader, err := gzip.NewReader(gzipfile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer reader.Close()

			newfilename := strings.TrimSuffix(fileName, ".gz")

			writer, err := os.Create(newfilename)

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			defer writer.Close()

			if _, err = io.Copy(writer, reader); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("Opening %s\n", newfilename)
			gtfFile, _ = os.Open(newfilename)

			fmt.Printf("Deleting %s\n", fileName)
			err = os.Remove(fileName)
			if err != nil {
				fmt.Println(err)
			}
		}

		defer gtfFile.Close()

		fileScanner := bufio.NewScanner(gtfFile)
		fileScanner.Split(bufio.ScanLines)

		fmt.Printf("Ingesting %s\n", string(assId))

		var (
			chromHeaderKey     = 0
			startKey           = 3
			endKey             = 4
			nameHeaderKeys     = []int{3}
			geneNameHeaderKeys []int
		)

		var columnsToPrint []string
		if assId == assemblyId.GRCh38 {
			// GRCh38 dataset has multiple name fields (name, name2) and
			// also includes gene name fields (geneName, geneName2)
			columnsToPrint = append(columnsToPrint, "#chrom", "chromStart", "chromEnd", "name", "name2", "geneName", "geneName2")
			nameHeaderKeys = append(nameHeaderKeys, 4)
			geneNameHeaderKeys = append(geneNameHeaderKeys, 5, 6)
		} else {
			columnsToPrint = append(columnsToPrint, "chrom", "txStart", "txEnd", "#name")
		}

		for fileScanner.Scan() {
			rowText := fileScanner.Text()
			if rowText[:2] == "##" {
				// Skip header rows
				continue
			}

			geneWg.Add(1)
			go func(rowText string, _chromHeaderKey int,
				_startKey int, _endKey int,
				_nameHeaderKeys []int, _geneNameHeaderKeys []int,
				_assId constants.AssemblyId,
				_gwg *sync.WaitGroup) {
				// fmt.Printf("row : %s\n", row)

				var (
					start    int
					end      int
					geneName string
				)

				rowSplits := strings.Split(rowText, "\t")

				// skip this row if it's not a gene row
				// i.e, if it's an exon or transcript
				if rowSplits[2] != "gene" {
					defer _gwg.Done()
					return
				}

				//clean chromosome
				chromosomeClean := strings.ReplaceAll(rowSplits[_chromHeaderKey], "chr", "")

				if !chromosome.IsValidHumanChromosome(chromosomeClean) {
					defer _gwg.Done()
					return
				}
				// http://ftp.ebi.ac.uk/pub/databases/gencode/Gencode_human/release_38/gencode.v38.annotation.gtf.gz

				// clean start/end
				chromStartClean := strings.ReplaceAll(strings.ReplaceAll(rowSplits[_startKey], ",", ""), " ", "")
				start, _ = strconv.Atoi(chromStartClean)

				chromEndClean := strings.ReplaceAll(strings.ReplaceAll(rowSplits[_endKey], ",", ""), " ", "")
				end, _ = strconv.Atoi(chromEndClean)

				dataClumpSplits := strings.Split(rowSplits[len(rowSplits)-1], ";")
				for _, v := range dataClumpSplits {
					if strings.Contains(v, "gene_name") {
						cleanedItemSplits := strings.Split(strings.TrimSpace(strings.ReplaceAll(v, "\"", "")), " ")
						if len(cleanedItemSplits) > 0 {
							geneName = cleanedItemSplits[len(cleanedItemSplits)-1]
						}
						break
					}
				}
				if len(geneName) == 0 {
					fmt.Printf("No gene found in row %s\n", rowText)
					return
				}

				discoveredGene := &models.Gene{
					Name:       geneName,
					Chrom:      chromosomeClean,
					Start:      start,
					End:        end,
					AssemblyId: _assId,
				}

				//fmt.Printf("Keys :%d, %d, %d, %d, %d -- %s\n", _chromHeaderKey, _startKey, _endKey, _nameHeaderKeys, _geneNameHeaderKeys, discoveredGene)

				iz.GeneIngestionBulkIndexingQueue <- &structs.GeneIngestionQueueStructure{
					Gene:      discoveredGene,
					WaitGroup: _gwg,
				}
			}(rowText, chromHeaderKey, startKey, endKey, nameHeaderKeys, geneNameHeaderKeys, assId, &geneWg)

			// fmt.Printf("Stats : %d\n", iz.GeneIngestionBulkIndexer.Stats())
		}
		geneWg.Wait()

	}
}
