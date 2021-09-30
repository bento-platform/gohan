package api

import (
	"testing"
	common "tests/common"

	"github.com/stretchr/testify/assert"
)

const (
	GenesOverviewPath string = "%s/genes/overview"
)

func TestGenesOverview(t *testing.T) {
	cfg := common.InitConfig()

	overviewJson := getGenesOverview(t, cfg)
	assert.NotNil(t, overviewJson)
}
