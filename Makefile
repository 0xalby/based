include .env

.PHONY: build run clean up down status release docker

build:
	@env CGO_ENABLED=0 go build -o bin/base -trimpath .

debug: build
	@dlv exec ./bin/base

run: build
	@./bin/base

clean:
	@rm -r bin

up:
	@GOOSE_DRIVER=$(DATABASE_DRIVER) GOOSE_DBSTRING=$(DATABASE_ADDRESS) goose -dir=database/migrations up
down:
	@GOOSE_DRIVER=$(DATABASE_DRIVER) GOOSE_DBSTRING=$(DATABASE_ADDRESS) goose -dir=database/migrations down
status:
	@GOOSE_DRIVER=$(DATABASE_DRIVER) GOOSE_DBSTRING=$(DATABASE_ADDRESS) goose -dir=database/migrations status

release:
	@env CGO_ENABLED=0 GOOS="windows" GOARCH="amd64" go build -o bin/base_windows_amd64.exe -ldflags="-s -w -extldflags=-static" -trimpath .
	@env CGO_ENABLED=0 GOOS="windows" GOARCH="arm64" go build -o bin/base_windows_arm64.exe -ldflags="-s -w -extldflags=-static" -trimpath .
	@env CGO_ENABLED=0 GOOS="darwin" GOARCH="amd64" go build -o bin/base_macos_amd64 -ldflags="-s -w -extldflags=-static" -trimpath .
	@env CGO_ENABLED=0 GOOS="darwin" GOARCH="arm64" go build -o bin/base_macos_arm64 -ldflags="-s -w -extldflags=-static" -trimpath .
	@env CGO_ENABLED=0 GOOS="linux" GOARCH="amd64" go build -o bin/base_linux_amd64 -ldflags="-s -w -extldflags=-static" -trimpath .
	@env CGO_ENABLED=0 GOOS="linux" GOARCH="arm64" go build -o bin/base_linux_arm64 -ldflags="-s -w -extldflags=-static" -trimpath .
	@env CGO_ENABLED=0 GOOS="freebsd" GOARCH="amd64" go build -o bin/base_freebsd_amd64 -ldflags="-s -w -extldflags=-static" -trimpath .
	@env CGO_ENABLED=0 GOOS="openbsd" GOARCH="amd64" go build -o bin/base_openbsd_amd64 -ldflags="-s -w -extldflags=-static" -trimpath .

docker:
	@docker build -t based .