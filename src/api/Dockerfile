ARG BUILDER_BASE_IMAGE
ARG BASE_IMAGE

# Stage 1 - builder
FROM $BUILDER_BASE_IMAGE as builder

# Maintainer
LABEL maintainer="Brennan Brouillette <brennan.brouillette@computationalgenomics.ca>"

WORKDIR /build

COPY . .
    
# Build gohan api
RUN go mod vendor && \
    go build -ldflags="-s -w" -o gohan_api


# Stage two - executioner
FROM $BASE_IMAGE

# Debian updates
#  - tabix for indexing VCFs
#  - other base dependencies provided by the base image
RUN apt-get update -y && \
    apt-get upgrade -y && \
    apt-get install -y tabix && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy pre-built executable from builder stage
COPY --from=builder /build/gohan_api .

# Copy static workflow files
COPY workflows/*.wdl /app/workflows/

# Use base image entrypoint to set up user & gosu exec the command below
# Run
CMD [ "/app/gohan_api" ]
