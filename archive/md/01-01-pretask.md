Описание Github API конвертера markdown в HTML: https://docs.github.com/en/rest/markdown/markdown?apiVersion=2022-11-28

Пример вызова API:
```bash
curl -L \
  -X POST \
  -H "Accept: text/html" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  https://api.github.com/markdown \
  -d '{"text":"## Title 2\nHello **world**"}'
```
Ответ:
```html
<div class="markdown-heading"><h2 class="heading-element">Title 2</h2><a id="user-content-title-2" class="anchor" aria-label="Permalink: Title 2" href="#title-2"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<p>Hello <strong>world</strong></p>
```
Нужен простой python скрипт, который будет:
1. Принимать на вход путь к markdown файлу `..path/example.md`
2. Через Github API конвертировать его в html
3. Формировать новый html файл по шаблону `md\template.html` и сохранять результат рядом в `..path/example.html`

===

У меня есть готовый пайтон скрипт md_to_html.py, который умеет конвертировать markdown в html с помощью Github API. 

Мне нужно переделать его в простое Streamlit приложение, которое будет иметь следующий интерфейс:
1. Поле для загрузки markdown файла
2. Кнопка для конвертации
3. Поле для отображения результата в виде HTML и возможность скачать результат в виде HTML файла

Так же хочу этот проект запускать в докере.

Было бы классно еще иметь один публичный API ендпоинт, который будет принимать markdown текст и возвращать html результат, чтобы можно было использовать этот сервис в других приложениях.

Задай вопросы, если что-то не понятно или есть неоднозночности или неопределенности.

Создай репозиторий на GitHub для этого проекта.
Введи версии релизов.
Настрой на гитхаб Actions для автоматической сборки и публикации докер образа при каждом релизе.


