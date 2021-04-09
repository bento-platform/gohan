ARG BASE_IMAGE
ARG BASE_IMAGE_VERSION

FROM "${BASE_IMAGE}:${BASE_IMAGE_VERSION}"

# TODO: change USER
USER root

RUN apk add --no-cache --virtual \
	build-base \
	cargo \
	gcc \
	git \
	libffi-dev \
	musl-dev \
	python3-dev \
	postgresql-dev \
	postgresql-libs

RUN mkdir /drs

WORKDIR /drs

ARG REPO
ARG BRANCH
ARG TAG

RUN git clone "${REPO}"

WORKDIR /drs/chord_drs

RUN git fetch && \
    git checkout "${BRANCH}" && \
    git pull ${REPO}
#    git pull $(echo "${REPO} refs/tags/${TAG}:refs/tags/${TAG}") && \

RUN ["pip", "install", "-r", "requirements.txt"]

# Run
WORKDIR /drs/chord_drs/chord_drs

COPY startup.sh ./startup.sh

EXPOSE 5000

CMD ["sh", "startup.sh"]
