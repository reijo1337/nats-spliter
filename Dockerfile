FROM golang:1.13 AS stage
ENV CGO_ENABLED 0
WORKDIR /nats-spliter
COPY . .
RUN go build -mod=vendor -o ./application cmd/*.go

FROM alpine:3.7
WORKDIR /nats-spliter
COPY --from=stage nats-spliter/application application
ENTRYPOINT [ "/nats-spliter/application" ]