# Variables
BINARY_NAME=kidsnoter
DOCKER_IMAGE_NAME=ghcr.io/karolistamutis/kidsnoter
VERSION=$(shell git rev-parse --short HEAD)

# Go command
GO=go

# Targets
all: build

build:
	$(GO) build -ldflags "-X github.com/karolistamutis/kidsnoter/cmd.Version=$(VERSION)" -o $(BINARY_NAME) .

clean:
	$(GO) clean
	rm -f $(BINARY_NAME)

run:
	$(GO) run .

docker-build:
	docker build --build-arg VERSION=$(VERSION) -t $(DOCKER_IMAGE_NAME):$(VERSION) .
	docker tag $(DOCKER_IMAGE_NAME):$(VERSION) $(DOCKER_IMAGE_NAME):latest

docker-push:
	docker push $(DOCKER_IMAGE_NAME):$(VERSION)
	docker push $(DOCKER_IMAGE_NAME):latest

.PHONY: all build clean run docker-build docker-push