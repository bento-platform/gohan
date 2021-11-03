debug: ${GOHAN_DEBUG}
api:
  url: "${GOHAN_PUBLIC_URL}"
  port: "${GOHAN_API_INTERNAL_PORT}"
  vcfPath: "${GOHAN_API_CONTAINERIZED_VCF_PATH}"
  localVcfPath: "${GOHAN_API_VCF_PATH}"

elasticsearch:
  url: "${GOHAN_ES_PUBLIC_URL}"
  username: "${GOHAN_ES_USERNAME}"
  password: "${GOHAN_ES_PASSWORD}"

drs:
  url: "${GOHAN_DRS_PUBLIC_URL}"
  username: "${GOHAN_DRS_BASIC_AUTH_USERNAME}"
  password: "${GOHAN_DRS_BASIC_AUTH_PASSWORD}"

authX:
  isAuthorizationEnabled: ${GOHAN_API_AUTHZ_ENABLED}
  oidcPublicJwksUrl: "${GOHAN_PUBLIC_AUTHN_JWKS_URL}"
  opaUrl: "${GOHAN_PRIVATE_AUTHZ_URL}"
  requiredHeadersCommaSep: "${GOHAN_API_AUTHZ_REQHEADS}"