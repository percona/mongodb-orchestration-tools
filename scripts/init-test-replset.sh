#!/bin/bash

set -x

tries=1
max_tries=10
sleep_secs=5

sleep $sleep_secs
while [ $tries -lt $max_tries ]; do
	/usr/bin/mongo --quiet \
		--port 65017 \
		--eval 'rs.initiate({
			_id: "rs",
			version: 1,
			members: [
				{ _id: 0, host: "127.0.0.1:65017", priority: 2 },
				{ _id: 1, host: "127.0.0.1:65018" },
				{ _id: 2, host: "127.0.0.1:65019" }
			]})'
	[ $? == 0 ] && break
	echo "# INFO: retrying in $sleep_secs secs (try $tries/$max_tries)"
	sleep $sleep_secs
	tries=$(($tries + 1))
done
if [ $tries -ge $max_tries ]; then
	echo "# ERROR: reached max tries, exiting"
	exit 1
fi

sleep $sleep_secs
tries=0
while [ $tries -lt $max_tries ]; do
	ISMASTER=$(/usr/bin/mongo --quiet \
		--port 65017 \
		--eval 'printjson(db.isMaster().ismaster)' 2>/dev/null)
	[ "$ISMASTER" == "true" ] && break
	echo "# INFO: retrying isMaster check in $sleep_secs secs (try $tries/$max_tries)"
	sleep $sleep_secs
	tries=$(($tries + 1))
done
if [ $tries -ge $max_tries ]; then
	echo "# ERROR: reached max tries, exiting"
	exit 1
fi

echo "# INFO: done init"
