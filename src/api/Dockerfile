ARG BASE_IMAGE

FROM $BASE_IMAGE

# Maintainer
LABEL maintainer="Brennan Brouillette <brennan.brouillette@computationalgenomics.ca>"

WORKDIR /app

COPY . .
    
# Build gohan api
RUN go mod vendor && \
    go build -ldflags="-s -w" -o gohan_api

# Debian updates
#  - tabix for indexing VCFs
#  - other base dependencies provided by the base image
RUN apt-get update -y && \
    apt-get upgrade -y && \
    apt-get install -y tabix && \
    rm -rf /var/lib/apt/lists/*

# Copy static workflow files
COPY workflows/*.wdl /app/workflows/

# Use base image entrypoint to set up user & gosu exec the command below
# Run
CMD [ "/app/gohan_api" ]
