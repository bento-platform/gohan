package main

import (
	"api/models"
	"api/models/constants"
	assemblyId "api/models/constants/assembly-id"
	"api/models/ingest/structs"
	"api/services"
	"api/utils"
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
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
		assemblyId.GRCh37: "gencode.v19.annotation.gtf_withproteinids",
		// SKIP
		// assemblyId.NCBI36: "hg18",
		// assemblyId.NCBI35: "hg17",
		// assemblyId.NCBI34: "hg16",
	}

	var geneWg sync.WaitGroup

	for assId, fileName := range assemblyIdMap {
		// Read one file at a time

		gtfFile, err := os.Open(fileName)
		if err != nil {
			log.Fatalf("failed to open file: %s", err)
		}

		defer gtfFile.Close()

		fileScanner := bufio.NewScanner(gtfFile)
		fileScanner.Split(bufio.ScanLines)

		fmt.Printf("%s :\n", fileName)

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
				// TODO: fomarmalize
				// if chromosome MT, set to 0
				// if chromosome X, set to 101
				// if chromosome Y, set to 102
				// if strings.Contains(strings.ToUpper(chromosomeClean), "MT") {
				// 	chromosome = 0
				// } else if strings.ToUpper(chromosomeClean) == "X" {
				// 	chromosome = 101
				// } else if strings.ToUpper(chromosomeClean) == "Y" {
				// 	chromosome = 102
				// } else {
				// 	chromosome, _ = strconv.Atoi(chromosomeClean)
				// }

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
