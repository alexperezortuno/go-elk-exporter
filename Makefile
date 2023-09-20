BINARY=elastic-query-export
VERSION=1.1.0
BUILD_DIR=./build
BUILD_TIME=`date +%FT%T%z`
GOX_OSARCH="darwin/amd64 darwin/arm64 linux/386 linux/amd64 windows/386 windows/amd64"

.PHONY: default
default: build

.PHONY: clean
clean:
	rm -rf ./build

.PHONY: build
build:
	CGO_ENABLED=0 \
	go build -a -o ${BUILD_DIR}/${BINARY} *.go

.PHONY: build-version
build-version:
	CGO_ENABLED=0 \
	go build -a -o ${BUILD_DIR}/${BINARY}-${VERSION} *.go

.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 \
	GOARCH=amd64 \
	GOOS=linux \
	go build -ldflags "-X main.Version=${VERSION}" -a -o ${BUILD_DIR}/${BINARY}-${VERSION} *.go

.PHONY: build-gox
build-gox:
	gox -ldflags "-X main.Version=${VERSION}" -osarch=${GOX_OSARCH} -output="/build/${VERSION}/{{.Dir}}_{{.OS}}_{{.Arch}}"

.PHONY: deps
deps:
	dep ensure;

.PHONY: test
test:
	go test
