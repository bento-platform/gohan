package constants

var ValidTableDataTypes = []string{"variant"}
var VcfHeaders = []string{"chrom", "pos", "id", "ref", "alt", "qual", "filter", "info", "format"}

/*
Defines a set of base level
constants and enums to be used
throughout Gohan and it's
associated services.
*/
type AssemblyId string
type Chromosome string
type GenotypeQuery string
type SearchOperation string
type SortDirection string

type Zygosity int
type Ploidy int
