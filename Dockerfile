# vim:set ft=dockerfile:
FROM scratch
LABEL maintainer="Vitaly Kovalyshyn"

ENV VERSION
ENV WEBITEL_MAJOR 3
ENV WEBITEL_REPO_BASE https://github.com/webitel

ADD /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ADD src/acr.go /

CMD ["/acr.go"]
