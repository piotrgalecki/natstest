#build stage
FROM golang:alpine AS builder
RUN apk add --no-cache git
WORKDIR /go/src/natstest-producer
COPY . .
RUN go get -d -v ./...
RUN go build -o /go/bin/natstest-producer -v ./...

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/natstest-producer /natstest-producer
ENTRYPOINT /natstest-producer
LABEL Name=natstest-producer Version=0.0.1
