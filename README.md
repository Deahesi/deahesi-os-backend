# Backend

Взаимодействие с ИИ реализовано на Go SDK Genkit v1.13.1 через официальный
плагин OpenRouter. Один экземпляр Genkit создаётся при запуске приложения;
генерация и стриминг выполняются в flow `sendMessage`.

## Запуск

Скопируйте `.env.example` в `.env` и задайте `OPENROUTER_API_KEY`.
Существующий `OPEN_ROUTER_API_KEY` также поддерживается; если заданы оба,
приоритет у `OPENROUTER_API_KEY`.

- `LLM_MODEL`: ID модели OpenRouter, например `qwen/qwen3.8-27b:free`.
  Допускается префикс Genkit: `openrouter/qwen/qwen3.8-27b:free`.
- `LLM_SYSTEM_PROMPT`: системные инструкции модели.
- `OPENROUTER_BASE_URL`: необязательный альтернативный endpoint.

```powershell
go run ./cmd/server
```

Для просмотра flow и трассировок через установленный Genkit CLI:

```powershell
genkit start -- go run ./cmd/server
```

## API

`POST /ai/message` принимает `multipart/form-data`: обязательное поле `text`
и необязательный `file`. Ответ передаётся через SSE в прежнем формате:

```text
data: {"delta":{"content":"Привет","role":"assistant"},"finish_reason":null,"index":0}

data: {"delta":{},"finish_reason":"stop","index":0}
```

Рассуждения модели передаются в `delta.reasoning`. Ошибка до начала стрима
возвращается как JSON с HTTP 500; ошибка после начала — как SSE-сообщение
`{"code":"internal_error","message":"Internal Error"}`. Некорректный ввод
возвращает HTTP 400. Отключение клиента или ошибка записи отменяют генерацию.

В Genkit v1.13.1 плагин `compat_oai` преобразует все media parts в `image_url`.
Адаптер `internal/ai/files.go` восстанавливает формат OpenRouter `file`
с исходным именем и data URL, сохраняя передачу документов и изображений.
Поддержка конкретного типа файла зависит от выбранной модели и провайдера.

## Проверка

```powershell
go test ./...
go vet ./...
go build ./...
```

Интеграционные тесты используют настоящий Genkit и локальный mock OpenRouter:
проверяют запросы, вложения, SSE, ошибки и отмену. Ключ API и внешние запросы
для тестов не нужны.

Документация: [Genkit для Go](https://genkit.dev/docs/go/get-started/),
[плагин OpenRouter](https://genkit.dev/docs/integrations/openrouter/).
