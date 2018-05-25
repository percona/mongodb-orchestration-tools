PLATFORM?=linux
GIT_COMMIT?=$(shell git rev-parse HEAD)
GIT_BRANCH?=$(shell git rev-parse --abbrev-ref HEAD)
GO_LDFLAGS?=-s -w
GO_LDFLAGS_FULL="${GO_LDFLAGS} -X main.GitCommit=${GIT_COMMIT} -X main.GitBranch=${GIT_BRANCH}"
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
ifeq ($(TEST_CODECOV),true)
	TEST_GO_EXTRA=-race -coverprofile=cover.out -covermode=atomic
endif

all: bin/mongodb-healthcheck-$(PLATFORM) bin/mongodb-controller-$(PLATFORM) bin/mongodb-executor-$(PLATFORM) bin/mongodb-watchdog-$(PLATFORM)

$(GOPATH)/bin/glide:
	go get github.com/Masterminds/glide

vendor: $(GOPATH)/bin/glide glide.yaml glide.lock
	$(GOPATH)/bin/glide install --strip-vendor

bin/mongodb-healthcheck-$(PLATFORM): vendor cmd/mongodb-healthcheck/main.go healthcheck/*.go common/*.go common/*/*.go common/*/*/*.go
	CGO_ENABLED=0 GOCACHE=$(GOCACHE) GOOS=$(PLATFORM) GOARCH=386 go build -ldflags=$(GO_LDFLAGS_FULL) -o bin/mongodb-healthcheck-$(PLATFORM) cmd/mongodb-healthcheck/main.go

bin/mongodb-controller-$(PLATFORM): vendor cmd/mongodb-controller/main.go controller/*.go controller/*/*.go common/*.go common/*/*.go common/*/*/*.go
	CGO_ENABLED=0 GOCACHE=$(GOCACHE) GOOS=$(PLATFORM) GOARCH=386 go build -ldflags=$(GO_LDFLAGS_FULL) -o bin/mongodb-controller-$(PLATFORM) cmd/mongodb-controller/main.go

bin/mongodb-executor-$(PLATFORM): vendor cmd/mongodb-executor/main.go executor/*.go executor/*/*.go common/*.go common/*/*.go common/*/*/*.go
	CGO_ENABLED=0 GOCACHE=$(GOCACHE) GOOS=$(PLATFORM) GOARCH=386 go build -ldflags=$(GO_LDFLAGS_FULL) -o bin/mongodb-executor-$(PLATFORM) cmd/mongodb-executor/main.go

bin/mongodb-watchdog-$(PLATFORM): vendor cmd/mongodb-watchdog/main.go watchdog/*.go watchdog/*/*.go watchdog/*/*/*.go common/*.go common/*/*.go common/*/*/*.go
	CGO_ENABLED=0 GOCACHE=$(GOCACHE) GOOS=$(PLATFORM) GOARCH=386 go build -ldflags=$(GO_LDFLAGS_FULL) -o bin/mongodb-watchdog-$(PLATFORM) cmd/mongodb-watchdog/main.go

test: vendor
	GOCACHE=$(GOCACHE) ENABLE_MONGODB_TESTS=$(ENABLE_MONGODB_TESTS) go test -v $(TEST_GO_EXTRA) ./...

test/test-mongod.key:
	openssl rand -base64 512 >test/test-mongod.key
	chown $(TEST_MONGODB_DOCKER_UID):0 test/test-mongod.key
	chmod 0600 test/test-mongod.key

test/test-rootCA.crt: test/ssl/rootCA.crt
	cp test/ssl/rootCA.crt test/test-rootCA.crt
	chown $(TEST_MONGODB_DOCKER_UID):0 test/test-rootCA.crt
	chmod 0600 test/test-rootCA.crt

test/test-mongod.pem: test/ssl/mongodb.pem
	cp test/ssl/mongodb.pem test/test-mongod.pem
	chown $(TEST_MONGODB_DOCKER_UID):0 test/test-mongod.pem
	chmod 0600 test/test-mongod.pem

test-full-keys: test/test-mongod.key test/test-rootCA.crt test/test-mongod.pem

test-full-prepare: test/test-mongod.key test/test-rootCA.crt test/test-mongod.pem
	TEST_RS_NAME=$(TEST_RS_NAME) \
	TEST_PSMDB_VERSION=$(TEST_PSMDB_VERSION) \
	TEST_ADMIN_USER=$(TEST_ADMIN_USER) \
	TEST_ADMIN_PASSWORD=$(TEST_ADMIN_PASSWORD) \
	TEST_PRIMARY_PORT=$(TEST_PRIMARY_PORT) \
	TEST_SECONDARY1_PORT=$(TEST_SECONDARY1_PORT) \
	TEST_SECONDARY2_PORT=$(TEST_SECONDARY2_PORT) \
	docker-compose up -d
	test/init-test-replset-wait.sh

test-full-clean:
	docker-compose down

test-full: vendor
	ENABLE_MONGODB_TESTS=true \
	TEST_RS_NAME=$(TEST_RS_NAME) \
	TEST_ADMIN_USER=$(TEST_ADMIN_USER) \
	TEST_ADMIN_PASSWORD=$(TEST_ADMIN_PASSWORD) \
	TEST_PRIMARY_PORT=$(TEST_PRIMARY_PORT) \
	TEST_SECONDARY1_PORT=$(TEST_SECONDARY1_PORT) \
	TEST_SECONDARY2_PORT=$(TEST_SECONDARY2_PORT) \
	GOCACHE=$(GOCACHE) go test -v $(TEST_GO_EXTRA) ./...

clean:
	rm -rf bin coverage.txt test/test-*.* vendor 2>/dev/null || true
