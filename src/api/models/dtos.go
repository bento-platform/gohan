package models

type VariantsResponseDTO struct {
	Status   int                        `json:"status"`
	Message  string                     `json:"message"`
	Data     []VariantResponseDataModel `json:"data"`
	DataType string                     `json:"data_type"` // i.e.: "variants"
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
