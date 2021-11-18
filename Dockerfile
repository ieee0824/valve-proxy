FROM golang:alpine AS install


RUN set -e \
	&& apk update \
	&& apk add git \
	&& go get github.com/ieee0824/valve-proxy \
	&& go install github.com/ieee0824/valve-proxy/cmd/vproxy

FROM alpine

COPY --from=install /go/bin/vproxy /bin/vproxy

EXPOSE 3128

ENTRYPOINT [ "vproxy" ]

CMD [ "-port", "3128", "-d", "10G", "-u", "10G" ]