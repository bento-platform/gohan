package services

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"gohan/api/models"
	"gohan/api/models/constants"
	"gohan/api/models/constants/chromosome"
	z "gohan/api/models/constants/zygosity"
	"gohan/api/models/ingest"
	"gohan/api/models/ingest/structs"
	"gohan/api/utils"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gohan/api/models/indexes"

	"github.com/Jeffail/gabs"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/mitchellh/mapstructure"
)

type (
	IngestionService struct {
		Initialized                    bool
		IngestRequestChan              chan *ingest.VariantIngestRequest
		IngestRequestMap               map[string]*ingest.VariantIngestRequest
		GeneIngestRequestChan          chan *ingest.GeneIngestRequest
		GeneIngestRequestMap           map[string]*ingest.GeneIngestRequest
		IngestionBulkIndexingCapacity  int
		IngestionBulkIndexingQueue     chan *structs.IngestionQueueStructure
		IngestionBulkIndexer           esutil.BulkIndexer
		GeneIngestionBulkIndexingQueue chan *structs.GeneIngestionQueueStructure
		GeneIngestionBulkIndexer       esutil.BulkIndexer
		ConcurrentFileIngestionQueue   chan bool
		ElasticsearchClient            *elasticsearch.Client
	}
)

func NewIngestionService(es *elasticsearch.Client, cfg *models.Config) *IngestionService {

	iz := &IngestionService{
		Initialized:                    false,
		IngestRequestChan:              make(chan *ingest.VariantIngestRequest),
		IngestRequestMap:               map[string]*ingest.VariantIngestRequest{},
		GeneIngestRequestChan:          make(chan *ingest.GeneIngestRequest),
		GeneIngestRequestMap:           map[string]*ingest.GeneIngestRequest{},
		IngestionBulkIndexingCapacity:  cfg.Api.BulkIndexingCap,
		IngestionBulkIndexingQueue:     make(chan *structs.IngestionQueueStructure, cfg.Api.BulkIndexingCap),
		GeneIngestionBulkIndexingQueue: make(chan *structs.GeneIngestionQueueStructure, 10),
		ConcurrentFileIngestionQueue:   make(chan bool, cfg.Api.FileProcessingConcurrencyLevel),
		ElasticsearchClient:            es,
	}

	//see: https://www.elastic.co/blog/why-am-i-seeing-bulk-rejections-in-my-elasticsearch-cluster
	var numWorkers = iz.IngestionBulkIndexingCapacity / 100
	//the lower the denominator (the number of documents per bulk upload). the higher
	//the chances of 100% successful upload, though the longer it may take (negligible)

	bi, _ := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:      "variants",
		Client:     iz.ElasticsearchClient,
		NumWorkers: numWorkers,
		// FlushBytes:    int(flushBytes),  // The flush threshold in bytes (default: 5MB ?)
		// FlushInterval: time.Second, // The periodic flush interval
	})
	iz.IngestionBulkIndexer = bi

	gbi, _ := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:      "genes",
		Client:     iz.ElasticsearchClient,
		NumWorkers: numWorkers,
		// FlushBytes: int(64), // The flush threshold in bytes (default: 5MB ?)
		// FlushInterval: 3 * time.Second, // The periodic flush interval
	})
	iz.GeneIngestionBulkIndexer = gbi

	iz.Init()

	return iz
}

func (i *IngestionService) Init() {
	// safeguard to prevent multiple initilizations
	if !i.Initialized {
		// spin up a go routine acting as a listener for variant and
		// gene ingest request updates, and variant and gene bulk indexing
		go func() {
			for {
				select {
				case variantIngestionRequest := <-i.IngestRequestChan:
					if variantIngestionRequest.State == ingest.Queued {
						fmt.Printf("Queueing a new variant ingestion request for %s\n", variantIngestionRequest.Filename)
					}

					variantIngestionRequest.UpdatedAt = time.Now().String()
					i.IngestRequestMap[variantIngestionRequest.Id.String()] = variantIngestionRequest

				case geneIngestionRequest := <-i.GeneIngestRequestChan:
					if geneIngestionRequest.State == ingest.Queued {
						fmt.Printf("Queueing a new gene ingestion request for %s\n", geneIngestionRequest.Filename)
					}

					geneIngestionRequest.UpdatedAt = time.Now().String()
					i.GeneIngestRequestMap[geneIngestionRequest.Filename] = geneIngestionRequest

				case queuedVariantItem := <-i.IngestionBulkIndexingQueue:

					queuedVariant := queuedVariantItem.Variant
					wg := queuedVariantItem.WaitGroup

					// Prepare the data payload: encode article to JSON
					variantData, marshallErr := json.Marshal(queuedVariant)
					if marshallErr != nil {
						log.Fatalf("Cannot encode variant %s: %s\n", queuedVariant.Id, marshallErr)
					}

					// Add an item to the BulkIndexer
					marshallErr = i.IngestionBulkIndexer.Add(
						context.Background(),
						esutil.BulkIndexerItem{
							// Action field configures the operation to perform (index, create, delete, update)
							Action: "index",

							// Body is an `io.Reader` with the payload
							Body: bytes.NewReader(variantData),

							// OnSuccess is called for each successful operation
							OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
								defer wg.Done()
								//fmt.Printf("Successfully indexed: %s", item)
								//atomic.AddUint64(&countSuccessful, 1)
							},

							// OnFailure is called for each failed operation
							OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
								defer wg.Done()
								//atomic.AddUint64(&countFailed, 1)
								if err != nil {
									fmt.Printf("ERROR: %s", err)
								} else {
									fmt.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
								}
							},
						},
					)
					if marshallErr != nil {
						fmt.Printf("Unexpected error: %s", marshallErr)
						wg.Done()
					}

				case queuedGeneItem := <-i.GeneIngestionBulkIndexingQueue:

					queuedGene := queuedGeneItem.Gene
					wg := queuedGeneItem.WaitGroup

					// Prepare the data payload: encode article to JSON
					geneData, marshallErr := json.Marshal(queuedGene)
					if marshallErr != nil {
						log.Fatalf("Cannot encode gene %+v: %s\n", queuedGene, marshallErr)
					}

					// Add an item to the BulkIndexer
					marshallErr = i.GeneIngestionBulkIndexer.Add(
						context.Background(),
						esutil.BulkIndexerItem{
							// Action field configures the operation to perform (index, create, delete, update)
							Action: "index",

							// Body is an `io.Reader` with the payload
							Body: bytes.NewReader(geneData),

							// OnSuccess is called for each successful operation
							OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
								defer wg.Done()
								//atomic.AddUint64(&countSuccessful, 1)
							},

							// OnFailure is called for each failed operation
							OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
								defer wg.Done()
								fmt.Printf("Failure Repsonse: %s", res.Error)
								//atomic.AddUint64(&countFailed, 1)
								if err != nil {
									fmt.Printf("ERROR: %s", err)
								} else {
									fmt.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
								}
							},
						},
					)
					if marshallErr != nil {
						fmt.Printf("Unexpected error: %s", marshallErr)
						wg.Done()
					}
				}
			}
		}()

		i.Initialized = true
	}
}

func (i *IngestionService) ExtractVcfGz(gzippedFilePath string, gzipStream io.Reader, vcfTmpPath string) string {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Fatal("ExtractTarGz: NewReader failed - ", err)
		return ""
	}

	// ---	 store to disk (temporarily)
	// new file name
	vcfFilePath := strings.Replace(gzippedFilePath, ".gz", "", -1)
	vcfFilePathSplits := strings.Split(vcfFilePath, "/")
	vcfFileName := vcfFilePathSplits[len(vcfFilePathSplits)-1]
	newVcfFilePath := vcfTmpPath + "/" + vcfFileName

	fmt.Printf("Creating new temporary VCF file: %s\n", newVcfFilePath)
	f, err := os.Create(newVcfFilePath)
	if err != nil {
		fmt.Println("Something went wrong:  ", err)
		return ""
	}

	fmt.Printf("Writing to new temporary VCF file: %s\n", newVcfFilePath)
	w := bufio.NewWriter(f)
	io.Copy(w, uncompressedStream)

	uncompressedStream.Close()
	f.Close()

	return newVcfFilePath
}

func (i *IngestionService) GenerateTabix(gzippedFilePath string) (string, string, error) {
	cmd := exec.Command("tabix", "-f", gzippedFilePath)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	err := cmd.Run()
	if err != nil {
		fmt.Println(cmdOutput.String())
		fmt.Println(err.Error())
		os.Stderr.WriteString(err.Error())
		return "", "", err
	}
	fmt.Print(cmdOutput.String())

	dir, file := path.Split(fmt.Sprintf("%s.tbi", gzippedFilePath))
	return dir, file, nil
}

func (i *IngestionService) UploadVcfGzToDrs(cfg *models.Config, drsBridgeDirectory string, gzippedFileName string, drsUrl, drsUsername, drsPassword string) string {

	if cfg.Debug {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	data := fmt.Sprintf("{\"path\": \"%s/%s\"}", drsBridgeDirectory, gzippedFileName)

	var (
		drsId           string
		drsResp         *http.Response
		drsErr          error
		attemptCount    int = 0
		maxAttempts     int = 5
		waitTimeSeconds int = 3
	)
	for {
		// prepare upload request to drs
		r, _ := http.NewRequest("POST", drsUrl+"/private/ingest", bytes.NewBufferString(data))

		r.SetBasicAuth(drsUsername, drsPassword)
		r.Header.Add("Content-Type", "application/json")

		client := &http.Client{}

		// perform request
		drsResp, drsErr = client.Do(r)

		// check for errors, possibly try again
		if drsErr != nil {
			fmt.Printf("Upload to DRS error: %s\n", drsErr)

			if attemptCount < maxAttempts {
				// increment attempt counter
				attemptCount++

				// give it a few seconds break
				time.Sleep(time.Duration(waitTimeSeconds * int(time.Second)))

				fmt.Printf("trying again...\n")
				continue
			} else {
				fmt.Printf("exiting upload loop...\n")
				return "" // empty drs-id string
			}
		}

		// check for simple upload error (like db locked) and try again
		fmt.Printf("Got a %d status code on DRS upload \n", drsResp.StatusCode)
		if drsResp.StatusCode == 201 {
			fmt.Printf("File %s upload to DRS succeeded: %d\n", gzippedFileName, drsResp.StatusCode)

			// proceed with vcf processing
			break
		} else if drsResp.StatusCode == 401 {
			// exit right away on 'unauthorized' status code
			fmt.Printf("Received a '401 Unauthorized' from DRS -- exiting upload loop...\n")
			return "" // empty drs-id string
		} else {
			// print response message
			unsuccessfulAttemptResponseBody, unsuccessfulAttemptResponseErr := ioutil.ReadAll(drsResp.Body)
			if unsuccessfulAttemptResponseErr != nil {
				fmt.Printf("Error reading unsuccessful attempt response body: %v", unsuccessfulAttemptResponseErr)
			} else {
				fmt.Printf("Received from after failed attempt: %s\n", string(unsuccessfulAttemptResponseBody))
			}
			// continue the loop in the event this failure to parse response body is intermittent

			if attemptCount < maxAttempts {
				// increment attempt counter
				attemptCount++

				// give it a few seconds break
				time.Sleep(time.Duration(waitTimeSeconds * int(time.Second)))

				fmt.Printf("Failed to upload to DRS after %d attempts.. Trying again...\n", attemptCount)
				continue
			} else {
				fmt.Printf("After %d failed attempts, exiting upload loop...\n", attemptCount)
				return "" // empty drs-id string
			}
		}
	}

	responsebody, bodyerr := ioutil.ReadAll(drsResp.Body)
	if bodyerr != nil {
		fmt.Printf("Error reading body: %v\n", bodyerr)
		return ""
	}

	jsonParsed, err := gabs.ParseJSON(responsebody)
	if err != nil {
		fmt.Printf("Parsing error: %s\n", err)
		return ""
	}
	drsId = jsonParsed.Path("id").Data().(string)

	fmt.Println("Get DRS ID: ", drsId)

	return drsId
}

func (i *IngestionService) ProcessVcf(
	vcfFilePath string, drsFileId string, tableId string,
	assemblyId constants.AssemblyId, filterOutHomozygousReferences bool,
	lineProcessingConcurrencyLevel int) {

	f, err := os.Open(vcfFilePath)
	if err != nil {
		fmt.Println("Failed to open file - ", err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var discoveredHeaders bool = false
	var headers []string
	headerSampleIds := make(map[int]string)

	skippedHomozygousReferencesCount := int32(0)

	var _fileWG sync.WaitGroup

	// "line ingestion queue"
	// - manage # of lines being concurrently processed per file at any given time
	lineProcessingQueue := make(chan bool, lineProcessingConcurrencyLevel)

	for scanner.Scan() {
		//fmt.Println(scanner.Text())

		// Gather Header row by seeking the CHROM string
		line := scanner.Text()
		if !discoveredHeaders {
			if line[0] == '#' {
				if strings.Contains(line, "CHROM") {
					// Split the string by tabs
					headers = strings.Split(line, "\t")

					for id, header := range headers {
						// determine if header is a default VCF header.
						// if it is not, assume it's a sampleId and keep
						// track of it with an id
						if !utils.StringInSlice(strings.ToLower(strings.TrimSpace(strings.ReplaceAll(header, "#", ""))), constants.VcfHeaders) {
							headerSampleIds[len(constants.VcfHeaders)-id] = header
						}
					}

					discoveredHeaders = true

					fmt.Println("Found the headers: ", headers)
					continue
				}
				continue
			}
		}

		// take a spot in the queue
		lineProcessingQueue <- true
		_fileWG.Add(1)
		go func(line string, drsFileId string, fileWg *sync.WaitGroup) {
			// free up a spot in the queue
			defer func() { <-lineProcessingQueue }()

			// ----  break up line
			rowComponents := strings.Split(line, "\t")

			// ----  process more...
			var tmpSamples []map[string]interface{}
			tmpSamplesMutex := sync.RWMutex{}

			tmpVariant := make(map[string]interface{})
			tmpVariantMapMutex := sync.RWMutex{}

			tmpVariant["fileId"] = drsFileId
			tmpVariant["assemblyId"] = assemblyId
			tmpVariant["tableId"] = tableId

			// skip this call if need be
			skipThisCall := false

			var rowWg sync.WaitGroup
			rowWg.Add(len(rowComponents))

			for rowIndex, rowComponent := range rowComponents {
				go func(i int, rc string, rwg *sync.WaitGroup) {
					defer rwg.Done()
					key := strings.ToLower(strings.TrimSpace(strings.Replace(headers[i], "#", "", -1)))
					value := strings.TrimSpace(rc)

					// if not a vcf header, assume it's a sampleId header
					if utils.StringInSlice(key, constants.VcfHeaders) {

						// filter field type by column name
						if key == "chrom" {
							// Strip out all non-numeric characters
							value = strings.ReplaceAll(value, "chr", "")

							// ems if value is valid chromosome
							if chromosome.IsValidHumanChromosome(value) {
								tmpVariantMapMutex.Lock()
								tmpVariant[key] = value
								tmpVariantMapMutex.Unlock()
							} else {
								// skip this call
								skipThisCall = true

								// redundant?
								tmpVariantMapMutex.Lock()
								tmpVariant[key] = "err"
								tmpVariantMapMutex.Unlock()
							}
						} else if key == "pos" || key == "qual" {

							// // Convert string's to int's, if possible
							value, err := strconv.ParseInt(value, 10, 0)
							if err == nil {
								tmpVariantMapMutex.Lock()
								tmpVariant[key] = value
								tmpVariantMapMutex.Unlock()
							} else {
								tmpVariantMapMutex.Lock()
								tmpVariant[key] = -1 // here to simulate a null value (such as when the string value is empty, or
								//                      is something as arbitrary as a single period '.')
								tmpVariantMapMutex.Unlock()
							}

						} else if key == "alt" || key == "ref" {
							// Split all alleles by comma
							tmpVariantMapMutex.Lock()
							tmpVariant[key] = strings.Split(value, ",")
							tmpVariantMapMutex.Unlock()
						} else if key == "format" {
							// Split all formats by colon
							tmpVariantMapMutex.Lock()
							tmpVariant[key] = strings.Split(value, ":")
							tmpVariantMapMutex.Unlock()
						} else if key == "id" {
							// check for "empty" IDs (i.e, those with a period) and tokenize with "none"
							if value == "." {
								value = "none"
							}
							tmpVariantMapMutex.Lock()
							tmpVariant[key] = value
							tmpVariantMapMutex.Unlock()
						} else if key == "info" {
							var allInfos []*indexes.Info

							// Split all alleles by semi-colon
							semiColonSeparations := strings.Split(value, ";")

							for _, scSep := range semiColonSeparations {
								// Split by equality symbol
								equalitySeparations := strings.Split(scSep, "=")

								if len(equalitySeparations) == 2 {
									allInfos = append(allInfos, &indexes.Info{
										Id:    equalitySeparations[0],
										Value: equalitySeparations[1],
									})
								} else { // len(equalitySeparations) == 1
									allInfos = append(allInfos, &indexes.Info{
										Id:    "",
										Value: equalitySeparations[0],
									})
								}
							}

							tmpVariantMapMutex.Lock()
							tmpVariant[key] = allInfos
							tmpVariantMapMutex.Unlock()

						} else {
							tmpVariantMapMutex.Lock()
							tmpVariant[key] = value
							tmpVariantMapMutex.Unlock()
						}
					} else { // assume its a sampleId header
						allValues := strings.Split(value, ":")

						// ---- filter out homozygous reference calls
						// support for multi-sampled calls
						// assume first component of allValues is the genotype
						genoTypeValue := allValues[0]
						if filterOutHomozygousReferences && (genoTypeValue == "0|0" || genoTypeValue == "0/0") {
							// skip adding this sample to the 'tmpSamples' list which
							// then goes to be further processed into a variant document

							// increase count of skipped calls
							atomic.AddInt32(&skippedHomozygousReferencesCount, 1)

							return
						}

						tmpSamplesMutex.Lock()
						tmpSamples = append(tmpSamples, map[string]interface{}{
							"key":    key,
							"values": allValues,
						})
						tmpSamplesMutex.Unlock()
					}
				}(rowIndex, rowComponent, &rowWg)
			}

			rowWg.Wait()

			if skipThisCall {
				// This variant call has been deemed unnecessary to ingest
				defer fileWg.Done()
				return
			}

			// --- prep formats + samples
			var samples []*indexes.Sample

			// ---- get genotype stuff
			var (
				hasGenotype      bool = false
				genotypePosition int  = 0

				hasGenotypeProbability      bool = false
				genotypeProbabilityPosition int  = 0

				hasPhredScaleLikelyhood      bool = false
				phredScaleLikelyhoodPosition int  = 0
			)

			// error checking --
			if tmpVariant == nil {
				fmt.Printf("Something went wrong, but was caught:\ntmpVariant is nil for file with DRS fileId `%s` at line `%s`  \n\n", drsFileId, line)
				return
			}
			if utils.KeyExists(tmpVariant, "format") {
				for i, f := range tmpVariant["format"].([]string) {
					// ----- check formats
					switch f {
					case "GT":
						hasGenotype = true
						genotypePosition = i
					case "GP":
						hasGenotypeProbability = true
						genotypeProbabilityPosition = i
					case "PL":
						hasPhredScaleLikelyhood = true
						phredScaleLikelyhoodPosition = i
					}
				}
			} else {
				fmt.Printf("Something went wrong, but was caught:\ntmpVariant[\"format\"] doesnt exist for file with DRS fileId `%s` at line `%s`  \n\n", drsFileId, line)
			}
			// --

			for _, ts := range tmpSamples {
				sample := &indexes.Sample{}
				variation := &indexes.Variation{}

				tmpKeyString := ts["key"].(string)
				tmpValueStrings := ts["values"].([]string)
				for k := range tmpValueStrings {
					if hasGenotype && k == genotypePosition {
						// create genotype from value
						gtString := tmpValueStrings[k]

						// TODO: check and handle when 'gtString' is '.'

						phased := strings.Contains(gtString, "|")

						var (
							alleleStringSplits []string
							alleleLeft         int
							alleleRight        int
						)
						if phased {
							alleleStringSplits = strings.Split(gtString, "|")
						} else {
							alleleStringSplits = strings.Split(gtString, "/")
						}

						// convert string to int
						// -- if error, assume it's a period and assign -1
						alleleLeft, errLeft := strconv.Atoi(alleleStringSplits[0])
						if errLeft != nil {
							alleleLeft = -1
						}

						alleleRight, errRight := strconv.Atoi(alleleStringSplits[1])
						if errRight != nil {
							alleleRight = -1
						}

						// -- zygosity:
						var zyg constants.Zygosity
						if alleleLeft == -1 || alleleRight == -1 {
							zyg = z.Unknown
						} else {
							switch alleleLeft == alleleRight {
							case true:
								switch alleleLeft * alleleRight {
								case 0:
									zyg = z.HomozygousReference
								default:
									zyg = z.HomozygousAlternate
								}
							case false:
								zyg = z.Heterozygous
							}
						}

						//   By this point, tmpVariant["alt"] is populated with
						//   an array of strings, i.e ["C", "CTT", "CTTTT", ...] .
						//   Using the values in 'alleleLeft' and 'alleleRight' as
						//   reference to which alleles are "most-likely", format and
						//   store alleles specific to each sample

						// indexing ref/alt in a vcf row:
						//
						//       0       1, 2, 3, ...
						// ...  REF		ALT			...
						// ...  G		CT,CTT,CTTT

						var alleles []string
						// hold a temporary pointer to the current state of this-variant's 'alt' and 'ref' for brevity
						tmpVariantAlt := tmpVariant["alt"].([]string)
						tmpVariantRef := tmpVariant["ref"].([]string)

						if alleleLeft > 0 {
							alleles = append(alleles, tmpVariantAlt[alleleLeft-1])
						} else {
							alleles = append(alleles, tmpVariantRef[0])
						}
						if alleleRight > 0 {
							alleles = append(alleles, tmpVariantAlt[alleleRight-1])
						} else {
							alleles = append(alleles, tmpVariantRef[0])
						}

						variation.Genotype = indexes.Genotype{
							Phased:   phased,
							Zygosity: zyg,
						}
						variation.Alleles = alleles

					} else if hasGenotypeProbability && k == genotypeProbabilityPosition {
						// create genotype probability from value
						probValStrings := strings.Split(tmpValueStrings[k], ",")
						for _, val := range probValStrings {
							if n, err := strconv.ParseFloat(val, 64); err == nil {
								variation.GenotypeProbability = append(variation.GenotypeProbability, n)
							} else {
								variation.GenotypeProbability = append(variation.GenotypeProbability, -1)
							}
						}
					} else if hasPhredScaleLikelyhood && k == phredScaleLikelyhoodPosition {
						// create phred scale likelyhood from value
						likelyhoodValStrings := strings.Split(tmpValueStrings[k], ",")
						for _, val := range likelyhoodValStrings {
							if n, err := strconv.ParseFloat(val, 64); err == nil {
								variation.PhredScaleLikelyhood = append(variation.PhredScaleLikelyhood, n)
							} else {
								variation.PhredScaleLikelyhood = append(variation.PhredScaleLikelyhood, -1)
							}
						}

					}
				}

				sample.Id = tmpKeyString
				sample.Variation = *variation

				samples = append(samples, sample)
			}

			// Determine if this variant is worth ingesting (if it has
			// any samples after having maybe filtered out all homozygous
			// references, and thus maybe all samples from the call
			// [i.e. if this is a single-sample VCF])
			if len(samples) > 0 {

				// for multi-sample vcfs, add 1 to the waitgroup for
				// each sample (minus 1 given the initial addition)
				fileWg.Add(len(samples) - 1)

				// Create a whole variant document for each sample found on this VCF line
				// TODO: revisit this model as it is surely not storage efficient
				for _, sample := range samples {
					tmpVariant["sample"] = sample
					// ---	 push to a bulk "queue"
					var resultingVariant indexes.Variant
					mapstructure.Decode(tmpVariant, &resultingVariant)

					// pass variant (along with a waitgroup) to the channel
					i.IngestionBulkIndexingQueue <- &structs.IngestionQueueStructure{
						Variant:   &resultingVariant,
						WaitGroup: fileWg,
					}
				}
			} else {
				// This variant call has been deemed unnecessary to ingest
				defer fileWg.Done()
				return
			}
		}(line, drsFileId, &_fileWG)
	}

	// allowing all lines to be queued up and waited for
	for i := 0; i < lineProcessingConcurrencyLevel; i++ {
		lineProcessingQueue <- true
	}

	// let all lines be queued up and processed
	_fileWG.Wait()

	fmt.Printf("File %s waited for and complete!\n\t- Number of skipped Homozygous Reference calls: %d\n", vcfFilePath, skippedHomozygousReferencesCount)
}

func (i *IngestionService) FilenameAlreadyRunning(filename string) bool {
	for _, v := range i.IngestRequestMap {
		if v.Filename == filename && (v.State == ingest.Queued || v.State == ingest.Running) {
			return true
		}
	}
	return false
}
