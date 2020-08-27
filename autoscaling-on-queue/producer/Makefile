RELEASE_VERSION  =v0.10.3
SERVICE_NAME    ?=autoscaling-producer
DOCKER_USERNAME ?=$(DOCKER_USER)

.PHONY: all
all: help

.PHONY: tidy
tidy: ## Updates the go modules and vendors all dependencies 
	go mod tidy
	go mod vendor

.PHONY: run
run: tidy ## Runs uncompiled code
	NUMBER_OF_THREADS=10 go run main.go

.PHONY: image
image: tidy ## Builds and publishes docker image 
	docker build -t "$(DOCKER_USERNAME)/$(SERVICE_NAME):$(RELEASE_VERSION)" .
	docker push "$(DOCKER_USERNAME)/$(SERVICE_NAME):$(RELEASE_VERSION)"

.PHONY: exec
exec: ## Execute producer 
	kubectl apply -n data -f conf/producer.yaml
	kubectl wait -n data --for=condition=ready pod autoscaling-producer --timeout=120s
	kubectl logs autoscaling-producer -n data -f

.PHONY: stop
stop: ## Stod producer 
	kubectl delete -n data -f conf/producer.yaml

.PHONY: lint
lint: ## Lints the entire project 
	golangci-lint run --timeout=3m

.PHONY: tag
tag: ## Creates release tag 
	git tag $(RELEASE_VERSION)
	git push origin $(RELEASE_VERSION)

.PHONY: kafka 
kafka: ## Creates local Kafka server 
	docker-compose -f ./conf/local-kafka.yaml up -d

.PHONY: clean
clean: ## Cleans up generated files 
	go clean
	rm -fr ./bin
	rm -fr ./vendor

.PHONY: help
help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'