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

	"bufio"
	"io/ioutil"
	"regexp"
	"time"

	"api/contexts"
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
	es := c.(*contexts.EsContext).Client

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
	es := c.(*contexts.EsContext).Client

	// Create 'variants' indices
	// variants_index_req := esapi.IndexRequest{Index: "variants"}

	// variants_index_res, vierr := variants_index_req.Do(context.Background(), es)
	// if vierr != nil {
	// 	fmt.Printf("Failed: %s\n", vierr)
	// 	return vierr
	// } else {
	// 	fmt.Printf("Variants Index created!: %s\n", variants_index_res)
	// }

	fmt.Printf("Ingest Start: %s\n", time.Now())

	// get vcf files
	// TODO: refactor
	vcfDirPath := "/home/brennan/Public/McGill/gohan/Gohan.Console/vcfs"
	var vcfGzfiles []string

	// Read all files
	fileInfo, err := ioutil.ReadDir(vcfDirPath)
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

	// ingest
	// -- foreach compressed vcf file:
	// ---	 decompress vcf.gz
	// ---	 store to disk (temporarily)
	// ---   push compressed to DRS
	// ---	 load back into memory and process
	// ---	 push to a bulk "queue"
	// ---	 once the queue is maxed out, wait for the bulk queue to be processed before continuing
	// ---   delete the temporary vcf file
	for _, vcfGzFile := range vcfGzfiles {
		go func(file string) {
			// ---	 decompress vcf.gz
			gzippedFilePath := fmt.Sprintf("%s%s%s", vcfDirPath, "/", file)
			r, err := os.Open(gzippedFilePath)
			if err != nil {
				fmt.Printf("error opening %s: %s\n", file, err)
				return
			}
			defer r.Close()

			vcfFilePath := extractVcfGz(gzippedFilePath, r)
			if vcfFilePath == "" {
				fmt.Println("Something went wrong: filepath is empty for ", file)
				return
			}

			// ---   push compressed to DRS
			drsId := uploadVcfGzToDrs(gzippedFilePath, r)
			if drsId == "" {
				fmt.Println("Something went wrong: DRS Id is empty for ", file)
				return
			}

			// ---	 load back into memory and process
			processVcf(vcfFilePath, drsId, es)

			// ---   delete the temporary vcf file
			os.Remove(vcfFilePath)

			fmt.Printf("Ingest End for file at %s : %s\n", vcfFilePath, time.Now())
		}(vcfGzFile)
	}

	return c.JSON(http.StatusOK, "{\"ingest\" : \"Done! Maybe it succeeded, maybe it failed!\"}")
}

func extractVcfGz(gzippedFilePath string, gzipStream io.Reader) string {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Fatal("ExtractTarGz: NewReader failed - ", err)
		return ""
	}

	// ---	 store to disk (temporarily)
	// new File name
	vcfFilePath := strings.Replace(gzippedFilePath, ".gz", "", -1)

	fmt.Printf("Creating File: %s\n", vcfFilePath)
	f, err := os.Create(vcfFilePath)
	if err != nil {
		fmt.Println("Something went wrong:  ", err)
		return ""
	}

	fmt.Printf("Writing to file: %s\n", vcfFilePath)
	w := bufio.NewWriter(f)
	io.Copy(w, uncompressedStream)

	uncompressedStream.Close()
	f.Close()

	return vcfFilePath
}

func uploadVcfGzToDrs(gzippedFilePath string, gzipStream *os.File) string {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(gzipStream.Name()))
	io.Copy(part, gzipStream)
	writer.Close()

	// TEMP: SECURITY RISK
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	//

	// TODO: Parameterize DRS Url
	r, _ := http.NewRequest("POST", "https://drs.gohan.local/public/ingest", body)
	r.SetBasicAuth("drsadmin", "gohandrspassword123")

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

func processVcf(vcfFilePath string, drsId string, es *elasticsearch.Client) {
	f, err := os.Open(vcfFilePath)
	if err != nil {
		fmt.Println("Failed to open file - ", err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var discoveredHeaders bool = false
	var headers []string

	var variants []*utils.Variant

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

		go func(line string, dId string) {

			// ----  break up line
			rowComponents := strings.Split(line, "\t")

			// ----  process more
			var samples []*utils.Sample
			tmpVariant := make(map[string]interface{})

			tmpVariant["fileId"] = dId

			for i, rc := range rowComponents {
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
								tmpVariant[key] = value
							} else {
								tmpVariant[key] = -1 // here to simulate a null value (such as when the string value is empty, or
								//                      is something as arbitrary as a single period '.')
							}

						} else if key == "alt" || key == "ref" {
							// Split all alleles by comma
							tmpVariant[key] = strings.Split(value, ",")
						} else if key == "info" {
							var allInfos []*utils.Info

							// Split all alleles by semi-colon
							semiColonSeparations := strings.Split(value, ";")

							for _, scSep := range semiColonSeparations {
								// Split by equality symbol
								equalitySeparations := strings.Split(scSep, "=")

								if len(equalitySeparations) == 2 {
									allInfos = append(allInfos, &utils.Info{
										Id:    equalitySeparations[0],
										Value: equalitySeparations[1],
									})
								} else { // len(equalitySeparations) == 1
									allInfos = append(allInfos, &utils.Info{
										Id:    "",
										Value: equalitySeparations[0],
									})
								}
							}

							tmpVariant[key] = allInfos

						} else {
							tmpVariant[key] = value
						}
					} else { // assume its a sampleId header
						samples = append(samples, &utils.Sample{
							SampleId:  key,
							Variation: value,
						})
					}
				}
			}

			tmpVariant["samples"] = samples

			// newVariant := &utils.Variant{
			// 	fileId: dId,
			// }

			// ---	 push to a bulk "queue"
			var resultingVariant utils.Variant
			mapstructure.Decode(tmpVariant, &resultingVariant)

			variants = append(variants, &resultingVariant)

		}(line, drsId)

	}

	// ---	 push all data to the bulk indexer
	fmt.Printf("Number of CPUs available: %d\n", runtime.NumCPU())

	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:      "variants",
		Client:     es,
		NumWorkers: runtime.NumCPU(), // The number of worker goroutines
	})

	var wg sync.WaitGroup
	wg.Add(len(variants))

	for _, v := range variants {

		// Prepare the data payload: encode article to JSON
		data, err := json.Marshal(v)
		if err != nil {
			log.Fatalf("Cannot encode variant %d: %s\n", v.Id, err)
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
					//log.Printf("Added item: %s", item.DocumentID)
				},

				// OnFailure is called for each failed operation
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					defer wg.Done()
					if err != nil {
						log.Printf("ERROR: %s", err)
					} else {
						log.Printf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
					}
				},
			},
		)
		if err != nil {
			defer wg.Done()
			log.Fatalf("Unexpected error: %s", err)
		}
	}

	wg.Wait()

	fmt.Printf("Done processing %s with %d variants!\n", vcfFilePath, len(variants))

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
