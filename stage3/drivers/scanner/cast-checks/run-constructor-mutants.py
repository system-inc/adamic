"""Omit constructor admission and each single-producer backend check independently."""
from pathlib import Path
import json
import subprocess
import tempfile

root = Path(__file__).resolve().parents[4]
mutations = [
    ("admission", "internal/lower/cast_proof.go", "if l.scannerStringConstructorCast(node) {", "if false && l.scannerStringConstructorCast(node) {", "./internal/lower", "TestScannerStringConstructorView"),
    ("native-producer", "internal/native/view_unions_untagged.go", "root.Kind == ir.ViewCallable && root.ProducerCertified", "false && root.Kind == ir.ViewCallable && root.ProducerCertified", "./internal/oracle", "TestScannerStringConstructorProducerMutant"),
    ("javascript-producer", "internal/javascript/view_unions_untagged.go", "root.Kind == ir.ViewCallable && root.ProducerCertified", "false && root.Kind == ir.ViewCallable && root.ProducerCertified", "./internal/oracle", "TestScannerStringConstructorProducerMutant"),
]
for name, path, needle, replacement, package, test in mutations:
    source = root / path
    original = source.read_text()
    if original.count(needle) != 1:
        raise SystemExit("constructor mutant seam changed: " + name)
    with tempfile.TemporaryDirectory(prefix="scanner-constructor-mutant-") as scratch:
        directory = Path(scratch)
        mutant = directory / source.name
        mutant.write_text(original.replace(needle, replacement))
        overlay = directory / "overlay.json"
        overlay.write_text(json.dumps({"Replace": {str(source): str(mutant)}}))
        log = Path("/tmp/scanner-constructor-" + name + "-mutant.log")
        with log.open("w") as output:
            result = subprocess.run(["go", "test", "-overlay", str(overlay), package, "-run", "^" + test + "$", "-count=1", "-v", "-timeout", "10m"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        text = log.read_text()
        if result.returncode == 0 or "--- FAIL: " + test not in text or "build failed" in text:
            raise SystemExit("mutant not caught by intended test: " + name + "\n" + text)
        print(name + " omission caught: named test exit " + str(result.returncode))
        print("log:", log)
