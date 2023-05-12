package authorization

import (
	"encoding/json"
	mauthz "gohan/api/models/authorization"
)

type PermissionRequestDto struct {
	RequestedResource   mauthz.ResourceEverything // TODO: ResourceSpecific
	RequiredPermissions mauthz.PermissionsList
}

func (p *PermissionRequestDto) MarshalJSON() ([]byte, error) {
	// customize the request serialized format
	// as permissions representation is specific
	rpl, _ := json.Marshal(&p.RequiredPermissions)
	rrl, _ := json.Marshal(&p.RequestedResource)
	res := map[string]interface{}{
		"requested_resource":   string(rrl),
		"required_permissions": string(rpl),
	}

	return json.Marshal(&res)

}
