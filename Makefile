OS=linux
ARCH=amd64
BUILDKIT_PROGRESS=plain
VER=$(shell date +%Y-%m-%d-%H%M%S)
IMG_VER=

.PHONY: lint test build

docker-image-dev:
	docker build --build-arg VER=${VER} --build-arg SKIP_TEST=true --build-arg SKIP_LINTER=true -t ghcr.io/theshamuel/gemini-proxy .

docker-image-prod:
	docker build --build-arg VER=${VER} -t ghcr.io/theshamuel/gemini-proxy:${IMG_VER} .

clean:
	- docker ps -a | grep -i "/bin/sh -c" | awk '{print $$1}' | xargs -n1 docker rm
	- docker images | grep -i "ghcr.io/theshamuel/gemini-proxy" | awk '{print $$3}' | xargs -n1 docker rmi
	- docker rmi $$(docker images -q -f dangling=true)

deploy:
	docker-compose up -d

lint:
	$(GOPATH)/bin/golangci-lint run --config .golangci.yml ./...

test:
	go test ./...

test-report:
	go test -coverprofile=cover.out ./...
	go tool cover --html=cover.out

build:
	go build -mod=vendor -o gemini-proxy ./app
