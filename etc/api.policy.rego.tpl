package permissions

import input


default allowed = false

default authorized_username = "bob"

allowed = true {
    # retrieve authentication token parts
    [authN_token_header, authN_token_payload, authN_token_signature] := io.jwt.decode(input.authN_token)

    authN_token_is_valid := io.jwt.verify_rs256(input.authN_token, input.authN_jwks)

    all([
        authN_token_is_valid == true,
        authorized_username == authN_token_payload.preferred_username
    ])
}