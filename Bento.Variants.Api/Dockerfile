ARG BASE_IMAGE
FROM $BASE_IMAGE

# Maintainer
LABEL maintainer="Brennan Brouillette <brennan.brouillette@computationalgenomics.ca>"

# Alpine Updates
RUN apk update
RUN apk upgrade

# Copy pre-built executeables
ARG SRC_DIR
COPY $SRC_DIR /app/

# Run
WORKDIR /app
ENTRYPOINT dotnet Bento.Variants.Api.dll --urls http://0.0.0.0:5000
