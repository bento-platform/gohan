package workflows

type WorkflowSchema map[string]interface{}

var WORKFLOW_VARIANT_SCHEMA WorkflowSchema = map[string]interface{}{
	"ingestion": map[string]interface{}{
		"vcf_gz": map[string]interface{}{
			"name":        "Compressed VCF Elasticsearch Indexing",
			"description": "This ingestion workflow will validate and ingest a BGZip-compressed VCF into Elasticsearch.",
			"data_type":   "variant",
			"tags":        []string{"variant"},
			"file":        "vcf_gz.wdl",
			"type":        "ingestion",
			"inputs": []map[string]interface{}{
				// User inputs:
				{
					"id":       "project_dataset",
					"type":     "project:dataset",
					"required": true,
					"help":     "The dataset to ingest the variants into.",
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
					"values":   "{{ serviceUrls.reference }}/genomes?response_format=id_list",
				},
				{
					"id":       "filter_out_references",
					"type":     "boolean",
					"required": true,
					"help": "If this is checked, variant calls which are (0, 0) (i.e., homozygous reference " +
						"calls) will not be ingested.",
				},
				// Injected inputs:
				{
					"id":           "gohan_url",
					"type":         "service-url",
					"required":     true,
					"injected":     true,
					"service_kind": "gohan",
				},
				{
					"id":       "access_token",
					"type":     "secret",
					"required": true,
					"injected": true,
					"key":      "access_token",
				},
				{
					"id":       "validate_ssl",
					"type":     "config",
					"required": true,
					"injected": true,
					"key":      "validate_ssl",
				},
			},
		},
	},
	"analysis": map[string]interface{}{},
	"export":   map[string]interface{}{},
}
