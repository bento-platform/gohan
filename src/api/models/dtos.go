package models

type BentoV2CompatibleVariantsResponseDTO struct {
	Results []BentoV2CompatibleVariantResponseDataModel `json:"results"`
}
type BentoV2CompatibleVariantResponseDataModel struct {
	SampleId string `json:"sample_id"`
}

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

type GenesResponseDTO struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Term    string `json:"term"`
	Count   int    `json:"count"`
	Results []Gene `json:"results"` // []Gene
}
