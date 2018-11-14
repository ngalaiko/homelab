IMAGE ?= "ngalayko/dyn-dns"

run: build
	./bin/dyn-dns

build:
	go build -o ./bin/dyn-dns ./cmd/dyn-dns

docker-build:
	docker build . -t ${IMAGE}

docker-push: docker-build
	docker push ${IMAGE}

tests:
	go test -v ./app
