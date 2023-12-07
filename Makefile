VERSION :=v0.0.19

RELEASE_DIR = dist
IMPORT_PATH = github.com/vearne/consul-cache

BUILD_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date +%Y%m%d%H%M%S)
GITTAG = `git log -1 --pretty=format:"%H"`
LDFLAGS = -ldflags "-s -w -X ${IMPORT_PATH}/internal/consts.GitTag=${GITTAG} -X ${IMPORT_PATH}/internal/consts.BuildTime=${BUILD_TIME} -X ${IMPORT_PATH}/internal/consts.Version=${VERSION}"

TAG = ${VERSION}-${BUILD_TIME}-${BUILD_COMMIT}
#TAG = ${VERSION}
IMAGE_FETCHER = woshiaotian/consul-fetcher:${TAG}
IMAGE_CACHE = woshiaotian/consul-cache:${TAG}


.PHONY: clean
clean: ## Remove release binaries
	rm -rf ${RELEASE_DIR}

build-dirs: clean
	mkdir -p ${RELEASE_DIR}

.PHONY: git-tag
git-tag:
	git tag $(VERSION)
	git push origin $(VERSION)

.PHONY: build
build: build-dirs git-tag
	env GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${RELEASE_DIR}/consul-fetcher ./cmd/fetcher
	env GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${RELEASE_DIR}/consul-cache ./cmd/cache

.PHONY: image
image: build
	# fetcher
	docker build -f ./dockerfile/Dockerfile.fetcher \
 		--rm --no-cache -t ${IMAGE_FETCHER} .

	docker push ${IMAGE_FETCHER}
	# cache
	docker build -f ./dockerfile/Dockerfile.cache \
 		--rm --no-cache -t ${IMAGE_CACHE} .
	docker push ${IMAGE_CACHE}




