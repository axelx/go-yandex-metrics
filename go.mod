module github.com/axelx/go-yandex-metrics

go 1.20

replace internal/storage => ./internal/server/storage

replace internal/handlers => ./internal/server/handlers

require (
	github.com/dustin/go-humanize v1.0.1
	internal/handlers v0.0.0-00010101000000-000000000000
	internal/storage v0.0.0-00010101000000-000000000000
)
