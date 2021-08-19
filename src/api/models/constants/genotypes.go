package constants

type Genotype string

const (
	GT_UNCALLED Genotype = ""

	// # Haploid
	GT_REFERENCE Genotype = "REFERENCE"
	GT_ALTERNATE Genotype = "ALTERNATE"

	// # Diploid or higher
	GT_HOMOZYGOUS_REFERENCE Genotype = "HOMOZYGOUS_REFERENCE"
	GT_HETEROZYGOUS         Genotype = "HETEROZYGOUS"
	GT_HOMOZYGOUS_ALTERNATE Genotype = "HOMOZYGOUS_ALTERNATE"
)
