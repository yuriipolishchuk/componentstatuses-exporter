FROM golang:1.12.9 as builder

WORKDIR /app

COPY go.mod main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o exporter .


FROM alpine:3.10.3

COPY --from=builder /app/exporter .

CMD ["/exporter"]
