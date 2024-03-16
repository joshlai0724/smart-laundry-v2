.PHONY: create-network
create-network:
	docker network create db-net
	docker network create edge-db-net
	docker network create redis-net
	docker network create amqp-net
	docker network create mqtt-net
	docker network create nginx-net

.PHONY: build
build:
	cd ../backend && make build
	cd ../edge && make build
	cd ../frontend && make build

.PHONY: start
start: create-network
	cd infra && docker compose up -d --wait
	cd web && docker compose up -d --wait
	cd edge && docker compose up -d --wait

.PHONY: restart
restart: stop start

.PHONY: stop
stop:
	cd edge && docker compose down
	cd web && docker compose down
	cd infra && docker compose down
	docker network rm db-net edge-db-net redis-net amqp-net mqtt-net nginx-net

.PHONY: clean
clean:
	cd ../backend && make clean
	cd ../edge && make clean
	cd ../frontend && make clean