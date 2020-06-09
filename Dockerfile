#test1
FROM alpine:latest AS builder
RUN apk update
RUN apk upgrade
RUN apk add --update go
ENV GOPATH /app
ADD . $GOPATH/src/gocron
WORKDIR $GOPATH/src/gocron/master/main/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main

FROM alpine:latest
COPY --from=builder /go/src/gocron/master/main /app/
CMD ["/app/main", "--config=/app/config.json"]
EXPOSE 8070
