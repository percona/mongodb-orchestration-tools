PLATFORM?=linux
GO_LDFLAGS?="-s -w"

all: bin/mongodb-healthcheck-$(PLATFORM) bin/mongodb-controller-$(PLATFORM) bin/mongodb-executor-$(PLATFORM) bin/mongodb-watchdog-$(PLATFORM)

$(GOPATH)/bin/glide:
	go get github.com/Masterminds/glide

vendor: $(GOPATH)/bin/glide glide.yaml glide.lock
	$(GOPATH)/bin/glide install

bin/mongodb-healthcheck-$(PLATFORM): vendor cmd/mongodb-healthcheck/main.go healthcheck/*.go common/*.go
	CGO_ENABLED=0 GOOS=$(PLATFORM) GOARCH=386 go build -ldflags=$(GO_LDFLAGS) -o bin/mongodb-healthcheck-$(PLATFORM) cmd/mongodb-healthcheck/main.go

bin/mongodb-controller-$(PLATFORM): vendor cmd/mongodb-controller/main.go controller/*.go controller/*/*.go common/*.go common/api/*.go
	CGO_ENABLED=0 GOOS=$(PLATFORM) GOARCH=386 go build -ldflags=$(GO_LDFLAGS) -o bin/mongodb-controller-$(PLATFORM) cmd/mongodb-controller/main.go

bin/mongodb-executor-$(PLATFORM): vendor cmd/mongodb-executor/main.go executor/*.go executor/*/*.go common/*.go
	CGO_ENABLED=0 GOOS=$(PLATFORM) GOARCH=386 go build -ldflags=$(GO_LDFLAGS) -o bin/mongodb-executor-$(PLATFORM) cmd/mongodb-executor/main.go

bin/mongodb-watchdog-$(PLATFORM): vendor cmd/mongodb-watchdog/main.go watchdog/*.go watchdog/*/*.go common/*.go common/api/*.go
	CGO_ENABLED=0 GOOS=$(PLATFORM) GOARCH=386 go build -ldflags=$(GO_LDFLAGS) -o bin/mongodb-watchdog-$(PLATFORM) cmd/mongodb-watchdog/main.go

test: vendor
	go test -v ./...

clean:
	rm -rf bin vendor 2>/dev/null || true
