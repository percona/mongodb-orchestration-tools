FROM alpine:3.7
COPY bin/mongodb-controller-* /usr/bin/
COPY bin/mongodb-watchdog-* /usr/bin/
