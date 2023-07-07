package api

import (
	"encoding/json"
	"gohan/api/contexts"
	serviceInfo "gohan/api/models/constants/service-info"
	serviceInfoMvc "gohan/api/mvc/service-info"
	"gohan/api/tests/common"
	"io"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestGetVariantsOverview(t *testing.T) {
	cfg := common.InitConfig()

	setUpEcho := func(method string, path string) (*contexts.GohanContext, *httptest.ResponseRecorder) {
		e := echo.New()
		req := httptest.NewRequest(method, path, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		gc := &contexts.GohanContext{
			Context:          c,
			Es7Client:        nil, // todo mockup
			Config:           cfg,
			IngestionService: nil,
			VariantService:   nil,
		}
		return gc, rec
	}

	getJsonBody := func(rec *httptest.ResponseRecorder) map[string]interface{} {
		// - extract body bytes from response
		body, _ := io.ReadAll(rec.Body)
		// - unmarshal or decode the JSON to a declared empty interface.
		var bodyJson map[string]interface{}
		json.Unmarshal(body, &bodyJson)

		return bodyJson
	}
	t.Run("should return 200 status ok and internal error", func(t *testing.T) {
		//set up
		gc, rec := setUpEcho(http.MethodGet, "/variants/overview")

		// perform
		serviceInfoMvc.GetServiceInfo(gc)

		// verify response status
		assert.Equal(t, http.StatusOK, rec.Code)

		// verify body
		json := getJsonBody(rec)

		// - detailed
		assert.Equal(t, json["bento"].(map[string]interface{})["dataService"].(bool), true)

		assert.Equal(t, json["id"].(string), string(serviceInfo.SERVICE_ID))
		assert.Equal(t, json["name"].(string), string(serviceInfo.SERVICE_NAME))
		assert.Equal(t, json["description"].(string), string(serviceInfo.SERVICE_DESCRIPTION))
	})
}
