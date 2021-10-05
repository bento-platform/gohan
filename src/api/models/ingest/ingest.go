package ingest

import (
	"github.com/google/uuid"
)

type State string

const (
	Queued      State = "Queued"
	Downloading       = "Downloading"
	Running           = "Running"
	Done              = "Done"
	Error             = "Error"
)

type VariantIngestRequest struct {
	Id        uuid.UUID `json:"id"`
	Filename  string    `json:"filename"`
	State     State     `json:"state"`
	Message   string    `json:"message"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}

type GeneIngestRequest struct {
	Filename  string `json:"filename"`
	State     State  `json:"state"`
	Message   string `json:"message"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type IngestResponseDTO struct {
	Id       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	State    State     `json:"state"`
	Message  string    `json:"message"`
}
