package models

type Config struct {
	Debug          bool   `yaml:"debug" envconfig:"GOHAN_DEBUG"`
	SemVer         string `yaml:"semver" envconfig:"GOHAN_SEMVER"`
	ServiceContact string `yaml:"serviceContact" envconfig:"GOHAN_SERVICE_CONTACT"`
	ValidateSSL    bool   `yaml:"validateSSL" envconfig:"BENTO_VALIDATE_SSL"`

	Api struct {
		Url  string `yaml:"url" envconfig:"GOHAN_PUBLIC_URL"`
		Port string `yaml:"port" envconfig:"GOHAN_API_INTERNAL_PORT"`
		// TODO: rm and use Drop-Box calls
		VcfPath string `yaml:"vcfPath" envconfig:"GOHAN_API_VCF_PATH"`
		// TODO: rm and use Drop-Box calls
		LocalVcfPath                   string `yaml:"localVcfPath" envconfig:"GOHAN_API_VCF_PATH"`
		BulkIndexingCap                int    `yaml:"BulkIndexingCap" envconfig:"GOHAN_API_BULK_INDEXING_CAP"`
		FileProcessingConcurrencyLevel int    `yaml:"fileProcessingConcurrencyLevel" envconfig:"GOHAN_API_FILE_PROC_CONC_LVL"`
		LineProcessingConcurrencyLevel int    `yaml:"lineProcessingConcurrencyLevel" envconfig:"GOHAN_API_LINE_PROC_CONC_LVL"`
		GtfPath                        string `yaml:"gtfPath" envconfig:"GOHAN_API_GTF_PATH"`
	} `yaml:"api"`

	Elasticsearch struct {
		Url      string `yaml:"url" envconfig:"GOHAN_ES_URL"`
		Username string `yaml:"username" envconfig:"GOHAN_ES_USERNAME"`
		Password string `yaml:"password" envconfig:"GOHAN_ES_PASSWORD"`
	} `yaml:"elasticsearch"`

	Drs struct {
		Url      string `yaml:"url" envconfig:"GOHAN_DRS_URL"`
		Username string `yaml:"username" envconfig:"GOHAN_DRS_BASIC_AUTH_USERNAME"`
		Password string `yaml:"password" envconfig:"GOHAN_DRS_BASIC_AUTH_PASSWORD"`
	} `yaml:"drs"`

	DropBox struct {
		Url string `yaml:"url" envconfig:"GOHAN_DROP_BOX_URL"`
	} `yaml:"dropbox"`

	AuthX struct {
		IsAuthorizationEnabled  bool   `yaml:"isAuthorizationEnabled" envconfig:"GOHAN_AUTHZ_ENABLED"`
		OidcPublicJwksUrl       string `yaml:"oidcPublicJwksUrl" envconfig:"GOHAN_PUBLIC_AUTHN_JWKS_URL"`
		OpaUrl                  string `yaml:"opaUrl" envconfig:"GOHAN_PRIVATE_AUTHZ_URL"`
		RequiredHeadersCommaSep string `yaml:"requiredHeadersCommaSep" envconfig:"GOHAN_AUTHZ_REQHEADS"`
	} `yaml:"authx"`
}
