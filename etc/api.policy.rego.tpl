package permissions

import input

now := time.now_ns()/1000000000

default allowed = false

default iss = "${AUTHN_SERVICE_PUBLIC_URL}/auth/realms/${AUTHN_REALM}"
default aud = "${AUTHN_CLIENT_ID}"

default full_authn_pk=`-----BEGIN PUBLIC KEY-----
${AUTHN_PUBLIC_KEY}
-----END PUBLIC KEY-----`

#default full_authz_pk=`-----BEGIN PUBLIC KEY-----
#${AUTHZ_PUBLIC_KEY}
#-----END PUBLIC KEY-----`


##default authz_jwks=`${AUTHZ_JWKS}`

allowed = true {
    # retrieve authentication token parts
    [authN_token_header, authN_token_payload, authN_token_signature] := io.jwt.decode(input.authNToken)

    # retrieve authorization token parts
    [authZ_token_header, authZ_token_payload, authZ_token_signature] := io.jwt.decode(input.authZToken)

    # retrieve rotated authZ jwks from the arbiter
    rotated_authz_jwks := input.authZjwks

    # Verify authentication token signature
    authN_token_is_valid := io.jwt.verify_rs256(input.authNToken, full_authn_pk)

    # Verify authentication token signature
    # (disabled until vault key rotation accommodated for)
    authZ_token_is_valid := io.jwt.verify_rs256(input.authZToken, rotated_authz_jwks)


    all([
        # Authentication
        authN_token_is_valid == true, 
        authN_token_payload.aud == aud, 
        authN_token_payload.iss == iss, 
        authN_token_payload.iat < now,
        
        # Authorization
        authZ_token_is_valid == true,
        ##authZ_token_payload.aud == aud, 
        ##authZ_token_payload.iss == iss, 
        authZ_token_payload.iat < now,
    ])
}
