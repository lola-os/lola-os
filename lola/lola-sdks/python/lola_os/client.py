"""The LolaCore client: spawns (or connects to) the lola-core binary and
speaks newline-delimited JSON-RPC 2.0 over its stdin/stdout.

A process-wide singleton is exposed via `get_client()` for convenience,
mirroring the blueprint's "one decorator, zero boilerplate" promise — most
callers never need to construct a LolaCore directly.
"""

from __future__ import annotations

import asyncio
import itertools
import json
import os
import subprocess
import threading
import time
from typing import Any, Callable, Dict, Iterator, Optional

from .config import GlobalConfig, get_global_config, resolve_binary_path
from .exceptions import LolaError, RPCConnectionError, from_rpc_error

_id_counter = itertools.count(1)


class LolaCore:
    """Manages a single lola-core subprocess and its JSON-RPC channel.

    Thread-safe for concurrent `call()` invocations: each call gets a
    unique request ID and waits on its own event, so multiple threads (or
    an asyncio event loop running calls via a thread executor) can share
    one subprocess without interleaving responses incorrectly.
    """

    def __init__(
        self,
        binary_path: Optional[str] = None,
        vault_passphrase: Optional[str] = None,
        extra_args: Optional[list] = None,
        auto_start: bool = True,
        startup_timeout: Optional[float] = None,
    ):
        cfg: GlobalConfig = get_global_config()
        self._binary_path = binary_path or cfg.binary_path
        self._vault_passphrase = vault_passphrase or cfg.vault_passphrase
        self._extra_args = extra_args or []
        self._startup_timeout = startup_timeout or cfg.startup_timeout

        self._proc: Optional[subprocess.Popen] = None
        self._stdout_thread: Optional[threading.Thread] = None
        self._log_thread: Optional[threading.Thread] = None
        self._lock = threading.Lock()
        self._pending: Dict[Any, "_PendingCall"] = {}
        self._log_subscribers: list = []
        self._log_subscribers_lock = threading.Lock()
        self._started = False

        if auto_start:
            self.start()

    # -- lifecycle -----------------------------------------------------

    def start(self) -> None:
        if self._started:
            return
        resolved_path = resolve_binary_path(self._binary_path)

        env = os.environ.copy()
        if self._vault_passphrase:
            env["LOLA_VAULT_PASSPHRASE"] = self._vault_passphrase

        args = [resolved_path, "serve", *self._extra_args]

        try:
            self._proc = subprocess.Popen(
                args,
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                env=env,
                text=True,
                bufsize=1,  # line-buffered
            )
        except OSError as exc:
            raise RPCConnectionError(
                f"Failed to start lola-core at {resolved_path}: {exc}"
            ) from exc

        self._stdout_thread = threading.Thread(target=self._read_stdout, daemon=True)
        self._stdout_thread.start()
        self._log_thread = threading.Thread(target=self._read_stderr_logs, daemon=True)
        self._log_thread.start()

        self._started = True
        self._wait_for_ready()

    def _wait_for_ready(self) -> None:
        """Best-effort readiness check: try a ping with retries, since
        lola-core may take a moment to bind/initialize."""
        deadline = time.time() + self._startup_timeout
        last_exc: Optional[Exception] = None
        while time.time() < deadline:
            if self._proc and self._proc.poll() is not None:
                stderr = ""
                if self._proc.stderr:
                    try:
                        stderr = self._proc.stderr.read()
                    except Exception:
                        pass
                raise RPCConnectionError(
                    f"lola-core exited immediately (code {self._proc.returncode}). "
                    f"stderr: {stderr[-2000:]}"
                )
            try:
                # A lightweight call with a short effective wait; any
                # response (even an error) proves the channel is alive.
                self._call_raw("budget_status", {}, timeout=1.5)
                return
            except Exception as exc:  # noqa: BLE001 - we retry regardless of cause
                last_exc = exc
                time.sleep(0.2)
        raise RPCConnectionError(
            f"lola-core did not become ready within {self._startup_timeout}s: {last_exc}"
        )

    def stop(self) -> None:
        if self._proc and self._proc.poll() is None:
            try:
                self._proc.terminate()
                self._proc.wait(timeout=5)
            except Exception:
                self._proc.kill()
        self._started = False

    def __enter__(self) -> "LolaCore":
        return self

    def __exit__(self, *exc_info) -> None:
        self.stop()

    def __del__(self):  # pragma: no cover - best effort cleanup
        try:
            self.stop()
        except Exception:
            pass

    # -- transport -----------------------------------------------------

    def _read_stdout(self) -> None:
        assert self._proc is not None and self._proc.stdout is not None
        for line in self._proc.stdout:
            line = line.strip()
            if not line:
                continue
            try:
                msg = json.loads(line)
            except json.JSONDecodeError:
                continue
            req_id = msg.get("id")
            with self._lock:
                pending = self._pending.pop(req_id, None)
            if pending is not None:
                pending.response = msg
                pending.event.set()

    def _read_stderr_logs(self) -> None:
        """lola-core writes structured logs to stderr. In `json` log
        format these are one JSON object per line; we forward parsed
        entries to any stream_logs() subscribers and otherwise discard
        them (rich-format text logs are not parsed)."""
        assert self._proc is not None and self._proc.stderr is not None
        for line in self._proc.stderr:
            line = line.rstrip("\n")
            if not line:
                continue
            entry = None
            try:
                entry = json.loads(line)
            except json.JSONDecodeError:
                entry = {"raw": line}
            with self._log_subscribers_lock:
                subscribers = list(self._log_subscribers)
            for q in subscribers:
                try:
                    q.put_nowait(entry)
                except Exception:
                    pass

    def subscribe_logs(self):
        """Returns a queue.Queue that receives parsed log entries. Used by
        stream_logs(). Callers should call unsubscribe_logs() when done."""
        import queue

        q: "queue.Queue" = queue.Queue()
        with self._log_subscribers_lock:
            self._log_subscribers.append(q)
        return q

    def unsubscribe_logs(self, q) -> None:
        with self._log_subscribers_lock:
            if q in self._log_subscribers:
                self._log_subscribers.remove(q)

    class _PendingCallHolder:
        pass

    def _call_raw(self, method: str, params: Dict[str, Any], timeout: float) -> Any:
        if not self._proc or self._proc.stdin is None:
            raise RPCConnectionError("lola-core process is not running")

        req_id = next(_id_counter)
        pending = _PendingCall()
        with self._lock:
            self._pending[req_id] = pending

        payload = json.dumps({"jsonrpc": "2.0", "id": req_id, "method": method, "params": params})
        try:
            self._proc.stdin.write(payload + "\n")
            self._proc.stdin.flush()
        except (BrokenPipeError, OSError) as exc:
            with self._lock:
                self._pending.pop(req_id, None)
            raise RPCConnectionError(f"lola-core pipe closed unexpectedly: {exc}") from exc

        if not pending.event.wait(timeout=timeout):
            with self._lock:
                self._pending.pop(req_id, None)
            raise RPCConnectionError(f"lola-core did not respond to '{method}' within {timeout}s")

        resp = pending.response
        if resp is None:
            raise RPCConnectionError(f"lola-core returned an empty response for '{method}'")
        if "error" in resp and resp["error"] is not None:
            err = resp["error"]
            raise from_rpc_error(err.get("message", "unknown error"), err.get("code"), err.get("data"))
        return resp.get("result")

    def call(self, method: str, params: Optional[Dict[str, Any]] = None, timeout: Optional[float] = None) -> Any:
        """Synchronous JSON-RPC call. Raises a LolaError subclass on
        failure."""
        cfg = get_global_config()
        return self._call_raw(method, params or {}, timeout or cfg.request_timeout)

    async def call_async(self, method: str, params: Optional[Dict[str, Any]] = None, timeout: Optional[float] = None) -> Any:
        """Async wrapper around call(), run in a thread pool executor so
        it doesn't block the event loop."""
        loop = asyncio.get_running_loop()
        return await loop.run_in_executor(None, lambda: self.call(method, params, timeout))


class _PendingCall:
    __slots__ = ("event", "response")

    def __init__(self):
        import threading as _threading

        self.event = _threading.Event()
        self.response: Optional[Dict[str, Any]] = None


# -- process-wide singleton ------------------------------------------------

_singleton: Optional[LolaCore] = None
_singleton_lock = threading.Lock()


def get_client(**kwargs) -> LolaCore:
    """Returns the process-wide LolaCore singleton, creating (and
    starting) it on first use. Pass kwargs only on the very first call;
    they are ignored on subsequent calls once the singleton exists."""
    global _singleton
    with _singleton_lock:
        if _singleton is None:
            _singleton = LolaCore(**kwargs)
        return _singleton


def reset_client() -> None:
    """Stops and clears the singleton client. Mainly useful for tests."""
    global _singleton
    with _singleton_lock:
        if _singleton is not None:
            _singleton.stop()
        _singleton = None
