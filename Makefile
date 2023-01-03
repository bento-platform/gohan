# Makefile for Gohan

#>>>
# set default shell
#<<<
SHELL = bash

# export host user IDs for more secure
# containerization and volume mounting
export HOST_USER_UID=$(shell id -u)
export HOST_USER_GID=$(shell id -g)
export OS_NAME=$(shell uname -s | tr A-Z a-z)

export GOOS=${OS_NAME}
export GOARCH=$(shell if [ "$(uname -m)" == "aarch64" ]; then echo arm64; else echo $(uname -m); fi | tr A-Z a-z)

# import global variables
env ?= .env

include $(env)
export $(shell sed 's/=.*//' $(env))


# initialize services
init:
	@# Gateway: 
	@# - DRS Authentication
	@htpasswd -cb gateway/drs.htpasswd ${GOHAN_DRS_BASIC_AUTH_USERNAME} ${GOHAN_DRS_BASIC_AUTH_PASSWORD} 
	
	@echo
	
	@# Authorization:
	@# - API OPA policies 
	@echo Configuring authorzation policies
	@envsubst < ./etc/api.policy.rego.tpl > ./authorization/api.policy.rego
	
	@$(MAKE) init-data-dirs
	@$(MAKE) init-networks

init-dev:
	@$(MAKE) init
	@$(MAKE) init-vendor

init-networks:
	docker network create ${GOHAN_DOCKER_NET} &

init-vendor:
	@echo "Initializing Go Module Vendor"
	cd src/api && go mod tidy && go mod vendor

init-data-dirs:
	mkdir -p ${GOHAN_API_DRS_BRIDGE_HOST_DIR}
	chown -R ${HOST_USER_UID}:${HOST_USER_GID} ${GOHAN_API_DRS_BRIDGE_HOST_DIR}
	chmod -R 770 ${GOHAN_API_DRS_BRIDGE_HOST_DIR}

	mkdir -p ${GOHAN_DRS_DATA_DIR}
	chown -R ${HOST_USER_UID}:${HOST_USER_GID} ${GOHAN_DRS_DATA_DIR}
	chmod -R 770 ${GOHAN_DRS_DATA_DIR}

	mkdir -p ${GOHAN_ES_DATA_DIR}
	chown -R ${HOST_USER_UID}:${HOST_USER_GID} ${GOHAN_ES_DATA_DIR}
	chmod -R 770 ${GOHAN_ES_DATA_DIR}

	@# tmp:
	@# (setup for when gohan needs to preprocess vcf's at ingestion time):
	mkdir -p ${GOHAN_API_VCF_PATH}
	mkdir -p ${GOHAN_API_VCF_PATH}/tmp
	chown -R ${HOST_USER_UID}:${HOST_USER_GID} ${GOHAN_API_VCF_PATH}
	chmod -R 770 ${GOHAN_API_VCF_PATH}/tmp

	mkdir -p ${GOHAN_API_GTF_PATH}
	mkdir -p ${GOHAN_API_GTF_PATH}/tmp
	chown -R ${HOST_USER_UID}:${HOST_USER_GID} ${GOHAN_API_GTF_PATH}
	chmod -R 770 ${GOHAN_API_GTF_PATH}/tmp
	
	@echo ".. done!"


# Run
run-all:
	docker-compose -f docker-compose.yaml up -d

run-dev-all:
	docker-compose -f docker-compose.dev.yaml up -d --force-recreate

run-dev-%:
	docker-compose -f docker-compose.dev.yaml up -d --force-recreate $*

run-%:
	docker-compose -f docker-compose.yaml up -d $*


# Build
build-gateway: stop-gateway clean-gateway
	echo "-- Building Gateway Container --"
	docker-compose -f docker-compose.yaml build gateway

build-api-local-binaries:
	@echo "-- Building Golang-Api-Alpine Binaries --"
	
	cd src/api && \
	export CGO_ENABLED=0 && \
	\
	go build -ldflags="-s -w" -o ../../bin/api_${GOOS}_${GOARCH} && \
	\
	cd ../..
	# Temporarily disabling:
	# && \
	# upx --brute bin/api_${GOOS}_${GOARCH}

build-api: stop-api clean-api
	@echo "-- Building Golang-Api-Alpine Container --"
	docker-compose -f docker-compose.yaml build api

build-drs: stop-drs clean-drs
	@echo "-- Building DRS Container --"
	docker-compose -f docker-compose.yaml build drs

build-authz: stop-authz clean-authz
	@echo "-- Building Authorization Container --"
	docker-compose -f docker-compose.yaml build authorization


# Stop
stop-all:
	docker-compose -f docker-compose.yaml down

stop-%:
	docker-compose -f docker-compose.yaml stop $*



# Clean up
clean-all: clean-networks clean-api clean-gateway clean-drs

clean-networks:
	docker network remove ${GOHAN_DOCKER_NET} &

clean-gateway:
	docker rm ${GOHAN_GATEWAY_CONTAINER_NAME} --force; \
	docker rmi ${GOHAN_GATEWAY_IMAGE}:${GOHAN_GATEWAY_VERSION} --force;

clean-api:
	rm -f bin/api_${GOOS}_${GOARCH}
	docker rm ${GOHAN_API_CONTAINER_NAME} --force; \
	docker rmi ${GOHAN_API_IMAGE}:${GOHAN_API_VERSION} --force;

clean-drs:
	docker rm ${GOHAN_DRS_CONTAINER_NAME} --force; \
	docker rmi ${GOHAN_DRS_IMAGE}:${GOHAN_DRS_VERSION} --force;

clean-authz:
	docker rm ${GOHAN_AUTHZ_OPA_CONTAINER_NAME} --force; \
	docker rmi ${GOHAN_AUTHZ_OPA_IMAGE}:${GOHAN_AUTHZ_OPA_IMAGE_VERSION} --force;



## -- WARNING: DELETES ALL LOCAL ELASTICSEARCH DATA
clean-all-data:
	@read -p "Are you sure you want to clean out all data? (yes/no) : " answer; \
	if [ "$$answer" == "yes" ]; then \
		echo "-- Cleaning! --" ; \
		docker-compose -f docker-compose.yaml down && \
		sudo rm -rf ${GOHAN_DATA_ROOT} ; \
		echo "-- Done! --" ; \
	else \
		echo "-- Skipping.. --" ; \
	fi
clean-elastic-data:
	@read -p "Are you sure you want to clean out all elasticsearch data? (yes/no) : " answer; \
	if [ "$$answer" == "yes" ]; then \
		echo "-- Cleaning! --" ; \
		docker-compose -f docker-compose.yaml down && \
		sudo rm -rf ${GOHAN_ES_DATA_DIR} ; \
		echo "-- Done! --" ; \
	else \
		echo "-- Skipping.. --" ; \
	fi

## -- WARNING: DELETES ALL LOCAL DRS DATA
clean-drs-data:
	@read -p "Are you sure you want to clean out all drs data? (yes/no) : " answer; \
	if [ "$$answer" == "yes" ]; then \
		echo "-- Cleaning! --" ; \
		docker-compose -f docker-compose.yaml down && \
		sudo rm -rf ${GOHAN_DRS_DATA_DIR} ; \
		echo "-- Done! --" ; \
	else \
		echo "-- Skipping.. --" ; \
	fi

## -- WARNING: DELETES ALL LOCAL API-DRS-BRIDGE DATA
clean-api-drs-bridge-data:
	@read -p "Are you sure you want to clean out all api-drs-bridge data? (yes/no) : " answer; \
	if [ "$$answer" == "yes" ]; then \
		echo "-- Cleaning! --" ; \
		docker-compose -f docker-compose.yaml down && \
		sudo rm -rf ${GOHAN_API_DRS_BRIDGE_HOST_DIR} ; \
		echo "-- Done! --" ; \
	else \
		echo "-- Skipping.. --" ; \
	fi

	

## Tests
test-api-dev: prepare-test-config
	@# Run the tests
	cd src/api && \
	go clean -cache && \
	go test ./tests/integration/... -v

prepare-test-config:
	@# Prepare environment variables dynamically via a JSON file 
	@# since xUnit doens't support loading env variables natively
	envsubst < ./etc/test.config.yml.tpl > ./src/api/tests/common/test.config.yml

clean-tests:
	@# Clean up
	rm ./src/api/tests/common/test.config.yml