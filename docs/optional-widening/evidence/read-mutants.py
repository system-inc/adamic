from pathlib import Path
import subprocess
p = Path("internal/lower/interface_cast.go")
original = p.read_text()
try:
    p.write_text(original.replace("l.result.CheckedFields[field] = true", "l.result.CheckedFields[field] = false"))
    with open("/tmp/optional-read-mutant.log", "w") as out:
        result = subprocess.run(["go", "test", "./internal/oracle", "-run", "^TestOptionalCheckedReads$", "-count=1"], stdout=out, stderr=subprocess.STDOUT)
    log = Path("/tmp/optional-read-mutant.log").read_text()
    assert result.returncode != 0 and "exit codes differ" in log and "build failed" not in log, log
finally:
    p.write_text(original)
