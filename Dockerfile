FROM node:slim
MAINTAINER Vitaly Kovalyshyn "v.kovalyshyn@webitel.com"

ENV VERSION 3.1
ENV WEBITEL_MAJOR 3.1
ENV WEBITEL_REPO_BASE https://github.com/webitel

COPY src /acr
COPY docker-entrypoint.sh /

WORKDIR /acr

RUN npm install && npm cache clear

ENTRYPOINT ["/docker-entrypoint.sh"]