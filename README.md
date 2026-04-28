# Config Analyzer

Анализатор конфигурационных файлов (JSON/YAML) на предмет уязвимостей и проблем безопасности.

## Установка

### С помощью Go (рекомендуется)

```bash
go install github.com/ya/sec-config-scanner@latest
```

### Сборка из исходников

```bash
git clone https://github.com/ya/sec-config-scanner.git
cd sec-config-scanner
go build -o config-analyzer ./cmd/main.go
```

### Docker

```bash
docker build -t config-analyzer .
docker run --rm config-analyzer --help
```

## Использование

### Анализ файла

```bash
config-analyzer config.json
config-analyzer config.yaml
```

### Анализ из stdin

```bash
cat config.json | config-analyzer --stdin
```

### Рекурсивный анализ директории

```bash
config-analyzer --dir /path/to/configs
```

### Автоматическое исправление при ошибках

```bash
# Без флага -s: выход с кодом 1, если найдены проблемы
config-analyzer config.json

# С флагом -s: вывод результатов, но выход с кодом 0
config-analyzer --silent config.json
```

## HTTP Server

### Запуск

```bash
config-analyzer --server --port 8080
```

### Анализ через HTTP API

**JSON:**

```bash
curl -X POST http://localhost:8080/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "log": {"level": "debug"},
      "password": "secret123"
    }
  }'
```

**YAML:**

```bash
curl -X POST http://localhost:8080/analyze \
  -H "Content-Type: application/yaml" \
  -d 'data:
  log:
    level: debug
  password: secret123'
```

### Health check

```bash
curl http://localhost:8080/health
```

**Ответ:**

```json
{
  "success": false,
  "issues": [
    {
      "severity": "LOW",
      "field": "log.level",
      "description": "логирование в debug-режиме",
      "advice": "Поменяйте режим на более избирательный (info+)."
    },
    {
      "severity": "HIGH",
      "field": "password",
      "description": "пароль в открытом виде",
      "advice": "Используйте переменные окружения или vault для хранения секретов."
    }
  ]
}
```

## gRPC Server

### Запуск

```bash
config-analyzer --grpc --port 9090
```

### Пример клиента (Go)

**server.proto:**

```protobuf
service AnalyzerService {
  rpc Analyze(AnalyzeRequest) returns (AnalyzeResponse);
}

message AnalyzeRequest {
  string format = 1;
  bytes data = 2;
}
```

**Клиент:**

```go
package main

import (
	"context"
	"encoding/json"
	"log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := gen.NewAnalyzerServiceClient(conn)

	config := map[string]any{
		"log": map[string]any{"level": "debug"},
	}

	data, _ := json.Marshal(config)

	resp, err := client.Analyze(context.Background(), &gen.AnalyzeRequest{
		Format: "json",
		Data:   data,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Success: %v, Issues: %d\n", resp.Success, len(resp.Issues))
}
```

## 📸 Демонстрация работы

### CLI режим

![CLI - Анализ конфигурации с уязвимостями](screenshots/cli_bad.png)

**Анализ файла с проблемами:**
```bash
go run .\cmd\main.go testdata\bad.json
```

**Результат:**
```
HIGH: пароль в открытом виде. Используйте переменные окружения или vault для хранения секретов.
HIGH: TLS проверка отключена. Включите TLS в продакшн-окружении.
HIGH: слишком слабый алгоритм - RC4. Замените его на более безопасный.
MEDIUM: сервис слушает на 0.0.0.0 без ограничений. Ограничьте bind конкретным интерфейсом или внутренним IP.
LOW: логирование в debug-режиме. Поменяйте режим на более избирательный (info+).
```

![CLI - Безопасная конфигурация](screenshots/cli_safe.png)

**Анализ безопасной конфигурации:**
```bash
go run .\cmd\main.go testdata\safe.json
```

**Результат:**
```
✓ Конфигурация безопасна
```

### HTTP Server

![HTTP Server - Health Check](screenshots/http-health.png)

**Запуск и проверка:**
```bash
# Запуск сервера
Start-Process -FilePath "go" -ArgumentList "run .\cmd\main.go --server --port 8080" -NoNewWindow

# Health check
Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing
```

**Ответ:**
```json
{"status":"ok"}
```

![HTTP Server - Анализ через API](screenshots/http-analyze.png)

**Анализ через API:**
```powershell
$content = Get-Content "testdata\bad.json" -Raw
Invoke-RestMethod -Method Post -Uri "http://localhost:8080/analyze" -ContentType "application/json" -Body $content
```

**JSON ответ:**
```json
{
  "success": false,
  "issues": [
    {
      "severity": "HIGH",
      "field": "database.password",
      "description": "пароль в открытом виде",
      "advice": "Используйте переменные окружения или vault для хранения секретов."
    },
    {
      "severity": "HIGH",
      "field": "crypto.cipher",
      "description": "слишком слабый алгоритм - RC4",
      "advice": "Замените его на более безопасный."
    }
    // ... остальные проблемы
  ]
}
```

### gRPC Server

![gRPC Client - Полное тестирование](screenshots/grpc-client.png)

**Запуск сервера и тестирование через клиента:**
```bash
# Запуск gRPC сервера
Start-Process -FilePath "go" -ArgumentList "run .\cmd\main.go --grpc --port 9090" -NoNewWindow

# Запуск клиента
go run grpc_client.go
```

**Результат:**
```
=== Testing bad.json ===
bad.json - Success: false, Issues found: 5
  - LOW: логирование в debug-режиме
  - HIGH: пароль в открытом виде
  - MEDIUM: сервис слушает на 0.0.0.0 без ограничений
  - HIGH: TLS проверка отключена
  - HIGH: слишком слабый алгоритм - RC4

=== Testing bad.yaml ===
bad.yaml - Success: false, Issues found: 5
  - LOW: логирование в debug-режиме
  - HIGH: пароль в открытом виде
  - MEDIUM: сервис слушает на 0.0.0.0 без ограничений
  - HIGH: TLS отключён
  - HIGH: слишком слабый алгоритм - MD5

=== Testing safe.json ===
safe.json - Success: true, Issues found: 0

=== Testing safe.yaml ===
safe.yaml - Success: true, Issues found: 0

=== All gRPC tests completed! ===
```

**Полное тестовое покрытие:**
- ✅ CLI режим (JSON/YAML)
- ✅ Рекурсивный анализ директории
- ✅ HTTP Server (REST API)
- ✅ gRPC Server (RPC API)
- ✅ Все 5 правил безопасности

---

## Правила анализа

### 1. Debug Log (LOW)

Проверяет уровень логирования.

**Проблема:** `log.level: debug` или аналогичные поля.

**Совет:** Поменяйте режим на более избирательный (info+).

---

### 2. Plaintext Password (HIGH)

Ищет пароли в открытом виде.

**Проверяемые поля:** `password`, `passwd`, `pwd`, `secret`.

**Исключения:** Хеши (MD5, SHA1, SHA256, SHA512, bcrypt) и переменные окружения (`$VAR`).

**Пример проблемы:**

```yaml
database:
  password: "secret123"  # HIGH
```

**Пример безопасной конфигурации:**

```yaml
database:
  password: "$DB_PASSWORD"  # OK - переменная окружения
  # или
  password: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"  # OK - SHA256
```

**Совет:** Используйте переменные окружения или vault для хранения секретов.

---

### 3. Bind All (MEDIUM)

Проверяет, что сервис не слушает на всех интерфейсах.

**Проблема:** `host: 0.0.0.0` или аналогичные поля.

**Совет:** Ограничьте bind конкретным интерфейсом или внутренним IP.

---

### 4. TLS Disabled (HIGH)

Проверяет настройки TLS/SSL.

**Проблемы:**

- `tls.enabled: false`
- `tls.insecure_skip_verify: true` или аналогичные

**Совет:** Включите TLS в продакшн-окружении.

---

### 5. Weak Algorithm (HIGH)

Проверяет использование слабых алгоритмов шифрования и хеширования.

**Проверяемые поля:** `algorithm`, `algo`, `cipher`, `digest`, `hash`, `encryption`.

**Запрещённые алгоритмы:** MD5, SHA1, SHA-1, DES, 3DES, RC4, NULL.

**Пример проблемы:**

```json
{
  "crypto": {
    "algorithm": "md5"  # HIGH
  }
}
```

**Пример безопасной конфигурации:**

```json
{
  "crypto": {
    "algorithm": "sha256"  # OK
  }
}
```

**Совет:** Используйте SHA-256 или выше.

---

### 6. File Permissions (MEDIUM)

Проверяет права доступа к файлам, указанным в конфигурации.

**Только при использовании `--dir` или stdin.**

**Проблема:** Файлы с правами `mode & 0077 != 0` (чтение/запись/выполнение для "others").

**Совет:** Ограничьте права доступа (рекомендуется 0600 или 0640).

---

## Разработка

### Установка зависимостей

```bash
go mod download
```

### Тесты

```bash
go test ./...
```

### Линтер

```bash
golangci-lint run
```

### Сборка

```bash
make build
```

### Цели Makefile

- `make build` — собирает бинарник
- `make test` — запускает тесты
- `make lint` — запускает golangci-lint
- `make docker` — собирает Docker-образ
- `make run` — запускает бинарник
- `make run-server` — запускает HTTP сервер на порту 8080
- `make run-grpc` — запускает gRPC сервер на порту 8080
- `make proto` — пересобирает proto-файлы

---

## Декомпозиция severity


| Severity | Цвет | Описание                                                   |
| -------- | ---- | ---------------------------------------------------------- |
| HIGH     | 🔴   | Критические уязвимости, требующие немедленного исправления |
| MEDIUM   | 🟡   | Потенциальные проблемы безопасности                        |
| LOW      | 🟢   | Рекомендации по улучшению                                  |


---

## Декомпозиция output format

```bash
$ config-analyzer config.json
```

**Без проблем:**

```
✓ Конфигурация безопасна
```

**С проблемами:**

```
HIGH: пароль в открытом виде (database.password). Используйте переменные окружения или vault для хранения секретов.
HIGH: TLS отключён (tls.enabled). Включите TLS в продакшн-окружении.
MEDIUM: сервис слушает на 0.0.0.0 без ограничений (host). Ограничьте bind конкретным интерфейсом или внутренним IP.
LOW: логирование в debug-режиме (log.level). Поменяйте режим на более избирательный (info+).
```

---

## Docker

### Сборка образа

```bash
docker build -t config-analyzer .
```

### Запуск анализа файла

```bash
docker run --rm -v $(pwd):/app -w /app config-analyzer config.json
```

### Запуск HTTP сервера

```bash
docker run --rm -p 8080:8080 config-analyzer --server
```

---

## Примеры конфигураций

### Проблемная конфигурация (пример config.json)

```json
{
  "log": {
    "level": "debug"
  },
  "database": {
    "host": "0.0.0.0",
    "password": "secret123",
    "ssl": {
      "insecure_skip_verify": true,
      "enabled": false
    }
  },
  "crypto": {
    "algorithm": "md5"
  }
}
```

**Результат:**

```
HIGH: пароль в открытом виде (database.password). Используйте переменные окружения или vault для хранения секретов.
HIGH: TLS проверка отключена (database.ssl.insecure_skip_verify). Включите TLS в продакшн-окружении.
HIGH: TLS отключён (database.ssl.enabled). Включите TLS в продакшн-окружении.
HIGH: слабый алгоритм — md5 (crypto.algorithm). Используйте SHA-256 или выше.
MEDIUM: сервис слушает на 0.0.0.0 без ограничений (database.host). Ограничьте bind конкретным интерфейсом или внутренним IP.
LOW: логирование в debug-режиме (log.level). Поменяйте режим на более избирательный (info+).
```

---

## Лицензия

MIT


---

## Дополнительные ресурсы

- [Go Best Practices](https://go.dev/doc/effective_go)
- [CIS Benchmarks](https://www.cisecurity.org/benchmark)
- [OWASP ASVS](https://owasp.org/www-project-application-security-verification-standard)

