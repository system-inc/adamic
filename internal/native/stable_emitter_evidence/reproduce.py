# Run from the repository root; all test output goes to scratch logs.
from pathlib import Path
import json
import os
import subprocess

root = Path.cwd()
scratch = Path("/tmp/stable-emitter-reproduce")
scratch.mkdir(exist_ok=True)
emitter = root / "internal/native/emit.go"
for label, line in [("temporary", "\te.temporaries = 0\n"), ("cache", "\te.caches = 0\n"), ("region-facts", "\te.regionValues = map[string]bool{}\n")]:
    source = scratch / (label + ".go")
    assert line in emitter.read_text()
    source.write_text(emitter.read_text().replace(line, ""))
    overlay = scratch / (label + ".json")
    overlay.write_text(json.dumps({"Replace": {str(emitter): str(source)}}))
    package = "./internal/native"
    pattern = "^TestFunctionCountersDoNotMove$"
    if label == "region-facts":
        package = "./internal/oracle"
        pattern = "TestNativeAgreesWithNode/internal/oracle/testdata/regions.a$"
    with (scratch / (label + ".log")).open("w") as log:
        result = subprocess.run(["go", "test", "-overlay=" + str(overlay), package, "-run", pattern, "-count=1", "-timeout", "30m"], stdout=log, stderr=subprocess.STDOUT, env={**os.environ, "ADAMIC_GATE_UNCACHED": "1"})
    text = (scratch / (label + ".log")).read_text()
    expected = {"temporary": "moved second body", "cache": "cache moved", "region-facts": "heap-use-after-free"}[label]
    assert result.returncode != 0 and expected in text, (label, result.returncode)
    print(label, "caught", "exit", result.returncode)

# The reporting probe embeds the unmodified splitter through splitC from developer-tools.
probe = root / "internal/native/unit_changes_probe_test.go"
assert not probe.exists()
probe.write_text((root / "internal/native/stable_emitter_evidence/unit_changes_probe.go.txt").read_text())
try:
    replacements = {}
    for filename in ["emit.go", "emit_objects.go"]:
        source = scratch / ("before-" + filename)
        source.write_bytes(subprocess.check_output(["git", "show", "52f9baea1651bf52af5267b556ebf525504655bb:internal/native/" + filename]))
        replacements[str(root / "internal/native" / filename)] = str(source)
    overlay = scratch / "before.json"
    overlay.write_text(json.dumps({"Replace": replacements}))
    for label, arguments in [("before", ["-overlay=" + str(overlay)]), ("after", [])]:
        with (scratch / ("units-" + label + ".log")).open("w") as log:
            result = subprocess.run(["go", "test", *arguments, "./internal/native", "-run", "^TestReportLintUnitChanges$", "-count=1", "-v"], stdout=log, stderr=subprocess.STDOUT)
        assert result.returncode == 0
        print((scratch / ("units-" + label + ".log")).read_text())
finally:
    probe.unlink()
