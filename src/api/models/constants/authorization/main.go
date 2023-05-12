package authorization

import "gohan/api/models/constants"

const (
	QUERY    constants.PermissionVerb = "query"
	DOWNLOAD constants.PermissionVerb = "download"
	VIEW     constants.PermissionVerb = "view"
	CREATE   constants.PermissionVerb = "create"
	EDIT     constants.PermissionVerb = "edit"
	DELEVE   constants.PermissionVerb = "delete"
	INGEST   constants.PermissionVerb = "ingest"
	ANALYZE  constants.PermissionVerb = "analyze"
	EXPORT   constants.PermissionVerb = "export"
)

const (
	DATA constants.PermissionNoun = "data"
)
