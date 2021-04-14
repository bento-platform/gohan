package permissions

import input


default allowed = false

default authorized_username = "bob"

allowed = true {
    # retrieve authentication token parts
    username := input.username

    all([
        username == authorized_username
    ])
}