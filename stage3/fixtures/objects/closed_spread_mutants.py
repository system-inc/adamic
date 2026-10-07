"""Prove the shape and NUL refusals fail when their guards are weakened."""
import json
import subprocess
import tempfile
from pathlib import Path

bucket = Path(__file__).resolve().parent
repository = bucket.parents[2]
target = repository / "internal/lower/object.go"
source = target.read_text()
needle = "\t\t\tnames, closed := l.closedSpreadNames(sourceNode, 0)\n\t\t\tif !closed {"
replacement = "\t\t\tnames, closed := l.closedSpreadNames(sourceNode, 0)\n\t\t\tif !closed {\n\t\t\t\tfor _, field := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(sourceNode)) {\n\t\t\t\t\tnames = append(names, field.Name)\n\t\t\t\t}\n\t\t\t\tclosed = true\n\t\t\t}\n\t\t\tif !closed {"
assert source.count(needle) == 1
mutants = [
    ("hidden_shape", source.replace(needle, replacement, 1), "TestSpreadHiddenFieldsStayRefused", "want soundness refusal, got <nil>"),
    ("nul_prefix", source.replace("strings.ContainsRune(name, 0)", "strings.ContainsRune(name, 1)"), "TestSpreadNulNamesStayNotYet", "want NUL lowering limit, got <nil>"),
    ("opaque_ownership", source.replace("from.IsReference() != into.IsReference()", "from.IsReference() == into.IsReference()"), "TestSpreadOpaqueRepresentationChangesStayNotYet", "want slot-ownership lowering limit, got <nil>"),
]
rows = []
for name, mutated, test, expected in mutants:
    assert mutated != source
    with tempfile.TemporaryDirectory(prefix="closed-spread-mutant-") as scratch:
        scratch = Path(scratch)
        go_file = scratch / "object.go"
        go_file.write_text(mutated)
        overlay = scratch / "overlay.json"
        overlay.write_text(json.dumps({"Replace": {str(target): str(go_file)}}))
        command = ["go", "test", "-overlay", str(overlay), "-count=1", "-timeout", "30m", "./internal/oracle", "-run", "^" + test + "$", "-v"]
        result = subprocess.run(command, cwd=repository, capture_output=True, text=True)
        killed = result.returncode != 0 and expected in result.stdout
        row = dict(mutant=name, check=test, exit=result.returncode, killed_by_intended_check=killed, stdout=result.stdout, stderr=result.stderr)
        rows.append(row)
        print(json.dumps(row), flush=True)
        if not killed:
            raise RuntimeError("mutant not killed by its intended refusal: " + name)
(bucket / "spread_mutant_proofs.json").write_text(json.dumps(rows, indent=2) + "\n")
