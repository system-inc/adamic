import difflib
import os
from pathlib import Path
import subprocess

runtime = Path("internal/native/runtime/closure.c")
original = runtime.read_bytes()
source = original.decode()
before = "return value->view != NULL ? value->view->underlying : value;"
assert source.count(before) == 1
mutated = source.replace(before, "return value; /* mutant: compare adapters rather than their root */", 1)
evidence = Path("review/compiler/views-v4")
(evidence / "identity-root-omission.patch").write_text("".join(difflib.unified_diff(source.splitlines(True), mutated.splitlines(True), fromfile="a/" + str(runtime), tofile="b/" + str(runtime), n=0)))
try:
    runtime.write_text(mutated)
    with (evidence / "identity-root-omission.log").open("w") as log:
        result = subprocess.run(["go", "test", "./internal/oracle", "-run", "^TestV4EscapeAdapterIdentity$", "-count=1", "-v", "-timeout", "90s"], stdout=log, stderr=subprocess.STDOUT, timeout=90, env={**os.environ, "ADAMIC_GATE_UNCACHED": "1"})
    output = (evidence / "identity-root-omission.log").read_text()
    assert result.returncode != 0 and "--- FAIL: TestV4EscapeAdapterIdentity" in output and "stdout differs" in output, output
    assert "AddressSanitizer:" not in output and "LeakSanitizer:" not in output and "error:" not in output, output
    print("actual runtime root-normalization omission caught by Node stdout, without compile or sanitizer failure")
finally:
    runtime.write_bytes(original)
assert runtime.read_bytes() == original
print("closure.c restored byte-for-byte")
