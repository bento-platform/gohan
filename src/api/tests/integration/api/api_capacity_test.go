package api

// import (
// 	"gohan/api/models/dtos"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"regexp"
// 	"sync"
// 	"testing"
// 	common "gohan/api/tests/common"

// 	"github.com/stretchr/testify/assert"
// )

// func TestCanHandleIngestAllLocalVcfs(_t *testing.T) {
// 	cfg := common.InitConfig()

// 	var vcfGzfiles []string

// 	fileInfo, err := ioutil.ReadDir(cfg.Api.LocalVcfPath)
// 	if err != nil {
// 		fmt.Printf("Failed: %s\n", err)
// 	}

// 	// Filter only .vcf.gz files
// 	for _, file := range fileInfo {
// 		if matched, _ := regexp.MatchString(".vcf.gz", file.Name()); matched {
// 			vcfGzfiles = append(vcfGzfiles, file.Name())
// 		} else {
// 			fmt.Printf("Skipping %s\n", file.Name())
// 		}
// 	}

// 	var combWg sync.WaitGroup
// 	simultaneousRequestsCapacity := 1 // TODO: tweak
// 	simultaneousRequestsQueue := make(chan bool, simultaneousRequestsCapacity)

// 	// fire and forget
// 	for _, vcfgz := range vcfGzfiles {
// 		simultaneousRequestsQueue <- true
// 		combWg.Add(1)
// 		go func(_vcfgz string, _combWg *sync.WaitGroup) {
// 			defer _combWg.Done()
// 			defer func() { <-simultaneousRequestsQueue }()

// 			url := cfg.Api.Url + fmt.Sprintf("/variants/ingestion/run?fileNames=%s&assemblyId=GRCH37&filterOutHomozygousReferences=true", _vcfgz)

// 			fmt.Printf("Calling %s\n", url)
// 			request, _ := http.NewRequest("GET", url, nil)

// 			client := &http.Client{}
// 			response, responseErr := client.Do(request)
// 			assert.Nil(_t, responseErr)

// 			defer response.Body.Close()

// 			// this test (at the time of writing) will only work if authorization is disabled
// 			shouldBe := 200
// 			assert.Equal(_t, shouldBe, response.StatusCode, fmt.Sprintf("Error -- Api GET %s Status: %s ; Should be %d", url, response.Status, shouldBe))

// 			//	-- interpret array of ingestion requests from response
// 			respBody, respBodyErr := ioutil.ReadAll(response.Body)
// 			assert.Nil(_t, respBodyErr)

// 			//	--- transform body bytes to string
// 			respBodyString := string(respBody)

// 			//	-- convert to json and check for error
// 			var respDto dtos.VariantReponse
// 			jsonUnmarshallingError := json.Unmarshal([]byte(respBodyString), &respDto)
// 			assert.Nil(_t, jsonUnmarshallingError)

// 		}(vcfgz, &combWg)
// 	}

// 	// allowing all lines to be queued up and waited for
// 	for i := 0; i < simultaneousRequestsCapacity; i++ {
// 		simultaneousRequestsQueue <- true
// 	}

// 	combWg.Wait()
// }
