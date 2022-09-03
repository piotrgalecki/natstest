#build stage
FROM golang:alpine AS builder
RUN apk add --no-cache git
WORKDIR /go/src/natstest-consumer
COPY . .
RUN go get -d -v ./...
RUN go build -o /go/bin/natstest-consumer -v ./...

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/natstest-consumer /natstest-consumer
ENTRYPOINT /natstest-consumer
LABEL Name=natstest-consumer Version=0.0.1
