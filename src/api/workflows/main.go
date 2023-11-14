package workflows

import (
	c "gohan/api/models/constants"
	a "gohan/api/models/constants/assembly-id"
)

type WorkflowSchema map[string]interface{}

var WORKFLOW_VARIANT_SCHEMA WorkflowSchema = map[string]interface{}{
	"ingestion": map[string]interface{}{
		"vcf_gz": map[string]interface{}{
			"name":        "Compressed-VCF Elasticsearch Indexing",
			"description": "This ingestion workflow will validate and ingest a BGZip-Compressed-VCF into Elasticsearch.",
			"data_type":   "variant",
			"tags":        []string{"variant"},
			"file":        "vcf_gz.wdl",
			"type":        "ingestion",
			"inputs": []map[string]interface{}{
				{
					"id":       "project_dataset",
					"type":     "project:dataset",
					"required": true,
				},
				{
					"id":       "vcf_gz_file_names",
					"type":     "file[]",
					"required": true,
					"pattern":  "^.*\\.vcf\\.gz$",
				},
				{
					"id":       "assembly_id",
					"type":     "enum",
					"required": true,
					"values":   []c.AssemblyId{a.GRCh38, a.GRCh37},
					"default":  "GRCh38",
				},
				{
					"id":       "filter_out_references",
					"type":     "enum",
					"required": true,
					"values":   []string{"true", "false"}, // simulate boolean type
					"default":  "false",
				},
				{
					"id":           "gohan_url",
					"type":         "service-kind",
					"required":     true,
					"injected":     true,
					"service_kind": "gohan",
				},
			},
		},
	},
	"analysis": map[string]interface{}{},
	"export":   map[string]interface{}{},
}
