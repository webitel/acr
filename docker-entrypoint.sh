#!/bin/bash
set -e

echo 'Webitel ACR '$VERSION

if [ "$LOGLEVEL" ]; then
	sed -i 's/LOGLEVEL/'$LOGLEVEL'/g' /acr/conf/config.json
else
	sed -i 's/LOGLEVEL/warn/g' /acr/conf/config.json
fi

if [ "$LOGSTASH_ENABLE" ]; then
	sed -i 's/LOGSTASH_ENABLE/'$LOGSTASH_ENABLE'/g' /acr/conf/config.json
else
	sed -i 's/LOGSTASH_ENABLE/false/g' /acr/conf/config.json
fi

if [ "$LOGSTASH_HOST" ]; then
	sed -i 's/LOGSTASH_HOST/'$LOGSTASH_HOST'/g' /acr/conf/config.json
fi

if [ "$MONGODB_HOST" ]; then
	sed -i 's/MONGODB_HOST/'$MONGODB_HOST'/g' /acr/conf/config.json
fi

exec node app.js