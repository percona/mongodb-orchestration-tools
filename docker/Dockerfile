FROM busybox:latest

COPY bin/mongodb-controller-* /usr/local/bin/
COPY bin/mongodb-watchdog-* /usr/local/bin/

RUN adduser -D dcos-mongo-tools
USER dcos-mongo-tools
