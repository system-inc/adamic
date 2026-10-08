#!/usr/bin/env python3
"""Remove the capture proof; the refusal assertion must catch acceptance."""
from pathlib import Path
import subprocess
import tempfile
import json
root = Path(__file__).resolve().parents[3]
source = root / "internal/lower/library_regex_callback_shape.go"
original = source.read_text()
before = "return n.Kind != regex.Capturing && present(n.Body)"
assert original.count(before) == 1
log = Path("/tmp/library-small-regex-capture-proof-mutant.log")
with tempfile.TemporaryDirectory(prefix="regex-proof-mutant-") as directory:
    replacement = Path(directory) / source.name
    replacement.write_text(original.replace(before, "return present(n.Body)"))
    overlay = Path(directory) / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {str(source): str(replacement)}}))
    with log.open("w") as output:
        result = subprocess.run(["go", "test", "-overlay", str(overlay), "./internal/lower", "-run", "^TestLibraryRegexOffsetRequiresNoCaptures$", "-count=1", "-v"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
    text = log.read_text()
    assert result.returncode != 0 and "capture-free producer proof admitted an unproven pattern" in text, text[-4000:]
    assert "build failed" not in text, text[-4000:]
    print("capture-proof: caught by refusal assertion; " + str(log))
