ARG BUILDER_BASE_IMAGE

# Stage 1 - builder
FROM $BUILDER_BASE_IMAGE as builder

# Maintainer
LABEL maintainer="Brennan Brouillette <brennan.brouillette@computationalgenomics.ca>"

WORKDIR /app

COPY . .
    
# Build gohan api
RUN go mod vendor

# Debian updates
#  - tabix for indexing VCFs
#  - other base dependencies provided by the base image
RUN apt-get update -y && \
    apt-get upgrade -y && \
    apt-get install -y tabix && \
    rm -rf /var/lib/apt/lists/*

# Install air for hot-reload
RUN go get -u github.com/cosmtrek/air

# Copy static workflow files
COPY workflows/*.wdl /app/workflows/

# Use base image entrypoint to set up user & gosu exec the command below
# Run
CMD [ "air" ]
