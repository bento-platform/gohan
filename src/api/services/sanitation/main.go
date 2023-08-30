package sanitation

import (
	"fmt"
	"time"

	es7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/go-co-op/gocron"

	"gohan/api/models"
)

type (
	SanitationService struct {
		Initialized bool
		Es7Client   *es7.Client
		Config      *models.Config
	}
)

func NewSanitationService(es *es7.Client, cfg *models.Config) *SanitationService {
	ss := &SanitationService{
		Initialized: false,
		Es7Client:   es,
		Config:      cfg,
	}

	ss.Init()

	return ss
}

func (ss *SanitationService) Init() {
	// initialization if necessary
	if !ss.Initialized {
		// - spin up a go routine that will periodically
		//   run through a series of steps to ensure
		//   the system is "sanitary" ; i.e. in an elasticsearch
		//   context, that would mean performing something like
		//   - removing duplicate documents
		//   - cleaning documents that have broken pseudo-foreign keys
		//     - variants -> tables (no longer necessary)
		//   etc...
		go func() {
			// setup cron job
			s := gocron.NewScheduler(time.UTC)

			// clean variant documents with non-existing tables
			s.Every(1).Days().At("04:00:00").Do(func() { // 12am EST
				// nothing for now
			})

			// starts the scheduler in blocking mode, which blocks
			// the current execution path
			s.StartBlocking()
		}()

		ss.Initialized = true
		fmt.Println("Sanitation Service Initialized ..")
	}
}

func setDifference(a, b []string) (c []string) {
	m := make(map[string]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; !ok {
			c = append(c, item)
		}
	}
	return
}
