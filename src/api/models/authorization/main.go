package authorization

import (
	"encoding/json"
	"fmt"
	c "gohan/api/models/constants"
)

type ResourceEverything struct {
	Everything bool `json:"everything"`
}

// TODO: implement ResourceSpecific

type Permission struct {
	Verb  c.PermissionVerb
	Noun  c.PermissionNoun
	Level c.PermissionLevel
}

type PermissionsList struct {
	List []Permission
}

func (p *PermissionsList) MarshalJSON() ([]byte, error) {
	rpStrs := make([]interface{}, 0)
	for _, rp := range p.List {
		rpStrs = append(rpStrs, fmt.Sprintf("%s:%s", rp.Verb, rp.Noun))
	}

	return json.Marshal(&rpStrs)
}
