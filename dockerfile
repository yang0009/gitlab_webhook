ARG IMAGE=alpine:3.12

FROM golang:1.16-alpine as builder
WORKDIR ${GOPATH}/src/gitlab_webhook
ENV  GOPROXY https://goproxy.io,direct
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/bin/eventserver

FROM ${IMAGE}

RUN  sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories \
      &&  apk add --no-cache bash openssl curl
COPY --from=builder /usr/bin/eventserver /usr/bin/
ENTRYPOINT ["/usr/bin/gitlab_webhook"]