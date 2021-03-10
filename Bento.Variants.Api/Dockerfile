ARG BASE_IMAGE
FROM $BASE_IMAGE

# Maintainer
LABEL maintainer="Brennan Brouillette <brennan.brouillette@computationalgenomics.ca>"

# Alpine updates and executable dependencies
ARG IS_ALPINE
RUN if [ "$IS_ALPINE" = "true" ] ; then \
    apk update \
    && apk add --no-cache libstdc++ \
    && apk add --no-cache libintl \
    && apk upgrade; \
    else echo "Not Building Alpine" ; fi
#RUN apk update
# RUN apk update \
#     && apk add --no-cache libstdc++ \
#     && apk add --no-cache libintl
# RUN apk upgrade

# Copy pre-built executeables
ARG SRC_DIR
COPY $SRC_DIR /app/

ARG INTERNAL_PORT
ENV internal_port=$INTERNAL_PORT

EXPOSE $INTERNAL_PORT

# Run
WORKDIR /app
ENTRYPOINT ./Bento.Variants.Api --urls http://0.0.0.0:$internal_port
