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
    apk add git upx
    
RUN CGO_ENABLED=0 && \
    go build -ldflags="-s -w" -o gohan_api && \
    upx --brute gohan_api


# Stage two - executioner
FROM $BASE_IMAGE

# # Alpine updates
RUN apk update && \
    apk upgrade

# Copy pre-built executeables
COPY --from=builder /build/gohan_api /app/

# Copy static workflow files
COPY workflows/*.wdl /app/workflows/

# Run
ARG GOHAN_API_INTERNAL_PORT
ENV GOHAN_API_INTERNAL_PORT=$GOHAN_API_INTERNAL_PORT
EXPOSE $GOHAN_API_INTERNAL_PORT
ENTRYPOINT [ "/app/gohan_api" ]
