#!/usr/bin/env python3
"""Create a private staging folder. Does not deploy or contact GCP."""
import json
import os
from pathlib import Path
import shutil
import tempfile

from run_local import ROOT, local_values

values = local_values()
for key in values:
    values[key] = os.environ.get(key, values[key])
for key in ("ES_URL", "GCS_BUCKET", "JWT_SECRET"):
    if not values.get(key):
        raise SystemExit(f"Missing server configuration: {key}")
if len(values["JWT_SECRET"].encode()) < 32:
    raise SystemExit("JWT_SECRET must be at least 32 bytes")
stage = Path(tempfile.mkdtemp(prefix="socialai-teal-deploy-"))
stage.chmod(0o700)
for name in ("go.mod", "go.sum", "main.go"):
    shutil.copy2(ROOT / name, stage / name)
for folder in ("backend", "constants", "handler", "model", "service"):
    (stage / folder).mkdir()
    for source in (ROOT / folder).glob("*.go"):
        if not source.name.endswith("_test.go"):
            shutil.copy2(source, stage / folder / source.name)
# JSON strings are valid YAML scalars.
configuration = (ROOT / "app.yaml").read_text().rstrip() + "\nenv_variables:\n"
for key, value in values.items():
    if key not in {"HOST", "PORT"}:
        configuration += f"  {key}: {json.dumps(value)}\n"
(stage / "app.yaml").write_text(configuration)
(stage / "app.yaml").chmod(0o600)
print(stage)
print("Prepared locally only; no cloud deployment or traffic change was made.")
