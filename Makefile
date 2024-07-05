BINARY_NAME=kidsnoter
DOCKER_IMAGE_NAME=ghcr.io/karolistamutis/kidsnoter
VERSION=$(shell git describe --tags --always --dirty)
PLATFORMS=linux/amd64,linux/arm64
GO=go

.PHONY: all build clean run docker-build docker-push

all: build

build:
	$(GO) build -ldflags "-X github.com/karolistamutis/kidsnoter/cmd.Version=$(VERSION)" -o $(BINARY_NAME) .

clean:
	$(GO) clean
	rm -f $(BINARY_NAME)

run:
	$(GO) run .

docker-build:
	docker buildx create --use --name multi-arch-builder || true
	docker buildx build --platform $(PLATFORMS) \
		--build-arg VERSION=$(VERSION) \
		-t $(DOCKER_IMAGE_NAME):$(VERSION) \
		-t $(DOCKER_IMAGE_NAME):latest \
		.

docker-push:
	@echo "$(GITHUB_PAT)" | docker login ghcr.io -u karolistamutis --password-stdin
	docker buildx create --use --name multi-arch-builder || true
	docker buildx build --platform $(PLATFORMS) \
		--build-arg VERSION=$(VERSION) \
		-t $(DOCKER_IMAGE_NAME):$(VERSION) \
		-t $(DOCKER_IMAGE_NAME):latest \
		--push \
		.