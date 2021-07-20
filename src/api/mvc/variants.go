package mvc

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"bufio"
	"io/ioutil"
	"regexp"
	"time"

	"api/contexts"
	"api/utils"

	"github.com/labstack/echo"
)

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

	// check errors
	//fmt.Println(respBuf.String())

	// Declared an empty interface
	var result map[string]interface{}

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

	// Create 'files' and 'variants' indices
	// files_index_req := esapi.IndexRequest{Index: "files"}
	// variants_index_req := esapi.IndexRequest{Index: "variants"}

	// files_index_res, fierr := files_index_req.Do(context.Background(), es)
	// if fierr != nil {
	// 	fmt.Printf("Failed: %s\n", fierr)
	// 	return fierr
	// } else {
	// 	fmt.Printf("Files Index created!: %s\n", files_index_res)
	// }

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
	// ---	 load back into memory and process
	// ---	 push to a bulk "queue"
	// ---	 once the queue is maxed out, wait for the bulk queue to be processed before continuing
	for _, vcfGzFile := range vcfGzfiles {
		go func(file string) {
			gzippedFilePath := fmt.Sprintf("%s%s%s", vcfDirPath, "/", file)
			r, err := os.Open(gzippedFilePath)
			if err != nil {
				fmt.Printf("error opening %s: %s\n", file, err)
			}
			extractVcfGz(gzippedFilePath, r)
		}(vcfGzFile)
	}

	fmt.Printf("Query End: %s\n", time.Now())

	return c.JSON(http.StatusOK, "{\"ingest\" : \"Done! Maybe it succeeded, maybe it failed!\"}")
}

func extractVcfGz(gzippedFilePath string, gzipStream io.Reader) {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Fatal("ExtractTarGz: NewReader failed")
	}

	// new File name
	vcfFileName := strings.Replace(gzippedFilePath, ".gz", "", -1)

	fmt.Printf("Creating File: %s\n", vcfFileName)
	f, err := os.Create(vcfFileName)

	fmt.Printf("Writing to file: %s\n", vcfFileName)
	w := bufio.NewWriter(f)
	io.Copy(w, uncompressedStream)

	uncompressedStream.Close()
	f.Close()
}
