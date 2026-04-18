# md-to-html

Сервис конвертации Markdown в самодостаточный HTML. Полностью офлайн, без обращений к внешним API.

Текущая версия: `0.2.0` (Go + goldmark + templUI)

## Возможности

- GFM + footnote + emoji + подсветка кода через chroma.
- Якоря в заголовках с ASCII-транслитом: `## Установка` → `#ustanovka`.
- CLI: `md-to-html cli file.md`.
- HTTP API: `POST /convert` совместим с `v0.1.x`.
- Web UI на `http://localhost:8080/` с inline-preview в sandbox iframe и одноразовыми ссылками на preview/download.

## Запуск через Docker

```bash
docker run --rm -p 8080:8080 ghcr.io/fserg/md-to-html:latest
```

## Локальная разработка

Требования: Go 1.23+, `templ` CLI, Node.js для dev-режима Tailwind или standalone `tailwindcss`.

```bash
go install github.com/a-h/templ/cmd/templ@v0.3.1001
make tailwind
make build
./bin/md-to-html serve
```

Для live-reload:

```bash
make dev
```

## CLI

```bash
md-to-html cli file.md
md-to-html cli file.md -o out.html
md-to-html cli --stdin < file.md
md-to-html cli - --title "Заголовок"
```

## HTTP API

`POST /convert`

```bash
curl -X POST http://localhost:8080/convert \
  -H 'content-type: application/json' \
  -d '{"markdown":"# Привет"}'
```

Прочие эндпоинты:

- `GET /` — веб-интерфейс.
- `GET /health`, `GET /version`, `GET /ready` — служебные эндпоинты.
- `GET /preview/{id}`, `GET /download/{id}` — одноразовые ссылки из веб-формы.

## Env-переменные

| Переменная           | По умолчанию | Назначение |
|----------------------|--------------|------------|
| `ADDR`               | `:8080`      | Адрес прослушивания |
| `MAX_MARKDOWN_BYTES` | `1048576`    | Лимит размера markdown |
| `MAX_REQUEST_BYTES`  | `1200000`    | Лимит размера HTTP-запроса |
| `PREVIEW_TTL`        | `1h`         | TTL одноразовых ссылок |

## Миграция с v0.1.x

- API-контракт `POST /convert` не изменился, существующие клиенты продолжают работать.
- Якоря заголовков теперь используют ASCII-транслит. Ссылки вида `#установка` нужно заменить на `#ustanovka`.
- HTML-разметка упрощена: больше нет `<div class="markdown-heading">`, поэтому ручные CSS-оверрайды нужно пересмотреть.
- Переменная окружения `READY_CHECK_GITHUB` удалена: сервис больше не зависит от внешнего Markdown API.
- UI работает на том же порту `8080`, отдельный UI-порт `:8501` больше не нужен.

Python-реализация сохранена в `archive/`.

## Релизы

```bash
git commit -am "Release vX.Y.Z"
git tag vX.Y.Z
git push origin main --tags
```

GitHub Actions публикует Docker-образ для `linux/amd64` и `linux/arm64` в GHCR и прикладывает бинарники для `linux/amd64`, `linux/arm64` и `darwin/arm64` к GitHub Release.
