#!/bin/bash

cd /gohan-api || exit

# Create bento_user and home
source /create_service_user.bash

# Create dev build directory
mkdir -p src/api/tmp

# Set permissions / groups
chown -R bento_user:bento_user ./
chown -R bento_user:bento_user /app
chmod -R o-rwx src/api/tmp

# Drop into bento_user from root and execute the CMD specified for the image
exec gosu bento_user "$@"
