"""Add the owned sanitized-native comparison through a temporary test overlay."""
import json
import subprocess
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[6]
SOURCE = ROOT / "stage1/cohere/lint/lint_test.go"
OLD = 'sides = []runtimeSide{{"Node", node(t, directory, path, false)}, {"emitted JavaScript", emittedNode(t, directory, path, false)}}'
NEW = OLD + '\n\t\t\t\tsides = append(sides, runtimeSide{"sanitized native", execute(t, "", buildMutantPort(t, directory), "--manifest", path)})'
with tempfile.TemporaryDirectory(prefix="one-var-native-mutant-") as directory:
    directory = Path(directory)
    text = SOURCE.read_text()
    assert text.count(OLD) == 1, "shared syntax-mutant comparison anchor changed"
    patched = directory / "lint_test.go"
    patched.write_text(text.replace(OLD, NEW))
    overlay = directory / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {str(SOURCE): str(patched)}}))
    subprocess.run([
        "go", "test", "-overlay=" + str(overlay), "./stage1/cohere/lint",
        "-run", "^TestMutants$/^one-var-combine-message$",
        "-count=1", "-v", "-timeout=15m",
    ], cwd=ROOT, check=True)
