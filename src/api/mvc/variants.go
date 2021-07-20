package mvc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"time"

	"api/contexts"
	"api/utils"

	"github.com/labstack/echo"
)

func VariantsSearchTest(c echo.Context) error {
	// Testing ES
	es := c.(*contexts.EsContext).Es

	fmt.Println("Query Start: %s", time.Now())

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
		fmt.Println("Error encoding query: %s", buffErr)
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
		fmt.Println("Error getting response: %s", searchErr)
	}

	respBuf := new(strings.Builder)
	_, respErr := io.Copy(respBuf, res.Body)
	if respErr != nil {
		fmt.Println("Error forming response: %s", respErr)
	}

	// check errors
	//fmt.Println(respBuf.String())

	// Declared an empty interface
	var result map[string]interface{}

	// Unmarshal or Decode the JSON to the interface.
	json.Unmarshal([]byte(respBuf.String()), &result)

	// Close the response
	defer res.Body.Close()

	fmt.Println("Query End: %s", time.Now())

	return c.JSON(http.StatusOK, result)
}
