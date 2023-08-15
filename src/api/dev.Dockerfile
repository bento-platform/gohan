ARG BUILDER_BASE_IMAGE

# Stage 1 - builder
FROM $BUILDER_BASE_IMAGE as builder

# Maintainer
LABEL maintainer="Brennan Brouillette <brennan.brouillette@computationalgenomics.ca>"

WORKDIR /app

# Debian updates
#  - tabix for indexing VCFs
#  - other base dependencies provided by the base image
RUN apt-get update -y && \
    apt-get upgrade -y && \
    apt-get install -y tabix && \
    rm -rf /var/lib/apt/lists/*

RUN go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./
RUN go mod download && go mod vendor

# Copy static workflow files
COPY workflows/*.wdl /app/workflows/

# Repository mounted to the container
# WORKDIR /app/repo/src/api
WORKDIR /gohan-api/src/api


CMD [ "air" ]
