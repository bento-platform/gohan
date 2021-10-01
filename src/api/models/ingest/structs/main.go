package structs

import (
	"api/models"
	"sync"
)

type IngestionQueueStructure struct {
	Variant   *models.Variant
	WaitGroup *sync.WaitGroup
}

type GeneIngestionQueueStructure struct {
	Gene      *models.Gene
	WaitGroup *sync.WaitGroup
}
