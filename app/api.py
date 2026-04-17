import os
from typing import Any
from urllib.error import URLError
from urllib.request import Request, urlopen

from fastapi import FastAPI, HTTPException, Request as FastAPIRequest, Response
from fastapi.exceptions import RequestValidationError
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from pydantic import BaseModel, ConfigDict, field_validator

from app.converter import convert, load_template_text

DEFAULT_MAX_MARKDOWN_BYTES = 1_048_576
DEFAULT_MAX_REQUEST_BYTES = 1_200_000


def get_int_env(name: str, default: int) -> int:
    raw_value = os.getenv(name)
    if raw_value is None:
        return default
    try:
        value = int(raw_value)
    except ValueError as exc:
        raise RuntimeError(f"{name} must be an integer.") from exc
    if value <= 0:
        raise RuntimeError(f"{name} must be positive.")
    return value


def get_bool_env(name: str, default: bool = False) -> bool:
    raw_value = os.getenv(name)
    if raw_value is None:
        return default
    return raw_value.strip().lower() in {"1", "true", "yes", "on"}


class ConvertRequest(BaseModel):
    model_config = ConfigDict(extra="forbid")

    markdown: str
    title: str | None = None

    @field_validator("markdown")
    @classmethod
    def validate_markdown_size(cls, value: str) -> str:
        max_markdown_bytes = get_int_env(
            "MAX_MARKDOWN_BYTES", DEFAULT_MAX_MARKDOWN_BYTES
        )
        if len(value.encode("utf-8")) > max_markdown_bytes:
            raise HTTPException(
                status_code=413,
                detail=f"markdown exceeds {max_markdown_bytes} bytes",
            )
        return value


class MaxRequestSizeMiddleware:
    def __init__(self, app: Any, max_request_bytes: int) -> None:
        self.app = app
        self.max_request_bytes = max_request_bytes

    async def __call__(self, scope, receive, send) -> None:
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return

        headers = {
            key.decode("latin1").lower(): value.decode("latin1")
            for key, value in scope.get("headers", [])
        }
        content_length = headers.get("content-length")
        if content_length:
            try:
                if int(content_length) > self.max_request_bytes:
                    await self._send_413(scope, receive, send)
                    return
            except ValueError:
                pass

        body = bytearray()
        while True:
            message = await receive()
            if message["type"] != "http.request":
                if message["type"] == "http.disconnect":
                    return
                continue

            chunk = message.get("body", b"")
            body.extend(chunk)
            if len(body) > self.max_request_bytes:
                await self._send_413(scope, receive, send)
                return

            if not message.get("more_body", False):
                break

        body_bytes = bytes(body)
        body_sent = False

        async def replay_receive():
            nonlocal body_sent
            if body_sent:
                return {"type": "http.request", "body": b"", "more_body": False}
            body_sent = True
            return {"type": "http.request", "body": body_bytes, "more_body": False}

        await self.app(scope, replay_receive, send)

    async def _send_413(self, scope, receive, send) -> None:
        response = JSONResponse(
            status_code=413,
            content={"detail": f"request exceeds {self.max_request_bytes} bytes"},
        )
        await response(scope, receive, send)


app = FastAPI(title="md-to-html")
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["POST", "GET"],
    allow_headers=["content-type"],
)
app.add_middleware(
    MaxRequestSizeMiddleware,
    max_request_bytes=get_int_env("MAX_REQUEST_BYTES", DEFAULT_MAX_REQUEST_BYTES),
)


@app.exception_handler(RequestValidationError)
async def request_validation_exception_handler(
    request: FastAPIRequest, exc: RequestValidationError
) -> JSONResponse:
    return JSONResponse(status_code=400, content={"detail": exc.errors()})


@app.post("/convert")
async def convert_markdown(payload: ConvertRequest) -> Response:
    if not payload.markdown.strip():
        raise HTTPException(status_code=400, detail="markdown must not be empty")

    fallback_title = payload.title or "Document"
    try:
        html_result = convert(payload.markdown, fallback_title=fallback_title)
    except RuntimeError as exc:
        raise HTTPException(status_code=502, detail=str(exc)) from exc

    return Response(content=html_result, media_type="text/html; charset=utf-8")


@app.get("/health")
async def health() -> dict[str, str]:
    return {"status": "ok"}


@app.get("/ready")
async def ready() -> dict[str, Any]:
    details: dict[str, Any] = {"status": "ok", "template_loaded": True}

    try:
        load_template_text()
    except Exception as exc:
        raise HTTPException(status_code=503, detail=f"Template load failed: {exc}") from exc

    if get_bool_env("READY_CHECK_GITHUB", default=False):
        request = Request(
            "https://api.github.com",
            headers={"User-Agent": "md-to-html-service-readiness"},
            method="HEAD",
        )
        try:
            with urlopen(request, timeout=5) as response:
                details["github_status"] = response.status
        except URLError as exc:
            raise HTTPException(
                status_code=503,
                detail=f"GitHub readiness check failed: {exc.reason}",
            ) from exc
    else:
        details["github_status"] = "skipped"

    return details
