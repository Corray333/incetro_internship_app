include .env
.SILENT:
goose-up:
	cd /api/migrations && goose postgres "user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) host=localhost port=5432 dbname=internship sslmode=disable" up
goose-down:
	cd /api/migrations && goose postgres "user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) host=localhost port=5432 dbname=internship sslmode=disable" down
goose-down-all:
	cd /api/migrations && goose postgres "user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) host=localhost port=5432 dbname=internship sslmode=disable" down-to 0