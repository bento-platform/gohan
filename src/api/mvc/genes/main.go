package genes

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"gohan/api/contexts"
	"gohan/api/models/constants"
	assemblyId "gohan/api/models/constants/assembly-id"
	"gohan/api/models/constants/chromosome"
	"gohan/api/models/dtos"
	"gohan/api/models/ingest"
	"gohan/api/models/ingest/structs"
	esRepo "gohan/api/repositories/elasticsearch"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gohan/api/models/indexes"

	"github.com/labstack/echo"
	"github.com/mitchellh/mapstructure"
)

func GenesIngestionStats(c echo.Context) error {
	fmt.Printf("[%s] - GenesIngestionStats hit!\n", time.Now())
	ingestionService := c.(*contexts.GohanContext).IngestionService

	return c.JSON(http.StatusOK, ingestionService.GeneIngestionBulkIndexer.Stats())
}

func GenesIngest(c echo.Context) error {
	fmt.Printf("[%s] - GenesIngest hit!\n", time.Now())
	// trigger global ingestion background process
	go func() {

		cfg := c.(*contexts.GohanContext).Config
		es7Client := c.(*contexts.GohanContext).Es7Client

		gtfPath := cfg.Api.GtfPath

		iz := c.(*contexts.GohanContext).IngestionService

		if cfg.Debug {
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}

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
				CreatedAt: fmt.Sprintf("%v", time.Now()),
			}

			go func(_assId constants.AssemblyId, _fileName string, _assemblyWg *sync.WaitGroup, reqStat *ingest.GeneIngestRequest) {
				defer _assemblyWg.Done()

				var (
					unzippedFileName string
					geneWg           sync.WaitGroup
				)
				gtfFile, err := os.Open(fmt.Sprintf("%s/%s", gtfPath, _fileName))
				if err != nil {
					// Download the file
					fullURLFile := assemblyIdGTFUrlMap[_assId]

					handleHardErr := func(err error) {
						msg := "Something went wrong:  " + err.Error()
						fmt.Println(msg)

						reqStat.State = ingest.Error
						reqStat.Message = msg
						iz.GeneIngestRequestChan <- reqStat
					}

					// Build fileName from fullPath
					fileURL, err := url.Parse(fullURLFile)
					if err != nil {
						handleHardErr(err)
						return
					}

					path := fileURL.Path
					segments := strings.Split(path, "/")
					_fileName = segments[len(segments)-1]

					// Create blank file
					file, err := os.Create(fmt.Sprintf("%s/%s", gtfPath, _fileName))
					if err != nil {
						handleHardErr(err)
						return
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
						handleHardErr(err)
						return
					}
					defer resp.Body.Close()

					size, err := io.Copy(file, resp.Body)
					if err != nil {
						handleHardErr(err)
						return
					}
					defer file.Close()

					fmt.Printf("Downloaded a file %s with size %d\n", _fileName, size)

					fmt.Printf("Unzipping %s...\n", _fileName)
					unzippedFile, err := os.Open(fmt.Sprintf("%s/%s", gtfPath, _fileName))
					if err != nil {
						handleHardErr(err)
						return
					}

					reader, err := gzip.NewReader(unzippedFile)
					if err != nil {
						handleHardErr(err)
						return
					}
					defer reader.Close()

					unzippedFileName = strings.TrimSuffix(_fileName, ".gz")

					writer, err := os.Create(fmt.Sprintf("%s/%s", gtfPath, unzippedFileName))
					if err != nil {
						handleHardErr(err)
						return
					}

					defer writer.Close()

					if _, err = io.Copy(writer, reader); err != nil {
						handleHardErr(err)
						return
					}

					fmt.Printf("Opening %s\n", unzippedFileName)
					gtfFile, _ = os.Open(fmt.Sprintf("%s/%s", gtfPath, unzippedFileName))

					fmt.Printf("Deleting %s\n", _fileName)
					err = os.Remove(fmt.Sprintf("%s/%s", gtfPath, _fileName))
					if err != nil {
						// "soft" error
						fmt.Println(err)
					}
				} else {
					// for the rare occurences where the file wasn't deleted
					// after ingestion (i.e. some kind of interruption), this ensures it does
					unzippedFileName = _fileName
				}

				defer gtfFile.Close()

				// clean out genes currently in elasticsearch by assembly id
				fmt.Printf("Cleaning out %s gene documents from genes index (if any)\n", string(_assId))
				esRepo.DeleteGenesByAssemblyId(cfg, es7Client, _assId)

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

						discoveredGene := &indexes.Gene{
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
					// "soft" error
					fmt.Println(err)
				}

				reqStat.State = ingest.Done
				iz.GeneIngestRequestChan <- reqStat

			}(assId, fileName, &assemblyWg, &newRequestState)
		}

		assemblyWg.Wait()
	}()

	return c.JSON(http.StatusOK, map[string]interface{}{"message": "please check in with /genes/overview !"})
}

func GetAllGeneIngestionRequests(c echo.Context) error {
	fmt.Printf("[%s] - GetAllGeneIngestionRequests hit!\n", time.Now())
	izMap := c.(*contexts.GohanContext).IngestionService.GeneIngestRequestMap

	// transform map of it-to-ingestRequests to an array
	m := make([]*ingest.GeneIngestRequest, 0, len(izMap))
	for _, val := range izMap {
		m = append(m, val)
	}
	return c.JSON(http.StatusOK, m)
}

func GenesGetByNomenclatureWildcard(c echo.Context) error {
	fmt.Printf("[%s] - GenesGetByNomenclatureWildcard hit!\n", time.Now())
	gc := c.(*contexts.GohanContext)
	cfg := gc.Config
	es := gc.Es7Client

	// Chromosome search term
	chromosomeSearchTerm := gc.Chromosome

	// Name search term
	term := c.QueryParam("term")

	// Assembly ID
	// perform wildcard search if empty/random parameter is passed
	// - set to Unknown to trigger it
	asmId := gc.AssemblyId

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

	fmt.Printf("Executing wildcard genes search for term %s, assemblyId %s (max size: %d)\n", term, asmId, size)

	// Execute
	docs, geneErr := esRepo.GetGeneDocumentsByTermWildcard(cfg, es, chromosomeSearchTerm, term, asmId, size)
	if geneErr != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  500,
			"message": "Something went wrong... Please contact the administrator!",
		})
	}

	docsHits := docs["hits"].(map[string]interface{})["hits"]
	allDocHits := []map[string]interface{}{}
	mapstructure.Decode(docsHits, &allDocHits)

	// grab _source for each hit
	var allSources []indexes.Gene

	for _, r := range allDocHits {
		source := r["_source"].(map[string]interface{})

		// cast map[string]interface{} to struct
		var resultingVariant indexes.Gene
		mapstructure.Decode(source, &resultingVariant)

		// accumulate structs
		allSources = append(allSources, resultingVariant)
	}

	fmt.Printf("Found %d docs!\n", len(allSources))

	geneResponseDTO := dtos.GenesResponseDTO{
		Term:    term,
		Count:   len(allSources),
		Results: allSources,
		Status:  200,
		Message: "Success",
	}

	return c.JSON(http.StatusOK, geneResponseDTO)
}

func GetGenesOverview(c echo.Context) error {
	fmt.Printf("[%s] - GetGenesOverview hit!\n", time.Now())

	resultsMap := map[string]interface{}{}
	resultsMux := sync.RWMutex{}

	es := c.(*contexts.GohanContext).Es7Client
	cfg := c.(*contexts.GohanContext).Config

	// retrieve aggregation of genes/chromosomes by assembly id
	results, geneErr := esRepo.GetGeneBucketsByKeyword(cfg, es)
	if geneErr != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":  500,
			"message": "Something went wrong... Please contact the administrator!",
		})
	}

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
