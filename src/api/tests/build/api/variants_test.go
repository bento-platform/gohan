package api

import (
	"fmt"
	ingest "gohan/api/models/ingest"
	common "gohan/api/tests/common"
	"gohan/api/utils"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	VariantsOverviewPath                      string = "%s/variants/overview"
	VariantsGetBySampleIdsPathWithQueryString string = "%s/variants/get/by/sampleId%s"
	IngestionRequestsPath                     string = "%s/variants/ingestion/requests"
)

func TestIngest(t *testing.T) {
	cfg := common.InitConfig()
	tableId := uuid.NewString()

	assert.True(t, t.Run("Ingest Demo VCF", func(t *testing.T) {
		// create demo vcf string
		sampleId := "abc1234"
		vcfString := `##fileformat=VCFv4.2
#CHROM	POS	ID	REF	ALT	QUAL	FILTER	INFO	FORMAT	S-1178-HAP
1	13656	.	CAG	C,<NON_REF>	868.60	.	BaseQRankSum=-5.505;DP=81;ExcessHet=3.0103;MLEAC=1,0;MLEAF=0.500,0.00;MQRankSum=-2.985;RAW_MQandDP=43993,81;ReadPosRankSum=-0.136	GT:AD:DP:GQ:PL:SB	0:50,25,0:75:99:876,0,2024,1026,2099,3126:4,46,5,20
10	28872481	.	CAAAA	C,CA,CAAA,CAAAAA,CAAAAAA,<NON_REF>	652.60	.	BaseQRankSum=0.029;DP=83;ExcessHet=3.0103;MLEAC=0,0,0,1,0,0;MLEAF=0.00,0.00,0.00,0.500,0.00,0.00;MQRankSum=-0.186;RAW_MQandDP=291409,83;ReadPosRankSum=-0.582	GT:AD:DP:GQ:PL:SB	0:19,3,2,5,29,9,0:67:99:660,739,2827,748,2714,2732,724,1672,1682,1587,0,340,338,249,265,321,956,929,699,245,898,866,1996,1991,1652,466,1006,1944:0,19,0,48
19	3619025	.	C	<NON_REF>	.	.	END=3619025	GT:DP:GQ:MIN_DP:PL	0:19:21:19:0,21,660
19	3619026	.	T	<NON_REF>	.	.	END=3619026	GT:DP:GQ:MIN_DP:PL	0:19:51:19:0,51,765`

		// - save string to vcf directory
		localDataRootPath := getRootGohanPath() + "/data"
		localVcfPath := localDataRootPath + "/vcfs"

		newFilePath := fmt.Sprintf("%s/%s.vcf", localVcfPath, sampleId)

		// - create file if not exists
		var (
			file *os.File
			err  error
		)

		file, err = os.Create(newFilePath)
		if isError(err) {
			return
		}
		defer file.Close()

		// - reopen file using READ & WRITE permission.
		file, err = os.OpenFile(newFilePath, os.O_RDWR, 0644)
		if isError(err) {
			return
		}
		defer file.Close()

		// - write some vcf string to file.
		_, err = file.WriteString(vcfString)
		if isError(err) {
			return
		}
		defer func() { os.Remove(newFilePath) }()

		// compress the vcf file with bgzip
		out, err := exec.Command("bgzip", newFilePath).Output()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println(string(out))

		newGzFile := newFilePath + ".gz"
		defer func() { os.Remove(newGzFile) }()

		// - ingest
		assemblyId := "GRCh38"
		containerizedVcfFilePath := "/data/" + filepath.Base(newGzFile)

		queryString := fmt.Sprintf("assemblyId=%s&fileNames=%s&tableId=%s", assemblyId, containerizedVcfFilePath, tableId)
		ingestUrl := fmt.Sprintf("%s/variants/ingestion/run?%s", cfg.Api.Url, queryString)

		initialIngestionDtos := utils.GetRequestReturnStuff[[]ingest.IngestResponseDTO](ingestUrl)
		assert.True(t, len(initialIngestionDtos) > 0)

		// check ingestion request
		for {
			fmt.Println("Checking state of the ingestion..")

			// make the call
			ingReqsUrl := fmt.Sprintf("%s/variants/ingestion/requests", cfg.Api.Url)
			ingReqDtos := utils.GetRequestReturnStuff[[]ingest.IngestResponseDTO](ingReqsUrl)
			assert.True(t, len(initialIngestionDtos) > 0)

			foundDone := false
			for _, dto := range ingReqDtos {
				if dto.Filename == filepath.Base(containerizedVcfFilePath) && dto.State == "Done" {
					foundDone = true
					break
				}
				if dto.Filename == filepath.Base(containerizedVcfFilePath) && dto.State == "Error" {
					log.Fatal(dto.Message)
				}
			}
			if foundDone {
				fmt.Println("Done, moving on..")
				break
			} else {
				// pause
				time.Sleep(3 * time.Second)
			}
		}

		// check ingestion stats
		for {
			fmt.Println("Checking ingestion stats..")
			// pause
			time.Sleep(3 * time.Second)

			// make the call
			statsReqUrl := fmt.Sprintf("%s/variants/ingestion/stats", cfg.Api.Url)
			stats := utils.GetRequestReturnStuff[ingest.IngestStatsDto](statsReqUrl)
			assert.NotNil(t, stats)

			fmt.Println(stats.NumAdded)
			fmt.Println(stats.NumFlushed)
			if stats.NumAdded == stats.NumFlushed {
				fmt.Println("Done, moving on..")
				break
			}
			if stats.NumFailed > 0 {
				log.Fatal("More than one variant failed to flush")
			}

			// pause
			time.Sleep(3 * time.Second)
		}
	}))

	// verify demo vcf was properly ingested
	// by pinging it with specific queries
	assert.True(t, t.Run("Check Demo VCF Ingestion", func(t *testing.T) {
		overviewJson := common.GetVariantsOverview(t, cfg)
		assert.NotNil(t, overviewJson)

		dtos := common.BuildQueryAndMakeGetVariantsCall("1", "*", true, "asc", "", "GRCh38", "", "", "", false, t, cfg)
		assert.True(t, len(dtos.Results[0].Calls) > 0)
	}))
}

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return (err != nil)
}

func getRootGohanPath() string {
	// check if file exists
	wd, err1 := os.Getwd()
	if err1 != nil {
		log.Println(err1)
	}
	fmt.Println(wd) // for example /home/user

	path := filepath.Dir(wd)
	for i := 1; i < 5; i++ {
		path = filepath.Dir(path)
	}

	return path
}
