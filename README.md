# Config Analyzer

CLI / HTTP / gRPC анализатор конфигураций **JSON и YAML**. Ориентирован на **DevSecOps / SRE**: выдаёт проблемы с `LOW`, `MEDIUM`, `HIGH` и возвращает ненулевой exit code при наличии проблем (кроме режима `--silent`).

[![Go Report Card](https://goreportcard.com/badge/github.com/lopatich-privet/sec-config-scanner)](https://goreportcard.com/report/github.com/lopatich-privet/sec-config-scanner)
![Go Version](https://img.shields.io/github/go-mod/go-version/lopatich-privet/sec-config-scanner)
![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)
[![golangci-lint](https://img.shields.io/badge/linters-golangci--lint-blue)](https://golangci-lint.run)

## Установка

```bash
go build -o config-analyzer ./cmd/main.go
./config-analyzer --help
```

## Использование

```bash
./config-analyzer ./config.yaml
./config-analyzer --stdin < ./config.json
./config-analyzer --dir ./configs
./config-analyzer --silent ./config.yaml
```

Порты (опционально): `PORT` (HTTP, по умолчанию `8080`), `GRPC_PORT` (gRPC, по умолчанию `50051`) — см. `.env.example`.

## HTTP API

```bash
./config-analyzer --server --port 8080
curl -sS -X POST http://localhost:8080/analyze -H "Content-Type: application/json" -d '{"log":{"level":"debug"}}'
curl -sS http://localhost:8080/health
```

## gRPC

```bash
./config-analyzer --grpc --grpc-port 50051
go run ./cmd/grpc-client/main.go
```

## Архитектура

**Чистая архитектура**: транспортный слой не содержит бизнес-правил; он валидирует вход и вызывает сервис анализа.


| Слой       | Ответственность                                | Код                                                 |
| ---------- | ---------------------------------------------- | --------------------------------------------------- |
| Транспорт  | HTTP / gRPC адаптеры, лимиты, проверка формата | `server/http/server.go`, `server/grpc/server.go`    |
| Приложение | Парсинг байтов, запуск правил, доменные ошибки | `internal/service/analyzer_service.go`              |
| Домен      | Проверки правил на распарсенной конфигурации   | `internal/analyzer/analyzer.go`, `internal/rules/*` |


**Внедрение зависимостей (DI)**: оба сервера принимают `service.ConfigAnalyzer` — `http.NewServer(port, svc)`, `grpc.NewServer(port, svc)`. Это упрощает тесты (mock) и держит транспортный слой тонким.

## Безопасность и валидация

### Рекурсивный обход (включая `[]any`)

`traverseAndCheck()` в `internal/rules/helpers.go` обходит `map[string]any`, `map[any]any` и `**[]any`**, собирая пути с индексами массивов (например `a.b[0].c`).

### Контекст TLS (меньше ложных срабатываний)

`internal/rules/tls_disabled.go` трактует `enabled: false` как TLS-проблему **только** когда путь выглядит TLS-специфичным (`tls`, `ssl`, `https`, `secure`). Это не флаггит `cache.enabled: false` и похожие случаи.

### Валидация транспортного слоя (кратко)

- **HTTP**: `POST`, непустое тело запроса, обязательный `Content-Type`, белый список через `parser.FormatFromContentType`, лимит размера тела (`MaxBytesReader`).
- **gRPC**: непустой `data`, белый список форматов через `parser.FormatFromString`.

## Контракт API

### gRPC: errors → status codes (как в `server/grpc/server.go`)

Клиентам важно ветвиться по `**status.Code(err)`**, а не по тексту ошибок.


| Источник     | Условие                                                  | gRPC code         |
| ------------ | -------------------------------------------------------- | ----------------- |
| Обработчик   | пустой `data`                                            | `InvalidArgument` |
| Обработчик   | неподдерживаемый `format`                                | `InvalidArgument` |
| Сервис       | `ErrEmptyData`, `ErrUnsupportedFormat`, `ErrParseFailed` | `InvalidArgument` |
| Контекст     | `context.Canceled`                                       | `Canceled`        |
| По умолчанию | любая другая ошибка                                      | `Internal`        |


## Политика прав доступа к файлам

`file_permissions` в `internal/rules/file_permissions.go`:

- Проверяет **сам конфиг-файл** через `cfg.FilePath` (подмена/утечка на диске).
- Проверяет **абсолютные пути** в полях, похожих на пути (сертификаты/ключи и т.п.).
- **Ошибки `os.Stat` не игнорируются**: создаётся issue `**LOW`** с текстом ошибки, чтобы операторы видели битые пути / deny по правам, а не «тихий» пропуск.
- Если `stat` успешен: `**HIGH**` при наличии прав на запись у группы/прочих; `**MEDIUM**` при наличии прав на чтение у группы/прочих (см. реализацию по бит-маскам).

## Разработка

```bash
go test ./...
golangci-lint run ./...
```

