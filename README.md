# Система постов и комментариев

Поддерживает чтение и создание постов, комментарии произвольной глубины, cursor-based пагинацию и доставку новых комментариев через GraphQL Subscriptions. В качестве хранилища может использоваться база данных PostgreSQL или in-memory структура. После подписки на новые комментарии к посту клиент может асинхронно получать комментарии без необходимости нового запроса.

## Документация

- [Архитектура](docs/architecture.md) - компоненты системы и путь запроса.
- [GraphQL API](docs/api.md) - queries, mutations, subscriptions и модель пагинации.
- [Масштабирование](docs/scaling.md) - горизонтальное масштабирование, шардирование бд и редиса.

## Возможности

- создание и чтение постов;
- включение и отключение комментариев автором поста;
- корневые комментарии и ответы без ограничения глубины хранения;
- cursor-based пагинация постов и комментариев и ленивая загрузка комментариев по необходимости;
- GraphQL Subscriptions для асинхронного получения новых комментариев;
- защита от n+1 в запросах посто и комментариев через DataLoader;
- Заполнение тестовыми данными PostgreSQL и in-memory реализации хранилища;
- Использование Docker для распространения системы.

## Запуск

Доступные команды:

```bash
task
```

Запуск приложения:

```bash
cp .env.example .env
docker-compose up -d postgres redis
task migrate
docker-compose up -d app
```

Для запуска тестов:

```bash
task test
```

## Структура проекта

internal/application, internal/storage/pg, internal/storage/mem, internal/transport/graph.

```text
cmd/server/                           точка входа
internal/domain/                      доменные модели и ошибки
internal/application/                 бизнес-логика
internal/platform/                    инициализация зависимостей (сервер, бд, редис и логгер)
internal/storage/pg/                  PostgreSQL-реализация
internal/storage/mem/                 in-memory реализация
internal/transport/graph/             схема и резолверы
internal/transport/graph/dataloader/  даталоадеры для запросов
internal/notifier/                    реализации подписок на уведомления (через Redis Pub/Sub)
migrations/                           SQL миграции
seeds/                                SQL скрипт для генерации тестовых данных
docs/                                 документация проекта
```
