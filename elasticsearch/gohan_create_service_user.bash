#!/bin/bash

# If set, use the local UID from outside the container (or default to 1001; 1000 is already created by the
# Elasticsearch container)
USER_ID=${GOHAN_UID:-1001}

echo "[gohan_elasticsearch] [/gohan_create_service_user.bash] using USER_ID=${USER_ID}"

# Add the user
useradd --shell /bin/bash -u "${USER_ID}" --non-unique -c "Bento container user" -m gohan_user
export HOME=/home/gohan_user
