package indexes

import (
	c "gohan/api/models/constants"
	"time"
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

	FileId      string    `json:"fileId"`
	Dataset     string    `json:"dataset"`
	AssemblyId  string    `json:"assemblyId"`
	CreatedTime time.Time `json:"createdTime"`
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

var MAPPING_KEYWORD = map[string]interface{}{"type": "keyword"}
var MAPPING_TEXT = map[string]interface{}{"type": "text"}
var MAPPING_INT = map[string]interface{}{"type": "integer"}
var MAPPING_FLOAT64 = map[string]interface{}{"type": "double"}
var MAPPING_BOOL = map[string]interface{}{"type": "boolean"}
var MAPPING_DATE = map[string]interface{}{"type": "date"}

var VARIANT_INDEX_MAPPING = map[string]interface{}{
	"properties": map[string]interface{}{
		"chrom":                                 MAPPING_KEYWORD,
		"pos":                                   MAPPING_INT,
		"id":                                    MAPPING_KEYWORD,
		"ref":                                   MAPPING_TEXT,
		"alt":                                   MAPPING_TEXT,
		"format":                                MAPPING_TEXT,
		"qual":                                  MAPPING_INT,
		"filter":                                MAPPING_KEYWORD,
		"info.id":                               MAPPING_KEYWORD,
		"info.value":                            MAPPING_TEXT,
		"sample.id":                             MAPPING_TEXT,
		"sample.variation.genotype.phased":      MAPPING_BOOL,
		"sample.variation.genotype.zygosity.":   MAPPING_INT,
		"sample.variation.genotypeProbability":  MAPPING_FLOAT64,
		"sample.variation.phredScaleLikelyhood": MAPPING_FLOAT64,
		"sample.variation.alleles.left":         MAPPING_TEXT,
		"sample.variation.alleles.right":        MAPPING_TEXT,
		"fileId":                                MAPPING_TEXT,
		"dataset":                               MAPPING_KEYWORD,
		"assemblyId":                            MAPPING_KEYWORD,
		"createdTime":                           MAPPING_DATE,
	},
}

type Gene struct {
	Name       string `json:"name"`
	Chrom      string `json:"chrom"`
	Start      int    `json:"start"`
	End        int    `json:"end"`
	AssemblyId string `json:"assemblyId"`
}
