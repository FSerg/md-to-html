# Прогресс миграции Python → Go

Источник истины по статусу фаз. Обновляется после каждого завершённого шага.

- Общий план: [plan-go-migration.md](plan-go-migration.md)
- Универсальный промпт для запуска фазы: [execute-phase-prompt.md](execute-phase-prompt.md)

## Статус

| #  | Фаза                                                | Статус       | Начата     | Завершена  | Commit/PR | Заметки |
|----|------------------------------------------------------|--------------|------------|------------|-----------|---------|
| 0  | [Архивирование Python](phases/phase-0-archive.md)    | ✅ done       | 2026-04-18 | 2026-04-18 | 425eae7   |         |
| 1  | [Go-скелет](phases/phase-1-skeleton.md)              | ✅ done       | 2026-04-18 | 2026-04-18 | 6b8d588   |         |
| 2  | [Converter (goldmark)](phases/phase-2-converter.md)  | ✅ done       | 2026-04-18 | 2026-04-18 | 8deba36   | Golden fixtures use relative/email links to keep generated HTML free of external resource URLs. |
| 3  | [HTTP-сервер](phases/phase-3-server.md)              | ✅ done       | 2026-04-18 | 2026-04-18 | 843d8dc   |         |
| 4  | [UI на templUI](phases/phase-4-ui.md)                | ✅ done       | 2026-04-18 | 2026-04-18 | d6aef55   |         |
| 5  | [CLI-подкоманда](phases/phase-5-cli.md)              | 🔄 in_progress | 2026-04-18 | —          | —         |         |
| 6  | [Docker + CI](phases/phase-6-docker-ci.md)           | ⏳ pending    | —          | —          | —         |         |
| 7  | [Документация + v0.2.0](phases/phase-7-docs.md)      | ⏳ pending    | —          | —          | —         |         |

Легенда статусов:
- ⏳ `pending` — не начата
- 🔄 `in_progress` — в работе
- ✅ `done` — завершена, acceptance criteria выполнены
- ⚠️ `blocked` — заблокирована, см. заметки

## Инварианты между фазами

- `git status` чист перед началом каждой фазы.
- Каждая фаза завершается отдельным commit в `main` (или PR с мёрджем). Сообщение в формате `phaseN: <краткое описание>`.
- Acceptance criteria фазы проверяются до смены статуса на `done`.
- Любое отклонение от плана документируется в колонке «Заметки» с ссылкой на commit.

## Лог ключевых решений (ADR lite)

| Дата       | Решение | Обоснование |
|------------|---------|-------------|
| 2026-04-18 | Goldmark + chroma inline + extension.Footnote + кастомный anchor-extender | См. `plan-go-migration.md` §11 |
| 2026-04-18 | ASCII-транслит id заголовков через `mozillazg/go-unidecode` | Решение пользователя (round-1) |
| 2026-04-18 | One-shot preview/download с UUIDv4 + TTL 1 ч | Решение пользователя (round-1) |
| 2026-04-18 | GitHub-style prefix-anchor (`<a>` как первый child `<h>`), не wrap-anchor | Закрытие F-01 round-3 — избегаем nested `<a>` |
| 2026-04-18 | `extractHeadingText` walker вместо deprecated `BaseNode.Text(src)` | Закрытие F-02 round-3 |
| 2026-04-18 | `<iframe sandbox srcdoc>` без `allow-same-origin` вместо `bluemonday` для inline preview | Меньше зависимостей, полная изоляция |
| 2026-04-18 | `POST /convert` сохраняется (не `/api/convert`), UI-форма на `POST /ui/convert` | Паритет API-контракта |
| 2026-04-18 | `html.WithUnsafe()` выключен; `parser.WithAttribute()` выключен | Безопасность + паритет |
| 2026-04-18 | Tailwind standalone binary в Docker (без Node) | Упрощение multi-stage build |
