import signal
import subprocess
import sys
from pathlib import Path

GRACEFUL_TIMEOUT_SECONDS = 10


def build_processes() -> list[subprocess.Popen[bytes]]:
    root = Path(__file__).resolve().parent
    return [
        subprocess.Popen(
            [
                sys.executable,
                "-m",
                "uvicorn",
                "app.api:app",
                "--host",
                "0.0.0.0",
                "--port",
                "8000",
            ],
            cwd=root,
        ),
        subprocess.Popen(
            [
                sys.executable,
                "-m",
                "streamlit",
                "run",
                "app/streamlit_app.py",
                "--server.port",
                "8501",
                "--server.address",
                "0.0.0.0",
                "--server.headless",
                "true",
                "--browser.gatherUsageStats",
                "false",
            ],
            cwd=root,
        ),
    ]


def stop_processes(processes: list[subprocess.Popen[bytes]], skip_pid: int | None = None) -> None:
    for process in processes:
        if process.pid == skip_pid:
            continue
        if process.poll() is None:
            process.terminate()


def reap_processes(
    processes: list[subprocess.Popen[bytes]], skip_pid: int | None = None
) -> None:
    for process in processes:
        if process.pid == skip_pid:
            continue
        if process.poll() is not None:
            continue
        try:
            process.wait(timeout=GRACEFUL_TIMEOUT_SECONDS)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait()


def main() -> int:
    processes = build_processes()
    exit_code = 0
    shutting_down = False

    def handle_signal(signum, _frame) -> None:
        nonlocal exit_code, shutting_down
        if shutting_down:
            return
        shutting_down = True
        exit_code = 128 + signum
        stop_processes(processes)

    signal.signal(signal.SIGTERM, handle_signal)
    signal.signal(signal.SIGINT, handle_signal)

    while True:
        try:
            pid, status = os_wait()
        except ChildProcessError:
            break
        except InterruptedError:
            continue

        process = next((item for item in processes if item.pid == pid), None)
        if process is not None:
            process.returncode = os_waitstatus_to_exitcode(status)

        if not shutting_down:
            exit_code = os_waitstatus_to_exitcode(status)
            shutting_down = True
            stop_processes(processes, skip_pid=pid)

        reap_processes(processes, skip_pid=pid)
        break

    reap_processes(processes)
    return exit_code


def os_wait() -> tuple[int, int]:
    import os

    return os.wait()


def os_waitstatus_to_exitcode(status: int) -> int:
    import os

    return os.waitstatus_to_exitcode(status)


if __name__ == "__main__":
    raise SystemExit(main())
