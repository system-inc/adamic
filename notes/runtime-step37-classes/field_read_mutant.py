#!/usr/bin/env python3
"""Prove the constructor-shared-field refusal pin catches lost field provenance."""

from pathlib import Path
import subprocess
import tempfile

repository = Path(__file__).resolve().parents[2]
target = repository / "internal/lower/parallel_private.go"
original = target.read_bytes()
old = b"""\t\t\tw.union(&result, value)
\t\t}
\t\treturn result, nil
\t}
\tif node.Kind == ast.KindElementAccessExpression {"""
changed = b"""\t\t\tw.union(&result, value)
\t\t}
\t\t// Mutant: a private parent incorrectly erases its field's shared origin.
\t\tif !receiver.shared && len(receiver.objects) > 0 {
\t\t\tresult.shared = false
\t\t}
\t\treturn result, nil
\t}
\tif node.Kind == ast.KindElementAccessExpression {"""
if original.count(old) != 1:
    raise SystemExit("missing or ambiguous private-field-read mutant seam")

logs = Path(tempfile.mkdtemp(prefix="step37-field-mutant-"))
command = [
    "go", "test", "./internal/oracle",
    "-run", "^TestConcurrencyRefusals/task_private_shared_field.a$",
    "-count=1", "-v", "-timeout", "10m",
]


def check(name):
    log = logs / (name + ".log")
    with log.open("wb") as output:
        result = subprocess.run(command, cwd=repository, stdout=output,
                                stderr=subprocess.STDOUT, timeout=660)
    return result.returncode, log.read_text()


baseline, _ = check("baseline")
if baseline != 0:
    raise SystemExit(f"baseline refusal pin failed; logs: {logs}")
try:
    target.write_bytes(original.replace(old, changed, 1))
    status, output = check("drop-shared-field-mark")
finally:
    target.write_bytes(original)

# A build error or unrelated failing test cannot certify the refusal pin.
if status == 0 or "--- FAIL: TestConcurrencyRefusals/task_private_shared_field.a" not in output:
    raise SystemExit(f"field-read mutant escaped the refusal pin; logs: {logs}")
if "want Refused" not in output and "want exact What" not in output:
    raise SystemExit(f"mutant failed for an unrelated reason; logs: {logs}")
restored, _ = check("restored")
if restored != 0:
    raise SystemExit(f"restored refusal pin failed; logs: {logs}")
print(f"drop-shared-field-mark caught by the refusal pin; logs: {logs}")
