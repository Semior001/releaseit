FROM golang:1.19-alpine AS builder
LABEL maintainer="Semior <ura2178@gmail.com>"

ENV CGO_ENABLED=0

LABEL maintainer="Semior <ura2178@gmail.com>"

WORKDIR /srv

RUN apk add --no-cache --update git bash curl tzdata && \
    cp /usr/share/zoneinfo/Europe/Moscow /etc/localtime && \
    rm -rf /var/cache/apk/*

COPY ./app /srv/app
COPY ./go.mod /srv/go.mod
COPY ./go.sum /srv/go.sum

COPY ./.git/ /srv/.git

RUN \
    export version="$(git describe --tags --long)" && \
    echo "version: $version" && \
    go build -o /go/build/app -ldflags "-X 'main.version=${version}' -s -w" /srv/app

FROM alpine:3.14
LABEL maintainer="Semior <ura2178@gmail.com>"

RUN apk add --no-cache --update tzdata && \
    cp /usr/share/zoneinfo/Europe/Moscow /etc/localtime && \
    rm -rf /var/cache/apk/*

COPY --from=builder /go/build/app /usr/bin/releaseit

ENTRYPOINT ["/usr/bin/releaseit"]