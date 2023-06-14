debug: false
api:
  url: http://0.0.0.0:5000
  port: ${GOHAN_API_INTERNAL_PORT}
  vcfPath: "${GOHAN_API_CONTAINERIZED_VCF_PATH}"
  localVcfPath: "${GOHAN_API_VCF_PATH}"

elasticsearch:
  url: elasticsearch:${GOHAN_API_ES_PORT}
  username: "${GOHAN_ES_USERNAME}"
  password: "${GOHAN_ES_PASSWORD}"

drs:
  url: gohan-drs:${GOHAN_DRS_INTERNAL_PORT}
  username: "${GOHAN_DRS_BASIC_AUTH_USERNAME}"
  password: "${GOHAN_DRS_BASIC_AUTH_PASSWORD}"

authX:
  isAuthorizationEnabled: false
  oidcPublicJwksUrl: ""
  opaUrl: ""
  requiredHeadersCommaSep: ""