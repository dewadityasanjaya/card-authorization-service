DB_URL=postgres://cardauth:cardauth_secret@localhost:5432/card_auth_db?sslmode=disable

migrate-up:
	migrate -path ./migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DB_URL)" down

migrate-version:
	migrate -path ./migrations -database "$(DB_URL)" version