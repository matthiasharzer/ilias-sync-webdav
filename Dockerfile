FROM golang:1.27.2-alpine3.24 AS build

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
		-o /go/bin/<tool-name> \
		-ldflags "-X ${module_path}/cmd/version.version=$version" \
		.

FROM alpine:3.24

RUN addgroup -S app && adduser -S -G app app

COPY --from=build /go/bin/<tool-name> /usr/local/bin/<tool-name>

WORKDIR /var/lib/<tool-name>
RUN chown app:app /var/lib/<tool-name>

USER app

ENTRYPOINT ["/usr/local/bin/<tool-name>"]
