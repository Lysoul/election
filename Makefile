createdb:
	createdb --username=postgres --owner=postgres election

dropdb:
	dropdb --username=postgres election

migrateup:
	migrate -path db/migration -database "postgresql://postgres:secret@localhost:5432/election?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://postgres:secret@localhost:5432/election?sslmode=disable" -verbose down

sqlc:
	sqlc generate

mock:
	mockgen -package mockdb  -destination db/mock/store.go --build_flags=--mod=mod election/db/sqlc Store	

test:
	go test -v -cover ./...

.PHONY: createdb dropdb migrateup migratedown sqlc test