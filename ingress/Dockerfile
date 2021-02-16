ARG BASE_IMAGE
ARG BASE_IMAGE_VERSION

FROM "${BASE_IMAGE}:${BASE_IMAGE_VERSION}"

LABEL Maintainer="Brennan Brouillette <brennan.brouillette@computationalgenomics.ca>"


# Main NGINX config
COPY nginx.conf /etc/nginx/conf.d/nginx.conf

# Dev certs
# TODO: integrate production certs
COPY ./certs/dev/variants_fullchain1.crt /etc/ssl/nginx/variants_fullchain1.crt
COPY ./certs/dev/variants_privkey1.key /etc/ssl/nginx/variants_privkey1.key


# Update and install dependencies
RUN apk update; \
    apk upgrade
