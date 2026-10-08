"""Run this rule's mutant as the sanitized native canary without editing shared tests."""
import json
import subprocess
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[6]
SOURCE = ROOT / "stage1/cohere/lint/lint_test.go"
OLD = 'const nativeCanaryRule = "react/jsx-no-comment-textnodes"'
NEW = 'const nativeCanaryRule = "no-dupe-else-if"'
with tempfile.TemporaryDirectory(prefix="no-dupe-native-mutant-") as directory:
    directory = Path(directory)
    text = SOURCE.read_text()
    assert text.count(OLD) == 1, "shared native canary anchor changed"
    patched = directory / "lint_test.go"
    patched.write_text(text.replace(OLD, NEW))
    overlay = directory / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {str(SOURCE): str(patched)}}))
    subprocess.run([
        "go", "test", "-overlay=" + str(overlay), "./stage1/cohere/lint",
        "-run", "^TestMutants/duplicate-else-if-keeps-covered-group$",
        "-count=1", "-v", "-timeout=10m",
    ], cwd=ROOT, check=True)
