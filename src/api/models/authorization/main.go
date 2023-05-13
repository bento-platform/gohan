package authorization

import (
	c "gohan/api/models/constants"
)

type Resource interface{}
type ResourceEverything struct {
	Everything bool `json:"everything"`
}

// TODO: implement ResourceSpecific

type Permission struct {
	Verb  c.PermissionVerb
	Noun  c.PermissionNoun
	Level c.PermissionLevel
}
