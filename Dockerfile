FROM xataz/alpine:3.7
MAINTAINER xataz <https://github.com/xataz>

ENV LISTEN=0.0.0.0:5000 \
    WEBROOT="" \
    PATH_FILE="/home" \
    HIDDEN_FILE=false \
    FORCE_URL="" \
    FORCE_SSL=false \
    SHARE_PATH="share" \
    UID=991 \
    GID=991

ADD . /app/gobrowser

RUN apk add --no-cache go \
		su-exec \
		tini \
		musl-dev \
	&& cd /app/gobrowser \
        && go build app.go \
	&& apk del --no-cache go musl-dev \
	&& chmod +x /app/gobrowser/startup 

EXPOSE 5000

CMD ["/app/gobrowser/startup"]
