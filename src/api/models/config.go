package models

type Config struct {
	Api struct {
		Port    string `envconfig:"GOHAN_API_INTERNAL_PORT"`
		VcfPath string `envconfig:"GOHAN_API_VCF_PATH"`
	}
	Elasticsearch struct {
		Url      string `envconfig:"GOHAN_ES_URL"`
		Username string `envconfig:"GOHAN_ES_USERNAME"`
		Password string `envconfig:"GOHAN_ES_PASSWORD"`
	}
	Drs struct {
		Url      string `envconfig:"GOHAN_DRS_URL"`
		Username string `envconfig:"GOHAN_DRS_BASIC_AUTH_USERNAME"`
		Password string `envconfig:"GOHAN_DRS_BASIC_AUTH_PASSWORD"`
	}
	AuthX struct {
		IsAuthorizationEnabled  bool   `envconfig:"GOHAN_AUTHZ_ENABLED"`
		OidcPublicJwksUrl       string `envconfig:"GOHAN_PUBLIC_AUTHN_JWKS_URL"`
		OpaUrl                  string `envconfig:"GOHAN_PRIVATE_AUTHZ_URL"`
		RequiredHeadersCommaSep string `envconfig:"GOHAN_AUTHZ_REQHEADS"`
	}
}
