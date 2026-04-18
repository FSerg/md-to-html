1. Функциональный паритет с GitHub API. Сейчас HTML от GitHub даёт: таблицы, task-list, strikethrough, autolinks, footnotes, подсветку кода, emoji       
  :name:, и главное — обёртки <div class="markdown-heading"> с якорями <a class="anchor"> (на них завязан CSS в template.html). Что нужно сохранить 1-в-1? 
  - a) Полный GFM (goldmark поддерживает через extension.GFM) — да/нет?                                                                                    
  - b) Подсветку кода chroma встроить в <pre><code>? Или оставить «просто теги» без классов?                                                               
  - c) Emoji-shortcodes (yuin/goldmark-emoji)?                                              
  - d) Обёртки heading’ов с якорями (делается через abhinav/goldmark-anchor или кастомный renderer). Если убрать — придётся править CSS в шаблоне.         
  - e) Frontmatter (---) — парсить/игнорировать/использовать для title?                                                                                    
                                                                                                                                                           
  2. Архитектура Go-приложения. Предлагаю один бинарник с подкомандами:                                                                                    
  - serve — единый HTTP-сервер: / и /convert (форма на templUI, HTMX-превью), /api/convert, /preview/{id}, /health, /version, /ready, /download/{id}.      
  - cli <file.md> — режим CLI (заменяет md_to_html.py).                                                                                                    
                                                                                                                                                           
     Подходит, или надо разделить два бинарника (api и ui)?                                                                                                
                                                                                                                                                           
  3. templUI-стек. templUI = templ + Tailwind + Alpine.js + HTMX. Подтвердите:                                                                             
  - a) Tailwind CSS сборку (tailwindcss CLI) встраиваем в Docker/Makefile?                                                                                 
  - b) HTMX для live-превью (без полной перезагрузки) — желательно?                                                                                        
  - c) Тёмная тема / языковой переключатель — нужны или оставляем русский-only как сейчас?                              
                                                                                                                                                           
  4. Структура репозитория. Ваш вариант — archive/ для Python, корень для Go. Подтвердите:                                                                 
  /archive/        # текущий Python-проект целиком                                                                                                         
  /cmd/md-to-html/ # main.go                                                                                                                               
  /internal/converter/   # goldmark-рендер                                                                                                                 
  /internal/server/      # HTTP handlers                                                                                                                   
  /internal/ui/          # .templ файлы (templUI components)                                                                                               
  /web/static/           # Tailwind output, favicon                                                                                                        
  /template/             # самодостаточный HTML-шаблон для итога                                                                                           
  go.mod / Makefile / Dockerfile                                                                                                                           
     Go module path — github.com/fserg/md-to-html?                                                                                                         
                                                                                                                                                           
  5. Шаблон итогового HTML. template.html сейчас рассчитан на GitHub-разметку (классы .markdown-heading, .heading-element, .anchor). Варианты:             
  - a) Сохранить визуал идентично → реализовать goldmark-renderer, генерирующий ту же разметку.                                                            
  - b) Упростить шаблон под «чистый» goldmark-вывод (меньше кода, чуть другой вид h2/h3).                                                                  
                                                                                                                                                           
     Какой?                                                                                                                                                
                                                                                                                        
  6. Версионирование и релизы. Сейчас v0.1.2, GitHub Actions собирает Docker в GHCR. После переписывания:                                                  
  - a) Бампнуть до v0.2.0 (или v1.0.0)?                                                                                 
  - b) CI: заменить на Go-сборку (тесты + cross-compile linux/amd64,arm64, darwin/arm64) + Docker multi-stage?                                             
  - c) Публиковать бинарники в GitHub Releases?                                                                         
                                                                                                                                                           
  7. Нефункциональное.                                                                                                                                     
  - Лимиты MAX_MARKDOWN_BYTES, MAX_REQUEST_BYTES — переносим (envs)?                                                                                       
  - Go 1.23+?                                                                                                                                              
  - Тесты: golden-файлы (MD→HTML diff против эталона) + smoke-тесты HTTP?                                               
  - air / templ generate --watch для dev-режима? 

# Ответы
1. Полный GFM и хотелось бы подсветку кода. Шаблон можно править как угодно под новый рендер. Якоря в заголовках хотелось бы сохранить, так как они полезны для навигации по документу.

2. Один бинарник, включая cli режим.

3. templUI можно встраивать в бинарник? Лайв-превью было бы круто, но не критично. Тёмная тема и языковой переключатель не нужны, так как целевая аудитория русскоязычная.

4. archive/ для Python, корень для Go

5. Шаблон поменяй под новый проект

6. После перехода на Go предлагаю бампнуть до v0.2.0

7. Про лимиты не знаю, на твое усмотрение и Го - на твой выбор. Тесты с golden-файлами звучат отлично, а для dev-режима air / templ generate --watch будет удобно.

===

Сохрани подробный план как md/02-01-plan.md
