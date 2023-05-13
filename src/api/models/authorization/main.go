package authorization

import (
	"encoding/json"
	"fmt"
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

type PermissionsList struct {
	List []Permission
}

func (p *PermissionsList) MarshalJSON() ([]byte, error) {
	// serialize struct as a simple json list of the contents
	// rather than an object with a "list" key and [...] value
	rpStrs := make([]interface{}, 0)
	for _, rp := range p.List {
		rpStrs = append(rpStrs, fmt.Sprintf("%s:%s", rp.Verb, rp.Noun))
	}
	// i.e : '["verb1:noun1", "verb2:noun2", ...]'
	// instead of '{"list": [{"verb":<verb1>, "noun": <noun1>},{"verb":<verb2>, "noun": <noun2>}]}'

	return json.Marshal(&rpStrs)
}
