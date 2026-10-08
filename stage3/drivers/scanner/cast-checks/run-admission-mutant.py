"""Remove only optional-presence admission through a Go source overlay."""
from pathlib import Path
import json
import subprocess
import tempfile

root = Path(__file__).resolve().parents[4]
source = root / "internal/lower/view_optional_presence_cast.go"
original = source.read_text()
needle = "return changed\n"
if original.count(needle) != 1:
    raise SystemExit("optional-presence admission seam changed")
with tempfile.TemporaryDirectory(prefix="scanner-cast-mutant-") as scratch:
    directory = Path(scratch)
    mutant = directory / source.name
    mutant.write_text(original.replace(needle, "return changed && false\n"))
    overlay = directory / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {str(source): str(mutant)}}))
    log = Path("/tmp/scanner-cast-admission-mutant.log")
    with log.open("w") as output:
        result = subprocess.run(["go", "test", "-overlay", str(overlay), "./internal/lower", "-run", "^TestOptionalPresenceViewCast$", "-count=1", "-v", "-timeout", "10m"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
    text = log.read_text()
    if result.returncode == 0 or "--- FAIL: TestOptionalPresenceViewCast" not in text or "build failed" in text:
        raise SystemExit("admission mutant was not caught by the intended test: " + text)
    print("optional-presence admission omission caught: test exit", result.returncode)
    print("log:", log)
