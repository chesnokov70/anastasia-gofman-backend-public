APP_NAME = art_gallery_backend
GO = go
GOPATH = $(shell go env GOPATH)
MAIN_PATH = ./cmd/api/main.go
BUILD_DIR = ./build
DOCKER_IMAGE = art-gallery-backend
BINARY = $(BUILD_DIR)/$(APP_NAME)
DB_DSN = "postgres://postgres:postgres@localhost:5433/art_gallery?sslmode=disable"

GREEN = \033[0;32m
NC = \033[0m

.PHONY: all build run clean test lint fmt vet goimports help docker migrations-up migrations-down

all: clean fmt vet lint build test

build:
	@echo "${GREEN}Сборка приложения...${NC}"
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -o $(BINARY) $(MAIN_PATH)
	@echo "${GREEN}Сборка завершена: $(BINARY)${NC}"

run:
	@echo "${GREEN}Запуск приложения...${NC}"
	@$(GO) run $(MAIN_PATH)

clean:
	@echo "${GREEN}Очистка...${NC}"
	@rm -rf $(BUILD_DIR)
	@$(GO) clean
	@echo "${GREEN}Очистка завершена${NC}"

# Команды для тестирования
test:
	@echo "${GREEN}Запуск тестов...${NC}"
	@$(GO) test -v ./internal/...

test-cover:
	@echo "${GREEN}Запуск тестов с покрытием...${NC}"
	@$(GO) test -cover -coverprofile=coverage.out ./internal/...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "${GREEN}Отчет о покрытии создан: coverage.html${NC}"

# Команды для работы с кодом
lint:
	@echo "${GREEN}Проверка линтером...${NC}"
	@golangci-lint run

fmt:
	@echo "${GREEN}Форматирование кода...${NC}"
	@$(GO) fmt ./...

vet:
	@echo "${GREEN}Статический анализ кода...${NC}"
	@$(GO) vet ./...

goimports:
	@echo "${GREEN}Проверка импортов...${NC}"
	@goimports -w .

# Команды для работы с базой данных
migrate-create:
	@echo "${GREEN}Создание миграции...${NC}"
	@migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	@echo "${GREEN}Применение миграций...${NC}"
	@migrate -path migrations -database $(DB_DSN) up

migrate-down:
	@echo "${GREEN}Откат миграций...${NC}"
	@migrate -path migrations -database $(DB_DSN) down

# Docker команды для запуска без Compose (не рекомендуется для разработки)
docker-standalone-build:
	@echo "${GREEN}Сборка Docker образа (standalone)...${NC}"
	@docker build -t $(DOCKER_IMAGE) .

docker-standalone-run:
	@echo "${GREEN}Запуск Docker контейнера (standalone)...${NC}"
	@echo "ВНИМАНИЕ: Этот метод запускает только бэкенд без базы данных. Используйте 'make compose-up' для полного запуска."
	@docker run -p 8080:8010 $(DOCKER_IMAGE)

# Docker Compose команды (рекомендуемый способ)
compose-up:
	@echo "${GREEN}Запуск проекта с Docker Compose...${NC}"
	@docker-compose up --build -d

compose-down:
	@echo "${GREEN}Остановка проекта Docker Compose...${NC}"
	@docker-compose down

compose-logs:
	@echo "${GREEN}Просмотр логов Docker Compose...${NC}"
	@docker-compose logs -f

compose-nocashe:
	@echo "${GREEN}Запуск проекта с Docker Compose без кеша...${NC}"
	@docker-compose down && docker-compose build --no-cache backend && docker-compose up -d

# Вспомогательные команды
setup:
	@echo "${GREEN}Установка зависимостей...${NC}"
	@$(GO) mod download
	@go install -v github.com/golang/mock/mockgen@latest
	@go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install -v golang.org/x/tools/cmd/goimports@latest
	@go install -v github.com/golang-migrate/migrate/v4/cmd/migrate@latest

help:
	@echo "Доступные команды:"
	@echo "  make build         - сборка приложения"
	@echo "  make run           - запуск приложения"
	@echo "  make test          - запуск тестов"
	@echo "  make test-cover    - запуск тестов с отчетом о покрытии"
	@echo "  make lint          - проверка линтером"
	@echo "  make fmt           - форматирование кода"
	@echo "  make clean         - очистка сборочных артефактов"
	@echo "  make setup         - установка зависимостей для разработки"
	@echo "  make migrate-up    - применение миграций базы данных"
	@echo "  make migrate-down  - откат миграций базы данных"
	@echo "  make compose-up    - запуск проекта с Docker Compose (рекомендуется)"
	@echo "  make compose-down  - остановка проекта Docker Compose"
	@echo "  make compose-logs  - просмотр логов Docker Compose"
	@echo "  make docker-standalone-build - сборка standalone Docker образа"
	@echo "  make docker-standalone-run   - запуск standalone Docker контейнера"
	@echo "  make help          - показать справку"

default: help
