package elasticsearch

import (
	"github.com/elastic/go-elasticsearch"
)

func GetDocumentsContainerVariantOrSampleIdInPositionRange(esClient *elasticsearch.Client,
	chromosome int, lowerBound int, upperBound int,
	variantId string, sampleId string,
	reference string, alternative string,
	size int, sortByPosition string,
	includeSamplesInResultSet bool) {
	// TODO : implement query
}
