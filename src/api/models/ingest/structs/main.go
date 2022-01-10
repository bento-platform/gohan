package structs

import (
	"api/models/indexes"
	"sync"
)

type IngestionQueueStructure struct {
	Variant   *indexes.Variant
	WaitGroup *sync.WaitGroup
}

type GeneIngestionQueueStructure struct {
	Gene      *indexes.Gene
	WaitGroup *sync.WaitGroup
}
