package elasticsearch

import (
	"api/contexts"
	"api/models"
	"api/models/indexes"
	"api/models/schemas"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/esapi"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/google/uuid"
	"github.com/labstack/echo"
)

const tablesIndex = "tables"

func CreateTable(c echo.Context) { //(map[string]interface{}, error)

	es := c.(*contexts.GohanContext).Es7Client

	// Create struct instance of the Elasticsearch fields struct object
	docStruct := indexes.Table{
		Id:       uuid.New().String(),
		DataType: "variant",
		Name:     "Fake Table", // TODO: provide table name parameter
		AssemblyIds: []string{
			"GRCh38",
			"GRCh37",
			"NCBI36",
			"Other",
		},
		Metadata: map[string]interface{}{
			"created_at": time.Now(),
			"updated_at": time.Now(),
			"name":       "Fake Table", // TODO: provide table name parameter
		},
		Schema: schemas.VARIANT_SCHEMA,
	}

	fmt.Println("\ndocStruct:", docStruct)
	fmt.Println("docStruct TYPE:", reflect.TypeOf(docStruct))

	// Marshal the struct to JSON and check for errors
	b, err := json.Marshal(docStruct)
	if err != nil {
		fmt.Println("json.Marshal ERROR:", err)
		return // nil, err
	}

	// Instantiate a request object
	req := esapi.IndexRequest{
		Index:   tablesIndex,
		Body:    strings.NewReader(string(b)),
		Refresh: "true",
	}
	fmt.Println(reflect.TypeOf(req))

	// Return an API response object from request
	res, err := req.Do(c.Request().Context(), es)
	if err != nil {
		log.Fatalf("IndexRequest ERROR: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("%s ERROR", res.Status())
	} else {

		// Deserialize the response into a map.
		var resMap map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&resMap); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		} else {
			log.Printf("\nIndexRequest() RESPONSE:")
			// Print the response status and indexed document version.
			fmt.Println("Status:", res.Status())
			fmt.Println("Result:", resMap["result"])
			fmt.Println("Version:", int(resMap["_version"].(float64)))
			fmt.Println("resMap:", resMap)
			fmt.Println("\n")
		}
	}

	// TODO: return new table object
}

func GetTables(c echo.Context, tableId string, dataType string) (map[string]interface{}, error) {

	cfg := c.(*contexts.GohanContext).Config
	es := c.(*contexts.GohanContext).Es7Client

	// TEMP: SECURITY RISK
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	//

	// get table by "any combination of any applicable parameter" query structure
	must := make([]map[string]interface{}, 0)

	if tableId != "" {
		must = append(must, map[string]interface{}{
			"query_string": map[string]string{
				"query": fmt.Sprintf("id:%s", tableId),
			},
		})
	}
	if dataType != "" {
		must = append(must, map[string]interface{}{
			"query_string": map[string]string{
				"query": fmt.Sprintf("data_type:%s", dataType),
			},
		})
	}
	// if `must` remains an empty array, this will effecetively act as a "wildcard" query

	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{{
					"bool": map[string]interface{}{
						"must": must,
					},
				}},
			},
		},
	}

	// encode the query
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s\n", err)
		return nil, err
	}
	// Perform the search request.
	res, searchErr := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(tablesIndex),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if searchErr != nil {
		fmt.Printf("Error getting response: %s\n", searchErr)
		return nil, searchErr
	}

	defer res.Body.Close()

	resultString := res.String()
	if cfg.Debug {
		fmt.Println(resultString)
	}

	// Declared an empty interface
	result := make(map[string]interface{})

	// Unmarshal or Decode the JSON to the interface.
	// Known bug: response comes back with a preceding '[200 OK] ' which needs trimming (hence the [9:])
	umErr := json.Unmarshal([]byte(resultString[9:]), &result)
	if umErr != nil {
		fmt.Printf("Error unmarshalling response: %s\n", umErr)
		return nil, umErr
	}

	fmt.Printf("Query End: %s\n", time.Now())

	return result, nil
}

func DeleteTableById(cfg *models.Config, es *elasticsearch.Client) { //(map[string]interface{}, error)
	// TODO : implement
}
