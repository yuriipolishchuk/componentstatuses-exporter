image=polishchuk/componentstatuses-exporter

build:
	gofmt -w .
	docker build -t ${image} .

push:
	docker push ${image}:latest
ifdef tag
	docker tag ${image} ${image}:${tag}
	docker push ${image}:${tag}
endif

