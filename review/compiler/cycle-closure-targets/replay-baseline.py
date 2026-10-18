from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
evidence = root / "review/compiler/cycle-closure-targets"
files = ["internal/ir/call_targets.go", "internal/ir/ir.go", "internal/lower/cycles.go"]
originals = {name: (root / name).read_bytes() for name in files}
try:
    for name in files:
        (root / name).write_bytes(subprocess.check_output(["git", "show", "88ec8b6e:" + name], cwd=root))
    with evidence.joinpath("base-build.log").open("w") as log:
        subprocess.run(["go", "build", "-o", "/tmp/adamic-cycle-base", "./cmd/adamic"], cwd=root, stdout=log, stderr=subprocess.STDOUT, check=True, timeout=120)
    with evidence.joinpath("red-replayed.log").open("w") as log:
        result = subprocess.run(["go", "test", "./internal/oracle", "-run", "^TestGettersCensus", "-count=1", "-timeout", "90s", "-parallel", "4", "-v"], cwd=root, stdout=log, stderr=subprocess.STDOUT, timeout=120)
    output = evidence.joinpath("red-replayed.log").read_text()
    if result.returncode != 1 or output.count("adamic/cycle-capable") < 10 or "build failed" in output:
        raise RuntimeError("baseline did not fail all ten fixtures on cycle membership")
    print("main baseline: all ten R2 fixtures red on cycle membership", flush=True)
finally:
    for name, data in originals.items():
        (root / name).write_bytes(data)
with evidence.joinpath("head-build.log").open("w") as log:
    subprocess.run(["go", "build", "-o", "/tmp/adamic-cycle-head", "./cmd/adamic"], cwd=root, stdout=log, stderr=subprocess.STDOUT, check=True, timeout=120)
print("base and head compilers built", flush=True)
