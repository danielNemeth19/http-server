.PHONY: conn-db
conn-db:
	psql "postgres://postgres:admin@localhost:5432/chirpy"


.PHONY: goose-up
goose-up:
	cd sql/schema && goose postgres "postgres://postgres:admin@localhost:5432/chirpy" up

.PHONY: goose-down
goose-down:
	cd sql/schema && goose postgres "postgres://postgres:admin@localhost:5432/chirpy" down

.PHONY: run-db
run-db:
	podman run -d --rm --name chirpy-db \
		-e POSTGRES_PASSWORD=admin \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_DB=chirpy \
		-p 5429:5432 \
		-v chirpy_storage:/var/lib/postgresql/data \
		docker.io/library/postgres:15
