BIN := haproxy-connect
SOURCES := $(shell find . -name '*.go')
GOLIST := $(shell go list ./... | egrep -v "/state$$|/haproxy-consul-connect$$")

all: test bin

$(BIN): $(SOURCES)
	go build -o haproxy-connect -ldflags "-X main.BuildTime=`date -u '+%Y-%m-%dT%H:%M:%SZ'` -X main.GitHash=`git rev-parse HEAD` -X main.Version=`git describe --tags`"

bin: $(BIN)

test:
	HAPROXY=/usr/sbin/haproxy DATAPLANEAPI=/usr/local/bin/dataplaneapi go test -timeout 600s ${gobuild_args} ./...

test_wo_binaries:
	go test $(GOLIST)

check:
	go fmt ./...
	go vet ./...
	git diff --exit-code
travis: check bin test
