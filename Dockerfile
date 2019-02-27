FROM golang:1.12 as builder

WORKDIR /go/src/github.com/yuriipolishchuk/componentstatuses-exporter/

RUN go get -d -v \
    k8s.io/client-go/kubernetes \
    k8s.io/client-go/rest \
    github.com/prometheus/client_golang/prometheus \
    github.com/sirupsen/logrus \
    ;

COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o exporter .


FROM alpine:3.9

COPY --from=builder /go/src/github.com/yuriipolishchuk/componentstatuses-exporter/exporter .

CMD ["/exporter"]
