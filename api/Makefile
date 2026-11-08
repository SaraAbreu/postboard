.PHONY: run test vet

run:            ## Ejecuta la API en memoria (sin base de datos)
	go run ./cmd/api

test:           ## Ejecuta los tests
	go test ./... -v

vet:            ## Análisis estático
	go vet ./...
