"""Compare registration of identical existing .ts and .a source in scratch."""
import pathlib
import shutil
import subprocess
import tempfile

repository = pathlib.Path(__file__).resolve().parents[5]
with tempfile.TemporaryDirectory(prefix="lint-wave1-15-") as scratch:
    root = pathlib.Path(scratch)
    shutil.copytree(repository / "stage1/cohere/lint/rules", root / "rules")
    command = ["go", "run", "./cmd/lint-registry", "-root", str(root)]
    baseline = subprocess.run(command, cwd=repository, capture_output=True, text=True)
    print(f"existing .ts baseline: exit={baseline.returncode}")
    print(baseline.stdout, baseline.stderr, sep="", end="")
    assert baseline.returncode == 0
    module = root / "rules/no-debugger/rule.ts"
    module.rename(module.with_suffix(".a"))
    renamed = subprocess.run(command, cwd=repository, capture_output=True, text=True)
    print(f"same source renamed .a: exit={renamed.returncode}")
    print(renamed.stdout, renamed.stderr, sep="", end="")
    assert renamed.returncode != 0
    assert "rule.ts: no such file or directory" in renamed.stderr
