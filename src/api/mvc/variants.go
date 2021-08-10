package mvc

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"io/ioutil"
	"regexp"
	"time"

	"api/contexts"
	"api/models"
	esRepo "api/repositories/elasticsearch"
	"api/utils"

	"github.com/elastic/go-elasticsearch"
	"github.com/google/uuid"

	"github.com/labstack/echo"
	"github.com/mitchellh/mapstructure"
)

func VariantsGetByVariantId(c echo.Context) error {
	// retrieve variant Ids from query parameter (comma separated)
	variantIds := strings.Split(c.QueryParam("ids"), ",")
	if len(variantIds[0]) == 0 {
		// if no ids were provided, assume "wildcard" search
		variantIds = []string{"*"}
	}

	return executeGetByIds(c, variantIds, true)
}

func VariantsGetBySampleId(c echo.Context) error {
	// retrieve sample Ids from query parameter (comma separated)
	sampleIds := strings.Split(c.QueryParam("ids"), ",")
	if len(sampleIds[0]) == 0 {
		// if no ids were provided, assume "wildcard" search
		sampleIds = []string{"*"}
	}

	return executeGetByIds(c, sampleIds, false)
}

func VariantsCountByVariantId(c echo.Context) error {
	// retrieve variant Ids from query parameter (comma separated)
	variantIds := strings.Split(c.QueryParam("ids"), ",")
	if len(variantIds[0]) == 0 {
		// if no ids were provided, assume "wildcard" search
		variantIds = []string{"*"}
	}

	return executeCountByIds(c, variantIds, true)
}

func VariantsCountBySampleId(c echo.Context) error {
	// retrieve sample Ids from query parameter (comma separated)
	sampleIds := strings.Split(c.QueryParam("ids"), ",")
	if len(sampleIds[0]) == 0 {
		// if no ids were provided, assume "wildcard" search
		sampleIds = []string{"*"}
	}

	return executeCountByIds(c, sampleIds, false)
}

func VariantsIngest(c echo.Context) error {
	es := c.(*contexts.GohanContext).Es7Client
	vcfPath := c.(*contexts.GohanContext).VcfPath
	drsUrl := c.(*contexts.GohanContext).DrsUrl
	drsUsername := c.(*contexts.GohanContext).DrsUsername
	drsPassword := c.(*contexts.GohanContext).DrsPassword

	ingestionService := c.(*contexts.GohanContext).IngestionService

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

			newRequestState := &models.IngestRequest{
				Id:        uuid.New(),
				Filename:  file,
				CreatedAt: fmt.Sprintf("%s", startTime),
			}

			newRequestState.State = "Running"
			ingestionService.IngestRequestChan <- newRequestState

			// ---	 decompress vcf.gz
			gzippedFilePath := fmt.Sprintf("%s%s%s", vcfPath, "/", file)
			r, err := os.Open(gzippedFilePath)
			if err != nil {
				msg := fmt.Sprintf("error opening %s: %s\n", file, err)
				fmt.Printf(msg)

				newRequestState.State = "Error"
				newRequestState.Message = msg
				ingestionService.IngestRequestChan <- newRequestState

				return
			}
			defer r.Close()

			vcfFilePath := ingestionService.ExtractVcfGz(gzippedFilePath, r, vcfTmpPath)
			if vcfFilePath == "" {
				msg := "Something went wrong: filepath is empty for " + file
				fmt.Println(msg)

				newRequestState.State = "Error"
				newRequestState.Message = msg
				ingestionService.IngestRequestChan <- newRequestState

				return
			}

			// ---   push compressed to DRS
			drsFileId := ingestionService.UploadVcfGzToDrs(gzippedFilePath, r, drsUrl, drsUsername, drsPassword)
			if drsFileId == "" {
				msg := "Something went wrong: DRS File Id is empty for " + file
				fmt.Println(msg)

				newRequestState.State = "Error"
				newRequestState.Message = msg
				ingestionService.IngestRequestChan <- newRequestState

				return
			}

			// ---	 load back into memory and process
			ingestionService.ProcessVcf(vcfFilePath, drsFileId, es)

			// ---   delete the temporary vcf file
			os.Remove(vcfFilePath)

			// ---   delete full tmp path and all contents
			// 		 (WARNING : Only do this when running over a single file)
			//os.RemoveAll(vcfTmpPath)

			fmt.Printf("Ingest duration for file at %s : %s\n", vcfFilePath, time.Now().Sub(startTime))

			newRequestState.State = "Done"
			ingestionService.IngestRequestChan <- newRequestState

		}(fileName)
	}

	// TODO: create a standard response object
	return c.JSON(http.StatusOK, "{\"ingest\" : \"Done! Maybe it succeeded, maybe it failed.. Check the debug logs!\"}")
}

func executeGetByIds(c echo.Context, ids []string, isVariantIdQuery bool) error {

	var es, chromosome, lowerBound, upperBound, reference, alternative = retrieveCommonElements(c)

	// retrieve other query parameters relevent to this 'get' query ---
	sizeQP := c.QueryParam("size")
	var (
		defaultSize = 100
		size        int
	)

	size = defaultSize
	if len(sizeQP) > 0 {
		parsedSize, sErr := strconv.Atoi(sizeQP)

		if sErr == nil && parsedSize != 0 {
			size = parsedSize
		}
	}

	sortByPosition := c.QueryParam("sortByPosition")

	includeSamplesInResultSetQP := c.QueryParam("includeSamplesInResultSet")
	var (
		includeSamplesInResultSet bool
		isirsErr                  error
	)
	if len(includeSamplesInResultSetQP) > 0 {
		includeSamplesInResultSet, isirsErr = strconv.ParseBool(includeSamplesInResultSetQP)
		if isirsErr != nil {
			log.Fatal(isirsErr)
		}
	}
	// ---

	// prepare response
	respDTO := models.VariantsResponseDTO{}
	respDTOMux := sync.RWMutex{}

	// TODO: optimize - make 1 repo call with all variantIds at once
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)

		go func(_id string) {
			defer wg.Done()

			variantRespDataModel := models.VariantResponseDataModel{}

			var docs map[string]interface{}
			if isVariantIdQuery {
				variantRespDataModel.VariantId = _id

				fmt.Printf("Executing Get-Variants for VariantId %s\n", _id)

				docs = esRepo.GetDocumentsContainerVariantOrSampleIdInPositionRange(es,
					chromosome, lowerBound, upperBound,
					_id, "", // note : "" is for sampleId
					reference, alternative,
					size, sortByPosition, includeSamplesInResultSet)
			} else {
				// implied sampleId query
				variantRespDataModel.SampleId = _id

				fmt.Printf("Executing Get-Samples for SampleId %s\n", _id)

				docs = esRepo.GetDocumentsContainerVariantOrSampleIdInPositionRange(es,
					chromosome, lowerBound, upperBound,
					"", _id, // note : "" is for variantId
					reference, alternative,
					size, sortByPosition, includeSamplesInResultSet)
			}

			// query for each id

			docsHits := docs["hits"].(map[string]interface{})["hits"]
			allDocHits := []map[string]interface{}{}
			mapstructure.Decode(docsHits, &allDocHits)

			// grab _source for each hit
			var allSources []map[string]interface{}

			for _, r := range allDocHits {
				source := r["_source"].(map[string]interface{})
				allSources = append(allSources, source)
			}

			fmt.Printf("Found %d docs!\n", len(allSources))

			variantRespDataModel.Count = len(allSources)
			variantRespDataModel.Results = allSources

			respDTOMux.Lock()
			respDTO.Data = append(respDTO.Data, variantRespDataModel)
			respDTOMux.Unlock()

		}(id)
	}

	wg.Wait()

	respDTO.Status = 200
	respDTO.Message = "Success"

	return c.JSON(http.StatusOK, respDTO)
}

func executeCountByIds(c echo.Context, ids []string, isVariantIdQuery bool) error {

	var es, chromosome, lowerBound, upperBound, reference, alternative = retrieveCommonElements(c)

	respDTO := models.VariantsResponseDTO{}
	respDTOMux := sync.RWMutex{}

	// TODO: optimize - make 1 repo call with all variantIds at once
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)

		go func(_id string) {
			defer wg.Done()

			variantRespDataModel := models.VariantResponseDataModel{}

			var docs map[string]interface{}
			if isVariantIdQuery {
				variantRespDataModel.VariantId = _id

				fmt.Printf("Executing Count-Variants for VariantId %s\n", _id)

				docs = esRepo.CountDocumentsContainerVariantOrSampleIdInPositionRange(es,
					chromosome, lowerBound, upperBound,
					_id, "", // note : "" is for sampleId
					reference, alternative)
			} else {
				// implied sampleId query
				variantRespDataModel.SampleId = _id

				fmt.Printf("Executing Count-Samples for SampleId %s\n", _id)

				docs = esRepo.CountDocumentsContainerVariantOrSampleIdInPositionRange(es,
					chromosome, lowerBound, upperBound,
					"", _id, // note : "" is for variantId
					reference, alternative)
			}

			variantRespDataModel.Count = int(docs["count"].(float64))

			respDTOMux.Lock()
			respDTO.Data = append(respDTO.Data, variantRespDataModel)
			respDTOMux.Unlock()

		}(id)
	}

	wg.Wait()

	respDTO.Status = 200
	respDTO.Message = "Success"

	return c.JSON(http.StatusOK, respDTO)
}

func retrieveCommonElements(c echo.Context) (*elasticsearch.Client, string, int, int, string, string) {
	es := c.(*contexts.GohanContext).Es7Client

	chromosome := c.QueryParam("chromosome")
	if len(chromosome) == 0 {
		// if no chromosome is provided, assume "wildcard" search
		chromosome = "*"
	}

	lowerBoundQP := c.QueryParam("lowerBound")
	var (
		lowerBound int
		lbErr      error
	)
	if len(lowerBoundQP) > 0 {
		lowerBound, lbErr = strconv.Atoi(lowerBoundQP)
		if lbErr != nil {
			log.Fatal(lbErr)
		}
	}

	upperBoundQP := c.QueryParam("upperBound")
	var (
		upperBound int
		ubErr      error
	)
	if len(upperBoundQP) > 0 {
		upperBound, ubErr = strconv.Atoi(upperBoundQP)
		if ubErr != nil {
			log.Fatal(ubErr)
		}
	}

	reference := c.QueryParam("reference")

	alternative := c.QueryParam("alternative")

	return es, chromosome, lowerBound, upperBound, reference, alternative
}
