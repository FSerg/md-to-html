import json
import os
import re
from functools import lru_cache
from html.parser import HTMLParser
from pathlib import Path
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

API_URL = "https://api.github.com/markdown"
API_VERSION = "2022-11-28"
TEMPLATE_PATH = Path(__file__).resolve().parent.parent / "template.html"


class FirstHeadingParser(HTMLParser):
    def __init__(self) -> None:
        super().__init__()
        self._capture = False
        self._done = False
        self._parts: list[str] = []

    def handle_starttag(self, tag: str, attrs) -> None:
        if self._done:
            return
        if tag in {"h1", "h2", "h3", "h4", "h5", "h6"}:
            self._capture = True

    def handle_endtag(self, tag: str) -> None:
        if self._capture and tag in {"h1", "h2", "h3", "h4", "h5", "h6"}:
            self._capture = False
            self._done = True

    def handle_data(self, data: str) -> None:
        if self._capture and not self._done:
            self._parts.append(data)

    def title(self) -> str:
        return "".join(self._parts).strip()


def render_markdown(markdown_text: str) -> str:
    payload = json.dumps({"text": markdown_text}).encode("utf-8")
    headers = {
        "Accept": "text/html",
        "Content-Type": "application/json",
        "User-Agent": "md-to-html-service",
        "X-GitHub-Api-Version": API_VERSION,
    }

    github_token = os.getenv("GITHUB_TOKEN")
    if github_token:
        headers["Authorization"] = f"Bearer {github_token}"

    request = Request(API_URL, data=payload, headers=headers, method="POST")
    try:
        with urlopen(request, timeout=30) as response:
            return response.read().decode("utf-8")
    except HTTPError as exc:
        error_body = exc.read().decode("utf-8", errors="replace")
        raise RuntimeError(
            f"GitHub API error: {exc.code} {exc.reason}\n{error_body}"
        ) from exc
    except URLError as exc:
        raise RuntimeError(f"Failed to reach GitHub API: {exc.reason}") from exc


def extract_title(html_text: str, fallback: str) -> str:
    parser = FirstHeadingParser()
    parser.feed(html_text)
    return parser.title() or fallback


def apply_template(template_text: str, html_text: str, title: str) -> str:
    updated = re.sub(
        r"<title>.*?</title>",
        f"<title>{title}</title>",
        template_text,
        flags=re.DOTALL,
    )
    output_lines = []
    inserted = False
    html_lines = [f"        {line}" if line else "" for line in html_text.splitlines()]
    for line in updated.splitlines():
        if not inserted and "Markdown -->" in line:
            output_lines.extend(html_lines)
            inserted = True
            continue
        output_lines.append(line)
    if not inserted:
        raise RuntimeError("Template placeholder not found.")
    return "\n".join(output_lines) + "\n"


@lru_cache(maxsize=1)
def load_template_text() -> str:
    return TEMPLATE_PATH.read_text(encoding="utf-8")


def convert(markdown_text: str, fallback_title: str = "Document") -> str:
    html_text = render_markdown(markdown_text)
    title = extract_title(html_text, fallback_title)
    template_text = load_template_text()
    return apply_template(template_text, html_text, title)
