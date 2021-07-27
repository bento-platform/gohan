package models

var VcfHeaders = []string{"chrom", "pos", "id", "ref", "alt", "qual", "filter", "info", "format"}

type Variant struct {
	Chrom  int      `json:"chrom"`
	Pos    int      `json:"pos"`
	Id     string   `json:"id"`
	Ref    []string `json:"ref"`
	Alt    []string `json:"alt"`
	Format string   `json:"format"`
	Qual   int      `json:"qual"`
	Filter string   `json:"filter"`
	Info   []Info   `json:"info"`

	Samples []Sample `json:"samples"`
	FileId  string   `json:"fileId"`
}

type Info struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}

type Sample struct {
	SampleId  string `json:"sampleId"`
	Variation string `json:"variation"`
}

type VariantResponseDataModel struct {
	VariantId string                   `json:"variantId"`
	SampleId  string                   `json:"sampleId"`
	Count     int                      `json:"count"`
	Results   []map[string]interface{} `json:"results"` // []Variant
}

type VariantsResponseDTO struct {
	Status  int                        `json:"status"`
	Message string                     `json:"message"`
	Data    []VariantResponseDataModel `json:"data"`
}
