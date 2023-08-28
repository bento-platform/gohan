ARG BUILDER_BASE_IMAGE

# Stage 1 - builder
FROM $BUILDER_BASE_IMAGE as builder

LABEL org.opencontainers.image.description="Local development image for Bento Gohan."
LABEL devcontainer.metadata='[{ \
  "customizations": { \
    "vscode": { \
      "extensions": ["golang.go", "eamodio.gitlens"], \
      "settings": {"workspaceFolder": "/gohan-api"} \
    } \
  } \
}]'

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

# Repository mounted to the container
WORKDIR /gohan-api/src/api

CMD [ "air" ]
