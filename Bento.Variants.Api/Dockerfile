ARG BASE_IMAGE
FROM $BASE_IMAGE

# Maintainer
LABEL maintainer="Brennan Brouillette <brennan.brouillette@computationalgenomics.ca>"

# Alpine updates and executable dependencies
#RUN apk update
RUN apk update \
    && apk add --no-cache libstdc++ \
    && apk add --no-cache libintl
RUN apk upgrade

# Copy pre-built executeables
ARG SRC_DIR
COPY $SRC_DIR /app/

# Run
WORKDIR /app
ENTRYPOINT ./Bento.Variants.Api --urls http://0.0.0.0:5000
#dotnet Bento.Variants.Api.dll --urls http://0.0.0.0:5000
