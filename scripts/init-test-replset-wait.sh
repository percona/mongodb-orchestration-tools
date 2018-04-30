#!/bin/bash

tries=1
max_tries=30
sleep_secs=5

while [ $tries -lt $max_tries ]; do
	docker-compose ps init 2>/dev/null | grep -q 'Exit 0'
	if [ $? == 0 ]; then
		docker-compose logs init 2>/dev/null | tail -1 | grep -q 'INFO: done init'
		[ $? == 0 ] && break
	fi
        echo "# INFO: retrying check in $sleep_secs secs (try $tries/$max_tries)"
	sleep $sleep_secs
	tries=$(($tries + 1))
done
if [ $tries -ge $max_tries ]; then
        echo "# ERROR: reached max tries, exiting"
        exit 1
fi
