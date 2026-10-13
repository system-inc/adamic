#!/usr/bin/env python3
"""Run only the reported gate's a-check policy on this unit's Adamic sources."""
import argparse
import json
import pathlib
import subprocess
import types

parser = argparse.ArgumentParser()
parser.add_argument("--tree", default=str(pathlib.Path(__file__).resolve().parents[2]))
parser.add_argument("--out", required=True)
parser.add_argument("--tools-sha", default="3e339bbf06e0695f1b1223065ee8d4eef4813ba8")
arguments = parser.parse_args()
tree = pathlib.Path(arguments.tree).resolve()
output = pathlib.Path(arguments.out)
output.mkdir(parents=True, exist_ok=True)
source = subprocess.check_output(["git", "-C", str(tree), "show", arguments.tools_sha + ":cloud/fast-gate/run.py"], text=True)
namespace = {"__name__": "checked_write_gate_tools"}
exec(compile(source, "gate-tools/cloud/fast-gate/run.py", "exec"), namespace)
runner_type = next(value for value in namespace.values() if isinstance(value, type) and hasattr(value, "aCheck"))
runner = object.__new__(runner_type)
runner.arguments = types.SimpleNamespace(tree=str(tree))
runner.result, runner.steps, runner.exits = {}, {}, {}
failures = []

def step(name, command):
    with (output / (name + ".log")).open("w") as log:
        result = subprocess.run(command, cwd=tree, stdout=log, stderr=subprocess.STDOUT)
    return result.returncode == 0

runner.step = step
runner.spawn = lambda command, stdout, stderr: subprocess.Popen(command, cwd=tree, stdout=stdout, stderr=stderr, text=True)
runner.fail = lambda name, message: failures.append({"step": name, "message": message})
paths = sorted(str(path.relative_to(tree)) for path in (tree / "stage3/checked-writes").glob("*.a"))
runner.aCheck(paths)
report = {"tools_sha": arguments.tools_sha, "a_check": runner.result.get("a_check", {}), "exits": runner.exits, "failures": failures}
(output / "a-check.json").write_text(json.dumps(report, indent=2) + "\n")
print(json.dumps(report, indent=2))
raise SystemExit(0 if runner.exits.get("a-check") == 0 else 1)
