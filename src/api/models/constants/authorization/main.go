package authorization

type PermissionVerb string
type PermissionNoun string
type PermissionLevel string

const (
	QUERY    PermissionVerb = "query"
	DOWNLOAD PermissionVerb = "download"
	CREATE   PermissionVerb = "create"
	EDIT     PermissionVerb = "edit"
	DELETE   PermissionVerb = "delete"
	INGEST   PermissionVerb = "ingest"
	ANALYZE  PermissionVerb = "analyze"
	EXPORT   PermissionVerb = "export"
)

const (
	DATA PermissionNoun = "data"
)
