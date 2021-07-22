package mvc

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"bufio"
	"io/ioutil"
	"regexp"
	"time"

	"api/contexts"
	"api/models"
	"api/utils"

	"github.com/elastic/go-elasticsearch"
	"github.com/elastic/go-elasticsearch/esutil"
	"github.com/mitchellh/mapstructure"

	"github.com/Jeffail/gabs"
	"github.com/labstack/echo"
)

var vcfHeaders = []string{"chrom", "pos", "id", "ref", "alt", "qual", "filter", "info", "format"}

func VariantsSearchTest(c echo.Context) error {
	// Testing ES
	es := c.(*contexts.GohanContext).Es7Client

	fmt.Printf("Query Start: %s", time.Now())

	// Build the request body.
	var queryString = `
	"bool": {
		"filter": [
			{
				"bool": {
					"must": [
						{
							"query_string": {
							"query": "chrom:*"
							}
						}
					]
				}
			}
		]
	}
	`
	query := utils.ConstructQuery(queryString, 2)
	var buf bytes.Buffer
	if buffErr := json.NewEncoder(&buf).Encode(query); buffErr != nil {
		fmt.Printf("Error encoding query: %s", buffErr)
	}

	// Perform the search request.
	res, searchErr := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("variants"),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if searchErr != nil {
		fmt.Printf("Error getting response: %s", searchErr)
	}

	respBuf := new(strings.Builder)
	_, respErr := io.Copy(respBuf, res.Body)
	if respErr != nil {
		fmt.Printf("Error forming response: %s", respErr)
	}

	// Declared an empty interface
	result := make(map[string]interface{})

	// Unmarshal or Decode the JSON to the interface.
	json.Unmarshal([]byte(respBuf.String()), &result)

	// Close the response
	defer res.Body.Close()

	fmt.Printf("Query End: %s", time.Now())

	return c.JSON(http.StatusOK, result)
}

func VariantsIngestTest(c echo.Context) error {
	// Testing ES
	es := c.(*contexts.GohanContext).Es7Client
	vcfPath := c.(*contexts.GohanContext).VcfPath
	drsUrl := c.(*contexts.GohanContext).DrsUrl
	drsUsername := c.(*contexts.GohanContext).DrsUsername
	drsPassword := c.(*contexts.GohanContext).DrsPassword

	// retrieve query parameters (comman separated)
	fileNames := strings.Split(c.QueryParam("fileNames"), ",")
	for _, fileName := range fileNames {
		if fileName == "" {
			// TODO: create a standard response object
			return c.JSON(http.StatusBadRequest, "{\"error\" : \"Missing 'fileNames' query parameter!\"}")
		}
	}

	startTime := time.Now()

	fmt.Printf("Ingest Start: %s\n", startTime)

	// get vcf files
	var vcfGzfiles []string

	// Read all files
	fileInfo, err := ioutil.ReadDir(vcfPath)
	if err != nil {
		fmt.Printf("Failed: %s\n", err)
		return err
	}

	// Filter only .vcf.gz files
	for _, file := range fileInfo {
		if matched, _ := regexp.MatchString(".vcf.gz", file.Name()); matched {
			vcfGzfiles = append(vcfGzfiles, file.Name())
		} else {
			fmt.Printf("Skipping %s\n", file.Name())
		}
	}

	fmt.Printf("Found .vcf.gz files: %s\n", vcfGzfiles)

	// Locate fileName from request inside found files
	for _, fileName := range fileNames {
		if utils.StringInSlice(fileName, vcfGzfiles) == false {
			return c.JSON(http.StatusBadRequest, "{\"error\" : \"file "+fileName+" not found! Aborted -- \"}")
		}
	}

	// create temporary directory for unzipped vcfs
	vcfTmpPath := vcfPath + "/tmp"
	_, err = os.Stat(vcfTmpPath)
	if os.IsNotExist(err) {
		fmt.Println("VCF /tmp folder does not exist, creating...")
		err = os.Mkdir(vcfTmpPath, 0755)
		if err != nil {
			fmt.Println(err)

			// TODO: create a standard response object
			return err
		}
	}

	// ingest vcf
	// TODO: create long-polling for status check
	// of long-running process
	for _, fileName := range fileNames {
		go func(file string) {
			// ---	 decompress vcf.gz
			gzippedFilePath := fmt.Sprintf("%s%s%s", vcfPath, "/", file)
			r, err := os.Open(gzippedFilePath)
			if err != nil {
				fmt.Printf("error opening %s: %s\n", file, err)
				return
			}
			defer r.Close()

			vcfFilePath := extractVcfGz(gzippedFilePath, r, vcfTmpPath)
			if vcfFilePath == "" {
				fmt.Println("Something went wrong: filepath is empty for ", file)
				return
			}

			// ---   push compressed to DRS
			drsFileId := uploadVcfGzToDrs(gzippedFilePath, r, drsUrl, drsUsername, drsPassword)
			if drsFileId == "" {
				fmt.Println("Something went wrong: DRS File Id is empty for ", file)
				return
			}

			// ---	 load back into memory and process
			processVcf(vcfFilePath, drsFileId, es)

			// ---   delete the temporary vcf file
			os.Remove(vcfFilePath)

			// ---   delete full tmp path and all contents
			// 		 (WARNING : Only do this when running over a single file)
			//os.RemoveAll(vcfTmpPath)

			fmt.Printf("Ingest duration for file at %s : %s\n", vcfFilePath, time.Now().Sub(startTime))
		}(fileName)
	}

	// TODO: create a standard response object
	return c.JSON(http.StatusOK, "{\"ingest\" : \"Done! Maybe it succeeded, maybe it failed.. Check the debug logs!\"}")
}

func extractVcfGz(gzippedFilePath string, gzipStream io.Reader, vcfTmpPath string) string {
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

func uploadVcfGzToDrs(gzippedFilePath string, gzipStream *os.File, drsUrl, drsUsername, drsPassword string) string {
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

func processVcf(vcfFilePath string, drsFileId string, es *elasticsearch.Client) {
	f, err := os.Open(vcfFilePath)
	if err != nil {
		fmt.Println("Failed to open file - ", err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var discoveredHeaders bool = false
	var headers []string

	var variants []*models.Variant
	variantsMutex := sync.RWMutex{}

	nonNumericRegexp := regexp.MustCompile("[^.0-9]")

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

		go func(line string, drsFileId string) {

			// ----  break up line
			rowComponents := strings.Split(line, "\t")

			// ----  process more...
			var samples []*models.Sample
			samplesMutex := sync.RWMutex{}
			tmpVariant := make(map[string]interface{})
			tmpVariantMapMutex := sync.RWMutex{}

			tmpVariant["fileId"] = drsFileId

			var rowWg sync.WaitGroup
			rowWg.Add(len(rowComponents))

			for rowIndex, rowComponent := range rowComponents {
				go func(i int, rc string, rwg *sync.WaitGroup) {
					defer rwg.Done()
					if rc != "0|0" {
						key := strings.ToLower(strings.TrimSpace(strings.Replace(headers[i], "#", "", -1)))
						value := strings.TrimSpace(rc)

						// if not a vcf header, assume it's a sampleId header
						if utils.StringInSlice(key, vcfHeaders) {

							// filter field type by column name
							if key == "chrom" || key == "pos" || key == "qual" {
								if key == "chrom" {
									// Strip out all non-numeric characters
									value = nonNumericRegexp.ReplaceAllString(value, "")
								}

								// // Convert string's to int's, if possible
								value, err := strconv.ParseInt(value, 10, 0)
								if err != nil {
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
							samplesMutex.Lock()
							samples = append(samples, &models.Sample{
								SampleId:  key,
								Variation: value,
							})
							samplesMutex.Unlock()
						}
					}
				}(rowIndex, rowComponent, &rowWg)
			}

			rowWg.Wait()

			tmpVariant["samples"] = samples

			// ---	 push to a bulk "queue"
			var resultingVariant models.Variant
			mapstructure.Decode(tmpVariant, &resultingVariant)

			variantsMutex.Lock()
			variants = append(variants, &resultingVariant)
			variantsMutex.Unlock()

		}(line, drsFileId)
	}

	// ---	 push all data to the bulk indexer
	fmt.Printf("Number of CPUs available: %d\n", runtime.NumCPU())

	var countSuccessful uint64
	var countFailed uint64

	// see: https://www.elastic.co/blog/why-am-i-seeing-bulk-rejections-in-my-elasticsearch-cluster
	var numWorkers = len(variants) / 50
	// the lower the denominator (the number of documents per bulk upload). the higher
	// the chances of 100% successful upload, though the longer it may take (negligible)

	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:      "variants",
		Client:     es,
		NumWorkers: numWorkers,
		// FlushBytes:    int(flushBytes),  // The flush threshold in bytes (default: 50MB ?)
		FlushInterval: 30 * time.Second, // The periodic flush interval
	})

	var wg sync.WaitGroup

	for _, v := range variants {

		wg.Add(1)

		// Prepare the data payload: encode article to JSON
		data, err := json.Marshal(v)
		if err != nil {
			log.Fatalf("Cannot encode variant %s: %s\n", v.Id, err)
		}

		// Add an item to the BulkIndexer
		err = bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				// Action field configures the operation to perform (index, create, delete, update)
				Action: "index",

				// Body is an `io.Reader` with the payload
				Body: bytes.NewReader(data),

				// OnSuccess is called for each successful operation
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					defer wg.Done()
					atomic.AddUint64(&countSuccessful, 1)

					//log.Printf("Added item: %s", item.DocumentID)
				},

				// OnFailure is called for each failed operation
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					defer wg.Done()
					atomic.AddUint64(&countFailed, 1)
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

	wg.Wait()

	fmt.Printf("Done processing %s with %d variants, with %d stats!\n", vcfFilePath, len(variants), bi.Stats())

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
