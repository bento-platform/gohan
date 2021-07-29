ARG BASE_IMAGE
FROM $BASE_IMAGE

# Maintainer
LABEL maintainer="Brennan Brouillette <brennan.brouillette@computationalgenomics.ca>"

# Alpine updates and executable dependencies
RUN apk update \
    && apk upgrade

# Copy pre-built executeables
ARG API_EXECUTABLE_FILENAME
ENV api_executable_filename=$API_EXECUTABLE_FILENAME
COPY $API_EXECUTABLE_FILENAME /app/

ARG GOHAN_API_INTERNAL_PORT
ENV GOHAN_API_INTERNAL_PORT=$GOHAN_API_INTERNAL_PORT
EXPOSE $GOHAN_API_INTERNAL_PORT

# Run
WORKDIR /app
ENTRYPOINT ./$api_executable_filename
