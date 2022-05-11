package serviceInfo

import "fmt"

type ServiceInfo string

var (
	SERVICE_NAME        ServiceInfo = "Bento Gohan Service"
	SERVICE_WELCOME     ServiceInfo = "Welcome to the next generation Gohan v2 API using Golang!"
	SERVICE_DESCRIPTION ServiceInfo = "Gohan Variant service for a Bento platform node."

	SERVICE_ARTIFACT    ServiceInfo = "gohan"
	SERVICE_TYPE_NO_VER ServiceInfo = ServiceInfo(fmt.Sprintf("ca.c3g.bento:%s", SERVICE_ARTIFACT))
	SERVICE_ID          ServiceInfo = SERVICE_TYPE_NO_VER
)
