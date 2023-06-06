package indexes

import (
	c "gohan/api/models/constants"
)

type Variant struct {
	Chrom  string   `json:"chrom"`
	Pos    int      `json:"pos"`
	Id     string   `json:"id"`
	Ref    []string `json:"ref"`
	Alt    []string `json:"alt"`
	Format []string `json:"format"`
	Qual   int      `json:"qual"`
	Filter string   `json:"filter"`
	Info   []Info   `json:"info"`

	Sample Sample `json:"sample"`

	FileId     string       `json:"fileId"`
	Dataset    string       `json:"dataset"`
	AssemblyId c.AssemblyId `json:"assemblyId"`
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
	Genotype             Genotype   `json:"genotype"`
	GenotypeProbability  []float64  `json:"genotypeProbability"`  // -1 = no call (equivalent to a '.')
	PhredScaleLikelyhood []float64  `json:"phredScaleLikelyhood"` // -1 = no call (equivalent to a '.')
	Alleles              AllelePair `json:"alleles"`
}
type AllelePair struct {
	Left  string `json:"left"`
	Right string `json:"right"`
}

type Genotype struct {
	Phased   bool       `json:"phased"`
	Zygosity c.Zygosity `json:"zygosity"`
}

type Gene struct {
	Name       string       `json:"name"`
	Chrom      string       `json:"chrom"`
	Start      int          `json:"start"`
	End        int          `json:"end"`
	AssemblyId c.AssemblyId `json:"assemblyId"`
}
