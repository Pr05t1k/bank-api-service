#  Bank API Service

[![Go Version](https://img.shields.io/badge/Go-1.23-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-24.0-blue.svg)](https://www.docker.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-17-green.svg)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

REST API для банковского сервиса на языке Go. Полнофункциональное приложение с поддержкой банковских счетов, карт, переводов и кредитов.

##  Содержание

- [Возможности](#-возможности)
- [Технологии](#-технологии)
- [Установка и запуск](#-установка-и-запуск)
- [API Эндпоинты](#-api-эндпоинты)
- [Примеры запросов](#-примеры-запросов)
- [Безопасность](#-безопасность)
- [Структура проекта](#-структура-проекта)
- [Тестирование](#-тестирование)
- [Roadmap](#-roadmap)

##  Возможности

### Основные функции
-  **Регистрация и аутентификация** - JWT токены с 24-часовым сроком действия
- **Управление счетами** - создание и просмотр банковских счетов
-  **Банковские карты** - генерация карт с валидными номерами (алгоритм Луна)
-  **Переводы** - безопасные переводы между счетами с историей транзакций
-  **Кредиты** - оформление кредитов с расчетом аннуитетных платежей
-  **График платежей** - детальный график всех платежей по кредиту

### Безопасность
-  **JWT аутентификация** - защита всех API эндпоинтов
-  **Bcrypt хеширование** - пароли пользователей и CVV коды
-  **HMAC** - проверка целостности данных карт
-  **PGP шифрование** - защита номеров карт (готово к интеграции)

### Инфраструктура
-  **Docker** - полная контейнеризация приложения и БД
-  **PostgreSQL** - надежное хранение данных
-  **Logrus** - структурированное логирование
-  **Миграции** - автоматическое создание схемы БД

## 🛠 Технологии

- **Go 1.23** - основной язык разработки
- **Gorilla Mux** - маршрутизация
- **JWT** - аутентификация
- **PostgreSQL 17** - база данных
- **Docker & Docker Compose** - контейнеризация
- **bcrypt** - хеширование
- **Logrus** - логирование

##  Установка и запуск

### Предварительные требования
- Go 1.23+
- Docker & Docker Compose
- PostgreSQL 17 (опционально, при локальном запуске)
- Make (опционально)

### Вариант 1: Запуск через Docker (рекомендуется)


# Клонируйте репозиторий
git clone https://github.com/YOUR_USERNAME/bank-api-service.git
cd bank-api-service

# Создайте файл .env из примера
cp .env.example .env

# Запустите приложение
docker-compose up --build -d

# Проверьте статус
docker-compose ps

# Посмотрите логи
docker-compose logs -f

## Вариант 2: Локальный запуск
# Установите зависимости
go mod download

# Создайте базу данных PostgreSQL
createdb bank

# Примените миграции
psql -d bank -f migrations/001_create_tables.sql

# Создайте файл .env
cat > .env << EOF
PORT=8080
DATABASE_URL=postgres://postgres:password@localhost:5432/bank?sslmode=disable
JWT_SECRET=your-secret-key
EOF

# Запустите приложение
go run cmd/main.go
Приложение будет доступно по адресу: http://localhost:8081
## API Эндпоинты
# Публичные эндпоинты
Метод	Эндпоинт	Описание
GET	/health	Проверка статуса сервера
POST	/register	Регистрация нового пользователя
POST	/login	Аутентификация и получение JWT токена
## Защищенные эндпоинты (требуют Bearer токен)
Метод	Эндпоинт	Описание
POST	/api/accounts	Создание нового счета
GET	/api/accounts	Получение всех счетов пользователя
GET	/api/accounts/{id}	Получение счета по ID
POST	/api/accounts/{account_id}/cards	Выпуск новой карты
GET	/api/accounts/{account_id}/cards	Получение карт счета
POST	/api/transfer	Перевод между счетами
GET	/api/transactions?account_id={id}	История транзакций
POST	/api/credits	Оформление кредита
GET	/api/credits/{credit_id}/schedule	График платежей
## Примеры запросов
# Регистрация пользователя
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "secure_password"
  }'
  Ответ:
  {
  "message": "User registered successfully",
  "user": {
    "id": 1,
    "username": "john_doe",
    "email": "john@example.com",
    "created_at": "2026-01-11T10:00:00Z"
  }
}

# Аутентификация
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "secure_password"
  }'

  Ответ:
  {
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "john_doe",
    "email": "john@example.com"
  }
}

# Создание счета
curl -X POST http://localhost:8080/api/accounts \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'

  # Выпуск карты
  curl -X POST http://localhost:8080/api/accounts/1/cards \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
Ответ:
{
  "card": {
    "id": 1,
    "last4": "1234",
    "expiry_month": 1,
    "expiry_year": 2029
  },
  "cvv": "123"
}

# Перевод средств
curl -X POST http://localhost:8080/api/transfer \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "from_account_id": 1,
    "to_account_id": 2,
    "amount": 1000,
    "description": "Monthly payment"
  }'

  # Оформление кредита
  curl -X POST http://localhost:8080/api/credits \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": 1,
    "amount": 50000,
    "term_months": 12
  }'
  Ответ:
  {
  "id": 1,
  "amount": 50000,
  "interest_rate": 12.5,
  "term_months": 12,
  "monthly_payment": 4454.14,
  "total_payment": 53449.72,
  "remaining_amount": 50000,
  "status": "active",
  "created_at": "2026-01-11T10:00:00Z"
}

# График платежей
curl -X GET http://localhost:8080/api/credits/1/schedule \
  -H "Authorization: Bearer YOUR_TOKEN"

  ## Безопасность
  # Реализованные меры

JWT аутентификация - все защищенные эндпоинты требуют валидный токен

Bcrypt хеширование - пароли и CVV коды хранятся в хешированном виде

HMAC - обеспечение целостности данных карт

SQL инъекции - все запросы параметризованы

Проверка прав - пользователь может получить доступ только к своим счетам

Транзакции - переводы выполняются в атомарных транзакциях

# Переменные окружения для безопасности

JWT_SECRET=your-super-secret-key-min-32-chars
HMAC_SECRET=your-hmac-secret-for-card-integrity
PGP_PUBLIC_KEY=your-pgp-public-key
PGP_PRIVATE_KEY=your-pgp-private-key

## Тестирование

# Все тесты
go test ./...

# Тесты с покрытием
go test -cover ./...

# Конкретный пакет
go test ./internal/service -v

## Пример тестового скрипта

#!/bin/bash

# Регистрация
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@test.com","password":"test123"}'

# Логин
TOKEN=$(curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"test123"}' \
  | jq -r '.token')

# Создание счета
curl -X POST http://localhost:8080/api/accounts \
  -H "Authorization: Bearer $TOKEN"

# Получение счетов
curl -X GET http://localhost:8080/api/accounts \
  -H "Authorization: Bearer $TOKEN"

  ## Roadmap
  # Реализовано
  Регистрация и аутентификация

Управление банковскими счетами

Генерация карт (алгоритм Луна)

Переводы между счетами

Кредиты с аннуитетными платежами

История транзакций

Docker контейнеризация

JWT аутентификация

Логирование