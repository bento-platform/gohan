ARG BUILDER_BASE_IMAGE
ARG BASE_IMAGE

# Stage 1 - builder
FROM $BUILDER_BASE_IMAGE as builder

# Maintainer
LABEL maintainer="Brennan Brouillette <brennan.brouillette@computationalgenomics.ca>"

WORKDIR /build

COPY . /build/

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
RUN CGO_ENABLED=0 && \
    go build -ldflags="-s -w" -o gohan_api 
    # Temporarily disabling: 
    #&& \
    # Compress the final binary :
    #upx --brute gohan_api


# Stage two - executioner
FROM $BASE_IMAGE

# Alpine updates
RUN apk update && \
    apk upgrade && \
    # Dependencies :
    apk add \
    # - for healthchecks
    curl \
    # - for tabix
    xz-dev zlib-dev bzip2-dev

# Security :
ARG HOST_USER_UID
ARG HOST_USER_GID
ARG OS_NAME
RUN if [ $OS_NAME != darwin ]; \
    then addgroup -S apigroup -g $HOST_USER_GID || true; \
    else addgroup -S apigroup || true; \
    fi && \
    adduser -S apiuser -u $HOST_USER_UID -G apigroup || true
    # note: ' || true' ignores possible minor errors

USER apiuser

# Copy pre-built executables
COPY --from=builder /build/gohan_api /app/
COPY --from=builder /usr/local/bin/tabix /usr/local/bin/

# Copy static workflow files
COPY workflows/*.wdl /app/workflows/

# Run
ARG GOHAN_API_INTERNAL_PORT
ENV GOHAN_API_INTERNAL_PORT=$GOHAN_API_INTERNAL_PORT
EXPOSE $GOHAN_API_INTERNAL_PORT
ENTRYPOINT [ "/app/gohan_api" ]
