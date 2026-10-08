from pathlib import Path
import subprocess
p = Path("internal/lower/optional_widening.go")
original = p.read_text()
try:
    p.write_text(original.replace("source = reducedOptionalSource(l.checker, source)", "// mutant: judge the unreduced source"))
    with open("/tmp/optional-reduced-mutant.log", "w") as out:
        result = subprocess.run(["go", "test", "./internal/lower", "-run", "^TestOptionalWideningReducedSource$", "-count=1"], stdout=out, stderr=subprocess.STDOUT)
    log = Path("/tmp/optional-reduced-mutant.log").read_text()
    assert result.returncode != 0 and "absent from structural source never" in log, log
finally:
    p.write_text(original)
