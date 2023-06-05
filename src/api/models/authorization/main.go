package authorization

import (
	authzConstants "gohan/api/models/constants/authorization"
)

type Resource interface{}
type ResourceEverything struct {
	Resource
	Everything bool `json:"everything"`
}
type ResourceSpecific struct {
	Resource
	Project  string `json:"project"`   // cannot be empty, should be a UUID
	Dataset  string `json:"dataset"`   // maybe empty, use "" to check for 'zero' value
	DataType string `json:"data_type"` // maybe empty, use "" to check for 'zero' value
}

type Permission struct {
	Verb  authzConstants.PermissionVerb
	Noun  authzConstants.PermissionNoun
	Level authzConstants.PermissionLevel
}

type Issuer struct {
	Iss string `json:"iss"`
}
type IssuerClient struct {
	Issuer
	Client string `json:"client"`
}
type IssuerSubject struct {
	Issuer
	Sub string `json:"sub"`
}

type SubjectEveryone struct {
	Everyone bool `json:"everyone"`
}
type SubjectGroup struct {
	Group int `json:"group"`
}

type GroupMembershipExpr struct {
	Expr []interface{} `json:"expr"`
	// TODO: ensure each element is either
	// of type '[]interface{}' or 'string'
}
