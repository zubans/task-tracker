# Переменные
APP_NAME = task-tracker
DOCKER_COMPOSE_FILE = docker-compose.yml
PROTOC_CONTAINER = protoc-gen

.PHONY: all build up down restart logs proto

# Сборка бинарного файла
build:
	docker build -t $(APP_NAME) .

# Генерация protobuf кода
proto:
	docker build -t $(PROTOC_CONTAINER) -f Dockerfile.protoc .
	docker run --rm -v $(shell pwd):/workspace $(PROTOC_CONTAINER)

# Запуск docker-compose
up:
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

# Остановка docker-compose
down:
	docker-compose -f $(DOCKER_COMPOSE_FILE) down

# Перезапуск сервисов
restart: down up

# Логи сервиса
logs:
	docker-compose logs -f