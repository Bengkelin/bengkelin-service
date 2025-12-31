.PHONY: run build test test-coverage swagger-install swagger-gen swagger-serve

run:
	go run cmd/app/main.go

build:
	go build -o cmd/app cmd/app/main.go

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Swagger documentation
swagger-install:
	go install github.com/swaggo/swag/cmd/swag@latest

swagger-gen:
	swag init -g main.go -o docs --parseDependency --parseInternal

swagger-serve:
	@echo "Swagger documentation available at: http://localhost:3000/swagger/index.html"
	@echo "Make sure to run 'make swagger-gen' first to generate the documentation"
	@echo "Then start the server with 'make run'"

swagger-clean:
	rm -f docs/docs.go docs/swagger.json docs/swagger.yaml