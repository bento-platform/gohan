package mvc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"io/ioutil"
	"regexp"
	"time"

	"api/contexts"
	esRepo "api/repositories/elasticsearch"
	"api/services"
	"api/utils"

	"github.com/labstack/echo"
)

func VariantsGetByVariantId(c echo.Context) error {
	// Testing ES
	es := c.(*contexts.GohanContext).Es7Client

	// retrieve query parameters
	// variantIds (comma separated)
	variantIds := strings.Split(c.QueryParam("ids"), ",")
	if len(variantIds) == 0 {
		// if no ids were provided, assume "wildcard" search
		variantIds = []string{"*"}
	}

	// TODO: optimize - make 1 repo call with all variantIds at once
	var wg sync.WaitGroup
	for i, vId := range variantIds {
		wg.Add(1)

		fmt.Printf("Queuing Get-Variants Query #%d for VariantId %s", i, vId)

		go func(vid string) {
			defer wg.Done()

			// query for each id
			docs := esRepo.GetDocumentsContainerVariantOrSampleIdInPositionRange()
			fmt.Printf("Found %d docs!\n", len(docs))
		}(vId)
	}

	wg.Wait()
}

func VariantsGetBySampleId(c echo.Context) error {}

func VariantsCountByVariantId(c echo.Context) error {}

func VariantsCountBySampleId(c echo.Context) error {}

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

			vcfFilePath := services.ExtractVcfGz(gzippedFilePath, r, vcfTmpPath)
			if vcfFilePath == "" {
				fmt.Println("Something went wrong: filepath is empty for ", file)
				return
			}

			// ---   push compressed to DRS
			drsFileId := services.UploadVcfGzToDrs(gzippedFilePath, r, drsUrl, drsUsername, drsPassword)
			if drsFileId == "" {
				fmt.Println("Something went wrong: DRS File Id is empty for ", file)
				return
			}

			// ---	 load back into memory and process
			services.ProcessVcf(vcfFilePath, drsFileId, es)

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
