ARG BASE_IMAGE
ARG BASE_IMAGE_VERSION

FROM "${BASE_IMAGE}:${BASE_IMAGE_VERSION}"

ARG HOST_USER_UID
ARG HOST_USER_GID
RUN groupadd --system elasticsearchgroup -g $HOST_USER_GID && \
    usermod -a -G elasticsearchgroup elasticsearch

USER elasticsearch