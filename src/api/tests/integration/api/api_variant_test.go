package api

import (
	"fmt"
	common "gohan/api/tests/common"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithInvalidAuthenticationToken(t *testing.T) {
	cfg := common.InitConfig()

	request, _ := http.NewRequest("GET", cfg.Api.Url, nil)

	request.Header.Add("X-AUTHN-TOKEN", "gibberish")

	client := &http.Client{}
	response, responseErr := client.Do(request)
	assert.Nil(t, responseErr)

	defer response.Body.Close()

	// default response without a valid authentication token is is 401; consider it a pass
	var shouldBe int
	if cfg.AuthX.IsAuthorizationEnabled {
		shouldBe = 401
	} else {
		shouldBe = 200
	}

	assert.Equal(t, shouldBe, response.StatusCode, fmt.Sprintf("Error -- Api GET / Status: %s ; Should be %d", response.Status, shouldBe))
}

func TestVariantsOverview(t *testing.T) {
	cfg := common.InitConfig()

	overviewJson := common.GetVariantsOverview(t, cfg)
	assert.NotNil(t, overviewJson)
}

func TestCanGetVariantsWithWildcardAlternatives(t *testing.T) {
	cfg := common.InitConfig()
	allele := "ATTN" // example allele - TODO: render more sophisticated randomization
	// TODO: improve variant call testing from being 1 call to many random ones
	dtos := common.BuildQueryAndMakeGetVariantsCall("14", "*", true, "asc", "HETEROZYGOUS", "GRCh37", "", allele, "", false, t, cfg)
	for _, dto := range dtos.Results {
		for _, call := range dto.Calls {
			// ensure, for each call, that at least
			// 1 of the alt's present matches the allele
			// queried for
			allNonWildcardCharactersMatch := true
			// iterate over all 'alt's in the call
			for _, alt := range call.Alt {
				// iterate over all characters for each alt
				for altIndex, altChar := range alt {
					// ensure the index is within bounds (length of the allele)
					// 'alt's are slices of strings, and not all 'alt's in these slices need to match
					if altIndex <= len(allele) {
						// obtain the character at the index for the iteration
						alleleChar := []rune(allele)[altIndex]
						if string(alleleChar) != "N" && altChar != alleleChar {
							// if the non-wildcard characters don't match, test fails
							allNonWildcardCharactersMatch = false
							break
						}
					}
				}
				if !allNonWildcardCharactersMatch {
					break
				}
			}
			assert.True(t, allNonWildcardCharactersMatch)
		}
	}

}
func TestCanGetVariantsWithWildcardReferences(t *testing.T) {
	cfg := common.InitConfig()
	allele := "ATTN" // example allele - TODO: render more sophisticated randomization
	// TODO: improve variant call testing from being 1 call to many random ones
	dtos := common.BuildQueryAndMakeGetVariantsCall("14", "*", true, "asc", "HETEROZYGOUS", "GRCh37", allele, "", "", false, t, cfg)
	for _, dto := range dtos.Results {
		for _, call := range dto.Calls {
			// ensure, for each call, that at least
			// 1 of the ref's present matches the allele
			// queried for
			allNonWildcardCharactersMatch := true
			// iterate over all 'ref's in the call
			for _, ref := range call.Ref {
				// iterate over all characters for each ref
				for refIndex, refChar := range ref {
					// ensure the index is within bounds (length of the allele)
					// 'ref's are slices of strings, and not all 'ref's in these slices need to match
					if refIndex <= len(allele) {
						// obtain the character at the index for the iteration
						alleleChar := []rune(allele)[refIndex]
						if string(alleleChar) != "N" && refChar != alleleChar {
							// if the non-wildcard characters don't match, test fails
							allNonWildcardCharactersMatch = false
							break
						}
					}
				}
				if !allNonWildcardCharactersMatch {
					break
				}
			}
			assert.True(t, allNonWildcardCharactersMatch)
		}
	}
}

func TestCanGetVariantsWithWildcardAlleles(t *testing.T) {
	cfg := common.InitConfig()
	// iterate over all 'allele's queried for
	qAlleles := []string{"N", "NN", "NNN", "NNNN", "NNNNN"} // wildcard alleles of different lengths
	for _, qAllele := range qAlleles {
		dtos := common.BuildQueryAndMakeGetVariantsCall("", "*", true, "asc", "", "GRCh38", "", "", qAllele, false, t, cfg)
		for _, dto := range dtos.Results {
			fmt.Printf("Got %d calls from allele query %s \n", len(dto.Calls), qAllele)
			if len(dto.Calls) == 0 {
				continue
			}

			for _, call := range dto.Calls {
				// ensure, for each call, that at least
				// 1 of the alleles present matches one of
				// the alleles queried for
				wildcardCharactersMatch := false

				// - iterate over all 'allele's in the call
				for _, allele := range call.Alleles {
					if len(qAllele) == len(allele) {
						wildcardCharactersMatch = true
						break
					}
				}

				assert.True(t, wildcardCharactersMatch)
			}
		}
	}
}
func TestCanGetVariantsWithWildcardAllelePairs(t *testing.T) {
	cfg := common.InitConfig()

	// wildcard allele pairs of different lengths
	qAllelePairs := [][]string{
		{"N", "N"},
		{"N", "NN"},
		{"NN", "N"},
		{"N", "NNN"},
		{"NNN", "N"}}

	// iterate over all 'allele pairs'
	for _, qAllelePair := range qAllelePairs {
		dtos := common.BuildQueryAndMakeGetVariantsCall("", "*", true, "asc", "", "GRCh38", "", "", strings.Join(qAllelePair, ","), false, t, cfg)
		for _, dto := range dtos.Results {
			if len(dto.Calls) == 0 {
				continue
			}

			for _, call := range dto.Calls {
				// ensure, for each call, that the length
				// of both alleles in the pair match either
				// wildcard query allele-pair lengths
				bothAllelesMatchesEitherQueriedAllele := (len(qAllelePair[0]) == len(call.Alleles[0]) && len(qAllelePair[1]) == len(call.Alleles[1])) ||
					(len(qAllelePair[1]) == len(call.Alleles[1]) && len(qAllelePair[0]) == len(call.Alleles[0])) ||
					(len(qAllelePair[0]) == len(call.Alleles[1]) && len(qAllelePair[1]) == len(call.Alleles[0])) ||
					(len(qAllelePair[1]) == len(call.Alleles[0]) && len(qAllelePair[0]) == len(call.Alleles[1]))

				if !bothAllelesMatchesEitherQueriedAllele {
					fmt.Print(qAllelePair, call.Alleles)
				}

				assert.True(t, bothAllelesMatchesEitherQueriedAllele)
			}
		}
	}
}

func TestGetVariantsCanHandleInvalidWildcardAlleleQuery(t *testing.T) {
	cfg := common.InitConfig()
	// iterate over all 'allele's queried for
	qAlleles := []string{"N", "NN", "NNN", "NNNN", "NNNNN"} // wildcard alleles of different lengths
	for i, _ := range qAlleles {
		if i <= 2 {
			continue
		} // skip valid calls

		limitedAlleles := strings.Join(qAlleles[:i], ",")
		invalidReqResObj := common.BuildQueryAndMakeGetVariantsCall("", "*", true, "asc", "", "GRCh38", "", "", limitedAlleles, true, t, cfg)

		// make sure only an error was returned
		assert.True(t, invalidReqResObj.Status == 400)
		assert.True(t, len(invalidReqResObj.Message) != 0)
		assert.True(t, len(invalidReqResObj.Results) == 0)
	}
}

// -- Common utility functions for api tests
