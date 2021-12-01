ARG BASE_IMAGE
ARG BASE_IMAGE_VERSION

FROM "${BASE_IMAGE}:${BASE_IMAGE_VERSION}"

LABEL Maintainer="Brennan Brouillette <brennan.brouillette@computationalgenomics.ca>"


# Main NGINX config
COPY nginx.conf /etc/nginx/conf.d/nginx.conf

# DRS Basic Auth
ARG DRS_BASIC_AUTH_FILE_PATH
COPY drs.htpasswd $DRS_BASIC_AUTH_FILE_PATH

# Dev certs
# TODO: integrate production certs
COPY ./certs/dev/gohan_fullchain1.crt /etc/ssl/nginx/gohan_fullchain1.crt
COPY ./certs/dev/gohan_privkey1.key /etc/ssl/nginx/gohan_privkey1.key
COPY ./certs/dev/es_gohan_fullchain1.crt /etc/ssl/nginx/es_gohan_fullchain1.crt
COPY ./certs/dev/es_gohan_privkey1.key /etc/ssl/nginx/es_gohan_privkey1.key


# Update and install dependencies
RUN apk update; \
    apk upgrade
