FROM busybox:1.27

COPY bin/mongodb-controller-* /usr/bin/
COPY bin/mongodb-watchdog-* /usr/bin/

USER nobody
