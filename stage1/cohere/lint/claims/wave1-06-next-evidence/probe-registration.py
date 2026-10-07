"""A positive control and an identical-source .a registration refusal."""
from pathlib import Path
import os
import shutil
import subprocess
import tempfile

repository = Path(__file__).resolve().parents[5]
with tempfile.TemporaryDirectory(prefix="wave1-06-registration-") as scratch:
    root = Path(scratch)
    shutil.copytree(repository / "stage1/cohere/lint/rules", root / "rules")
    command = ["go", "run", "./cmd/lint-registry", "-root", str(root)]
    for extension in ("ts", "a"):
        if extension == "a":
            module = root / "rules/no-debugger/rule.ts"
            module.rename(module.with_suffix(".a"))
        with (root / f"{extension}.stdout").open("w") as out, (root / f"{extension}.stderr").open("w") as err:
            result = subprocess.run(command, cwd=repository, stdout=out, stderr=err)
        stdout = (root / f"{extension}.stdout").read_text()
        stderr = (root / f"{extension}.stderr").read_text()
        print(f"identical source, .{extension}: exit={result.returncode}")
        print(stdout, stderr, sep="", end="")
        if extension == "ts":
            assert result.returncode == 0
        elif os.environ.get("WAVE06_EXPECT_A") == "accepted":
            assert result.returncode == 0
            assert "no-debugger/rule.a" in (root / ".generated/registry.ts").read_text()
        else:
            assert result.returncode != 0
            assert "rule.ts: no such file or directory" in stderr
print("PASS: expected registration outcomes observed")
