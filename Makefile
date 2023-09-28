# GIT_COMMIT = $(shell git log -1 --pretty=format:%h)
# BUILD_DATE = $(shell date +%Y%m%d%H%M%S)
# IMAGE_TAG = ${GIT_COMMIT}-${BUILD_DATE}
IMAGE_TAG = $(shell git log -1 --pretty=format:%cd-%h --date=format:%Y%m%d%H%M%S)

IMAGE = lunettes

# Image URL to use all building/pushing image targets
IMAGE_FULL ?= ${IMAGE}:${IMAGE_TAG}

build: fmt vet
	go build -o bin/aggregator cmd/aggregator/main.go

build-aggregator: fmt vet
	go build -o bin/aggregator cmd/aggregator/main.go

# Run tests
test: fmt vet
	mkdir -p tmp
	go test -v -count 1 ./pkg/... ./cmd/... -coverprofile tmp/cover.out


# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Build docker image
docker-build:
	docker build . -f Dockerfile -t ${IMAGE_FULL}

# Push the docker image
docker-push:
	docker push ${IMAGE_FULL}

docker: docker-build docker-push
	echo "image pushed"
