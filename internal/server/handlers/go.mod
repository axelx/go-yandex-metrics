module handlers

go 1.20

replace internal/service => ../service

replace internal/storage => ../storage

require (
	github.com/go-chi/chi/v5 v5.0.8
	internal/service v0.0.0-00010101000000-000000000000
	internal/storage v0.0.0-00010101000000-000000000000
)
