package authorization

import (
	"encoding/json"
	mauthz "gohan/api/models/authorization"
)

type PermissionRequestDto struct {
	RequestedResource   mauthz.Resource
	RequiredPermissions mauthz.PermissionsList
}

func (p *PermissionRequestDto) MarshalJSON() ([]byte, error) {
	// customize the request serialized format:
	// - serialize and then deserialize permissions
	rpl, _ := json.Marshal(&p.RequiredPermissions)
	var reqPermStrArr []string
	json.Unmarshal([]byte(rpl), &reqPermStrArr)
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
