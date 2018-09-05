NAME?=dcos-mongo-tools
PLATFORM?=linux
BASE_DIR?=$(shell readlink -f $(CURDIR))
VERSION?=$(shell grep -oP '"\d+\.\d+\.\d+(-\S+)?"' version.go | tr -d \")
GIT_COMMIT?=$(shell git rev-parse HEAD)
GIT_BRANCH?=$(shell git rev-parse --abbrev-ref HEAD)
GITHUB_REPO?=percona/$(NAME)
RELEASE_CACHE_DIR?=/tmp/$(NAME)_release.cache

DOCKERHUB_REPO?=percona/$(NAME)
DOCKERHUB_TAG?=$(VERSION)
ifneq ($(GIT_BRANCH), master)
	DOCKERHUB_TAG=$(VERSION)-$(GIT_BRANCH)
endif

GO_VERSION?=1.10
GO_VERSION_MAJ_MIN=$(shell echo $(GO_VERSION) | cut -d. -f1-2)
GO_LDFLAGS?=-s -w
GO_LDFLAGS_FULL="${GO_LDFLAGS} -X main.GitCommit=${GIT_COMMIT} -X main.GitBranch=${GIT_BRANCH}"
GO_TEST_PATH?=./...
GOCACHE?=

ENABLE_MONGODB_TESTS?=false
TEST_PSMDB_VERSION?=latest
TEST_RS_NAME?=rs
TEST_MONGODB_DOCKER_UID?=1001
TEST_ADMIN_USER?=admin
TEST_ADMIN_PASSWORD?=123456
TEST_PRIMARY_PORT?=65217
TEST_SECONDARY1_PORT?=65218
TEST_SECONDARY2_PORT?=65219

TEST_CODECOV?=false
TEST_GO_EXTRA?=
ifeq ($(TEST_CODECOV), true)
	TEST_GO_EXTRA=-coverprofile=cover.out -covermode=atomic
endif

all: bin/mongodb-healthcheck-$(PLATFORM) bin/mongodb-controller-$(PLATFORM) bin/mongodb-executor-$(PLATFORM) bin/mongodb-watchdog-$(PLATFORM)

$(GOPATH)/bin/glide:
	go get github.com/Masterminds/glide

vendor: $(GOPATH)/bin/glide glide.yaml glide.lock
	$(GOPATH)/bin/glide install --strip-vendor

bin/mongodb-healthcheck-$(PLATFORM): vendor cmd/mongodb-healthcheck/main.go healthcheck/*.go internal/*.go internal/*/*.go internal/*/*/*.go
	CGO_ENABLED=0 GOCACHE=$(GOCACHE) GOOS=$(PLATFORM) GOARCH=386 go build -ldflags=$(GO_LDFLAGS_FULL) -o bin/mongodb-healthcheck-$(PLATFORM) cmd/mongodb-healthcheck/main.go

bin/mongodb-controller-$(PLATFORM): vendor cmd/mongodb-controller/main.go controller/*.go controller/*/*.go controller/*/*/*.go internal/*.go internal/*/*.go internal/*/*/*.go
	CGO_ENABLED=0 GOCACHE=$(GOCACHE) GOOS=$(PLATFORM) GOARCH=386 go build -ldflags=$(GO_LDFLAGS_FULL) -o bin/mongodb-controller-$(PLATFORM) cmd/mongodb-controller/main.go

bin/mongodb-executor-$(PLATFORM): vendor cmd/mongodb-executor/main.go executor/*.go executor/*/*.go internal/*.go internal/*/*.go internal/*/*/*.go
	CGO_ENABLED=0 GOCACHE=$(GOCACHE) GOOS=$(PLATFORM) GOARCH=386 go build -ldflags=$(GO_LDFLAGS_FULL) -o bin/mongodb-executor-$(PLATFORM) cmd/mongodb-executor/main.go

bin/mongodb-watchdog-$(PLATFORM): vendor cmd/mongodb-watchdog/main.go watchdog/*.go watchdog/*/*.go internal/*.go internal/*/*.go internal/*/*/*.go
	CGO_ENABLED=0 GOCACHE=$(GOCACHE) GOOS=$(PLATFORM) GOARCH=386 go build -ldflags=$(GO_LDFLAGS_FULL) -o bin/mongodb-watchdog-$(PLATFORM) cmd/mongodb-watchdog/main.go

test: vendor
	GOCACHE=$(GOCACHE) ENABLE_MONGODB_TESTS=$(ENABLE_MONGODB_TESTS) go test -v $(TEST_GO_EXTRA) $(GO_TEST_PATH)

test-race: vendor
	GOCACHE=$(GOCACHE) ENABLE_MONGODB_TESTS=$(ENABLE_MONGODB_TESTS) go test -v -race $(TEST_GO_EXTRA) $(GO_TEST_PATH)

test-full-prepare:
	TEST_RS_NAME=$(TEST_RS_NAME) \
	TEST_PSMDB_VERSION=$(TEST_PSMDB_VERSION) \
	TEST_ADMIN_USER=$(TEST_ADMIN_USER) \
	TEST_ADMIN_PASSWORD=$(TEST_ADMIN_PASSWORD) \
	TEST_PRIMARY_PORT=$(TEST_PRIMARY_PORT) \
	TEST_SECONDARY1_PORT=$(TEST_SECONDARY1_PORT) \
	TEST_SECONDARY2_PORT=$(TEST_SECONDARY2_PORT) \
	docker-compose up -d --force-recreate
	docker/test/init-test-replset-wait.sh

test-full-clean:
	docker-compose down --volumes

test-full: vendor
	ENABLE_MONGODB_TESTS=true \
	TEST_RS_NAME=$(TEST_RS_NAME) \
	TEST_ADMIN_USER=$(TEST_ADMIN_USER) \
	TEST_ADMIN_PASSWORD=$(TEST_ADMIN_PASSWORD) \
	TEST_PRIMARY_PORT=$(TEST_PRIMARY_PORT) \
	TEST_SECONDARY1_PORT=$(TEST_SECONDARY1_PORT) \
	TEST_SECONDARY2_PORT=$(TEST_SECONDARY2_PORT) \
	GOCACHE=$(GOCACHE) go test -v -race $(TEST_GO_EXTRA) $(GO_TEST_PATH)
ifeq ($(TEST_CODECOV), true)
	curl -s https://codecov.io/bash | bash -s - -t ${CODECOV_TOKEN}
endif

release: clean
	docker build --build-arg GOLANG_DOCKERHUB_TAG=$(GO_VERSION_MAJ_MIN)-stretch -t $(NAME)_release -f docker/Dockerfile.release .
	docker run --rm --network=host \
	-v $(BASE_DIR)/bin:/go/src/github.com/$(GITHUB_REPO)/bin \
	-v $(RELEASE_CACHE_DIR)/glide:/root/.glide/cache \
	-e ENABLE_MONGODB_TESTS=$(ENABLE_MONGODB_TESTS) \
	-e TEST_CODECOV=$(TEST_CODECOV) \
	-e CODECOV_TOKEN=$(CODECOV_TOKEN) \
	-e TEST_RS_NAME=$(TEST_RS_NAME) \
	-e TEST_ADMIN_USER=$(TEST_ADMIN_USER) \
	-e TEST_ADMIN_PASSWORD=$(TEST_ADMIN_PASSWORD) \
	-e TEST_PRIMARY_PORT=$(TEST_PRIMARY_PORT) \
	-e TEST_SECONDARY1_PORT=$(TEST_SECONDARY1_PORT) \
	-e TEST_SECONDARY2_PORT=$(TEST_SECONDARY2_PORT) \
	-i $(NAME)_release

release-clean:
	rm -rf $(RELEASE_CACHE_DIR) 2>/dev/null
	docker rmi -f $(NAME)_release 2>/dev/null
	docker rmi -f $(NAME):$(DOCKERHUB_TAG) 2>/dev/null

docker-build: release
	docker build -t $(NAME):$(DOCKERHUB_TAG) -f docker/Dockerfile
	docker run --rm -i $(NAME):$(DOCKERHUB_TAG) mongodb-controller-$(PLATFORM) --version
	docker run --rm -i $(NAME):$(DOCKERHUB_TAG) mongodb-watchdog-$(PLATFORM) --version

docker-push:
	docker tag $(NAME):$(DOCKERHUB_TAG) $(DOCKERHUB_REPO):$(DOCKERHUB_TAG)
	docker push $(DOCKERHUB_REPO):$(DOCKERHUB_TAG)
ifeq ($(GIT_BRANCH), master)
	docker tag $(NAME):$(DOCKERHUB_TAG) $(DOCKERHUB_REPO):latest
	docker push $(DOCKERHUB_REPO):latest
endif

clean:
	rm -rf bin cover.out vendor 2>/dev/null || true
