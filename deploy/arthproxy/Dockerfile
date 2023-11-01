FROM alpine:3.18

RUN apk add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone


ENV WORKDIR                 /app
ENV APP_NAME                arthproxy
WORKDIR                     $WORKDIR
ADD ./$APP_NAME             $WORKDIR/$APP_NAME
RUN chmod +xr $WORKDIR/$APP_NAME


CMD ["./$APP_NAME"]

