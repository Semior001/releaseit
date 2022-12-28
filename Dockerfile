FROM golang:1.19-alpine
LABEL maintainer="Semior <ura2178@gmail.com>"

ENV CGO_ENABLED=0

LABEL maintainer="Semior <ura2178@gmail.com>"

RUN apk add --no-cache --update git bash curl tzdata && \
    cp /usr/share/zoneinfo/Europe/Moscow /etc/localtime && \
    rm -rf /var/cache/apk/*

WORKDIR /srv

COPY ./app /srv/app
COPY ./go.mod /srv/go.mod
COPY ./go.sum /srv/go.sum

COPY ./.git/ /srv/.git

RUN \
    export version="$(git describe --tags --long)" && \
    echo $version && \
    go build -o /go/build/app -ldflags "-X 'main.version=${version}' -s -w" /srv/app

ENTRYPOINT ["/go/build/app"]