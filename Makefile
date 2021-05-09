# Run go lint
lint:
	golangci-lint run -vc ./.golangci.yaml ./...

build-upgrade-plugin: 
	cd cmd/ && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ../bin/upgrade-plugin && cd -


