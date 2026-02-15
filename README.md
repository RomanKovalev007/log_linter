# loglint

Go-линтер для проверки лог-записей, совместимый с [golangci-lint](https://golangci-lint.run/).

Анализирует вызовы логгеров `log/slog` и `go.uber.org/zap` и проверяет сообщения на соответствие правилам.

## Правила

| # | Правило | Описание |
|---|---------|----------|
| 1 | Строчная буква | Лог-сообщения должны начинаться со строчной буквы |
| 2 | Английский язык | Лог-сообщения должны быть только на английском языке |
| 3 | Без спецсимволов | Лог-сообщения не должны содержать спецсимволы или эмодзи |
| 4 | Без чувствительных данных | Лог-сообщения не должны содержать конкатенацию с переменными, содержащими пароли, токены и т.д. |

## Примеры

```go
// Rule 1: lowercase start
slog.Info("Starting server") // BAD
slog.Info("starting server") // OK

// Rule 2: English only
slog.Info("запуск сервера")  // BAD
slog.Info("starting server") // OK

// Rule 3: no special characters or emoji
slog.Info("server started!") // BAD
slog.Info("server started")  // OK

// Rule 4: no sensitive data
slog.Info("user password " + password) // BAD
slog.Info("user authenticated")        // OK
```

## Поддерживаемые логгеры

- **log/slog** — `Debug`, `Info`, `Warn`, `Error`, `DebugContext`, `InfoContext`, `WarnContext`, `ErrorContext`, `Log`, `LogAttrs`
- **go.uber.org/zap** — `Logger` (`Debug`, `Info`, `Warn`, `Error`, `DPanic`, `Panic`, `Fatal`) и `SugaredLogger` (`Debugw`, `Infow`, `Warnw`, `Errorw`, `Debugf`, `Infof`, `Warnf`, `Errorf` и т.д.)

## Авто-исправление (SuggestedFixes)

Линтер предоставляет автоматические исправления для правил 1 (строчная буква) и 3 (спецсимволы). Исправления применяются только при запуске с флагом `-fix`:

```bash
./loglint -fix ./...
```

## Конфигурация

Правила можно настраивать через YAML-файл. Передайте путь к файлу через флаг `-config`:

```bash
./loglint -config .loglint.yml ./...
```

Пример `.loglint.yml`:

```yaml
rules:
  lowercase: true
  english_only: true
  no_special_chars: false    # отключить проверку спецсимволов
  sensitive_data: true

# пользовательские ключевые слова для правила 4
sensitive_keywords:
  - password
  - secret
  - token
  - api_key
  - my_custom_keyword
```

По умолчанию все правила включены. Если `sensitive_keywords` не указаны, используется встроенный список: `password`, `pwd`, `secret`, `token`, `api_key`, `apikey`, `private_key`, `privatekey`, `access_key`, `accesskey`, `credential`, `bearer`, `session_id`.

## Сборка и запуск

### Требования

- Go 1.24+

### Standalone

```bash
# Сборка
go build -o loglint ./cmd/loglint

# Запуск на проекте
./loglint ./...

# Запуск с авто-исправлением
./loglint -fix ./...

# Запуск с конфигурацией
./loglint -config .loglint.yml ./...
```

### Интеграция с golangci-lint (Go Plugin)

1. Соберите плагин:

```bash
go build -buildmode=plugin -o loglint.so ./plugin/
```

2. Добавьте в `.golangci.yml`:

```yaml
linters-settings:
  custom:
    loglint:
      path: ./loglint.so
      description: "Checks log messages for style and security issues"
```

3. Запустите:

```bash
golangci-lint run
```

## CI/CD

Проект включает GitHub Actions workflow (`.github/workflows/ci.yml`), который автоматически запускается на push и pull request в `main`/`master`:

- **Build** — сборка всех пакетов
- **Build plugin** — сборка плагина для golangci-lint
- **Test** — запуск всех тестов
- **Vet** — статический анализ через `go vet`

## Тестирование

```bash
go test ./... -v
```

Тесты включают:

- **Unit-тесты** (`rules_test.go`) — проверка каждой функции валидации (`isUppercaseStart`, `hasNonEnglish`, `hasSpecialChars`, `stripSpecialChars`, `containsSensitiveKeyword`, `collectLits`, `hasNonLiteralParts`)
- **Интеграционные тесты** (`analyzer_test.go`) — проверка анализатора на тестовых файлах через `analysistest.Run` и `analysistest.RunWithSuggestedFixes`
- **Тесты конфигурации** (`config_test.go`) — загрузка, парсинг и дефолтные значения конфигурации

## Структура проекта

```
├── cmd/
│   └── loglint/
│       └── main.go              # Standalone CLI (singlechecker)
├── plugin/
│   └── plugin.go                # Плагин для golangci-lint
├── loglint/
│   ├── analyzer.go              # Определение анализатора, обнаружение вызовов логгеров
│   ├── rules.go                 # Функции валидации и проверки правил
│   ├── config.go                # Загрузка и парсинг YAML-конфигурации
│   ├── analyzer_test.go         # Интеграционные тесты (analysistest)
│   ├── rules_test.go            # Unit-тесты для функций валидации
│   ├── config_test.go           # Тесты конфигурации
│   └── testdata/
│       └── src/
│           ├── testcases/
│           │   ├── testcases.go         # Тестовые кейсы с // want аннотациями
│           │   └── testcases.go.golden  # Ожидаемый результат после авто-исправления
│           └── go.uber.org/
│               └── zap/
│                   └── zap.go           # Stub-пакет zap для тестов
├── .github/
│   └── workflows/
│       └── ci.yml               # GitHub Actions CI pipeline
├── go.mod
└── go.sum
```
