build:
	docker build -t yuriipolishchuk/kube-componentstatuses-prometheus-exporter .

push:
	docker tag yuriipolishchuk/kube-componentstatuses-prometheus-exporter yuriipolishchuk/kube-componentstatuses-prometheus-exporter:${tag}
	docker push yuriipolishchuk/kube-componentstatuses-prometheus-exporter:${tag}
