ACR
====

Advanced Call Router - route calls that match configured rules.

## Default ports

`10022/tcp` - WebSocket Server port

## Environment Variables

The CDR image uses several environment variables

### Server variables

`SSL` - enable https (default: false)

`CONSOLE_HOST` - Webitel Console host or IP

`CONSOLE_PASSWORD` - Webitel Console password

`MONGODB_HOST` - MongoDB host or IP

`FS_HOST` - FreeSWITCH host or IP

`TOKEN_KEY` - application token key for storing session

### Logs

`LOGLEVEL` - log level (default: warn)

`LOGSTASH_ENABLE` - send logs to Logstash Server (default: false)

`LOGSTASH_HOST` - Logstash host or IP


## Supported Docker versions

This image is officially supported on Docker version `1.3.2` and newest.

## User Feedback

### Issues
If you have any problems with or questions about this image, please contact us through a [GitHub issue](https://github.com/webitel/core/issues).
