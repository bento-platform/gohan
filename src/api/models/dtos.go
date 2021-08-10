package models

import "github.com/google/uuid"

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

type IngestRequest struct {
	Id        uuid.UUID `json:"id"`
	Filename  string    `json:"filename"`
	State     string    `json:"state"`
	Message   string    `json:"message"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}

type IngestResponseDTO struct {
	Id       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	State    string    `json:"state"`
	Message  string    `json:"message"`
}
