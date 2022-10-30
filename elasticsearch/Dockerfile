ARG BASE_IMAGE
ARG BASE_IMAGE_VERSION

FROM "${BASE_IMAGE}:${BASE_IMAGE_VERSION}"

ARG HOST_USER_UID
ARG HOST_USER_GID
ARG OS_NAME
RUN if [ $OS_NAME != darwin ]; \
    then groupadd --system elasticsearchgroup -g $HOST_USER_GID || true; \
    else groupadd --system elasticsearchgroup || true; \
    fi && \
    usermod -a -G elasticsearchgroup elasticsearch || true
    # note: ' || true' ignores possible minor errors

USER elasticsearch