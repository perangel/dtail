# -- Build
FROM golang:1.11-alpine as build
LABEL maintainer="Angel Perez <perangel@gmail.com>"
RUN apk add --update git
COPY . /go/src/github.com/perangel/dtail
WORKDIR /go/src/github.com/perangel/dtail
RUN go get -u github.com/golang/dep/cmd/dep \
    && dep ensure -v
RUN go install

# -- Release
FROM alpine:3.9
COPY --from=build /go/bin/dtail /usr/bin/dtail
CMD ["/usr/bin/dtail"]
