.DEFAULT_GOAL := target/amp-httpd-rpi
.PHONY: clean fmt

go.mod:
	GOOS=linux GOARCH=arm64 docker-compose run --rm go mod init github.com/munkyboy/rpi-music/amp-httpd

go.sum: go.mod
	GOOS=linux GOARCH=arm64 docker-compose run --rm go mod download

target/amp-httpd-rpi: go.mod go.sum *.go
	@mkdir -p target &> /dev/null
	GOOS=linux GOARCH=arm64 docker-compose run --rm go build -v -o target/amp-httpd-rpi .

clean:
	rm -rf target

fmt:
	docker-compose run --rm go fmt .
