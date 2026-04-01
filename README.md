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

### 2. Локальный запуск приложения с PostgreSQL

Сначала нужно поднять PostgreSQL и применить миграции:

```bash
make postgres-init
```

После этого запустить сервиса локально:

```bash
make run-postgres
```

По умолчанию используются:
- PostgreSQL: `localhost:5433`
- сервис: `http://localhost:8081`

Если нужен свой адрес или свой DSN, можно передать их при запуске:

```bash
make run-postgres HTTP_ADDR=:8081 POSTGRES_DSN='postgres://postgres:postgres@localhost:5433/shortener?sslmode=disable'
```

### 3. Полный запуск через Docker

```bash
make docker-up
```

После запуска будут доступны:
- сервис: `http://localhost:8081`
- PostgreSQL: `localhost:5433`

### 4. Тесты

```bash
make test
```

Запустит все тесты сервиса

## Как пользоваться

При любом виде запуска обращаемся к `http://localhost:8081`, если его не заоверрайдить через --http-addr при мануальном запуске через go run или HTTP_ADDR в запуске через `make run-memory` или `make run-postgres`.

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
