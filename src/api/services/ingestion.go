package services

import (
	"api/models"
	"api/models/constants"
	z "api/models/constants/zygosity"
	"api/models/ingest"
	"api/utils"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esutil"
	"github.com/mitchellh/mapstructure"
)

type (
	IngestionService struct {
		Initialized                   bool
		IngestRequestChan             chan *ingest.IngestRequest
		IngestRequestMap              map[string]*ingest.IngestRequest
		IngestionBulkIndexingCapacity int
		IngestionBulkIndexingQueue    chan *IngestionQueueStructure
		ElasticsearchClient           *elasticsearch.Client
		IngestionBulkIndexer          esutil.BulkIndexer
	}

	IngestionQueueStructure struct {
		Variant   *models.Variant
		WaitGroup *sync.WaitGroup
	}
)

const defaultBulkIndexingCap int = 10000

func NewIngestionService(es *elasticsearch.Client) *IngestionService {

	iz := &IngestionService{
		Initialized:                   false,
		IngestRequestChan:             make(chan *ingest.IngestRequest),
		IngestRequestMap:              map[string]*ingest.IngestRequest{},
		IngestionBulkIndexingCapacity: defaultBulkIndexingCap,
		IngestionBulkIndexingQueue:    make(chan *IngestionQueueStructure, defaultBulkIndexingCap),
		ElasticsearchClient:           es,
	}

	// see: https://www.elastic.co/blog/why-am-i-seeing-bulk-rejections-in-my-elasticsearch-cluster
	var numWorkers = defaultBulkIndexingCap / 100
	// the lower the denominator (the number of documents per bulk upload). the higher
	// the chances of 100% successful upload, though the longer it may take (negligible)

	bi, _ := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:      "variants",
		Client:     iz.ElasticsearchClient,
		NumWorkers: numWorkers,
		// FlushBytes:    int(flushBytes),  // The flush threshold in bytes (default: 50MB ?)
		FlushInterval: 30 * time.Second, // The periodic flush interval
	})

	iz.IngestionBulkIndexer = bi

	iz.Init()

	return iz
}

func (i *IngestionService) Init() {
	// safeguard to prevent multiple initilizations
	if !i.Initialized {
		// spin up a listener for ingest request updates
		go func() {
			for {
				select {
				case newRequest := <-i.IngestRequestChan:
					if newRequest.State == ingest.Queued {
						fmt.Printf("Received new request for %s", newRequest.Filename)
					}

					newRequest.UpdatedAt = fmt.Sprintf("%s", time.Now())
					i.IngestRequestMap[newRequest.Id.String()] = newRequest
				}
			}
		}()

		// spin up a listener for bulk indexing
		go func() {
			for {
				select {
				case queuedItem := <-i.IngestionBulkIndexingQueue:

					v := queuedItem.Variant
					wg := queuedItem.WaitGroup

					// Prepare the data payload: encode article to JSON
					data, err := json.Marshal(v)
					if err != nil {
						log.Fatalf("Cannot encode variant %s: %s\n", v.Id, err)
					}

					// Add an item to the BulkIndexer
					err = i.IngestionBulkIndexer.Add(
						context.Background(),
						esutil.BulkIndexerItem{
							// Action field configures the operation to perform (index, create, delete, update)
							Action: "index",

							// Body is an `io.Reader` with the payload
							Body: bytes.NewReader(data),

							// OnSuccess is called for each successful operation
							OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
								defer wg.Done()
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
					if err != nil {
						defer wg.Done()
						fmt.Printf("Unexpected error: %s", err)
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

func (i *IngestionService) UploadVcfGzToDrs(gzippedFilePath string, gzipStream *os.File, drsUrl, drsUsername, drsPassword string) string {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(gzipStream.Name()))
	io.Copy(part, gzipStream)
	writer.Close()

	// TEMP: SECURITY RISK
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	//

	// TODO: Parameterize DRS Url and credentials
	r, _ := http.NewRequest("POST", drsUrl+"/public/ingest", body)
	r.SetBasicAuth(drsUsername, drsPassword)

	r.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		fmt.Printf("Upload to DRS error: %s\n", err)
		return ""
	} else {
		fmt.Printf("Upload to DRS succeeded: %d\n", resp.StatusCode)
	}

	responsebody, bodyerr := ioutil.ReadAll(resp.Body)
	if bodyerr != nil {
		log.Printf("Error reading body: %v", err)
		return ""
	}

	jsonParsed, err := gabs.ParseJSON(responsebody)
	if err != nil {
		fmt.Printf("Parsing error: %s\n", err)
		return ""
	}
	id := jsonParsed.Path("id").Data().(string)

	fmt.Println("Get DRS ID: ", id)

	return id
}

func (i *IngestionService) ProcessVcf(vcfFilePath string, drsFileId string, assemblyId constants.AssemblyId) {
	f, err := os.Open(vcfFilePath)
	if err != nil {
		fmt.Println("Failed to open file - ", err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var discoveredHeaders bool = false
	var headers []string

	nonNumericRegexp := regexp.MustCompile("[^.0-9]")

	var _fileWG sync.WaitGroup

	for scanner.Scan() {
		//fmt.Println(scanner.Text())

		// Gather Header row by seeking the CHROM string
		line := scanner.Text()
		if !discoveredHeaders {
			if line[0] == '#' {
				if strings.Contains(line, "CHROM") {
					// Split the string by tabs
					headers = strings.Split(line, "\t")
					discoveredHeaders = true

					fmt.Println("Found the headers: ", headers)
					continue
				}
				continue
			}
		}

		_fileWG.Add(1)
		go func(line string, drsFileId string, fileWg *sync.WaitGroup) {
			// ----  break up line
			rowComponents := strings.Split(line, "\t")

			// ----  process more...
			var tmpSamples []map[string]interface{}
			tmpSamplesMutex := sync.RWMutex{}

			tmpVariant := make(map[string]interface{})
			tmpVariantMapMutex := sync.RWMutex{}

			tmpVariant["fileId"] = drsFileId
			tmpVariant["assemblyId"] = assemblyId

			var rowWg sync.WaitGroup
			rowWg.Add(len(rowComponents))

			for rowIndex, rowComponent := range rowComponents {
				go func(i int, rc string, rwg *sync.WaitGroup) {
					defer rwg.Done()
					key := strings.ToLower(strings.TrimSpace(strings.Replace(headers[i], "#", "", -1)))
					value := strings.TrimSpace(rc)

					// if not a vcf header, assume it's a sampleId header
					if utils.StringInSlice(key, models.VcfHeaders) {

						// filter field type by column name
						if key == "chrom" || key == "pos" || key == "qual" {
							if key == "chrom" {
								// Strip out all non-numeric characters
								value = nonNumericRegexp.ReplaceAllString(value, "")
							}

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
						} else if key == "info" {
							var allInfos []*models.Info

							// Split all alleles by semi-colon
							semiColonSeparations := strings.Split(value, ";")

							for _, scSep := range semiColonSeparations {
								// Split by equality symbol
								equalitySeparations := strings.Split(scSep, "=")

								if len(equalitySeparations) == 2 {
									allInfos = append(allInfos, &models.Info{
										Id:    equalitySeparations[0],
										Value: equalitySeparations[1],
									})
								} else { // len(equalitySeparations) == 1
									allInfos = append(allInfos, &models.Info{
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
						tmpSamplesMutex.Lock()

						tmpSamples = append(tmpSamples, map[string]interface{}{
							"key":    key,
							"values": strings.Split(value, ":"),
						})

						tmpSamplesMutex.Unlock()
					}
				}(rowIndex, rowComponent, &rowWg)
			}

			rowWg.Wait()

			// --- TODO: prep formats + samples
			var samples []*models.Sample

			// ---- get genotype stuff
			var (
				hasGenotype      bool = false
				genotypePosition int  = 0

				hasGenotypeProbability      bool = false
				genotypeProbabilityPosition int  = 0

				hasPhredScaleLikelyhood      bool = false
				phredScaleLikelyhoodPosition int  = 0
			)

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

			for _, ts := range tmpSamples {
				sample := &models.Sample{}
				variation := &models.Variation{}

				tmpKeyString := ts["key"].(string)
				tmpValueStrings := ts["values"].([]string)
				for k := range tmpValueStrings {
					if hasGenotype && k == genotypePosition {
						// create genotype from value
						gtString := tmpValueStrings[k]
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
						var zygosity constants.Zygosity

						if alleleLeft == -1 || alleleRight == -1 {
							zygosity = z.Unknown
						} else {
							switch alleleLeft == alleleRight {
							case true:
								zygosity = z.Homozygous
							case false:
								zygosity = z.Heterozygous
							}
						}

						variation.Genotype = models.Genotype{
							Phased:      phased,
							AlleleLeft:  alleleLeft,
							AlleleRight: alleleRight,
							Zygosity:    zygosity,
						}
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

			tmpVariant["samples"] = samples
			// ---	 push to a bulk "queue"
			var resultingVariant models.Variant
			mapstructure.Decode(tmpVariant, &resultingVariant)

			// pass variant (along with a waitgroup) to the channel
			i.IngestionBulkIndexingQueue <- &IngestionQueueStructure{
				Variant:   &resultingVariant,
				WaitGroup: fileWg,
			}

		}(line, drsFileId, &_fileWG)
	}

	// let all lines be queued up and processed
	_fileWG.Wait()
}

func (i *IngestionService) FilenameAlreadyRunning(filename string) bool {
	for _, v := range i.IngestRequestMap {
		if v.Filename == filename && (v.State == ingest.Queued || v.State == ingest.Running) {
			return true
		}
	}
	return false
}
