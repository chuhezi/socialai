#!/usr/bin/env python3
"""Load server configuration without evaluating it as shell code."""
import os
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
ALLOWED = {"ES_URL", "ES_USERNAME", "ES_PASSWORD", "GCS_BUCKET", "JWT_SECRET",
           "OPENAI_API_KEY", "OPENAI_IMAGE_MODEL", "ALLOWED_ORIGINS", "PORT", "HOST"}


def local_values():
    values = {}
    source = ROOT / ".env.local"
    if source.exists():
        for line in source.read_text().splitlines():
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            key, separator, value = line.partition("=")
            if separator and key.strip() in ALLOWED:
                value = value.strip()
                if len(value) >= 2 and value[0] == value[-1] and value[0] in (chr(34), chr(39)):
                    value = value[1:-1]
                values[key.strip()] = value
    return values


if __name__ == "__main__":
    for key, value in local_values().items():
        os.environ.setdefault(key, value)
    os.environ.setdefault("HOST", "127.0.0.1")
    os.chdir(ROOT)
    os.execvp("go", ["go", "run", "."])
