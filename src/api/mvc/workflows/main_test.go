package workflows

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func serve(t *testing.T, name string) *httptest.ResponseRecorder {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/workflows/"+name, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/workflows/:file")
	c.SetParamNames("file")
	c.SetParamValues(name)

	if err := WorkflowsServeFile(c); err != nil {
		e.HTTPErrorHandler(err, c)
	}
	return rec
}

func TestServeFileJSONDefinition(t *testing.T) {
	rec := serve(t, "vcf_gz")
	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Inputs []struct {
			ID string `json:"id"`
		} `json:"inputs"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.NotEmpty(t, body.Inputs)

	ids := make([]string, 0, len(body.Inputs))
	for _, in := range body.Inputs {
		assert.NotEmpty(t, in.ID)
		ids = append(ids, in.ID)
	}
	assert.Contains(t, ids, "project_dataset")
	assert.Contains(t, ids, "vcf_gz_file_names")
}

func TestServeFileUnknownWorkflow(t *testing.T) {
	assert.Equal(t, http.StatusNotFound, serve(t, "nope").Code)
}

func TestServeFileWDL(t *testing.T) {
	// the repo's real WDL lives in ../../workflows
	orig := workflowsDir
	workflowsDir = "../../workflows"
	defer func() { workflowsDir = orig }()

	want, err := os.ReadFile("../../workflows/vcf_gz.wdl")
	require.NoError(t, err)

	rec := serve(t, "vcf_gz.wdl")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, string(want), rec.Body.String())
}

func TestServeFileRejectsTraversal(t *testing.T) {
	assert.Equal(t, http.StatusBadRequest, serve(t, "../secret.wdl").Code)
}
