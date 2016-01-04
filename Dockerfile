FROM xataz/alpine:3.3
MAINTAINER xataz <https://github.com/xataz>

ADD . /gobrowser

RUN apk add -U go@community && \
	cd /gobrowser && go build app.go && \
	apk del go && rm -rf /var/cache/apk/*

WORKDIR /gobrowser

CMD ["./app"]
