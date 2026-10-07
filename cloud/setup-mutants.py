#!/usr/bin/env python3
"""Drop each warming-key component independently and require its regression to fail."""
import os
from pathlib import Path
import subprocess
import tempfile

source = Path(__file__).resolve().parent
scratch = Path(tempfile.mkdtemp(prefix="setup-mutants-"))
print("mutant logs:", scratch)
original = (source / "setup-key.py").read_text()
for component in ["head", "sums", "version", "environment", "packages", "cache", "mode"]:
    lines = original.splitlines(keepends=True)
    mutant = "".join(line for line in lines if not line.lstrip().startswith(f'"{component}":'))
    assert mutant != original
    module = scratch / f"drop-{component}.py"
    module.write_text(mutant)
    environment = dict(os.environ, ADAMIC_SETUP_KEY_MODULE=str(module))
    environment.pop("ADAMIC_SETUP_INTEGRATION", None)
    log = scratch / f"drop-{component}.log"
    with log.open("wb") as output:
        result = subprocess.run(["python3", str(source / "test_setup.py"), "WarmingKey"],
                                env=environment, stdout=output, stderr=subprocess.STDOUT)
    text = log.read_text()
    if result.returncode != 1 or f"component='{component}'" not in text:
        raise SystemExit(f"mutant {component} escaped or failed for the wrong reason: {log}")
    print(f"drop {component}: caught by test_every_component_invalidates (exit 1)")
