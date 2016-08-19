FROM alpine:3.4
MAINTAINER Wantedly, Inc. Infrastructure Team <dev@wantedly.com>

EXPOSE 8080

ENTRYPOINT ["/root/kubenetes-slack"]
