package schemas

import (
	c "gohan/api/models/constants"
	so "gohan/api/models/constants/search"
)

type Schema map[string]interface{}

var VARIANT_TABLE_METADATA_SCHEMA Schema = map[string]interface{}{
	"$id":         "variant:table_metadata", // TODO: Real ID
	"$schema":     "http://json-schema.org/draft-07/schema#",
	"description": "Bento variant data type metadata schema",
	"type":        "object",
	"required":    []string{},
	"properties": map[string]interface{}{
		"created": map[string]interface{}{
			"type":                "string",
			"chord_autogenerated": true, // TODO: Extend schema
		},
		"updated": map[string]interface{}{
			"type":                "string",
			"chord_autogenerated": true, // TODO: Extend schema
		},
	},
}

var VARIANT_SCHEMA Schema = map[string]interface{}{
	"$id":         "variant:variant", // TODO: Real ID
	"$schema":     "http://json-schema.org/draft-07/schema#",
	"description": "Bento variant data type",
	"type":        "object",
	"required": []string{
		"assembly_id",
		"chromosome",
		"start",
		"end",
		"calls",
	},
	"search": map[string]interface{}{
		"operations": []c.SearchOperation{},
	},
	"properties": map[string]interface{}{
		"assembly_id": map[string]interface{}{
			"description": "Reference genome assembly ID.",
			"enum": []string{
				"GRCh38",
				"GRCh37",
				"NCBI36",
				"Other",
			},
			"search": map[string]interface{}{
				"canNegate": false,
				"operations": []string{
					"eq",
				},
				"order":     0,
				"queryable": "all",
				"required":  true,
				"type":      "single",
			},
			"type": "string",
		},
		"calls": map[string]interface{}{
			"type":        "array",
			"description": "Called instances of this variant on samples.",
			"items":       VARIANT_CALL_SCHEMA,
			"search": map[string]interface{}{
				"required": false,
				"type":     "unlimited",
				"order":    7,
			},
		},
		"chromosome": map[string]interface{}{
			"description": "Reference genome chromosome identifier (e.g. 17 or X)",
			"search": map[string]interface{}{
				"canNegate": false,
				"operations": []string{
					"eq",
				},
				"order":     1,
				"queryable": "all",
				"required":  true,
				"type":      "single",
			},
			"type": "string",
		},
		"alternative": map[string]interface{}{
			"description": "Alternate allele",
			"search": map[string]interface{}{
				"canNegate": false,
				"operations": []string{
					"eq",
				},
				"order":     1,
				"queryable": "all",
				"required":  false,
				"type":      "single",
			},
			"type": "string",
		},
		"reference": map[string]interface{}{
			"description": "Reference allele",
			"search": map[string]interface{}{
				"canNegate": false,
				"operations": []string{
					"eq",
				},
				"order":     1,
				"queryable": "all",
				"required":  false,
				"type":      "single",
			},
			"type": "string",
		},
		"start": map[string]interface{}{
			"description": "1-indexed start location of the variant on the chromosome.",
			"search": map[string]interface{}{
				"canNegate": false,
				"operations": []string{
					"eq",
					"lt",
					"le",
					"gt",
					"ge",
				},
				"order":     2,
				"queryable": "all",
				"required":  true,
				"type":      "unlimited",
			},
			"type": "integer",
		},
		"end": map[string]interface{}{
			"description": "1-indexed end location (exclusive) of the variant on the chromosome, in terms of the number of bases in the reference sequence for the variant.",
			"search": map[string]interface{}{
				"canNegate": true,
				"operations": []string{
					"eq",
					"lt",
					"le",
					"gt",
					"ge",
				},
				"order":     3,
				"queryable": "all",
				"required":  false,
				"type":      "unlimited",
			},
			"type": "integer",
		},
	},
}

var VARIANT_CALL_SCHEMA Schema = map[string]interface{}{
	"id":          "variant:variant_call", // TODO: Real ID
	"type":        "object",
	"description": "An object representing a called instance of a variant.",
	"required": []string{
		"sample_id",
		"genotype_type",
	},
	"properties": map[string]interface{}{
		"sample_id": map[string]interface{}{
			"type":        "string",
			"description": "Variant call sample ID.", // TODO: More detailed?
			"search": map[string]interface{}{
				"operations": []c.SearchOperation{so.SEARCH_OP_EQ},
				"queryable":  "internal",
				"canNegate":  true,
				"required":   false,
				"type":       "single",
				"order":      0,
			},
		},
		"genotype_type": map[string]interface{}{
			"description": "Variant call genotype type.",
			"enum": []string{
				"MISSING",
				"MISSING_UPSTREAM_DELETION",
				"REFERENCE",
				"ALTERNATE",
				"HOMOZYGOUS_REFERENCE",
				"HETEROZYGOUS",
				"HOMOZYGOUS_ALTERNATE",
			},
			"search": map[string]interface{}{
				"canNegate": true,
				"operations": []string{
					"eq",
				},
				"order":     2.0,
				"queryable": "all",
				"required":  true,
				"type":      "single",
			},
			"type": "string",
		},
	},
}
