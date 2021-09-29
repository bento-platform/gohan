package mvc

import (
	"api/contexts"
	"api/models"
	assemblyId "api/models/constants/assembly-id"
	esRepo "api/repositories/elasticsearch"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/mitchellh/mapstructure"
)

func GenesGetByNomenclatureWildcard(c echo.Context) error {
	cfg := c.(*contexts.GohanContext).Config
	es := c.(*contexts.GohanContext).Es7Client

	term := c.QueryParam("term")

	// perform wildcard search if empty/random parameter is passed
	// - set to Unknown to trigger it
	assId := assemblyId.Unknown
	assIdQP := c.QueryParam("assemblyId")
	if assemblyId.CastToAssemblyId(assIdQP) != assemblyId.Unknown {
		// retrieve passed parameter if is valid
		assId = assemblyId.CastToAssemblyId(assIdQP)
	}

	fmt.Printf("Executing wildcard genes search for term %s, assemblyId %s\n", term, assId)

	docs := esRepo.GetGeneDocumentsByTermWildcard(cfg, es, term, assId)

	docsHits := docs["hits"].(map[string]interface{})["hits"]
	allDocHits := []map[string]interface{}{}
	mapstructure.Decode(docsHits, &allDocHits)

	// grab _source for each hit
	var allSources []models.Gene

	for _, r := range allDocHits {
		source := r["_source"].(map[string]interface{})

		// cast map[string]interface{} to struct
		var resultingVariant models.Gene
		mapstructure.Decode(source, &resultingVariant)

		// accumulate structs
		allSources = append(allSources, resultingVariant)
	}

	fmt.Printf("Found %d docs!\n", len(allSources))

	geneResponseDTO := models.GenesResponseDTO{
		Term:    term,
		Count:   len(allSources),
		Results: allSources,
		Status:  200,
		Message: "Success",
	}

	return c.JSON(http.StatusOK, geneResponseDTO)
}
