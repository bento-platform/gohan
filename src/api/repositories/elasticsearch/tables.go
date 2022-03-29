package elasticsearch

import (
	"api/contexts"
	"api/models"
	"api/models/dtos"
	"api/models/indexes"
	"api/models/schemas"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
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
	"github.com/mitchellh/mapstructure"
)

const tablesIndex = "tables"

func CreateTable(c echo.Context, t dtos.CreateTableRequestDto) (indexes.Table, error) {

	es := c.(*contexts.GohanContext).Es7Client
	now := time.Now()

	// TODO: improve checks and balances..

	// merge inbound metadata if any
	defaultMeta := map[string]interface{}{
		"created_at": now,
		"updated_at": now,
		"name":       t.Name,
	}

	defaultAssemblyIds := []string{
		"GRCh38",
		"GRCh37",
		"NCBI36",
		"Other",
	}

	// Create struct instance of the Elasticsearch fields struct object
	docStruct := indexes.Table{
		Id:          uuid.New().String(),
		Name:        t.Name,
		DataType:    t.DataType,
		Dataset:     t.Dataset,
		AssemblyIds: defaultAssemblyIds,
		Metadata:    defaultMeta,
		Schema:      schemas.VARIANT_SCHEMA,
	}

	fmt.Println("\ndocStruct:", docStruct)
	fmt.Println("docStruct TYPE:", reflect.TypeOf(docStruct))

	// Marshal the struct to JSON and check for errors
	b, err := json.Marshal(docStruct)
	if err != nil {
		fmt.Println("json.Marshal ERROR:", err)
		return docStruct, err
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
		fmt.Printf("IndexRequest ERROR: %s\n", err)
		return docStruct, err
	}
	defer res.Body.Close()

	if res.IsError() {
		msg := fmt.Sprintf("%s ERROR", res.Status())
		fmt.Println(msg)
		return docStruct, errors.New(msg)
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
			fmt.Println()
		}
	}

	return docStruct, nil
}

func GetTables(c echo.Context, tableId string, dataType string) (map[string]interface{}, error) {

	cfg := c.(*contexts.GohanContext).Config
	es := c.(*contexts.GohanContext).Es7Client

	// TEMP: SECURITY RISK
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	//

	// get table by "any combination of any applicable parameter" query structure
	filter := make([]map[string]interface{}, 0)

	if tableId != "" {

		filter = append(filter, map[string]interface{}{
			"term": map[string]string{
				"id.keyword": tableId,
			},
		})
	}
	if dataType != "" {
		filter = append(filter, map[string]interface{}{
			"term": map[string]string{
				"data_type.keyword": dataType,
			},
		})
	}
	// if `must` remains an empty array, this will effecetively act as a "wildcard" query

	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": filter,
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

func GetTablesByName(c echo.Context, tableName string) ([]indexes.Table, error) {

	cfg := c.(*contexts.GohanContext).Config
	es := c.(*contexts.GohanContext).Es7Client

	allTables := make([]indexes.Table, 0)

	// overall query structure
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{{
					"term": map[string]interface{}{
						"name.keyword": tableName,
					},
				}},
			},
		},
	}

	// encode the query
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s\n", err)
		return allTables, err
	}

	if cfg.Debug {
		// view the outbound elasticsearch query
		myString := string(buf.Bytes()[:])
		fmt.Println(myString)
	}

	fmt.Printf("Query Start: %s\n", time.Now())

	// TEMP: SECURITY RISK
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	//
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
		return allTables, searchErr
	}

	defer res.Body.Close()

	resultString := res.String()
	if cfg.Debug {
		fmt.Println(resultString)
	}

	// TODO: improve stability
	// - check for 404 Not Found : assume index simply doesnt exist, return 0 results
	if strings.Contains(resultString[0:15], "Not Found") {
		return allTables, nil
	}

	// Declared an empty interface
	result := make(map[string]interface{})

	// Unmarshal or Decode the JSON to the interface.
	// Known bug: response comes back with a preceding '[200 OK] ' which needs trimming (hence the [9:])
	umErr := json.Unmarshal([]byte(resultString[9:]), &result)
	if umErr != nil {
		fmt.Printf("Error unmarshalling response: %s\n", umErr)
		return allTables, umErr
	}

	fmt.Printf("Query End: %s\n", time.Now())

	// gather data from "hits"
	docsHits := result["hits"].(map[string]interface{})["hits"]
	allDocHits := []map[string]interface{}{}
	mapstructure.Decode(docsHits, &allDocHits)

	// grab _source for each hit

	for _, r := range allDocHits {
		source := r["_source"]
		byteSlice, _ := json.Marshal(source)

		// cast map[string]interface{} a table
		var resultingTable indexes.Table
		if err := json.Unmarshal(byteSlice, &resultingTable); err != nil {
			fmt.Println("failed to unmarshal:", err)
		}

		// accumulate structs
		allTables = append(allTables, resultingTable)
	}

	return allTables, nil
}

func DeleteTableById(cfg *models.Config, es *elasticsearch.Client) { //(map[string]interface{}, error)
	// TODO : implement
}
