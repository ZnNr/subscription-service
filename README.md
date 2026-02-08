# subscription-service
A REST service for aggregating data about users' online subscriptions.


# Subscription Service

REST API для управления подписками пользователей.


## Технологии
- **Go** + **Gin** - HTTP сервер
- **PostgreSQL** - база данных
- **Docker** + **Docker Compose** - контейнеризация
- **Swagger/OpenAPI** - документация
- **Logrus** - логирование

## Быстрый запуск

```bash
# Запуск через Docker Compose
docker-compose up --build
```
### API будет доступно по адресу: http://localhost:8080
### Swagger UI: http://localhost:8080/swagger/index.html

## API Endpoints
### Подписки
POST /api/v1/subscriptions - Создать подписку

GET /api/v1/subscriptions - Список подписок (фильтры: user_id, service_name)

GET /api/v1/subscriptions/:id - Получить подписку по ID

PUT /api/v1/subscriptions/:id - Обновить подписку

DELETE /api/v1/subscriptions/:id - Удалить подписку

Отчеты
POST /api/v1/subscriptions/summary - Подсчет суммы подписок за период

Примеры запросов
# Создать подписку
curl -X POST http://localhost:8080/api/v1/subscriptions \
-H "Content-Type: application/json" \
-d '\''{
"service_name": "Netflix",
"price": 599,
"user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
"start_date": "01-2025"
}'\''

# Получить все подписки
curl http://localhost:8080/api/v1/subscriptions

# Подсчитать сумму за 2025 год для пользователя
curl -X POST http://localhost:8080/api/v1/subscriptions/summary \
-H "Content-Type: application/json" \
-d '\''{
"start_date": "01-2025",
"end_date": "12-2025",
"user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba"
}'\''