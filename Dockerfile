FROM node:slim
MAINTAINER Vitaly Kovalyshyn "v.kovalyshyn@webitel.com"

ENV VERSION 3.0.5024

COPY src /acr
COPY docker-entrypoint.sh /

WORKDIR /acr

RUN npm install && npm cache clear

ENTRYPOINT ["/docker-entrypoint.sh"]