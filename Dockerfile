FROM alpine

COPY . /var/tgateway

WORKDIR /var/tgateway

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --update go git libc-dev \
    && export GOPATH=`pwd` \
    && export GOBIN=/usr/bin \
    && go get \
    && go install \
    && echo $GOPATH \
    && rm -r /var/tgateway \
    && apk del --purge go git libc-dev

WORKDIR /var/config
