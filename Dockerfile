FROM golang:1.27.3-alpine AS build

ARG version=unknown

RUN apk update && \
		apk add git

WORKDIR /go/src

COPY go.mod go.sum ./
RUN go mod download && \
		go mod verify

COPY . .

RUN module_path=$(go list -m) && \
	go build \
		-o /go/bin/ilias-sync-webdav \
		-ldflags "-X ${module_path}/cmd/version.version=$version" \
		.

FROM alpine:3.24

RUN addgroup -S app && adduser -S -G app app

COPY --from=build /go/bin/ilias-sync-webdav /usr/local/bin/ilias-sync-webdav

WORKDIR /var/lib/ilias-sync-webdav
RUN chown app:app /var/lib/ilias-sync-webdav

USER app

ENTRYPOINT ["/usr/local/bin/ilias-sync-webdav"]
