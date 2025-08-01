.PHONY: migrate-up migrate_down migrate_create migrate-version

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL) down 1"

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

migrate-version:
	migrate -path migrations -database "$(DATABASE_URL)" version

# Example usage:
# make migrate -up DATABASE_URL = "postgres://user:pass@localhost:5432/talkify?sslmode=disable"
