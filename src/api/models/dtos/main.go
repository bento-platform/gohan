package dtos

import (
	"api/models/constants"
	"api/models/indexes"
)

// ---- Variants
type VariantReponse struct {
	Status  int    `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}
type VariantGetReponse struct {
	VariantReponse
	Results []VariantGetResult `json:"results"`
}
type VariantCountReponse struct {
	VariantReponse
	Results []VariantCountResult `json:"results"`
}

type VariantResult struct {
	Query      string               `json:"query,omitempty"`
	AssemblyId constants.AssemblyId `json:"assembly_id"`
	Chromosome string               `json:"chromosome"`
	Start      int                  `json:"start"`
	End        int                  `json:"end"`
}

type VariantGetResult struct {
	VariantResult
	Calls []VariantCall `json:"calls"`
}
type VariantCountResult struct {
	VariantResult
	Count int `json:"count"`
}

type VariantCall struct {
	Chrom  string   `json:"chrom,omitempty"`
	Pos    int      `json:"pos,omitempty"`
	Id     string   `json:"id,omitempty"`
	Ref    []string `json:"ref,omitempty"`
	Alt    []string `json:"alt,omitempty"`
	Format []string `json:"format,omitempty"`
	Qual   int      `json:"qual,omitempty"`
	Filter string   `json:"filter,omitempty"`

	Info []indexes.Info `json:"info,omitempty"` // TODO; refactor?

	SampleId     string `json:"sample_id"`
	GenotypeType string `json:"genotype_type"`
	// TODO: GenotypeProbability, PhredScaleLikelyhood ?

	AssemblyId constants.AssemblyId `json:"assemblyId,omitempty"`
}

// -- Genes
type GenesResponseDTO struct {
	Status  int            `json:"status"`
	Message string         `json:"message"`
	Term    string         `json:"term"`
	Count   int            `json:"count"`
	Results []indexes.Gene `json:"results"` // []Gene
}
