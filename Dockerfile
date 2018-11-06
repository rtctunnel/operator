FROM golang:1.11-alpine as builder
RUN apk --no-cache add git gcc musl-dev

ENV GO111MODULE=on
RUN mkdir -p /go/src/github.com/rtctunnel/rtctunnel
WORKDIR /go/src/github.com/rtctunnel/rtctunnel

COPY go.mod .
COPY go.sum . 
RUN go mod download

COPY . .
RUN go build -v -ldflags '-extldflags "-static"' \
    -o /bin/operator ./cmd/operator

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root
COPY --from=0 /bin/operator /bin/operator
CMD ["/bin/operator", "--bind-addr=:9100"]

EXPOSE 9100