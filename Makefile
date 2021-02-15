# Makefile for Bento Variants

# # import global variables
# env ?= .env

# include $(env)
# export $(shell sed 's/=.*//' $(env))



# Run
run:
	docker-compose -f docker-compose.yaml up -d

run-api:
	docker-compose -f docker-compose.yaml up -d api

run-elasticsearch:
	docker-compose -f docker-compose.yaml up -d elasticsearch

run-kibana:
	docker-compose -f docker-compose.yaml up -d kibana



# Build
build-api:
	docker-compose -f docker-compose.yaml build api


# Stop
stop:
	docker-compose -f docker-compose.yaml down


# Clean up
clean: clean-api

# TODO: use env variables for container versions
clean-api:
	docker rm variants-api --force; \
	docker rmi variants-api:latest --force;


## WARNING: DELETES ALL LOCAL ELASTICSEARCH DATA
clean-elastic-data:
	docker-compose -f docker-compose.yaml down
	rm -rf data/elasticsearch/nodes