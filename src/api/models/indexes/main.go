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

var MAPPING_FIELDS_KEYWORD_IG256 = map[string]interface{}{
	"keyword": map[string]interface{}{
		"type":         "keyword",
		"ignore_above": 256,
	},
}
var MAPPING_TEXT = map[string]interface{}{"type": "text", "fields": MAPPING_FIELDS_KEYWORD_IG256}
var MAPPING_LONG = map[string]interface{}{"type": "long"}
var MAPPING_FLOAT64 = map[string]interface{}{"type": "double"}
var MAPPING_BOOL = map[string]interface{}{"type": "boolean"}
var MAPPING_DATE = map[string]interface{}{"type": "date"}

// This mapping is derived from the one exported by Victor from the ICHANGE instance on 2024-11-01,
// using the following commands:
// ./bentoctl.bash shell gohan-api
// --> inside gohan-api container
// curl -u $GOHAN_ES_USERNAME:$GOHAN_ES_PASSWORD bentov2-gohan-elasticsearch:9200/_mapping
var VARIANT_INDEX_MAPPING = map[string]interface{}{
	"properties": map[string]interface{}{
		"chrom":  MAPPING_TEXT,
		"pos":    MAPPING_LONG,
		"id":     MAPPING_TEXT,
		"ref":    MAPPING_TEXT,
		"alt":    MAPPING_TEXT,
		"format": MAPPING_TEXT,
		"qual":   MAPPING_LONG,
		"filter": MAPPING_TEXT,
		"info": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":    MAPPING_TEXT,
				"value": MAPPING_TEXT,
			},
		},
		"sample": map[string]interface{}{
			"properties": map[string]interface{}{
				"id": MAPPING_TEXT,
				"variation": map[string]interface{}{
					"properties": map[string]interface{}{
						"genotype": map[string]interface{}{
							"properties": map[string]interface{}{
								"phased":   MAPPING_BOOL,
								"zygosity": MAPPING_LONG,
							},
						},
						"alleles": map[string]interface{}{
							"properties": map[string]interface{}{
								"left":  MAPPING_TEXT,
								"right": MAPPING_TEXT,
							},
						},
						"phredScaleLikelyhood": MAPPING_LONG,
						"genotypeProbability":  MAPPING_FLOAT64,
					},
				},
			},
		},
		"fileId":      MAPPING_TEXT,
		"dataset":     MAPPING_TEXT,
		"assemblyId":  MAPPING_TEXT,
		"createdTime": MAPPING_DATE,
	},
}

type Gene struct {
	Name       string `json:"name"`
	Chrom      string `json:"chrom"`
	Start      int    `json:"start"`
	End        int    `json:"end"`
	AssemblyId string `json:"assemblyId"`
}
