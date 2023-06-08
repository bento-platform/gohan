package dtos

import (
	"gohan/api/models/constants"
	"gohan/api/models/indexes"
	"time"
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

	SampleId     string   `json:"sample_id"`
	GenotypeType string   `json:"genotype_type,omitempty"`
	Alleles      []string `json:"alleles,omitempty"`
	// TODO: GenotypeProbability, PhredScaleLikelyhood ?

	AssemblyId constants.AssemblyId `json:"assemblyId,omitempty"`
	Dataset    string               `json:"dataset,omitempty"`
	DocumentId string               `json:"documentId,omitempty"`
}

// --- Dataset
type DatasetSummaryResponseDto struct {
	Count            int                    `json:"count"`
	DataTypeSpecific map[string]interface{} `json:"data_type_specific"` // TODO: type-safety?
}

// -- Genes
type GenesResponseDTO struct {
	Status  int            `json:"status"`
	Message string         `json:"message"`
	Term    string         `json:"term"`
	Count   int            `json:"count"`
	Results []indexes.Gene `json:"results"` // []Gene
}

// -- Errors
type GeneralErrorResponseDto struct {
	Status    int            `json:"status,omitempty"`
	Message   string         `json:"message,omitempty"`
	Errors    []GeneralError `json:"errors,omitempty"`
	Timestamp time.Time      `json:"timestamp,omitempty"`
}
type GeneralError struct {
	Message string `json:"message,omitempty"`
}
