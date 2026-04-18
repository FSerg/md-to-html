import threading
import uuid
from collections import OrderedDict
from html.parser import HTMLParser
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
import sys

import streamlit as st

try:
    from app.converter import convert
    from app.version import __version__
except ModuleNotFoundError:
    sys.path.append(str(Path(__file__).resolve().parent.parent))
    from app.converter import convert
    from app.version import __version__

MAX_PREVIEW_STORE_ITEMS = 20


class BodyInnerHTMLParser(HTMLParser):
    def __init__(self) -> None:
        super().__init__(convert_charrefs=False)
        self._inside_body = False
        self._depth = 0
        self._parts: list[str] = []

    def handle_starttag(self, tag: str, attrs) -> None:
        rendered = self.get_starttag_text()
        if tag == "body":
            self._inside_body = True
            self._depth = 0
            return
        if self._inside_body and rendered is not None:
            self._parts.append(rendered)
            self._depth += 1

    def handle_endtag(self, tag: str) -> None:
        if tag == "body" and self._inside_body:
            self._inside_body = False
            self._depth = 0
            return
        if self._inside_body:
            self._parts.append(f"</{tag}>")
            if self._depth > 0:
                self._depth -= 1

    def handle_startendtag(self, tag: str, attrs) -> None:
        if self._inside_body:
            rendered = self.get_starttag_text()
            if rendered is not None:
                self._parts.append(rendered)

    def handle_data(self, data: str) -> None:
        if self._inside_body:
            self._parts.append(data)

    def handle_entityref(self, name: str) -> None:
        if self._inside_body:
            self._parts.append(f"&{name};")

    def handle_charref(self, name: str) -> None:
        if self._inside_body:
            self._parts.append(f"&#{name};")

    def handle_comment(self, data: str) -> None:
        if self._inside_body:
            self._parts.append(f"<!--{data}-->")

    def body_html(self) -> str:
        return "".join(self._parts).strip()


def extract_body_html(document_html: str) -> str:
    parser = BodyInnerHTMLParser()
    parser.feed(document_html)
    parser.close()
    return parser.body_html()


@st.cache_resource
def get_preview_runtime() -> dict[str, object]:
    store: OrderedDict[str, str] = OrderedDict()
    lock = threading.Lock()

    class PreviewHandler(BaseHTTPRequestHandler):
        def do_GET(self) -> None:
            prefix = "/preview/"
            if not self.path.startswith(prefix):
                self.send_error(404)
                return

            preview_id = self.path[len(prefix) :].split("?", 1)[0]
            with lock:
                document_html = store.get(preview_id)

            if document_html is None:
                self.send_error(404)
                return

            payload = document_html.encode("utf-8")
            self.send_response(200)
            self.send_header("Content-Type", "text/html; charset=utf-8")
            self.send_header("Content-Length", str(len(payload)))
            self.send_header("Cache-Control", "no-store")
            self.end_headers()
            self.wfile.write(payload)

        def log_message(self, format: str, *args) -> None:
            return

    server = ThreadingHTTPServer(("127.0.0.1", 0), PreviewHandler)
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    return {
        "base_url": f"http://127.0.0.1:{server.server_port}",
        "store": store,
        "lock": lock,
    }


def register_preview(document_html: str) -> str:
    runtime = get_preview_runtime()
    preview_id = uuid.uuid4().hex
    store = runtime["store"]
    lock = runtime["lock"]

    with lock:
        store[preview_id] = document_html
        while len(store) > MAX_PREVIEW_STORE_ITEMS:
            store.popitem(last=False)

    return f"{runtime['base_url']}/preview/{preview_id}"


st.set_page_config(
    page_title="Markdown to HTML",
    page_icon=":material/description:",
    layout="centered",
)

if "html_result" not in st.session_state:
    st.session_state["html_result"] = None
if "output_name" not in st.session_state:
    st.session_state["output_name"] = "document.html"
if "preview_url" not in st.session_state:
    st.session_state["preview_url"] = None

st.title("Markdown → HTML")
st.caption(
    f"Версия {__version__}. Загрузите markdown-файл или вставьте текст, проверьте превью и скачайте готовый HTML."
)

input_mode = st.segmented_control(
    "Источник Markdown",
    options=["Файл", "Текст"],
    default="Файл",
)

uploaded_file = None
pasted_markdown = ""

if input_mode == "Файл":
    uploaded_file = st.file_uploader(
        "Загрузите .md файл",
        type=["md", "markdown"],
    )
else:
    pasted_markdown = st.text_area(
        "Вставьте Markdown из буфера обмена",
        placeholder="# Заголовок\n\nВставьте сюда markdown-текст.",
        height=260,
    )

html_result = st.session_state["html_result"]
is_convert_disabled = (
    uploaded_file is None if input_mode == "Файл" else not pasted_markdown.strip()
)

with st.container(border=True):
    action_col, preview_col, download_col = st.columns(
        [1.1, 1, 1],
        vertical_alignment="center",
    )

    with action_col:
        convert_clicked = st.button(
            "Конвертировать",
            disabled=is_convert_disabled,
            type="primary",
            icon=":material/auto_awesome:",
            use_container_width=True,
        )

    with preview_col:
        if html_result and st.session_state["preview_url"] is not None:
            st.link_button(
                "Открыть превью",
                url=st.session_state["preview_url"],
                icon=":material/open_in_new:",
                use_container_width=True,
            )
        else:
            st.button(
                "Открыть превью",
                disabled=True,
                icon=":material/open_in_new:",
                use_container_width=True,
            )

    with download_col:
        if html_result:
            st.download_button(
                "Скачать HTML",
                data=html_result,
                file_name=st.session_state["output_name"],
                mime="text/html",
                icon=":material/download:",
                use_container_width=True,
            )
        else:
            st.button(
                "Скачать HTML",
                disabled=True,
                icon=":material/download:",
                use_container_width=True,
            )

    if html_result:
        st.caption(":green-badge[Результат готов]")
    else:
        st.caption("После конвертации здесь появятся действия с готовым файлом.")

if convert_clicked and not is_convert_disabled:
    if input_mode == "Файл":
        markdown_bytes = uploaded_file.getvalue()
        markdown_text = markdown_bytes.decode("utf-8")
        fallback_title = Path(uploaded_file.name).stem or "Document"
        output_name = f"{fallback_title}.html"
    else:
        markdown_text = pasted_markdown
        fallback_title = "Document"
        output_name = "document.html"

    try:
        st.session_state["html_result"] = convert(
            markdown_text,
            fallback_title=fallback_title,
        )
        st.session_state["output_name"] = output_name
        st.session_state["preview_url"] = register_preview(st.session_state["html_result"])
        st.rerun()
    except RuntimeError as exc:
        st.session_state["html_result"] = None
        st.session_state["preview_url"] = None
        st.error(str(exc))

html_result = st.session_state["html_result"]
if html_result:
    body_html = extract_body_html(html_result)

    with st.container(border=True):
        st.caption(
            "Inline-превью без стилей. Для точного вида — «Открыть превью» или скачайте файл."
        )
        st.markdown(body_html, unsafe_allow_html=True)

    with st.expander("Показать исходный HTML", icon=":material/code:"):
        st.code(html_result, language="html")
