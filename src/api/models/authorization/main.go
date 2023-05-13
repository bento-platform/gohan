package authorization

import (
	authzConstants "gohan/api/models/constants/authorization"
)

type Resource interface{}
type ResourceEverything struct {
	Everything bool `json:"everything"`
}

// TODO: implement ResourceSpecific

type Permission struct {
	Verb  authzConstants.PermissionVerb
	Noun  authzConstants.PermissionNoun
	Level authzConstants.PermissionLevel
}
