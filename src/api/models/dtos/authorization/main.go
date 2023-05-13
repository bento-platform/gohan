package authorization

import (
	"encoding/json"
	"fmt"
	mauthz "gohan/api/models/authorization"
)

type PermissionRequestDto struct {
	RequestedResource   mauthz.Resource
	RequiredPermissions []mauthz.Permission
}

func (p *PermissionRequestDto) MarshalJSON() ([]byte, error) {
	// customize the request serialized format:

	// serialize permissions as a simple json list of the contents
	// concatenated with a colon ':'
	reqPermStrArr := make([]string, 0)
	for _, rp := range p.RequiredPermissions {
		reqPermStrArr = append(reqPermStrArr, fmt.Sprintf("%s:%s", rp.Verb, rp.Noun))
	}
	// i.e : '["verb1:noun1", "verb2:noun2", ...]'
	// instead of '{"requiredPermissions": [{"verb":<verb1>, "noun": <noun1>},{"verb":<verb2>, "noun": <noun2>}]}'

	// - serialize and then deserialize requested resource
	rrl, _ := json.Marshal(&p.RequestedResource)
	var reqResMap map[string]interface{}
	json.Unmarshal([]byte(rrl), &reqResMap)
	// - structure the request body using snake case
	res := map[string]interface{}{
		"requested_resource":   reqResMap,
		"required_permissions": reqPermStrArr,
	}
	// - reserialize the request body
	return json.Marshal(&res)
}
