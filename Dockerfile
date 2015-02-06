FROM node:slim
MAINTAINER Ihor Navrotskyi "support@webitel.com"

COPY src /acr

WORKDIR /acr
RUN npm install && npm cache clear

EXPOSE 10025 10026 10027 10028

ENTRYPOINT ["node", "app.js"]