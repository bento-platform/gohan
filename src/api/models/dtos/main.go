package dtos

import (
	"api/models/constants"
	"api/models/indexes"
)

// ---- PROTOTYPING
type VariantReponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}
type VariantGetReponse struct {
	VariantReponse
	Results []VariantQueryResult `json:"results"`
}
type VariantCountReponse struct {
	VariantReponse
	Count int `json:"count"`
}

type VariantQueryResult struct {
	QueryId    string               `json:"queryId"`
	AssemblyId constants.AssemblyId `json:"assembly_id"`

	Calls []VariantCall `json:"calls"`

	Chromosome string `json:"chromosome"`
	Start      int    `json:"start"`
	End        int    `json:"end"`
}

type VariantCall struct {
	Chrom  string   `json:"chrom"`
	Pos    int      `json:"pos"`
	Id     string   `json:"id"`
	Ref    []string `json:"ref"`
	Alt    []string `json:"alt"`
	Format []string `json:"format"`
	Qual   int      `json:"qual"`
	Filter string   `json:"filter"`

	Info []indexes.Info `json:"info"` // TODO; refactor?

	SampleId     string `json:"sample_id"`
	GenotypeType string `json:"genotype_type"`
	// TODO: GenotypeProbability, PhredScaleLikelyhood ?

	AssemblyId constants.AssemblyId `json:"assemblyId"` // redundant ?
}

// ----

type VariantsResponseDTO struct {
	Status  int                        `json:"status"`
	Message string                     `json:"message"`
	Data    []VariantResponseDataModel `json:"data"`
}
type VariantResponseDataModel struct {
	VariantId string      `json:"variantId"`
	SampleId  string      `json:"sampleId"`
	Count     int         `json:"count"`
	Results   interface{} `json:"results"` // i.e.: []Variant or []string
}

// -- --

type GenesResponseDTO struct {
	Status  int            `json:"status"`
	Message string         `json:"message"`
	Term    string         `json:"term"`
	Count   int            `json:"count"`
	Results []indexes.Gene `json:"results"` // []Gene
}
