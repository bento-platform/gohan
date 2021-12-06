package mvc

import (
	"fmt"
	"io"
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
	"api/models/constants"
	a "api/models/constants/assembly-id"
	gq "api/models/constants/genotype-query"
	s "api/models/constants/sort"
	"api/models/ingest"
	esRepo "api/repositories/elasticsearch"
	"api/utils"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/google/uuid"

	"github.com/labstack/echo"
	"github.com/mitchellh/mapstructure"
)

func VariantsGetByVariantId(c echo.Context) error {
	fmt.Printf("[%s] - VariantsGetByVariantId hit!\n", time.Now())
	// retrieve variant Ids from query parameter (comma separated)
	variantIds := strings.Split(c.QueryParam("ids"), ",")
	if len(variantIds[0]) == 0 {
		// if no ids were provided, assume "wildcard" search
		variantIds = []string{"*"}
	}

	return executeGetByIds(c, variantIds, true)
}

func VariantsGetBySampleId(c echo.Context) error {
	fmt.Printf("[%s] - VariantsGetBySampleId hit!\n", time.Now())
	// retrieve sample Ids from query parameter (comma separated)
	sampleIds := strings.Split(c.QueryParam("ids"), ",")
	if len(sampleIds[0]) == 0 {
		// if no ids were provided, assume "wildcard" search
		sampleIds = []string{"*"}
	}

	return executeGetByIds(c, sampleIds, false)
}

func VariantsCountByVariantId(c echo.Context) error {
	fmt.Printf("[%s] - VariantsCountByVariantId hit!\n", time.Now())
	// retrieve single variant id from query parameter and map to a list
	// to conform to function signature
	singleVariantIdSlice := []string{c.QueryParam("id")}
	if len(singleVariantIdSlice[0]) == 0 {
		// if no id was provided, assume "wildcard" search
		singleVariantIdSlice = []string{"*"}
	}

	return executeCountByIds(c, singleVariantIdSlice, true)
}

func VariantsCountBySampleId(c echo.Context) error {
	fmt.Printf("[%s] - VariantsCountBySampleId hit!\n", time.Now())
	// retrieve single sample id from query parameter and map to a list
	// to conform to function signature
	singleSampleIdSlice := []string{c.QueryParam("id")}
	if len(singleSampleIdSlice[0]) == 0 {
		// if no id was provided, assume "wildcard" search
		singleSampleIdSlice = []string{"*"}
	}

	return executeCountByIds(c, singleSampleIdSlice, false)
}

func VariantsIngest(c echo.Context) error {
	fmt.Printf("[%s] - VariantsIngest hit!\n", time.Now())
	cfg := c.(*contexts.GohanContext).Config
	vcfPath := cfg.Api.VcfPath
	drsUrl := cfg.Drs.Url
	drsUsername := cfg.Drs.Username
	drsPassword := cfg.Drs.Password

	ingestionService := c.(*contexts.GohanContext).IngestionService

	// retrieve query parameters (comman separated)
	fileNames := strings.Split(c.QueryParam("fileNames"), ",")
	for _, fileName := range fileNames {
		if fileName == "" {
			// TODO: create a standard response object
			return c.JSON(http.StatusBadRequest, "{\"error\" : \"Missing 'fileNames' query parameter!\"}")
		}
	}

	assemblyId := a.CastToAssemblyId(c.QueryParam("assemblyId"))

	// -- optional filter
	var (
		filterOutHomozygousReferences bool = false // default
		fohrErr                       error
	)
	filterOutHomozygousReferencesQP := c.QueryParam("filterOutHomozygousReferences")
	if len(filterOutHomozygousReferencesQP) > 0 {
		filterOutHomozygousReferences, fohrErr = strconv.ParseBool(filterOutHomozygousReferencesQP)
		if fohrErr != nil {
			// TODO: create a standard response object
			log.Fatal(fohrErr)
		}
	}

	startTime := time.Now()

	fmt.Printf("Ingest Start: %s\n", startTime)

	// get vcf files
	var vcfGzfiles []string

	// TODO: simply load files by filename provided
	// rather than load all available files and looping over them
	// -----
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

	// Locate fileName from request inside found files
	for _, fileName := range fileNames {
		if utils.StringInSlice(fileName, vcfGzfiles) == false {
			return c.JSON(http.StatusBadRequest, "{\"error\" : \"file "+fileName+" not found! Aborted -- \"}")
		}
	}
	// -----

	// create temporary directory for unzipped vcfs
	vcfTmpPath := fmt.Sprintf("%s/tmp", vcfPath)
	_, err = os.Stat(vcfTmpPath)
	if os.IsNotExist(err) {
		fmt.Printf("VCF %s folder does not exist -- creating...", vcfTmpPath)
		err = os.Mkdir(vcfTmpPath, 0755)
		if err != nil {
			fmt.Println(err)

			// TODO: create a standard response object
			return err
		}
	}

	// ingest vcf
	responseDtos := []ingest.IngestResponseDTO{}
	for _, fileName := range fileNames {

		// check if there is an already existing ingestion request state
		if ingestionService.FilenameAlreadyRunning(fileName) {
			responseDtos = append(responseDtos, ingest.IngestResponseDTO{
				Filename: fileName,
				State:    ingest.Error,
				Message:  "File already being ingested..",
			})
			continue
		}

		// if not, execute

		newRequestState := &ingest.VariantIngestRequest{
			Id:        uuid.New(),
			Filename:  fileName,
			State:     ingest.Queued,
			CreatedAt: fmt.Sprintf("%v", startTime),
		}
		ingestionService.IngestRequestChan <- newRequestState

		responseDtos = append(responseDtos, ingest.IngestResponseDTO{
			Id:       newRequestState.Id,
			Filename: newRequestState.Filename,
			State:    newRequestState.State,
			Message:  "Successfully queued..",
		})

		go func(_fileName string, _newRequestState *ingest.VariantIngestRequest) {

			// take a spot in the queue
			ingestionService.ConcurrentFileIngestionQueue <- true
			go func(gzippedFileName string, reqStat *ingest.VariantIngestRequest) {
				// free up a spot in the queue
				defer func() { <-ingestionService.ConcurrentFileIngestionQueue }()

				reqStat.State = ingest.Running
				ingestionService.IngestRequestChan <- reqStat

				// ---	 decompress vcf.gz
				gzippedFilePath := fmt.Sprintf("%s%s%s", vcfPath, "/", gzippedFileName)
				r, err := os.Open(gzippedFilePath)
				if err != nil {
					msg := fmt.Sprintf("error opening %s: %s\n", gzippedFileName, err)
					fmt.Println(msg)

					reqStat.State = ingest.Error
					reqStat.Message = msg
					ingestionService.IngestRequestChan <- reqStat

					return
				}

				// ---   copy gzipped file over to a temp folder that is common to DRS and gohan
				// 	     such that DRS can load the file into memory to process rather than receiving
				//       the file from an upload, thus utilizing it's already-exisiting /private/ingest endpoind
				// -----
				tmpDestinationFileName := fmt.Sprintf("%s%s%s", cfg.Api.BridgeDirectory, "/", gzippedFileName)
				destination, err := os.Create(tmpDestinationFileName)
				if err != nil {
					msg := fmt.Sprintf("error creating temporary bridge file for %s: %s\n", gzippedFileName, err)
					fmt.Println(msg)

					reqStat.State = ingest.Error
					reqStat.Message = msg
					ingestionService.IngestRequestChan <- reqStat

					return
				}
				defer destination.Close()

				_, err = io.Copy(destination, r)
				if err != nil {
					msg := fmt.Sprintf("error copying to temporary bridge file from %s to %s: %s\n", gzippedFileName, tmpDestinationFileName, err)
					fmt.Println(msg)

					reqStat.State = ingest.Error
					reqStat.Message = msg
					ingestionService.IngestRequestChan <- reqStat

					return
				}
				// -----

				// --- tabix generation
				tabixFileDir, tabixFileName, tabixErr := ingestionService.GenerateTabix(tmpDestinationFileName)
				if tabixErr != nil {
					msg := "Something went wrong: Tabix problem " + gzippedFileName
					fmt.Println(msg)

					reqStat.State = ingest.Error
					reqStat.Message = msg
					ingestionService.IngestRequestChan <- reqStat

					return
				}

				// ---   push compressed to DRS
				drsFileId := ingestionService.UploadVcfGzToDrs(cfg.Drs.BridgeDirectory, gzippedFileName, drsUrl, drsUsername, drsPassword)
				if drsFileId == "" {
					msg := "Something went wrong: DRS File Id is empty for " + gzippedFileName
					fmt.Println(msg)

					reqStat.State = ingest.Error
					reqStat.Message = msg
					ingestionService.IngestRequestChan <- reqStat

					return
				}

				// -- push tabix to DRS
				drsTabixFileId := ingestionService.UploadVcfGzToDrs(cfg.Drs.BridgeDirectory, tabixFileName, drsUrl, drsUsername, drsPassword)
				if drsTabixFileId == "" {
					msg := "Something went wrong: DRS Tabix File Id is empty for " + tabixFileName
					fmt.Println(msg)

					reqStat.State = ingest.Error
					reqStat.Message = msg
					ingestionService.IngestRequestChan <- reqStat

					return
				}

				// ---   remove temporary files now that they have been ingested successfully into DRS
				if tmpFileRemovalErr := os.Remove(tmpDestinationFileName); tmpFileRemovalErr != nil {
					msg := fmt.Sprintf("Something went wrong: trying to remove temporary file at %s : %s\n", tmpDestinationFileName, tmpFileRemovalErr)
					fmt.Println(msg)

					reqStat.State = ingest.Error
					reqStat.Message = msg
					ingestionService.IngestRequestChan <- reqStat

					return
				}
				tmpTabixFilePath := fmt.Sprintf("%s%s", tabixFileDir, tabixFileName)
				if tmpTabixFileRemovalErr := os.Remove(tmpTabixFilePath); tmpTabixFileRemovalErr != nil {
					msg := fmt.Sprintf("Something went wrong: trying to remove temporary file at %s : %s\n", tmpTabixFilePath, tmpTabixFileRemovalErr)
					fmt.Println(msg)

					reqStat.State = ingest.Error
					reqStat.Message = msg
					ingestionService.IngestRequestChan <- reqStat

					return
				}

				// ---   reopen gzipped file after having been copied to the temporary api-drs
				//       bridge directory, as the stream depletes and needs a refresh
				r, err = os.Open(gzippedFilePath)
				if err != nil {
					msg := fmt.Sprintf("error reopening %s: %s\n", gzippedFileName, err)
					fmt.Println(msg)

					reqStat.State = ingest.Error
					reqStat.Message = msg
					ingestionService.IngestRequestChan <- reqStat

					return
				}

				// ---   extract gzip compressed vcf file
				vcfFilePath := ingestionService.ExtractVcfGz(gzippedFilePath, r, vcfTmpPath)
				if vcfFilePath == "" {
					msg := "Something went wrong: filepath is empty for " + gzippedFileName
					fmt.Println(msg)

					reqStat.State = ingest.Error
					reqStat.Message = msg
					ingestionService.IngestRequestChan <- reqStat

					return
				}
				defer r.Close()

				// ---	 load vcf into memory and ingest the vcf file into elasticsearch
				ingestionService.ProcessVcf(vcfFilePath, drsFileId, assemblyId, filterOutHomozygousReferences, cfg.Api.LineProcessingConcurrencyLevel)

				// ---   delete the temporary vcf file
				os.Remove(vcfFilePath)

				// ---   delete full tmp path and all contents
				// 		 (WARNING : Only do this when running over a single file)
				//os.RemoveAll(vcfTmpPath)

				fmt.Printf("Ingest duration for file at %s : %s\n", vcfFilePath, time.Since(startTime))

				reqStat.State = ingest.Done
				ingestionService.IngestRequestChan <- reqStat
			}(_fileName, _newRequestState)
		}(fileName, newRequestState)

	}

	return c.JSON(http.StatusOK, responseDtos)
}

func GetVariantsOverview(c echo.Context) error {
	fmt.Printf("[%s] - GetVariantsOverview hit!\n", time.Now())

	resultsMap := map[string]interface{}{}
	resultsMux := sync.RWMutex{}

	var wg sync.WaitGroup
	es := c.(*contexts.GohanContext).Es7Client
	cfg := c.(*contexts.GohanContext).Config

	callGetBucketsByKeyword := func(key string, keyword string, _wg *sync.WaitGroup) {
		defer _wg.Done()

		results, bucketsError := esRepo.GetVariantsBucketsByKeyword(cfg, es, keyword)
		if bucketsError != nil {
			resultsMap[key] = map[string]interface{}{
				"error": "Something went wrong. Please contact the administrator!",
			}
			return
		}

		// retrieve aggregations.items.buckets
		bucketsMapped := []interface{}{}
		if aggs, ok := results["aggregations"]; ok {
			aggsMapped := aggs.(map[string]interface{})

			if items, ok := aggsMapped["items"]; ok {
				itemsMapped := items.(map[string]interface{})

				if buckets := itemsMapped["buckets"]; ok {
					bucketsMapped = buckets.([]interface{})
				}
			}
		}

		individualKeyMap := map[string]interface{}{}
		// push results bucket to slice
		for _, bucket := range bucketsMapped {
			doc_key := fmt.Sprint(bucket.(map[string]interface{})["key"]) // ensure strings and numbers are expressed as strings
			doc_count := bucket.(map[string]interface{})["doc_count"]

			individualKeyMap[doc_key] = doc_count
		}

		resultsMux.Lock()
		resultsMap[key] = individualKeyMap
		resultsMux.Unlock()
	}

	// get distribution of chromosomes
	wg.Add(1)
	go callGetBucketsByKeyword("chromosomes", "chrom.keyword", &wg)

	// get distribution of variant IDs
	wg.Add(1)
	go callGetBucketsByKeyword("variantIDs", "id.keyword", &wg)

	// get distribution of sample IDs
	wg.Add(1)
	go callGetBucketsByKeyword("sampleIDs", "sample.id.keyword", &wg)

	// get distribution of assembly IDs
	wg.Add(1)
	go callGetBucketsByKeyword("assemblyIDs", "assemblyId.keyword", &wg)

	wg.Wait()

	return c.JSON(http.StatusOK, resultsMap)
}

func GetAllVariantIngestionRequests(c echo.Context) error {
	fmt.Printf("[%s] - GetAllVariantIngestionRequests hit!\n", time.Now())
	izMap := c.(*contexts.GohanContext).IngestionService.IngestRequestMap

	// transform map of it-to-ingestRequests to an array
	m := make([]*ingest.VariantIngestRequest, 0, len(izMap))
	for _, val := range izMap {
		m = append(m, val)
	}
	return c.JSON(http.StatusOK, m)
}

func executeGetByIds(c echo.Context, ids []string, isVariantIdQuery bool) error {
	cfg := c.(*contexts.GohanContext).Config

	var es, chromosome, lowerBound, upperBound, reference, alternative, genotype, assemblyId = retrieveCommonElements(c)

	// retrieve other query parameters relevent to this 'get' query ---
	getSampleIdsOnlyQP := c.QueryParam("getSampleIdsOnly")
	var (
		getSampleIdsOnly bool = false
		getSioErr        error
	)
	// only respond sampleIds-only
	if isVariantIdQuery && len(getSampleIdsOnlyQP) > 0 {
		getSampleIdsOnly, getSioErr = strconv.ParseBool(getSampleIdsOnlyQP)
		if getSioErr != nil {
			log.Fatal(getSioErr)
		}
	}

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

	sortByPosition := s.CastToSortDirection(c.QueryParam("sortByPosition"))

	includeInfoInResultSetQP := c.QueryParam("includeInfoInResultSet")
	var (
		includeInfoInResultSet bool
		isirsErr               error
	)
	if len(includeInfoInResultSetQP) > 0 {
		includeInfoInResultSet, isirsErr = strconv.ParseBool(includeInfoInResultSetQP)
		if isirsErr != nil {
			log.Fatal(isirsErr)
		}
	}
	// ---

	// prepare response
	respDTO := models.VariantsResponseDTO{}
	respDTOMux := sync.RWMutex{}

	var errors []error
	errorMux := sync.RWMutex{}

	// TODO: optimize - make 1 repo call with all variantIds at once
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)

		go func(_id string) {
			defer wg.Done()

			variantRespDataModel := models.VariantResponseDataModel{}

			var (
				docs      map[string]interface{}
				searchErr error
			)
			if isVariantIdQuery {
				variantRespDataModel.VariantId = _id

				fmt.Printf("Executing Get-Variants for VariantId %s\n", _id)

				docs, searchErr = esRepo.GetDocumentsContainerVariantOrSampleIdInPositionRange(cfg, es,
					chromosome, lowerBound, upperBound,
					_id, "", // note : "" is for sampleId
					reference, alternative,
					size, sortByPosition,
					includeInfoInResultSet, genotype, assemblyId,
					getSampleIdsOnly)
			} else {
				// implied sampleId query
				variantRespDataModel.SampleId = _id

				fmt.Printf("Executing Get-Samples for SampleId %s\n", _id)

				docs, searchErr = esRepo.GetDocumentsContainerVariantOrSampleIdInPositionRange(cfg, es,
					chromosome, lowerBound, upperBound,
					"", _id, // note : "" is for variantId
					reference, alternative,
					size, sortByPosition,
					includeInfoInResultSet, genotype, assemblyId,
					false)
			}
			if searchErr != nil {
				errorMux.Lock()
				errors = append(errors, searchErr)
				errorMux.Unlock()
				return
			}

			// query for each id

			if !getSampleIdsOnly {
				docsHits := docs["hits"].(map[string]interface{})["hits"]
				allDocHits := []map[string]interface{}{}
				mapstructure.Decode(docsHits, &allDocHits)

				// grab _source for each hit
				var allSources []models.Variant

				for _, r := range allDocHits {
					source := r["_source"].(map[string]interface{})

					// cast map[string]interface{} to struct
					var resultingVariant models.Variant
					mapstructure.Decode(source, &resultingVariant)

					// accumulate structs
					allSources = append(allSources, resultingVariant)
				}

				fmt.Printf("Found %d docs!\n", len(allSources))

				variantRespDataModel.Count = len(allSources)
				variantRespDataModel.Results = allSources
			} else {
				// TODO: refactor this 'else' statement
				docsBuckets := docs["aggregations"].(map[string]interface{})["sampleIds"].(map[string]interface{})["buckets"]
				allDocBuckets := []map[string]interface{}{}
				mapstructure.Decode(docsBuckets, &allDocBuckets)

				var allSampleIdsOnly []string

				for _, r := range allDocBuckets {
					sampleId := r["key"].(string)

					// accumulate sample Id's
					allSampleIdsOnly = append(allSampleIdsOnly, sampleId)
				}

				fmt.Printf("Found %d docs!\n", len(allSampleIdsOnly))

				variantRespDataModel.Count = len(allSampleIdsOnly)
				variantRespDataModel.Results = allSampleIdsOnly
			}

			respDTOMux.Lock()
			respDTO.Data = append(respDTO.Data, variantRespDataModel)
			respDTOMux.Unlock()

		}(id)
	}

	wg.Wait()

	if len(errors) == 0 {
		respDTO.Status = 200
		respDTO.Message = "Success"
	} else {
		respDTO.Status = 500
		respDTO.Message = "Something went wrong.. Please contact the administrator!"
	}

	return c.JSON(http.StatusOK, respDTO)
}

func executeCountByIds(c echo.Context, ids []string, isVariantIdQuery bool) error {
	cfg := c.(*contexts.GohanContext).Config

	var es, chromosome, lowerBound, upperBound, reference, alternative, genotype, assemblyId = retrieveCommonElements(c)

	respDTO := models.VariantsResponseDTO{}
	respDTOMux := sync.RWMutex{}

	var errors []error
	errorMux := sync.RWMutex{}
	// TODO: optimize - make 1 repo call with all variantIds at once
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)

		go func(_id string) {
			defer wg.Done()

			variantRespDataModel := models.VariantResponseDataModel{}

			var (
				docs       map[string]interface{}
				countError error
			)
			if isVariantIdQuery {
				variantRespDataModel.VariantId = _id

				fmt.Printf("Executing Count-Variants for VariantId %s\n", _id)

				docs, countError = esRepo.CountDocumentsContainerVariantOrSampleIdInPositionRange(cfg, es,
					chromosome, lowerBound, upperBound,
					_id, "", // note : "" is for sampleId
					reference, alternative, genotype, assemblyId)
			} else {
				// implied sampleId query
				variantRespDataModel.SampleId = _id

				fmt.Printf("Executing Count-Samples for SampleId %s\n", _id)

				docs, countError = esRepo.CountDocumentsContainerVariantOrSampleIdInPositionRange(cfg, es,
					chromosome, lowerBound, upperBound,
					"", _id, // note : "" is for variantId
					reference, alternative, genotype, assemblyId)
			}

			if countError != nil {
				errorMux.Lock()
				errors = append(errors, countError)
				errorMux.Unlock()
				return
			}

			variantRespDataModel.Count = int(docs["count"].(float64))

			respDTOMux.Lock()
			respDTO.Data = append(respDTO.Data, variantRespDataModel)
			respDTOMux.Unlock()

		}(id)
	}

	wg.Wait()

	if len(errors) == 0 {
		respDTO.Status = 200
		respDTO.Message = "Success"
	} else {
		respDTO.Status = 500
		respDTO.Message = "Something went wrong.. Please contact the administrator!"
	}

	return c.JSON(http.StatusOK, respDTO)
}

func retrieveCommonElements(c echo.Context) (*elasticsearch.Client, string, int, int, string, string, constants.GenotypeQuery, constants.AssemblyId) {
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

	genotype := gq.UNCALLED
	genotypeQP := c.QueryParam("genotype")
	if len(genotypeQP) > 0 {
		if parsedGenotype, gErr := gq.CastToGenoType(genotypeQP); gErr == nil {
			genotype = parsedGenotype
		}
	}

	assemblyId := a.Unknown
	assemblyIdQP := c.QueryParam("assemblyId")
	if len(assemblyIdQP) > 0 && a.IsKnownAssemblyId(assemblyIdQP) {
		assemblyId = a.CastToAssemblyId(assemblyIdQP)
	}

	return es, chromosome, lowerBound, upperBound, reference, alternative, genotype, assemblyId
}
