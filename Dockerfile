#test1
FROM golang:1.13.12 AS builder
ADD . $GOPATH/src/gocron
WORKDIR $GOPATH/src/gocron/master/main/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main

FROM centos
COPY --from=builder /go/src/gocron/master/main /app/
CMD ["/app/main", "--config=/app/config.json"]
EXPOSE 8070
