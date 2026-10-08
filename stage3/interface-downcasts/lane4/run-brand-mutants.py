#!/usr/bin/env python3
"""Run source-backend undefined-admission mutants through isolated Go overlays."""
import json, pathlib, subprocess, tempfile
repo = pathlib.Path(__file__).resolve().parents[3]
logs = pathlib.Path(__file__).resolve().parent / "logs"
results = []
for backend in ("native", "javascript"):
    production = repo / "internal" / backend / "view_unions_mixed.go"
    text = production.read_text()
    original = "return property.Of == ir.String && id > 0 && int(id) <= len(e.program.ViewContracts) && e.program.ViewContracts[id-1].Undefined && e.program.ViewContracts[id-1].Unsupported == \"\""
    assert text.count(original) == 1
    mutant = text.replace(original, "_ = id\n return false")
    with tempfile.TemporaryDirectory(prefix="brand-mutant-") as scratch:
        scratch = pathlib.Path(scratch)
        altered = scratch / "view_unions_mixed.go"
        altered.write_text(mutant)
        overlay = scratch / "overlay.json"
        overlay.write_text(json.dumps({"Replace": {str(production): str(altered)}}))
        log = logs / ("brand-mutant-" + backend + "-undefined.log")
        with log.open("w") as output:
            result = subprocess.run(["go", "test", "-overlay=" + str(overlay), "./internal/oracle", "-run", "^TestCheckedViewCompleteBrand$/^undefined$", "-count=1", "-v", "-timeout", "10m"], cwd=repo, stdout=output, stderr=subprocess.STDOUT)
        observation = log.read_text()
        assert result.returncode != 0, "undefined admission mutant escaped"
        assert "--- FAIL: TestCheckedViewCompleteBrand/undefined" in observation
        assert "clang failed" not in observation and "[build failed]" not in observation
        results.append({"backend": backend, "mutation": "omit declared undefined member admission", "caught": True, "log": str(log.relative_to(repo))})
(logs.parent / "brand-mutants.json").write_text(json.dumps(results, indent=2) + "\n")
