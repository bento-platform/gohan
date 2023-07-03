package api

import (
	"fmt"
	c "gohan/api/models/constants"
	a "gohan/api/models/constants/assembly-id"
	gq "gohan/api/models/constants/genotype-query"
	s "gohan/api/models/constants/sort"
	z "gohan/api/models/constants/zygosity"
	"gohan/api/models/dtos"
	common "gohan/api/tests/common"
	testConsts "gohan/api/tests/common/constants"
	ratt "gohan/api/tests/common/constants/referenceAlternativeTestType"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"testing"

	. "github.com/ahmetb/go-linq"

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

func TestCanGetReferenceSamples(t *testing.T) {
	// trigger
	runAndValidateGenotypeQueryResults(t, gq.REFERENCE, validateReferenceSample)
}

func TestCanGetAlternateSamples(t *testing.T) {
	// trigger
	runAndValidateGenotypeQueryResults(t, gq.ALTERNATE, validateAlternateSample)
}

func TestCanGetHeterozygousSamples(t *testing.T) {
	// trigger
	runAndValidateGenotypeQueryResults(t, gq.HETEROZYGOUS, validateHeterozygousSample)
}

func TestCanGetHomozygousReferenceSamples(t *testing.T) {
	// trigger
	runAndValidateGenotypeQueryResults(t, gq.HOMOZYGOUS_REFERENCE, validateHomozygousReferenceSample)
}

func TestCanGetHomozygousAlternateSamples(t *testing.T) {
	// trigger
	runAndValidateGenotypeQueryResults(t, gq.HOMOZYGOUS_ALTERNATE, validateHomozygousAlternateSample)
}

func TestCanGetHomozygousAlternateVariantsWithVariousReferences(t *testing.T) {
	// setup
	specificValidation := func(__t *testing.T, call *dtos.VariantCall, referenceAllelePattern string, alternativeAllelePattern string) {
		// ensure test is formatted correctly
		assert.True(__t, alternativeAllelePattern == "")

		// validate variant
		assert.Contains(__t, call.Ref, referenceAllelePattern)

		validateHomozygousAlternateSample(__t, call)
	}

	executeReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HOMOZYGOUS_ALTERNATE, ratt.Reference, specificValidation)
}

func TestCanGetHomozygousReferenceVariantsWithVariousReferences(t *testing.T) {
	// setup
	specificValidation := func(__t *testing.T, call *dtos.VariantCall, referenceAllelePattern string, alternativeAllelePattern string) {
		// ensure test is formatted correctly
		assert.True(__t, alternativeAllelePattern == "")

		// validate variant
		assert.Contains(__t, call.Ref, referenceAllelePattern)

		validateHomozygousReferenceSample(__t, call)
	}

	executeReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HOMOZYGOUS_REFERENCE, ratt.Reference, specificValidation)
}

func TestCanGetHeterozygousVariantsWithVariousReferences(t *testing.T) {
	// setup
	specificValidation := func(__t *testing.T, call *dtos.VariantCall, referenceAllelePattern string, alternativeAllelePattern string) {
		// ensure test is formatted correctly
		assert.True(__t, alternativeAllelePattern == "")

		// validate variant
		assert.Contains(__t, call.Ref, referenceAllelePattern)

		validateHeterozygousSample(__t, call)
	}

	// trigger
	executeReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HETEROZYGOUS, ratt.Reference, specificValidation)
}

func TestCanGetHomozygousAlternateVariantsWithVariousAlternatives(t *testing.T) {
	// setup
	specificValidation := func(__t *testing.T, call *dtos.VariantCall, referenceAllelePattern string, alternativeAllelePattern string) {
		// ensure test is formatted correctly
		assert.True(__t, referenceAllelePattern == "")

		// validate variant
		assert.Contains(__t, call.Alt, alternativeAllelePattern)

		validateHomozygousAlternateSample(__t, call)
	}

	// trigger
	executeReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HOMOZYGOUS_ALTERNATE, ratt.Alternative, specificValidation)
}

func TestCanGetHomozygousReferenceVariantsWithVariousAlternatives(t *testing.T) {
	// setup
	specificValidation := func(__t *testing.T, call *dtos.VariantCall, referenceAllelePattern string, alternativeAllelePattern string) {
		// ensure test is formatted correctly
		assert.True(__t, referenceAllelePattern == "")

		// validate variant
		assert.Contains(__t, call.Alt, alternativeAllelePattern)

		validateHomozygousReferenceSample(__t, call)
	}

	// trigger
	executeReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HOMOZYGOUS_REFERENCE, ratt.Alternative, specificValidation)
}

func TestCanGetHeterozygousVariantsWithVariousAlternatives(t *testing.T) {
	// setup
	specificValidation := func(__t *testing.T, call *dtos.VariantCall, referenceAllelePattern string, alternativeAllelePattern string) {
		// ensure test is formatted correctly
		assert.True(__t, referenceAllelePattern == "")

		// validate variant
		assert.Contains(__t, call.Alt, alternativeAllelePattern)

		validateHeterozygousSample(__t, call)
	}

	// trigger
	executeReferenceOrAlternativeQueryTestsOfVariousPatterns(t, gq.HETEROZYGOUS, ratt.Alternative, specificValidation)
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
func executeReferenceOrAlternativeQueryTestsOfVariousPatterns(_t *testing.T,
	genotypeQuery c.GenotypeQuery, refAltTestType testConsts.ReferenceAlternativeTestType,
	specificValidation func(__t *testing.T, call *dtos.VariantCall, referenceAllelePattern string, alternativeAllelePattern string)) {

	// TODO: use some kind of Allele Enum
	patterns := []string{"A", "C", "T", "G"}
	var patWg sync.WaitGroup
	for _, pat := range patterns {
		patWg.Add(1)
		go func(_pat string, _patWg *sync.WaitGroup) {
			defer _patWg.Done()

			switch refAltTestType {
			case ratt.Reference:
				runAndValidateReferenceOrAlternativeQueryResults(_t, genotypeQuery, _pat, "", specificValidation)
			case ratt.Alternative:
				runAndValidateReferenceOrAlternativeQueryResults(_t, genotypeQuery, "", _pat, specificValidation)
			default:
				println("Skipping Test -- no Ref/Alt Test Type provided")
			}

		}(pat, &patWg)
	}
	patWg.Wait()
}

func runAndValidateReferenceOrAlternativeQueryResults(_t *testing.T,
	genotypeQuery c.GenotypeQuery,
	referenceAllelePattern string, alternativeAllelePattern string,
	specificValidation func(__t *testing.T, call *dtos.VariantCall, referenceAllelePattern string, alternativeAllelePattern string)) {

	allDtoResponses := getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(_t, true, s.Undefined, genotypeQuery, referenceAllelePattern, alternativeAllelePattern)

	// assert that all of the responses include sample sets with the appropriate zygosity
	// - * accumulate all variants into a single list using the set of SelectManyT's and the SelectT
	// - ** iterate over each variant in the ForEachT
	// var accumulatedVariants []*indexes.Variant
	var accumulatedCalls []*dtos.VariantCall

	From(allDtoResponses).SelectManyT(func(resp dtos.VariantGetReponse) Query { // *
		return From(resp.Results)
	}).SelectManyT(func(data dtos.VariantGetResult) Query {
		return From(data.Calls)
	}).ForEachT(func(call dtos.VariantCall) { // **
		accumulatedCalls = append(accumulatedCalls, &call)
	})

	if len(accumulatedCalls) == 0 {
		_t.Skip(fmt.Sprintf("No variants returned for patterns ref: '%s', alt: '%s'! Skipping --", referenceAllelePattern, alternativeAllelePattern))
	}

	for _, v := range accumulatedCalls {
		assert.NotNil(_t, v.Id)
		specificValidation(_t, v, referenceAllelePattern, alternativeAllelePattern)
	}

}

func runAndValidateGenotypeQueryResults(_t *testing.T, genotypeQuery c.GenotypeQuery, specificValidation func(__t *testing.T, call *dtos.VariantCall)) {

	allDtoResponses := getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(_t, true, s.Undefined, genotypeQuery, "", "")

	// assert that all of the responses include heterozygous sample sets
	// - * accumulate all samples into a single list using the set of SelectManyT's and the SelectT
	// - ** iterate over each sample in the ForEachT
	// var accumulatedSamples []*indexes.Sample
	var accumulatedCalls []*dtos.VariantCall

	From(allDtoResponses).SelectManyT(func(resp dtos.VariantGetReponse) Query { // *
		return From(resp.Results)
	}).SelectManyT(func(data dtos.VariantGetResult) Query {
		return From(data.Calls)
	}).ForEachT(func(call dtos.VariantCall) { // **
		accumulatedCalls = append(accumulatedCalls, &call)
	})

	if len(accumulatedCalls) == 0 {
		_t.Skip("No samples returned! Skipping --")
	}

	for _, c := range accumulatedCalls {
		assert.NotEmpty(_t, c.SampleId)
		assert.NotEmpty(_t, c.GenotypeType)

		specificValidation(_t, c)
	}
}

func getOverviewResultCombinations(chromosomeStruct interface{}, sampleIdsStruct interface{}, assemblyIdsStruct interface{}) [][]string {
	var allCombinations = [][]string{}

	for i, _ := range chromosomeStruct.(map[string]interface{}) {
		for j, _ := range sampleIdsStruct.(map[string]interface{}) {
			for k, _ := range assemblyIdsStruct.(map[string]interface{}) {
				allCombinations = append(allCombinations, []string{i, j, k})
			}
		}
	}

	return allCombinations
}

func getAllDtosOfVariousCombinationsOfChromosomesAndSampleIds(_t *testing.T, includeInfo bool, sortByPosition c.SortDirection, genotype c.GenotypeQuery, referenceAllelePattern string, alternativeAllelePattern string) []dtos.VariantGetReponse {
	cfg := common.InitConfig()

	// retrieve the overview
	overviewJson := common.GetVariantsOverview(_t, cfg)

	// ensure the response is valid
	// TODO: error check instead of nil check
	assert.NotNil(_t, overviewJson)

	// generate all possible combinations of
	// available samples, assemblys, and chromosomes
	overviewCombinations := getOverviewResultCombinations(overviewJson["chromosomes"], overviewJson["sampleIDs"], overviewJson["assemblyIDs"])

	// avoid overflow:
	// - shuffle all combinations and take top x
	x := 10
	croppedCombinations := make([][]string, len(overviewCombinations))
	perm := rand.Perm(len(overviewCombinations))
	for i, v := range perm {
		croppedCombinations[v] = overviewCombinations[i]
	}
	if len(croppedCombinations) > x {
		croppedCombinations = croppedCombinations[:x]
	}

	// initialize a common slice in which to
	// accumulate al responses asynchronously
	allDtoResponses := []dtos.VariantGetReponse{}
	allDtoResponsesMux := sync.RWMutex{}

	var combWg sync.WaitGroup
	for _, combination := range croppedCombinations {
		combWg.Add(1)
		go func(_wg *sync.WaitGroup, _combination []string) {
			defer _wg.Done()

			chrom := _combination[0]
			sampleId := _combination[1]
			assemblyId := a.CastToAssemblyId(_combination[2])

			// make the call
			dto := common.BuildQueryAndMakeGetVariantsCall(chrom, sampleId, includeInfo, sortByPosition, genotype, assemblyId, referenceAllelePattern, alternativeAllelePattern, "", false, _t, cfg)

			assert.Equal(_t, 1, len(dto.Results))

			// accumulate all response objects
			// to a common slice in an
			// asynchronous-safe manner
			allDtoResponsesMux.Lock()
			allDtoResponses = append(allDtoResponses, dto)
			allDtoResponsesMux.Unlock()
		}(&combWg, combination)
	}

	combWg.Wait()

	return allDtoResponses
}

// --- sample validation
func validateReferenceSample(__t *testing.T, call *dtos.VariantCall) {
	assert.True(__t, call.GenotypeType == z.ZygosityToString(z.Reference))
}

func validateAlternateSample(__t *testing.T, call *dtos.VariantCall) {
	assert.True(__t, call.GenotypeType == z.ZygosityToString(z.Alternate))
}

func validateHeterozygousSample(__t *testing.T, call *dtos.VariantCall) {
	assert.True(__t, call.GenotypeType == z.ZygosityToString(z.Heterozygous))
}

func validateHomozygousReferenceSample(__t *testing.T, call *dtos.VariantCall) {
	assert.True(__t, call.GenotypeType == z.ZygosityToString(z.HomozygousReference))
}

func validateHomozygousAlternateSample(__t *testing.T, call *dtos.VariantCall) {
	assert.True(__t, call.GenotypeType == z.ZygosityToString(z.HomozygousAlternate))
}

// --
