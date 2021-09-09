package models

var VcfHeaders = []string{"chrom", "pos", "id", "ref", "alt", "qual", "filter", "info", "format"}

type Variant struct {
	Chrom  int      `json:"chrom"`
	Pos    int      `json:"pos"`
	Id     string   `json:"id"`
	Ref    []string `json:"ref"`
	Alt    []string `json:"alt"`
	Format []string `json:"format"`
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
	Id        string    `json:"id"`
	Variation Variation `json:"variation"`
}

type Variation struct {
	Genotype             Genotype  `json:"genotype"`
	GenotypeProbability  []float64 `json:"genotypeProbability"`  // -1 = no call (equivalent to a '.')
	PhredScaleLikelyhood []float64 `json:"phredScaleLikelyhood"` // -1 = no call (equivalent to a '.')
}

type Genotype struct {
	Phased      bool `json:"phased"`
	AlleleLeft  int  `json:"alleleLeft"`  // -1 = no call (equivalent to a '.')
	AlleleRight int  `json:"alleleRight"` // -1 = no call (equivalent to a '.')
}
