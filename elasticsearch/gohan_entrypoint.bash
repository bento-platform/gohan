#!/bin/bash

source /gohan_create_service_user.bash

# Fix permissions on Elasticsearch directories
# See https://www.elastic.co/guide/en/elasticsearch/reference/7.17/docker.html#_configuration_files_must_be_readable_by_the_elasticsearch_user
#  - except we use a different user!
chown -R gohan_user:gohan_user /usr/share/elasticsearch/config
chown -R gohan_user:gohan_user /usr/share/elasticsearch/data
chown -R gohan_user:gohan_user /usr/share/elasticsearch/logs

# Drop into gohan_user from root and execute the CMD specified for the image
exec gosu gohan_user "$@"
