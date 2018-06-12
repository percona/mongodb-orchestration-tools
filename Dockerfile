FROM golang:1.10-alpine

WORKDIR /go/src/github.com/percona/dcos-mongo-tools
COPY . .

RUN apk update && apk add gcc git make
RUN make test
RUN make
