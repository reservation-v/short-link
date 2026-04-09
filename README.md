# short-link

`short-link` это сервис для сокращения ссылок. Он принимает длинный URL, выдаёт короткий код и по этому коду возвращает исходный URL в JSON.

Сервис поддерживает два режима запуска:
- `memory` для быстрого локального старта с in-memory хранилищем
- `postgres` для работы с PostgreSQL

## Что поддерживается

- `POST /links` создаёт короткую ссылку для URL или возвращает уже существующую для того же URL
- `GET /links/{code}` возвращает исходный URL по короткому коду

Закодированная ссылка:
- всегда имеет длину `10`
- строится из числового `id` через base63
  base63 здесь это запись числа в системе счисления с алфавитом из `63` символов: `a-z`, `A-Z`, `0-9` и `_`
- слева дополняется символом `a`
- не хранится в базе отдельно, а вычисляется из `id`

## Флаги запуска

- `--storage=memory|postgres`
- `--postgres-dsn=<dsn>`
- `--http-addr=<listen address>`

По умолчанию сервис слушает `:8081`.
Все доступные команды можно посмотреть через:

```bash
make help
```

## Запуск

### 1. Быстрый запуск с in-memory хранилищем

```bash
make run-memory
```

Сервис будет доступен на `http://localhost:8081`.

### 2. Полный запуск через Docker

```bash
make docker-up
```

После запуска будут доступны:
- сервис: `http://localhost:8081`
- PostgreSQL: `localhost:5433`

### 3. Ручной запуск

#### **In-memory** хранилище
```bash
go run ./cmd/short-link --storage=memory --http-addr=:<your_addr>
```
- сервис: `http://localhost:<your_addr>`

#### **Postgres**
```bash
# postgres: поднимаем только БД и миграции, приложение запускаем локально
docker compose up -d postgres
docker compose run --rm migrate
go run ./cmd/short-link \
  --storage=postgres \
  --postgres-dsn='postgres://postgres:postgres@localhost:5433/shortener?sslmode=disable' \
  --http-addr=:<your_addr>
```
- сервис: `http://localhost:<your_addr>`

Остановка контейнера
```bash
# обычная остановка
make docker-down
# или с полной очисткой volume
make reset
```

### 4. Тесты

#### Unit тесты
```bash
make test
```
#### Интеграционные тесты
```bash
make test-integration
```

## Как пользоваться

При любом виде запуска обращаемся к `http://localhost:8081`, если его не заоверрайдить через `--http-addr` при мануальном запуске через go run.

Создать короткую ссылку:

```bash
curl -i -X POST http://localhost:8081/links \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/some/path"}'
```

Пример ответа:

```json
{
  "url": "https://example.com/some/path",
  "code": "aaaaaaaaaB",
  "short_url": "http://localhost:8081/aaaaaaaaaB"
}
```

Посмотреть оригинальный URL, если такая короткая ссылка сохранена (aaaaaaaaaB как раз и является короткой ссылкой):

```bash
curl -i http://localhost:8081/links/aaaaaaaaaB
```

Остановить Docker Compose:

```bash
make docker-down
```
