# Архитектура первого среза

Control Plane — общее название управляющей части платформы.
Сейчас эту роль выполняет Deployment Service, отдельный процесс не создаём.

## Deployment Service
Владеет Project, Deployment и их данными в PostgreSQL.
Проверяет бизнес-правила и разрешённые переходы статусов.
Атомарно создаёт Deployment и событие в transactional outbox.
Publisher отправляет команды в Kafka; consumer принимает результаты worker.

## Deployment Worker
Получает команду с параметрами конкретной попытки деплоя.
Клонирует указанный commit, собирает образ по Dockerfile, запускает контейнер.
Проверяет health и сообщает прогресс и результат через события Kafka.
Не читает и не изменяет таблицы Deployment Service напрямую.

## MCP Server
Предоставляет tools для AI-агента и вызывает Deployment Service через gRPC.
Преобразует запросы и ответы, но не управляет состоянием деплоя.

## Границы текущей недели
Реализуем gRPC CreateDeployment и GetDeployment, PostgreSQL и outbox.
Kafka publisher и worker consumer начнём на следующем этапе.
Отдельную сущность Release пока не вводим.