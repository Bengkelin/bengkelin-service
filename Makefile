.PHONY: run build

run:
	go run cmd/app/main.go

build:
	go build -o cmd/app cmd/app/main.go