# md-to-html

Сервис конвертации Markdown в самодостаточный HTML. Конвертация выполняется локально, без внешних API.

![Превью интерфейса](screen.png)

Текущая версия: `0.2.2` (Go + goldmark + templUI)

## Возможности

- GFM + footnote + emoji + подсветка кода через chroma.
- Web UI на `http://localhost:8080/` с загрузкой файла или вставкой текста, HTMX-обновлением результата и одноразовыми ссылками на preview/download.
- CLI: `md-to-html cli file.md`.
- HTTP API: `POST /convert`, совместим с `v0.1.x`.
- Якоря в заголовках с ASCII-транслитом: `## Установка` → `#ustanovka`.

## Запуск через Docker

```bash
docker run --rm -p 8080:8080 ghcr.io/fserg/md-to-html:latest
```

## Быстрый старт

```bash
go install github.com/a-h/templ/cmd/templ@v0.3.1001
npm install
make build
./bin/md-to-html serve
```

## Локальная разработка

Требования: Go 1.24+, Node.js, `templ` CLI.

```bash
go install github.com/a-h/templ/cmd/templ@v0.3.1001
npm install
make tailwind
make build
./bin/md-to-html serve
```

Для live-reload:

```bash
make dev
```

## Релизная сборка

Локальный release-билд для текущей платформы:

```bash
make release
```

Скрипт:
- генерирует `templ`-код
- собирает Tailwind bundle
- прогоняет `go test ./...`
- собирает release-бинарь с версией из `VERSION`
- кладёт артефакты в `dist/`

Проверка готового release-билда:

```bash
./dist/md-to-html-$(go env GOOS)-$(go env GOARCH) serve
```

Сборка всех release-таргетов как в CI:

```bash
make release-all
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
