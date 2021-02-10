# Makefile for Bento Variants

# # import global variables
# env ?= .env

# include $(env)
# export $(shell sed 's/=.*//' $(env))



# Run
run-dev:
	docker-compose up -d

run-dev-api:
	docker-compose up -d api

run-dev-elasticsearch:
	docker-compose up -d elasticsearch

run-dev-kibana:
	docker-compose up -d kibana



# Build
build-dev-api:
	docker-compose build api



# Clean up
clean-dev: clean-dev-api

# TODO: use env variables for container versions
clean-dev-api:
	docker rm variants-api --force; \
	docker rmi variants-api:latest --force;
