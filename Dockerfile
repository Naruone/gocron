#test
FROM golang:1.2 AS builder

ADD ./ /app

WORKDIR /app

RUN GOOS=linux GOARCH=amd64 go build -o master/main/main master/main/main.go


FROM alpine:3
COPY --from=builder /app/master/main/main /app/
CMD ["/app/main", ""]
EXPOSE 8070
