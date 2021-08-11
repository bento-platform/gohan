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

run-dev-all:
	docker-compose -f docker-compose.dev.yaml up -d --force-recreate

run-dev-%:
	docker-compose -f docker-compose.dev.yaml up -d --force-recreate $*

run-%:
	docker-compose -f docker-compose.yaml up -d --force-recreate $*



# Build
build-gateway: stop-gateway clean-gateway
	echo "-- Building Gateway Container --"
	docker-compose -f docker-compose.yaml build gateway

build-api-go-alpine-binaries:
	@echo "-- Building Golang-Api-Alpine Binaries --"
	
	cd src/api && \
	export CGO_ENABLED=0 && \
	\
	go build -ldflags="-s -w" -o ../../bin/api_${GOOS}_${GOARCH} && \
	\
	cd ../.. && \
	upx --brute bin/api_${GOOS}_${GOARCH}

build-api-go-alpine-container: stop-api-go-alpine clean-api-go-alpine build-api-go-alpine-binaries
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

stop-%:
	docker-compose -f docker-compose.yaml stop $*



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
	@# Run the tests
	@# dotnet test -c Debug Gohan.Tests/Gohan.Tests.csproj
	go test tests/integration/...

# test-api-release:
# 	dotnet test -c Release Gohan.Tests/Gohan.Tests.csproj

prepare-test-config:
	@# Prepare environment variables dynamically via a JSON file 
	@# since xUnit doens't support loading env variables natively
	@# (see `./Gohan.Tests/IntegrationTestFixture.cs`)
	@# envsubst < ./etc/appsettings.test.json.tpl > ./Gohan.Tests/appsettings.test.json
	envsubst < ./etc/test.config.yml.tpl > ./src/tests/common/test.config.yml

clean-tests:
	@# Clean up
	rm ./Gohan.Tests/appsettings.test.json