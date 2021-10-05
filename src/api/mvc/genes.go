package mvc

import (
	"api/contexts"
	"api/models"
	"api/models/constants"
	assemblyId "api/models/constants/assembly-id"
	"api/models/constants/chromosome"
	"api/models/ingest"
	"api/models/ingest/structs"
	esRepo "api/repositories/elasticsearch"
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
	"time"

	"github.com/labstack/echo"
	"github.com/mitchellh/mapstructure"
)

func GenesIngest(c echo.Context) error {
	// trigger global ingestion background process
	go func() {

		cfg := c.(*contexts.GohanContext).Config
		gtfPath := cfg.Api.GtfPath

		iz := c.(*contexts.GohanContext).IngestionService

		// TEMP: SECURITY RISK
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		//
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

		var assemblyWg sync.WaitGroup

		for assId, fileName := range assemblyIdMap {
			assemblyWg.Add(1)

			newRequestState := ingest.GeneIngestRequest{
				Filename:  fileName,
				State:     ingest.Queued,
				CreatedAt: fmt.Sprintf("%s", time.Now()),
			}

			go func(_assId constants.AssemblyId, _fileName string, _assemblyWg *sync.WaitGroup, reqStat *ingest.GeneIngestRequest) {
				defer _assemblyWg.Done()

				var (
					unzippedFileName string
					geneWg           sync.WaitGroup
				)
				gtfFile, err := os.Open(fmt.Sprintf("%s/%s", gtfPath, _fileName))
				if err != nil {
					// log.Fatalf("failed to open file: %s", err)
					// Download the file
					fullURLFile := assemblyIdGTFUrlMap[_assId]

					// Build fileName from fullPath
					fileURL, err := url.Parse(fullURLFile)
					if err != nil {
						log.Fatal(err)
					}
					path := fileURL.Path
					segments := strings.Split(path, "/")
					_fileName = segments[len(segments)-1]

					// Create blank file
					file, err := os.Create(fmt.Sprintf("%s/%s", gtfPath, _fileName))
					if err != nil {
						msg := "Something went wrong:  " + err.Error()
						fmt.Println(msg)

						reqStat.State = ingest.Error
						reqStat.Message = msg
						iz.GeneIngestRequestChan <- reqStat
					}
					client := http.Client{
						CheckRedirect: func(r *http.Request, via []*http.Request) error {
							r.URL.Opaque = r.URL.Path
							return nil
						},
					}

					fmt.Printf("Downloading file %s ...\n", _fileName)
					reqStat.State = ingest.Downloading
					iz.GeneIngestRequestChan <- reqStat

					// Put content on file
					resp, err := client.Get(fullURLFile)
					if err != nil {
						msg := "Something went wrong:  " + err.Error()
						fmt.Println(msg)

						reqStat.State = ingest.Error
						reqStat.Message = msg
						iz.GeneIngestRequestChan <- reqStat
					}
					defer resp.Body.Close()

					size, err := io.Copy(file, resp.Body)
					if err != nil {
						msg := "Something went wrong:  " + err.Error()
						fmt.Println(msg)

						reqStat.State = ingest.Error
						reqStat.Message = msg
						iz.GeneIngestRequestChan <- reqStat
					}
					defer file.Close()

					fmt.Printf("Downloaded a file %s with size %d\n", _fileName, size)

					fmt.Printf("Unzipping %s...\n", _fileName)
					unzippedFile, err := os.Open(fmt.Sprintf("%s/%s", gtfPath, _fileName))
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}

					reader, err := gzip.NewReader(unzippedFile)
					if err != nil {
						msg := "Something went wrong:  " + err.Error()
						fmt.Println(msg)

						reqStat.State = ingest.Error
						reqStat.Message = msg
						iz.GeneIngestRequestChan <- reqStat
					}
					defer reader.Close()

					unzippedFileName = strings.TrimSuffix(_fileName, ".gz")

					writer, err := os.Create(fmt.Sprintf("%s/%s", gtfPath, unzippedFileName))

					if err != nil {
						msg := "Something went wrong:  " + err.Error()
						fmt.Println(msg)

						reqStat.State = ingest.Error
						reqStat.Message = msg
						iz.GeneIngestRequestChan <- reqStat
					}

					defer writer.Close()

					if _, err = io.Copy(writer, reader); err != nil {
						msg := "Something went wrong:  " + err.Error()
						fmt.Println(msg)

						reqStat.State = ingest.Error
						reqStat.Message = msg
						iz.GeneIngestRequestChan <- reqStat
					}

					fmt.Printf("Opening %s\n", unzippedFileName)
					gtfFile, _ = os.Open(fmt.Sprintf("%s/%s", gtfPath, unzippedFileName))

					fmt.Printf("Deleting %s\n", _fileName)
					err = os.Remove(fmt.Sprintf("%s/%s", gtfPath, _fileName))
					if err != nil {
						fmt.Println(err)
					}
				} else {
					// for the rare occurences where the file wasn't deleted
					// after ingestion (i.e. some kind of interruption), this ensures it does
					unzippedFileName = _fileName
				}

				defer gtfFile.Close()
				fileScanner := bufio.NewScanner(gtfFile)
				fileScanner.Split(bufio.ScanLines)

				fmt.Printf("Ingesting %s\n", string(_assId))
				reqStat.State = ingest.Running
				iz.GeneIngestRequestChan <- reqStat

				var (
					chromHeaderKey     = 0
					startKey           = 3
					endKey             = 4
					nameHeaderKeys     = []int{3}
					geneNameHeaderKeys []int
				)

				var columnsToPrint []string
				if _assId == assemblyId.GRCh38 {
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

						iz.GeneIngestionBulkIndexingQueue <- &structs.GeneIngestionQueueStructure{
							Gene:      discoveredGene,
							WaitGroup: _gwg,
						}
					}(rowText, chromHeaderKey, startKey, endKey, nameHeaderKeys, geneNameHeaderKeys, _assId, &geneWg)
				}

				geneWg.Wait()

				fmt.Printf("%s ingestion done!\n", _assId)
				fmt.Printf("Deleting %s\n", unzippedFileName)
				err = os.Remove(fmt.Sprintf("%s/%s", gtfPath, unzippedFileName))
				if err != nil {
					fmt.Println(err)
				}

				reqStat.State = ingest.Done
				iz.GeneIngestRequestChan <- reqStat

			}(assId, fileName, &assemblyWg, &newRequestState)
		}

		assemblyWg.Wait()
	}()

	return c.JSON(http.StatusOK, "{\"message\":\"please check in with /genes/overview !\"}")
}

func GetAllGeneIngestionRequests(c echo.Context) error {
	izMap := c.(*contexts.GohanContext).IngestionService.GeneIngestRequestMap

	// transform map of it-to-ingestRequests to an array
	m := make([]*ingest.GeneIngestRequest, 0, len(izMap))
	for _, val := range izMap {
		m = append(m, val)
	}
	return c.JSON(http.StatusOK, m)
}

func GenesGetByNomenclatureWildcard(c echo.Context) error {
	cfg := c.(*contexts.GohanContext).Config
	es := c.(*contexts.GohanContext).Es7Client

	// Chromosome search term
	chromosomeSearchTerm := c.QueryParam("chromosome")
	if len(chromosomeSearchTerm) == 0 {
		// if no chromosome is provided, assume "wildcard" search
		chromosomeSearchTerm = "*"
	}

	// Name search term
	term := c.QueryParam("term")

	// Assembly ID
	// perform wildcard search if empty/random parameter is passed
	// - set to Unknown to trigger it
	assId := assemblyId.Unknown
	if assemblyId.CastToAssemblyId(c.QueryParam("assemblyId")) != assemblyId.Unknown {
		// retrieve passed parameter if is valid
		assId = assemblyId.CastToAssemblyId(c.QueryParam("assemblyId"))
	}

	// Size
	var (
		size        int = 25
		sizeCastErr error
	)
	if len(c.QueryParam("size")) > 0 {
		sizeQP := c.QueryParam("size")
		size, sizeCastErr = strconv.Atoi(sizeQP)
		if sizeCastErr != nil {
			size = 25
		}
	}

	fmt.Printf("Executing wildcard genes search for term %s, assemblyId %s (max size: %d)\n", term, assId, size)

	// Execute
	docs := esRepo.GetGeneDocumentsByTermWildcard(cfg, es, chromosomeSearchTerm, term, assId, size)

	docsHits := docs["hits"].(map[string]interface{})["hits"]
	allDocHits := []map[string]interface{}{}
	mapstructure.Decode(docsHits, &allDocHits)

	// grab _source for each hit
	var allSources []models.Gene

	for _, r := range allDocHits {
		source := r["_source"].(map[string]interface{})

		// cast map[string]interface{} to struct
		var resultingVariant models.Gene
		mapstructure.Decode(source, &resultingVariant)

		// accumulate structs
		allSources = append(allSources, resultingVariant)
	}

	fmt.Printf("Found %d docs!\n", len(allSources))

	geneResponseDTO := models.GenesResponseDTO{
		Term:    term,
		Count:   len(allSources),
		Results: allSources,
		Status:  200,
		Message: "Success",
	}

	return c.JSON(http.StatusOK, geneResponseDTO)
}

func GetGenesOverview(c echo.Context) error {

	resultsMap := map[string]interface{}{}
	resultsMux := sync.RWMutex{}

	es := c.(*contexts.GohanContext).Es7Client
	cfg := c.(*contexts.GohanContext).Config

	// retrieve aggregation of genes/chromosomes by assembly id
	results := esRepo.GetGeneBucketsByKeyword(cfg, es)

	// begin mapping results
	geneChromosomeGroupBucketsMapped := []map[string]interface{}{}

	// loop over top level aggregation and
	// accumulated nested aggregations
	if aggs, ok := results["aggregations"]; ok {
		aggsMapped := aggs.(map[string]interface{})

		if items, ok := aggsMapped["genes_assembly_id_group"]; ok {
			itemsMapped := items.(map[string]interface{})

			if buckets := itemsMapped["buckets"]; ok {
				arrayMappedBuckets := buckets.([]interface{})

				for _, mappedBucket := range arrayMappedBuckets {
					geneChromosomeGroupBucketsMapped = append(geneChromosomeGroupBucketsMapped, mappedBucket.(map[string]interface{}))
				}
			}
		}
	}

	individualAssemblyIdKeyMap := map[string]interface{}{}

	// iterated over each assemblyId bucket
	for _, chromGroupBucketMap := range geneChromosomeGroupBucketsMapped {

		assemblyIdKey := fmt.Sprint(chromGroupBucketMap["key"])

		numGenesPerChromMap := map[string]interface{}{}
		bucketsMapped := map[string]interface{}{}

		if chromGroupItem, ok := chromGroupBucketMap["genes_chromosome_group"]; ok {
			chromGroupItemMapped := chromGroupItem.(map[string]interface{})

			for _, chromBucket := range chromGroupItemMapped["buckets"].([]interface{}) {
				doc_key := fmt.Sprint(chromBucket.(map[string]interface{})["key"]) // ensure strings and numbers are expressed as strings
				doc_count := chromBucket.(map[string]interface{})["doc_count"]

				// add to list of buckets by chromosome
				bucketsMapped[doc_key] = doc_count
			}
		}

		numGenesPerChromMap["numberOfGenesPerChromosome"] = bucketsMapped
		individualAssemblyIdKeyMap[assemblyIdKey] = numGenesPerChromMap
	}

	resultsMux.Lock()
	resultsMap["assemblyIDs"] = individualAssemblyIdKeyMap
	resultsMux.Unlock()

	return c.JSON(http.StatusOK, resultsMap)
}
