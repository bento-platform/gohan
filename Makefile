# Makefile for Bento Variants

# import global variables
env ?= .env

include $(env)
export $(shell sed 's/=.*//' $(env))



# Run
run-all:
	docker-compose -f docker-compose.yaml up -d

run-gateway:
	docker-compose -f docker-compose.yaml up -d gateway

run-api:
	docker-compose -f docker-compose.yaml up -d api

run-api-alpine:
	docker-compose -f docker-compose.yaml up -d api-alpine

run-elasticsearch:
	docker-compose -f docker-compose.yaml up -d elasticsearch

run-kibana:
	docker-compose -f docker-compose.yaml up -d kibana



# Build
build-gateway: stop-gateway clean-gateway
	echo "-- Building Gateway Container --"
	docker-compose -f docker-compose.yaml build gateway

build-api: stop-api clean-api
	echo "-- Building Api Binaries --"
	cd Bento.Variants.Api/;
	dotnet clean; dotnet restore; dotnet publish -c Release --self-contained;
	cd ..
	echo "-- Building Api Container --"
	docker-compose -f docker-compose.yaml build api

build-api-alpine: stop-api-alpine clean-api-alpine
	echo "-- Building Api-Alpine Binaries --"
	cd Bento.Variants.Api/;
	dotnet clean; dotnet restore; dotnet publish -c ReleaseAlpine --self-contained;
	cd ..
	echo "-- Building Api-Alpine Container --"
	docker-compose -f docker-compose.yaml build api-alpine



# Stop
stop-all:
	docker-compose -f docker-compose.yaml down

stop-gateway:
	docker-compose -f docker-compose.yaml stop gateway

stop-api:
	docker-compose -f docker-compose.yaml stop api

stop-api-alpine:
	docker-compose -f docker-compose.yaml stop api-alpine



# Clean up
clean-all: clean-api clean-api-alpine clean-gateway

clean-gateway:
	docker rm ${BENTO_VARIANTS_GATEWAY_CONTAINER_NAME} --force; \
	docker rmi ${BENTO_VARIANTS_GATEWAY_IMAGE}:${BENTO_VARIANTS_GATEWAY_VERSION} --force;

clean-api:
	docker rm ${BENTO_VARIANTS_API_CONTAINER_NAME} --force; \
	docker rmi ${BENTO_VARIANTS_API_IMAGE}:${BENTO_VARIANTS_API_VERSION} --force;

clean-api-alpine:
	docker rm ${BENTO_VARIANTS_API_CONTAINER_NAME} --force; \
	docker rmi ${BENTO_VARIANTS_API_IMAGE}:${BENTO_VARIANTS_API_VERSION} --force;

## -- WARNING: DELETES ALL LOCAL ELASTICSEARCH DATA
clean-elastic-data:
	docker-compose -f docker-compose.yaml down
	sudo rm -rf ${BENTO_VARIANTS_ES_DATA_DIR}/nodes




## Tests
test-api-dev: prepare-test-config
	# Run the tests
	dotnet test -c Debug Bento.Variants.Tests/Bento.Variants.Tests.csproj

# test-api-release:
# 	dotnet test -c Release Bento.Variants.Tests/Bento.Variants.Tests.csproj

prepare-test-config:
	# Prepare environment variables dynamically via a JSON file 
	# since xUnit doens't support loading env variables natively
	# (see `./Bento.Variants.Tests/IntegrationTestFixture.cs`)
	envsubst < ./etc/appsettings.test.json.tpl > ./Bento.Variants.Tests/appsettings.test.json

clean-tests:
	# Clean up
	rm ./Bento.Variants.Tests/appsettings.test.json