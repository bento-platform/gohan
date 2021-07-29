# Makefile for Gohan

# import global variables
env ?= .env

include $(env)
export $(shell sed 's/=.*//' $(env))

export GOOS=linux
export GOARCH=amd64

# initialize services
init:
	# Gateway: 
	@# - DRS Authentication
	@htpasswd -cb gateway/drs.htpasswd ${GOHAN_DRS_BASIC_AUTH_USERNAME} ${GOHAN_DRS_BASIC_AUTH_PASSWORD} 
	
	@echo
	
	# Authorization:
	@# - API OPA policies 
	@echo Configuring authorzation policies
	@envsubst < ./etc/api.policy.rego.tpl > ./authorization/api.policy.rego



# Run
run-all:
	docker-compose -f docker-compose.yaml up -d --force-recreate

run-gateway:
	docker-compose -f docker-compose.yaml up -d --force-recreate gateway

# run-api:
# 	docker-compose -f docker-compose.yaml up -d --force-recreate api

# run-api-alpine:
# 	docker-compose -f docker-compose.yaml up -d --force-recreate api-alpine

run-api-go-alpine:
	docker-compose -f docker-compose.yaml up -d --force-recreate api-go-alpine

run-elasticsearch:
	docker-compose -f docker-compose.yaml up -d --force-recreate elasticsearch

run-kibana:
	docker-compose -f docker-compose.yaml up -d --force-recreate kibana

run-drs:
	docker-compose -f docker-compose.yaml up -d --force-recreate drs

run-authz:
	docker-compose -f docker-compose.yaml up -d --force-recreate authorization


# DEV --
run-dev-all:
	docker-compose -f docker-compose.dev.yaml up -d --force-recreate

# run-dev-api:
# 	docker-compose -f docker-compose.dev.yaml up -d --force-recreate api

# run-dev-api-alpine:
# 	docker-compose -f docker-compose.dev.yaml up -d --force-recreate api-alpine

run-dev-api-go-alpine:
	docker-compose -f docker-compose.dev.yaml up -d --force-recreate api-go-alpine

run-dev-elasticsearch:
	docker-compose -f docker-compose.dev.yaml up -d --force-recreate elasticsearch

run-dev-kibana:
	docker-compose -f docker-compose.dev.yaml up -d --force-recreate kibana


run-dev-drs:
	docker-compose -f docker-compose.dev.yaml up -d --force-recreate drs

run-dev-authz:
	docker-compose -f docker-compose.dev.yaml up -d --force-recreate authorization
# --


# Build
build-gateway: stop-gateway clean-gateway
	echo "-- Building Gateway Container --"
	docker-compose -f docker-compose.yaml build gateway

# build-api: stop-api clean-api
# 	echo "-- Building Api Binaries --"
# 	cd Gohan.Api/;
# 	dotnet clean; dotnet restore; dotnet publish -c Release --self-contained;
# 	cd ..
# 	echo "-- Building Api Container --"
# 	docker-compose -f docker-compose.yaml build api

# build-api-alpine: stop-api-alpine clean-api-alpine
# 	echo "-- Building Api-Alpine Binaries --"
# 	cd Gohan.Api/;
# 	dotnet clean; dotnet restore; dotnet publish -c ReleaseAlpine --self-contained;
# 	cd ..
# 	echo "-- Building Api-Alpine Container --"
# 	docker-compose -f docker-compose.yaml build api-alpine

build-api-go-alpine: stop-api-go-alpine clean-api-go-alpine
	@echo "-- Building Golang-Api-Alpine Binaries --"
	
	cd src/api && \
	export CGO_ENABLED=0 && \
	\
	go build -ldflags="-s -w" -o ../../bin/api_${GOOS}_${GOARCH} && \
	\
	cd ../.. 

	@echo "-- Building Golang-Api-Alpine Container --"
	cp bin/api_${GOOS}_${GOARCH} src/api
	docker-compose -f docker-compose.yaml build api-go-alpine
	rm src/api/api_${GOOS}_${GOARCH}

build-drs: stop-drs clean-drs
	@echo "-- Building DRS Container --"
	docker-compose -f docker-compose.yaml build drs

build-authz: stop-authz clean-authz
	@echo "-- Building Authorization Container --"
	docker-compose -f docker-compose.yaml build authorization


# Stop
stop-all:
	docker-compose -f docker-compose.yaml down

stop-gateway:
	docker-compose -f docker-compose.yaml stop gateway

# stop-api:
# 	docker-compose -f docker-compose.yaml stop api

# stop-api-alpine:
# 	docker-compose -f docker-compose.yaml stop api-alpine

stop-api-go-alpine:
	docker-compose -f docker-compose.yaml stop api-go-alpine

stop-drs:
	docker-compose -f docker-compose.yaml stop drs

stop-authz:
	docker-compose -f docker-compose.yaml stop authorization



# Clean up
clean-all: clean-api clean-api-alpine clean-gateway clean-drs

clean-gateway:
	docker rm ${GOHAN_GATEWAY_CONTAINER_NAME} --force; \
	docker rmi ${GOHAN_GATEWAY_IMAGE}:${GOHAN_GATEWAY_VERSION} --force;

# clean-api:
# 	docker rm ${GOHAN_API_CONTAINER_NAME} --force; \
# 	docker rmi ${GOHAN_API_IMAGE}:${GOHAN_API_VERSION} --force;

# clean-api-alpine:
# 	docker rm ${GOHAN_API_CONTAINER_NAME} --force; \
# 	docker rmi ${GOHAN_API_IMAGE}:${GOHAN_API_VERSION} --force;

clean-api-go-alpine:
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
clean-elastic-data:
	docker-compose -f docker-compose.yaml down
	sudo rm -rf ${GOHAN_ES_DATA_DIR}

## -- WARNING: DELETES ALL LOCAL DRS DATA
clean-drs-data:
	docker-compose -f docker-compose.yaml down
	sudo rm -rf ${GOHAN_DRS_DATA_DIR}



## Tests
test-api-dev: prepare-test-config
	# Run the tests
	dotnet test -c Debug Gohan.Tests/Gohan.Tests.csproj

# test-api-release:
# 	dotnet test -c Release Gohan.Tests/Gohan.Tests.csproj

prepare-test-config:
	# Prepare environment variables dynamically via a JSON file 
	# since xUnit doens't support loading env variables natively
	# (see `./Gohan.Tests/IntegrationTestFixture.cs`)
	envsubst < ./etc/appsettings.test.json.tpl > ./Gohan.Tests/appsettings.test.json

clean-tests:
	# Clean up
	rm ./Gohan.Tests/appsettings.test.json