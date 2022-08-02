ARG BASE_IMAGE
ARG BASE_IMAGE_VERSION

FROM "${BASE_IMAGE}:${BASE_IMAGE_VERSION}"

# TODO: change USER
USER root
RUN apk update; \
	apk upgrade; \
	apk add --no-cache --virtual \
		autoconf \
		automake \
		bash \
		build-base \
		bzip2-dev \
		cargo \
		curl \
		curl-dev \
		gcc \
		git \
		libcurl \
		libressl-dev \
		libffi-dev \
		linux-headers \
		make \
		musl-dev \
		openssl-dev \
		python3-dev \
		postgresql-dev \
		postgresql-libs \
		perl \
		xz-dev \
		yaml-dev \
		zlib-dev
# RUN apk add --no-cache --virtual \
# 	build-base \
# 	cargo \
# 	gcc \
# 	git \
# 	libffi-dev \
# 	musl-dev \
# 	python3-dev \
# 	postgresql-dev \
# 	postgresql-libs

RUN mkdir /drs

WORKDIR /drs

ARG REPO
ARG BRANCH
ARG TAG

RUN git clone "${REPO}"

WORKDIR /drs/bento_drs

RUN git config --global user.email "you@example.com" && \
	git config --global user.name "Your Name" && \
	git fetch && \
    git checkout "${BRANCH}" && \
    git pull && \
	git checkout tags/${TAG}

RUN ["pip", "install", "gunicorn", "-r", "requirements.txt"]

# Security :
ARG HOST_USER_UID
ARG HOST_USER_GID
ARG OS_NAME
RUN if [ $OS_NAME != darwin ]; \
    then addgroup -S drsgroup -g $HOST_USER_GID; \
	else addgroup -S drsgroup; \
	fi && \
    adduser -S drsuser -u $HOST_USER_UID -G drsgroup

USER drsuser

# Run

# assume it is mounted
# COPY startup.sh ./startup.sh
EXPOSE 5000
CMD ["sh", "startup.sh"]
