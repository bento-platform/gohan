package workflows

import (
	c "api/models/constants"
	a "api/models/constants/assembly-id"
)

type WorkflowSchema map[string]interface{}

var WORKFLOW_VARIANT_SCHEMA WorkflowSchema = map[string]interface{}{
	"ingestion": map[string]interface{}{
		"gohan_vcf_gz": map[string]interface{}{
			"name":        "Compressed-VCF Elasticsearch Indexing",
			"description": "This ingestion workflow will validate and ingest a BGZip-Compressed-VCF into Elasticsearch.",
			"data_type":   "variant",
			"file":        "vcf_gz.wdl",
			"skipDrs":     true, // WIP
			"inputs": []map[string]interface{}{
				{
					"id":         "vcf_gz_files",
					"type":       "file[]",
					"required":   true,
					"extensions": []string{".vcf.gz"},
				},
				{
					"id":       "assembly_id",
					"type":     "enum",
					"required": true,
					"values":   []c.AssemblyId{a.GRCh38, a.GRCh37},
					"default":  "GRCh38",
				},
			},
			"outputs": []map[string]interface{}{},
		},
	},
	"analysis": map[string]interface{}{},
}
