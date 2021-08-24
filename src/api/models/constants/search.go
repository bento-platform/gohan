package constants

type SearchOperation string

const (
	SEARCH_OP_EQ SearchOperation = "eq"
	SEARCH_OP_LT SearchOperation = "lt"
	SEARCH_OP_LE SearchOperation = "le"
	SEARCH_OP_GT SearchOperation = "gt"
	SEARCH_OP_GE SearchOperation = "ge"

	SEARCH_OP_CO SearchOperation = "co"
)
