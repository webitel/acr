Advanced Call Router (ACR)
====

Advanced Call Router - route calls that match configured rules.

## Default ports

`10025/tcp` and `10026/tcp`

## Environment Variables

The ACR image uses several environment variables

### Server variables

`MONGODB_HOST` - MongoDB host or IP

### Logs

`LOGLEVEL` - log level (default: warn)

`LOGSTASH_ENABLE` - send logs to Logstash Server (default: false)

`LOGSTASH_HOST` - Logstash host or IP


## Supported Docker versions

This image is officially supported on Docker version `1.5` and newest.

## User Feedback

### Issues
If you have any problems with or questions about this image, please contact us through a [GitHub issue](https://github.com/webitel/acr/issues).
