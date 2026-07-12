"""Global SDK configuration: where the lola-core binary lives, default
connection settings, and the platform/arch detection used by binary
management.
"""

from __future__ import annotations

import os
import platform
import sys
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional


@dataclass
class GlobalConfig:
    """Process-wide defaults. Mutating this directly is supported but
    `lola_os.override()` (see context.py) is preferred for scoped changes."""

    binary_path: Optional[str] = None
    vault_passphrase: Optional[str] = field(
        default_factory=lambda: os.environ.get("LOLA_VAULT_PASSPHRASE")
    )
    startup_timeout: float = 15.0
    request_timeout: float = 60.0
    auto_start: bool = True


_global_config = GlobalConfig()


def get_global_config() -> GlobalConfig:
    return _global_config


def platform_binary_name() -> str:
    """Returns the expected lola-core binary filename for the current
    platform/architecture, matching the naming convention produced by
    lola-core's cross-compilation targets (see lola-core/README.md)."""
    system = platform.system().lower()
    machine = platform.machine().lower()

    if machine in ("x86_64", "amd64"):
        arch = "amd64"
    elif machine in ("arm64", "aarch64"):
        arch = "arm64"
    else:
        arch = machine

    if system == "darwin":
        return f"lola-darwin-{arch}"
    if system == "windows":
        return f"lola-windows-{arch}.exe"
    return f"lola-linux-{arch}"


def bundled_binary_path() -> Path:
    """Path to a binary bundled inside the installed wheel under
    lola_os/bin/, used as a fallback if no system-wide `lola` binary is
    found and nothing was downloaded."""
    return Path(__file__).parent / "bin" / platform_binary_name()


def resolve_binary_path(explicit: Optional[str] = None) -> str:
    """Resolution order:

    1. An explicit path passed to LolaCore(binary_path=...) or set via
       GlobalConfig.binary_path.
    2. The LOLA_CORE_BINARY environment variable.
    3. A `lola` binary bundled inside the installed wheel (lola_os/bin/).
    4. A `lola` binary found on $PATH.

    Raises FileNotFoundError with an actionable message if none is found.
    """
    candidates = []
    if explicit:
        candidates.append(explicit)
    if _global_config.binary_path:
        candidates.append(_global_config.binary_path)
    env_path = os.environ.get("LOLA_CORE_BINARY")
    if env_path:
        candidates.append(env_path)

    bundled = bundled_binary_path()
    candidates.append(str(bundled))

    for path_str in candidates:
        p = Path(path_str)
        if p.is_file() and os.access(p, os.X_OK):
            return str(p)

    # Fall back to $PATH lookup.
    import shutil

    found = shutil.which("lola")
    if found:
        return found

    raise FileNotFoundError(
        "Could not locate a lola-core binary.\n"
        f"Looked for: {', '.join(candidates)}, and 'lola' on $PATH.\n\n"
        "Fix this by either:\n"
        "  1. Building lola-core yourself: `cd lola-core && go build -o bin/lola ./cmd/lola`\n"
        "     then set LOLA_CORE_BINARY=/path/to/bin/lola, or\n"
        "  2. Placing a prebuilt binary at "
        f"{bundled} (this wheel ships without one — see lola-sdks/python/README.md), or\n"
        "  3. Installing `lola` on your $PATH."
    )
