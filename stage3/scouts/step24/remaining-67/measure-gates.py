#!/usr/bin/env python3
"""Sequential isolated oracles, followed by unchanged lane evidence checking.
Trees already have independently measured, successful complete applies.
"""
import json, shutil, subprocess, sys, time
from pathlib import Path
unit = Path(__file__).resolve().parent
repo = unit.parents[3]
cache = Path(sys.argv[1]); tools = cache / "remaining67-baseline-tools"
cases = (
    ("main", cache / "remaining67-main-clean", tools / "stage3", "031a1259bc7973934792dc6cb1bd4074fc2204b9"),
    ("extended", cache / "remaining67-final-lane/adapted-tree", repo / "stage3", subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=repo, text=True).strip()),
)
if len(sys.argv) > 2:
    assert sys.argv[2] == "fresh-extended"
    cases = [("extended-fresh", *cases[1][1:])]
for label, tree, runner, commit in cases:
    output = cache / ("remaining67-isolated-" + label)
    output.mkdir(exist_ok=False)
    (output / "adapted-tree").symlink_to(tree, target_is_directory=True)
    shutil.copyfile(tree / "patch-set.md", output / "patch-set.md")
    cache_note = "existing cache state retained"
    if label == "extended-fresh":
        perf = tree / ".parallelperf.json"
        assert perf.is_file()
        shutil.copyfile(perf, output / "performance-cache-before.json")
        perf.unlink()
        cache_note = "removed transient upstream timing cache to match clean-main initial file-size batching"
    start = time.monotonic()
    with (output / "oracle.log").open("w") as log:
        result = subprocess.run(["bash", str(runner / "oracle/run.sh"), str(tree), str(output / "oracle")], stdout=log, stderr=subprocess.STDOUT)
    platform = json.loads(subprocess.check_output(["node", "-p", "JSON.stringify({os:process.platform,arch:process.arch,node:process.version})"], text=True))
    execution = dict(commit=commit, apply_exit=0, oracle_exit=result.returncode, platform=platform,
                     oracle_wall_seconds=round(time.monotonic() - start, 3),
                     method="successful independently measured full apply; fresh isolated full oracle; unchanged lane checker replay",
                     performance_cache=cache_note,
                     apply_log=str(cache / "remaining67-final-lane/apply.log") if label != "main" else "/tmp/remaining67-main-clean-apply.log")
    (output / "execution.json").write_text(json.dumps(execution, indent=2) + "\n")
    with (output / "lane.log").open("w") as log:
        checked = subprocess.run(["python3", str(runner / "lane/check.py"), str(output)], stdout=log, stderr=subprocess.STDOUT)
    (output / "exits.json").write_text(json.dumps(dict(oracle=result.returncode, lane=checked.returncode)) + "\n")
    print(label, "oracle", result.returncode, "lane", checked.returncode, flush=True)
print("DONE", flush=True)
