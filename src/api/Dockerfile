ARG BUILDER_BASE_IMAGE
ARG BASE_IMAGE

# Stage 1 - builder
FROM $BUILDER_BASE_IMAGE as builder

# Maintainer
LABEL maintainer="Brennan Brouillette <brennan.brouillette@computationalgenomics.ca>"

WORKDIR /build

COPY . .

# Alpine updates and executable dependencies
RUN apk update && \
    apk upgrade && \
    # For compressing the final binary :
    apk add git
    # Temporarily disabling: 
    # \
    # upx
    
# Install & build samtools from github repos (htslib needed first)
RUN apk --update add ncurses-dev && \
    apk add git build-base xz-dev zlib-dev bzip2-dev curl-dev && \
    \
    git clone https://github.com/samtools/htslib.git && \
    \
    cd htslib && \
    git submodule update --init --recursive && \
    \
    make && \
    make install && \
    \
    cd .. && \
    git clone https://github.com/samtools/samtools.git && \
    \
    cd samtools && \
    make && \
    make install && \
    \
    cd .. && \
    rm -rf htslib samtools
    
# Build gohan api
RUN go mod vendor && \
    go build -ldflags="-s -w" -o gohan_api 
    # Temporarily disabling: 
    #&& \
    # Compress the final binary :
    #upx --brute gohan_api


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

# Copy pre-built executable
COPY --from=builder /build/gohan_api .

# Copy static workflow files
COPY workflows/*.wdl /app/workflows/

# Use base image entrypoint to set up user & gosu exec the command below
# Run
CMD [ "/app/gohan_api" ]
