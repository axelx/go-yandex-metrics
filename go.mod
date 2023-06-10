module github.com/axelx/go-yandex-metrics

go 1.20

replace internal/storage => ./internal/server/storage

replace internal/handlers => ./internal/server/handlers

require (
	github.com/go-chi/chi/v5 v5.0.8
	github.com/stretchr/testify v1.8.4
	internal/handlers v0.0.0-00010101000000-000000000000
	internal/storage v0.0.0-00010101000000-000000000000
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
