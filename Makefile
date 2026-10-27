include .env
export


migrate-up:
	@migrate -path ./migrations -database ${DATABASE_URL} up

migrate-down:
	@migrate -path ./migrations -database ${DATABASE_URL} down

migrate-force:
	@migrate -path ./migrations -database ${DATABASE_URL} force ${VERSION}


migrate-status:
	@migrate -path ./migrations -database ${DATABASE_URL} version

deployment-up:
	go run ./cmd/deployment-service

publisher-up:
	go run ./cmd/outbox-publisher

worker-up:
	go run ./cmd/build-worker