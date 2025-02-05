db_create:
	goose create -s $(name) sql

db_migrate_up:
	goose up 

db_migrate_down:
	goose down

run:
	go run ./cmd/server/main.go
