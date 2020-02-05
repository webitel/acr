# vim:set ft=dockerfile:
FROM golang:1.13

COPY src /go/src/github.com/webitel/acr/src
WORKDIR /go/src/github.com/webitel/acr/src/

ENV GO111MODULE=on
RUN go mod download
RUN GIT_COMMIT=$(git log -1 --format='%h %ci') && \
        CGO_ENABLED=0 GOOS=linux go build -ldflags "-X 'github.com/webitel/acr/src/model.BuildNumber=$GIT_COMMIT'" \
         -a -o acr .

FROM scratch
LABEL maintainer="Vitaly Kovalyshyn"

ENV VERSION
ENV WEBITEL_MAJOR 3
ENV WEBITEL_REPO_BASE https://github.com/webitel

WORKDIR /
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=0 /go/src/github.com/webitel/acr/src/acr .
COPY conf/config.json .

EXPOSE 10030
CMD ["./acr", "-c", "./config.json"]
