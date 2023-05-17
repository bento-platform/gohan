package variants

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"regexp"
	"time"

	"gohan/api/contexts"
	a "gohan/api/models/constants/assembly-id"
	s "gohan/api/models/constants/sort"
	"gohan/api/models/dtos"
	"gohan/api/models/indexes"
	"gohan/api/models/ingest"
	"gohan/api/mvc"
	esRepo "gohan/api/repositories/elasticsearch"
	variantService "gohan/api/services/variants"
	"gohan/api/utils"

	"gohan/api/models/constants/zygosity"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"

	"github.com/labstack/echo"
)

func VariantsIngestionStats(c echo.Context) error {
	fmt.Printf("[%s] - VariantsIngestionStats hit!\n", time.Now())
	ingestionService := c.(*contexts.GohanContext).IngestionService

	return c.JSON(http.StatusOK, ingestionService.IngestionBulkIndexer.Stats())
}

func VariantsGetByVariantId(c echo.Context) error {
	fmt.Printf("[%s] - VariantsGetByVariantId hit!\n", time.Now())
	// retrieve variant Ids from query parameter (comma separated)
	variantIds := strings.Split(c.QueryParam("ids"), ",")
	if len(variantIds[0]) == 0 {
		// if no ids were provided, assume "wildcard" search
		variantIds = []string{"*"}
	}

	return executeGetByIds(c, variantIds, true, false)
}
func VariantsGetBySampleId(c echo.Context) error {
	fmt.Printf("[%s] - VariantsGetBySampleId hit!\n", time.Now())
	// retrieve sample Ids from query parameter (comma separated)
	sampleIds := strings.Split(c.QueryParam("ids"), ",")
	if len(sampleIds[0]) == 0 {
		// if no ids were provided, assume "wildcard" search
		sampleIds = []string{"*"}
	}

	return executeGetByIds(c, sampleIds, false, false)
}
func VariantsGetByDocumentId(c echo.Context) error {
	fmt.Printf("[%s] - VariantsGetByDocumentId hit!\n", time.Now())
	// retrieve document Ids from query parameter (comma separated)
	docIds := strings.Split(c.QueryParam("ids"), ",")
	if len(docIds[0]) == 0 {
		// if no ids were provided, assume "wildcard" search
		docIds = []string{"*"}
	}

	return executeGetByIds(c, docIds, false, true)
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
	var fileNames []string
	// get vcf files
	var vcfGzfiles []string

	dirName := c.QueryParam("directory")
	if dirName != "" {
		if strings.HasPrefix(dirName, cfg.Drs.BridgeDirectory) {
			replaced := strings.Replace(dirName, cfg.Drs.BridgeDirectory, "", 1)

			replacedFullPath, replacedDirName := path.Split(replaced)
			// strip the leading '/' away
			if replacedFullPath == "/" {
				dirName = replacedDirName
			} else {
				dirName = replaced
			}
		}

		err := filepath.Walk(fmt.Sprintf("%s/%s", vcfPath, dirName),
			func(absoluteFileName string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if absoluteFileName == vcfPath {
					// skip
					return nil
				}

				// keep track of relative path
				relativePathFileName := strings.ReplaceAll(absoluteFileName, vcfPath, "")

				// verify if there is a relative path
				directoryPath, fileName := path.Split(relativePathFileName)
				if directoryPath == "/" {
					relativePathFileName = fileName // effectively strips the leading '/' away
				}

				// Filter only .vcf.gz files
				// if fileName != "" {
				if matched, _ := regexp.MatchString(".vcf.gz", relativePathFileName); matched {
					fileNames = append(fileNames, relativePathFileName)
				} else {
					fmt.Printf("Skipping %s\n", relativePathFileName)
				}
				// }
				return nil
			})
		if err != nil {
			log.Println(err)
		}
	} else {
		fileNames = strings.Split(c.QueryParam("fileNames"), ",")
		for i, fileName := range fileNames {
			if fileName == "" {
				// TODO: create a standard response object
				return c.JSON(http.StatusBadRequest, "{\"error\" : \"Missing 'fileNames' query parameter!\"}")
			} else {
				// remove DRS bridge directory base path from the requested filenames (if present)
				if strings.HasPrefix(fileName, cfg.Drs.BridgeDirectory) {
					replaced := strings.Replace(fileName, cfg.Drs.BridgeDirectory, "", 1)

					replacedDirectory, replacedFileName := path.Split(replaced)
					// strip the leading '/' away
					if replacedDirectory == "/" {
						fileNames[i] = replacedFileName
					} else {
						fileNames[i] = replaced
					}
				}
			}
		}

		// TODO: simply load files by filename provided
		// rather than load all available files and looping over them
		// -----
		// Read all files and temporarily catalog all .vcf.gz files
		err := filepath.Walk(vcfPath,
			func(absoluteFileName string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if absoluteFileName == vcfPath {
					// skip
					return nil
				}

				// keep track of relative path
				relativePathFileName := strings.ReplaceAll(absoluteFileName, vcfPath, "")

				// verify if there is a relative path
				directoryPath, fileName := path.Split(relativePathFileName)
				if directoryPath == "/" {
					relativePathFileName = fileName // effectively strips the leading '/' away
				}

				// Filter only .vcf.gz files
				// if fileName != "" {
				if matched, _ := regexp.MatchString(".vcf.gz", relativePathFileName); matched {
					vcfGzfiles = append(vcfGzfiles, relativePathFileName)
				} else {
					fmt.Printf("Skipping %s\n", relativePathFileName)
				}
				// }
				return nil
			})
		if err != nil {
			log.Println(err)
		}

		// Locate fileName from request inside found files
		for _, fileName := range fileNames {
			if !utils.StringInSlice(fileName, vcfGzfiles) {
				return c.JSON(http.StatusBadRequest, "{\"error\" : \"file "+fileName+" not found! Aborted -- \"}")
			}
		}
		// -----
	}

	assemblyId := a.CastToAssemblyId(c.QueryParam("assemblyId"))
	tableId := c.QueryParam("tableId")
	// TODO: validate table exists in elasticsearch

	// -- optional filter
	var (
		filterOutReferences bool = false // default
		fohrErr             error
	)
	filterOutReferencesQP := c.QueryParam("filterOutReferences")
	if len(filterOutReferencesQP) > 0 {
		filterOutReferences, fohrErr = strconv.ParseBool(filterOutReferencesQP)
		if fohrErr != nil {
			fmt.Printf("Error parsing filterOutReferences: %s, [%s] - defaulting to 'false'\n", filterOutReferencesQP, fohrErr)
			// defaults to false
		}
	}

	startTime := time.Now()
	fmt.Printf("Ingest Start: %s\n", startTime)

	// ingest vcf
	// ingserviceMux := sync.RWMutex{}
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
				defer func() {
					<-ingestionService.ConcurrentFileIngestionQueue
				}()

				fmt.Printf("Begin running %s !\n", gzippedFileName)
				reqStat.State = ingest.Running
				ingestionService.IngestRequestChan <- reqStat

				// ---	 open vcf.gz

				fmt.Printf("Opening %s !\n", gzippedFileName)
				var separator string
				if strings.HasPrefix(gzippedFileName, "/") {
					separator = ""
				} else {
					separator = "/"
				}

				gzippedFilePath := fmt.Sprintf("%s%s%s", vcfPath, separator, gzippedFileName)
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
				tmpDestinationFileName := fmt.Sprintf("%s%s%s", cfg.Api.BridgeDirectory, separator, gzippedFileName)

				// prepare directory inside bridge directory
				partialTmpDir, _ := path.Split(gzippedFileName)
				fullTmpDir, _ := path.Split(tmpDestinationFileName)
				if partialTmpDir != "" {
					if _, err := os.Stat(fullTmpDir); os.IsNotExist(err) {
						os.MkdirAll(fullTmpDir, 0700) // Create your file
					}
				}

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
				fmt.Printf("Generating Tabix %s !\n", tmpDestinationFileName)
				tabixFileDir, tabixFileName, tabixErr := ingestionService.GenerateTabix(tmpDestinationFileName)
				if tabixErr != nil {
					msg := "Something went wrong: Tabix problem " + gzippedFileName
					fmt.Println(msg)

					reqStat.State = ingest.Error
					reqStat.Message = msg
					ingestionService.IngestRequestChan <- reqStat

					return
				}
				tabixFileNameWithRelativePath := fmt.Sprintf("%s%s", partialTmpDir, tabixFileName)

				// ---   push compressed to DRS
				fmt.Printf("Uploading %s to DRS !\n", gzippedFileName)
				drsFileId := ingestionService.UploadVcfGzToDrs(cfg, cfg.Drs.BridgeDirectory, gzippedFileName, drsUrl, drsUsername, drsPassword)
				if drsFileId == "" {
					msg := "Something went wrong: DRS File Id is empty for " + gzippedFileName
					fmt.Println(msg)

					reqStat.State = ingest.Error
					reqStat.Message = msg
					ingestionService.IngestRequestChan <- reqStat

					return
				}

				// -- push tabix to DRS
				fmt.Printf("Uploading %s to DRS !\n", tabixFileNameWithRelativePath)
				drsTabixFileId := ingestionService.UploadVcfGzToDrs(cfg, cfg.Drs.BridgeDirectory, tabixFileNameWithRelativePath, drsUrl, drsUsername, drsPassword)
				if drsTabixFileId == "" {
					msg := "Something went wrong: DRS Tabix File Id is empty for " + tabixFileNameWithRelativePath
					fmt.Println(msg)

					reqStat.State = ingest.Error
					reqStat.Message = msg
					ingestionService.IngestRequestChan <- reqStat

					return
				}

				// ---   remove temporary files now that they have been ingested successfully into DRS
				fmt.Printf("Removing %s !\n", tmpDestinationFileName)
				if tmpFileRemovalErr := os.Remove(tmpDestinationFileName); tmpFileRemovalErr != nil {
					msg := fmt.Sprintf("Something went wrong: trying to remove temporary file at %s : %s\n", tmpDestinationFileName, tmpFileRemovalErr)
					fmt.Println(msg)

					reqStat.State = ingest.Error
					reqStat.Message = msg
					ingestionService.IngestRequestChan <- reqStat

					return
				}
				tmpTabixFilePath := fmt.Sprintf("%s%s", tabixFileDir, tabixFileName)
				fmt.Printf("Removing %s !\n", tmpTabixFilePath)
				if tmpTabixFileRemovalErr := os.Remove(tmpTabixFilePath); tmpTabixFileRemovalErr != nil {
					msg := fmt.Sprintf("Something went wrong: trying to remove temporary file at %s : %s\n", tmpTabixFilePath, tmpTabixFileRemovalErr)
					fmt.Println(msg)

					reqStat.State = ingest.Error
					reqStat.Message = msg
					ingestionService.IngestRequestChan <- reqStat

					return
				}

				defer r.Close()

				// ---	 load vcf into memory and ingest the vcf file into elasticsearch
				beginProcessingTime := time.Now()
				fmt.Printf("Begin processing %s at [%s]\n", gzippedFilePath, beginProcessingTime)
				ingestionService.ProcessVcf(gzippedFilePath, drsFileId, tableId, assemblyId, filterOutReferences, cfg.Api.LineProcessingConcurrencyLevel)
				fmt.Printf("Ingest duration for file at %s : %s\n", gzippedFilePath, time.Since(beginProcessingTime))

				reqStat.State = ingest.Done
				ingestionService.IngestRequestChan <- reqStat
			}(_fileName, _newRequestState)
		}(fileName, newRequestState)

	}

	return c.JSON(http.StatusOK, responseDtos)
}

func GetVariantsOverview(c echo.Context) error {
	fmt.Printf("[%s] - GetVariantsOverview hit!\n", time.Now())

	es := c.(*contexts.GohanContext).Es7Client
	cfg := c.(*contexts.GohanContext).Config

	// TODO: refactor to handle errors better
	resultsMap := variantService.GetVariantsOverview(es, cfg)

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

func executeGetByIds(c echo.Context, ids []string, isVariantIdQuery bool, isDocumentIdQuery bool) error {
	cfg := c.(*contexts.GohanContext).Config

	var es, chromosome, lowerBound, upperBound, reference, alternative, alleles, genotype, assemblyId, tableId = mvc.RetrieveCommonElements(c)

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
	respDTO := dtos.VariantGetReponse{
		Results: make([]dtos.VariantGetResult, 0),
	}
	respDTOMux := sync.RWMutex{}

	var errors []error
	errorMux := sync.RWMutex{}

	// TODO: optimize - make 1 repo call with all variantIds at once
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)

		go func(_id string) {
			defer wg.Done()

			variantResult := dtos.VariantGetResult{
				Calls: make([]dtos.VariantCall, 0),
			}

			var (
				docs      map[string]interface{}
				searchErr error
			)
			if isVariantIdQuery {
				fmt.Printf("Executing Get-Variants for VariantId %s\n", _id)

				// only set query string if
				// 'getSampleIdsOnly' is false
				// (current support for bentoV2 + bento_federation_service integration)
				if !getSampleIdsOnly {
					variantResult.Query = fmt.Sprintf("variantId:%s", _id) // TODO: Refactor
				}

				docs, searchErr = esRepo.GetDocumentsContainerVariantOrSampleIdInPositionRange(cfg, es,
					chromosome, lowerBound, upperBound,
					_id, "", // note : "" is for sampleId
					reference, alternative, alleles,
					size, sortByPosition,
					includeInfoInResultSet, genotype, assemblyId, tableId,
					getSampleIdsOnly)
			} else {

				if isDocumentIdQuery {
					variantResult.Query = fmt.Sprintf("documentId:%s", _id) // TODO: Refactor

					fmt.Printf("Executing Get-Samples for DocumentId %s\n", _id)
					docs, searchErr = esRepo.GetDocumentsByDocumentId(cfg, es, _id)
				} else {
					// implied sampleId query
					fmt.Printf("Executing Get-Samples for SampleId %s\n", _id)

					// only set query string if
					// 'getSampleIdsOnly' is false
					// (current support for bentoV2 + bento_federation_service integration)
					if !getSampleIdsOnly {
						variantResult.Query = fmt.Sprintf("variantId:%s", _id) // TODO: Refactor
					}

					docs, searchErr = esRepo.GetDocumentsContainerVariantOrSampleIdInPositionRange(cfg, es,
						chromosome, lowerBound, upperBound,
						"", _id, // note : "" is for variantId
						reference, alternative, alleles,
						size, sortByPosition,
						includeInfoInResultSet, genotype, assemblyId, tableId,
						false)
				}

			}

			if searchErr != nil {
				errorMux.Lock()
				errors = append(errors, searchErr)
				errorMux.Unlock()
				return
			}

			// -- map variant index models to appropriate variant result + call dto models
			variantResult.AssemblyId = assemblyId
			variantResult.Chromosome = chromosome
			variantResult.Start = lowerBound
			variantResult.End = upperBound

			if getSampleIdsOnly {
				// gather data from "aggregations"
				docsBuckets := docs["aggregations"].(map[string]interface{})["sampleIds"].(map[string]interface{})["buckets"]
				allDocBuckets := []map[string]interface{}{}
				mapstructure.Decode(docsBuckets, &allDocBuckets)

				for _, r := range allDocBuckets {
					sampleId := r["key"].(string)

					// TEMP : re-capitalize sampleIds retrieved from elasticsearch at response time
					// TODO: touch up elasticsearch ingestion/parsing settings
					// to not automatically force all sampleIds to lowercase when indexing
					sampleId = strings.ToUpper(sampleId)

					// accumulate sample Id's
					variantResult.Calls = append(variantResult.Calls, dtos.VariantCall{
						SampleId:     sampleId,
						GenotypeType: string(genotype),
					})
				}
			} else {
				// gather data from "hits"
				docsHits := docs["hits"].(map[string]interface{})["hits"]
				allDocHits := []map[string]interface{}{}
				mapstructure.Decode(docsHits, &allDocHits)

				// grab _source for each hit
				var allSources []interface{}
				// var allSources []indexes.Variant

				for _, r := range allDocHits {
					source := r["_source"].(map[string]interface{})
					docId := r["_id"].(string)

					// cast map[string]interface{} to struct
					var resultingVariant indexes.Variant
					mapstructure.Decode(source, &resultingVariant)

					// accumulate structs
					allSources = append(allSources, map[string]interface{}{
						"variant":    resultingVariant,
						"documentId": docId,
					})
				}

				fmt.Printf("Found %d docs!\n", len(allSources))

				for _, source := range allSources {
					// TEMP : re-capitalize sampleIds retrieved from elasticsearch at response time
					// TODO: touch up elasticsearch ingestion/parsing settings
					// to not automatically force all sampleIds to lowercase when indexing
					variant := source.(map[string]interface{})["variant"].(indexes.Variant)
					docId := source.(map[string]interface{})["documentId"].(string)

					alleles := variant.Sample.Variation.Alleles

					sampleId := strings.ToUpper(variant.Sample.Id)

					variantResult.Calls = append(variantResult.Calls, dtos.VariantCall{
						Chrom:  variant.Chrom,
						Pos:    variant.Pos,
						Id:     variant.Id,
						Ref:    variant.Ref,
						Alt:    variant.Alt,
						Format: variant.Format,
						Qual:   variant.Qual,
						Filter: variant.Filter,

						Info: variant.Info,

						SampleId:     sampleId,
						GenotypeType: zygosity.ZygosityToString(variant.Sample.Variation.Genotype.Zygosity),
						Alleles:      []string{alleles.Left, alleles.Right},

						AssemblyId: variant.AssemblyId,
						DocumentId: docId,
					})
				}
			}
			// --

			respDTOMux.Lock()
			respDTO.Results = append(respDTO.Results, variantResult)
			respDTOMux.Unlock()

		}(id)
	}

	wg.Wait()

	if len(errors) == 0 {
		// only set status and message if
		// 'getSampleIdsOnly' is false
		// (current support for bentoV2 + bento_federation_service integration)
		if !getSampleIdsOnly {
			respDTO.Status = 200
			respDTO.Message = "Success"
		}
	} else {
		respDTO.Status = 500
		respDTO.Message = "Something went wrong.. Please contact the administrator!"
	}

	return c.JSON(http.StatusOK, respDTO)
}

func executeCountByIds(c echo.Context, ids []string, isVariantIdQuery bool) error {
	cfg := c.(*contexts.GohanContext).Config

	var es, chromosome, lowerBound, upperBound, reference, alternative, alleles, genotype, assemblyId, tableId = mvc.RetrieveCommonElements(c)

	respDTO := dtos.VariantCountReponse{
		Results: make([]dtos.VariantCountResult, 0),
	}
	respDTOMux := sync.RWMutex{}

	var errors []error
	errorMux := sync.RWMutex{}
	// TODO: optimize - make 1 repo call with all variantIds at once
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)

		go func(_id string) {
			defer wg.Done()

			countResult := dtos.VariantCountResult{}

			var (
				docs       map[string]interface{}
				countError error
			)
			if isVariantIdQuery {
				fmt.Printf("Executing Count-Variants for VariantId %s\n", _id)
				countResult.Query = fmt.Sprintf("variantId:%s", _id) // TODO: Refactor

				docs, countError = esRepo.CountDocumentsContainerVariantOrSampleIdInPositionRange(cfg, es,
					chromosome, lowerBound, upperBound,
					_id, "", // note : "" is for sampleId
					reference, alternative, alleles, genotype, assemblyId, tableId)
			} else {
				// implied sampleId query
				fmt.Printf("Executing Count-Samples for SampleId %s\n", _id)
				countResult.Query = fmt.Sprintf("sampleId:%s", _id) // TODO: Refactor

				docs, countError = esRepo.CountDocumentsContainerVariantOrSampleIdInPositionRange(cfg, es,
					chromosome, lowerBound, upperBound,
					"", _id, // note : "" is for variantId
					reference, alternative, alleles, genotype, assemblyId, tableId)
			}

			if countError != nil {
				errorMux.Lock()
				errors = append(errors, countError)
				errorMux.Unlock()
				return
			}
			countResult.AssemblyId = assemblyId
			countResult.Chromosome = chromosome
			countResult.Start = lowerBound
			countResult.End = upperBound

			countResult.Count = int(docs["count"].(float64))

			respDTOMux.Lock()
			respDTO.Results = append(respDTO.Results, countResult)
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
