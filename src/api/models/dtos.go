package models

type VariantsResponseDTO struct {
	Status  int                        `json:"status"`
	Message string                     `json:"message"`
	Data    []VariantResponseDataModel `json:"data"`
}

type VariantResponseDataModel struct {
	VariantId string                   `json:"variantId"`
	SampleId  string                   `json:"sampleId"`
	Count     int                      `json:"count"`
	Results   []map[string]interface{} `json:"results"` // []Variant
}
