module github.com/axelx/go-yandex-metrics

go 1.20

replace internal/storage => ./internal/server/storage

replace internal/handlers => ./internal/server/handlers

replace internal/service => ./internal/server/service

replace internal/config => ./internal/server/config

replace internal/metrics => ./internal/server/metrics

require (
	github.com/stretchr/testify v1.8.4
	internal/config v0.0.0-00010101000000-000000000000
	internal/handlers v0.0.0-00010101000000-000000000000
	internal/metrics v0.0.0-00010101000000-000000000000
	internal/storage v0.0.0-00010101000000-000000000000
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-chi/chi/v5 v5.0.8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	internal/service v0.0.0-00010101000000-000000000000 // indirect
)
