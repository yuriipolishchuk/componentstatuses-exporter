FROM golang:1.10.3 as builder

WORKDIR /go/src/github.com/yuriipolishchuk/kube-componentstatuses-prometheus-exporter/

RUN go get -d -v \
    k8s.io/client-go/kubernetes \
    k8s.io/client-go/rest \
    github.com/prometheus/client_golang/prometheus \
    ;

COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .


FROM alpine

ENV COMPONENTSTATUSES_CHECK_RATE=30

COPY --from=builder /go/src/github.com/yuriipolishchuk/kube-componentstatuses-prometheus-exporter/app .

CMD ["/app"]
