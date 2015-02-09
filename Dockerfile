FROM node:slim
MAINTAINER Vitaly Kovalyshyn "v.kovalyshyn@webitel.com"

ENV VERSION 3.0.5021

COPY src /acr
COPY docker-entrypoint.sh /

WORKDIR /acr

ENTRYPOINT ["/docker-entrypoint.sh"]