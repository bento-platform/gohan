package models

type VariantsResponseDTO struct {
	Status  int                        `json:"status"`
	Message string                     `json:"message"`
	Data    []VariantResponseDataModel `json:"data"`
}

type VariantResponseDataModel struct {
	VariantId string    `json:"variantId"`
	SampleId  string    `json:"sampleId"`
	Count     int       `json:"count"`
	Results   []Variant `json:"results"`   // []Variant
	SampleIds []string  `json:"sampleIds"` // []string
}

type GenesResponseDTO struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Term    string `json:"term"`
	Count   int    `json:"count"`
	Results []Gene `json:"results"` // []Gene
}
