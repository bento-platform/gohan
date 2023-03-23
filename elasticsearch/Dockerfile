ARG BASE_IMAGE
ARG BASE_IMAGE_VERSION

FROM ${BASE_IMAGE}:${BASE_IMAGE_VERSION}

RUN apt-get update -y && \
    apt-get install -y bash gosu && \
    rm -rf /var/lib/apt/lists/*

COPY gohan_create_service_user.bash /gohan_create_service_user.bash
COPY gohan_entrypoint.bash /gohan_entrypoint.bash

ENTRYPOINT ["/bin/bash", "/gohan_entrypoint.bash"]
CMD ["/usr/local/bin/docker-entrypoint.sh"]
