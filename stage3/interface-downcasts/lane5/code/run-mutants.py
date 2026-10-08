#!/usr/bin/env python3
"""Remove each logical producer gate and require a semantic oracle failure."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[4]
evidence = root / "stage3/interface-downcasts/lane5/code/logs"
evidence.mkdir(exist_ok=True)
for backend in ("native", "javascript"):
    path = root / "internal" / backend / "view_callables_higher_order.go"
    original = path.read_text()
    marker = "{\n\tif property.ViewContract == 0 {"
    replacement = "{\n\treturn\n\tif property.ViewContract == 0 {" if backend == "native" else "{\n\treturn recorded\n\tif property.ViewContract == 0 {"
    assert original.count(marker) == 1
    try:
        path.write_text(original.replace(marker, replacement))
        log = evidence / (backend + "-logical-producer-mutant.log")
        with log.open("w") as output:
            result = subprocess.run(
                ["go", "test", "./internal/oracle", "-run", "^TestCheckedViewCallableCodeHigherOrder$", "-count=1", "-v", "-timeout", "10m"],
                cwd=root, env=dict(os.environ, ADAMIC_GATE_UNCACHED="1"), stdout=output, stderr=subprocess.STDOUT,
            )
        text = log.read_text()
        assert result.returncode != 0, f"{backend}: mutant survived"
        assert 'negative exit 0 stdout "completed\\n"' in text, f"{backend}: not a semantic failure: {log}"
        assert "[build failed]" not in text and "clang failed" not in text, f"{backend}: rejected build failure"
        print(f"{backend}: removed producer identity gate; same-representation wrong callback finished, caught by pinned exit 70")
    finally:
        path.write_text(original)
