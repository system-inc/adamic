"""Keep per-check and per-execution evidence from the fifty-run Go logs."""
import collections
import json
import pathlib
import re
import statistics
import sys

logs = pathlib.Path(sys.argv[1] if len(sys.argv) > 1 else "/workspace")
races = {"skip_items_share", "lazy_cache", "plain_shared_count", "field_cache", "remote_free"}
phases = {}
for phase, filename in [("baseline", "baseline"), ("baseline_loaded", "baseline-loaded"),
                        ("intermediate", "after"), ("cache_window", "cache-window"),
                        ("final", "final-rates")]:
    source = (logs / f"race-reliable-{filename}.log").read_text()
    checks = collections.defaultdict(list)
    executions = collections.defaultdict(list)
    first_report = {}
    current = None
    for line in source.splitlines():
        case = re.match(r"=== RUN\s+TestParallel(?:ChecksCatchMutants|ScalingGuardMutants)/(\w+)", line)
        if case:
            current = case[1]
        match = re.search(r"--- (PASS|FAIL): TestParallel(?:ChecksCatchMutants|ScalingGuardMutants)/(\w+) \(([\d.]+)s\)", line)
        if match:
            checks[match[2]].append({"caught": match[1] == "PASS", "check_seconds": float(match[3])})
        timing = re.search(r"mutant attempt=(\d+) elapsed=(\S+)", line)
        if timing:
            executions[current].append({"attempt": int(timing[1]), "elapsed": timing[2], "caught": False})
        report = re.search(r"(\w+) caught: (.*)", line)
        if report:
            first_report.setdefault(report[1], report[2])
            if phase == "baseline" or phase == "baseline_loaded":
                executions[report[1]].append({"attempt": 1, "caught": True})
            else:
                executions[report[1]][-1]["caught"] = True
        if "mutant not caught" in line and phase.startswith("baseline"):
            executions[current].append({"attempt": 1, "caught": False})
    phases[phase] = {
        "log": f"race-reliable-{filename}.log",
        "controls": {name: {"checks": values, "caught": sum(row["caught"] for row in values),
                             "total": len(values), "tsan": name in races,
                             "executions": executions[name], "first_report": first_report.get(name)}
                     for name, values in checks.items()}}

samples = collections.defaultdict(list)
current = None
for line in (logs / "race-reliable-harness-times.log").read_text().splitlines():
    if line.startswith("=== RUN") or line.startswith("=== NAME"):
        current = line.split()[-1]
    match = re.search(r"harness threads=(\S+) arguments=\[(.*?)\] runs=(\d+) elapsed=(\S+)", line)
    if not match:
        continue
    duration = match[4]
    for suffix, scale in [("ms", 0.001), ("µs", 0.000001), ("ns", 0.000000001), ("s", 1)]:
        if duration.endswith(suffix):
            seconds = float(duration[:-len(suffix)]) * scale
            break
    samples[(current, match[1], match[2], int(match[3]))].append(seconds)

final = phases["final"]["controls"]
assert len(final) == 13
assert all(row["total"] == row["caught"] == 50 for row in final.values())
assert all(len(final[name]["executions"]) == 150 and
           all(row["caught"] for row in final[name]["executions"]) for name in races)
assert all(len(values) == 5 for values in samples.values())
objects = json.loads((logs / "race-reliable-objects-final.json").read_text())
assert len(objects["objects"]) == 47 and all(row["identical"] for row in objects["objects"])
mutant = json.loads((logs / "race-reliable-object-mutant.json").read_text())
assert [row["file"] for row in mutant["objects"] if not row["identical"]] == ["parallel.c"]
result = {
    "parent": "0f672843d211fd40b7ba84d3207d8eac0772ac02",
    "implementation": "dbb48b3",
    "machine": {"platform": "Linux amd64", "processor": "AMD EPYC 9V74", "nproc": 5,
                "cpu_max": "400000 100000", "memory_gb": 17.6,
                "clang": "20.1.8", "go": "1.27.1", "node": "24.19.0",
                "observed_load_one_minute_range": [1.67, 12.21],
                "harness_timing_load_approximate": "9 to 12, concurrent gates",
                "setup_seconds": 53},
    "tsan_options": "halt_on_error=1:history_size=4:report_atomic_races=1",
    "perturbation": "ADAMIC_TSAN_PERTURB=1; 100 microseconds, first 64 visits per point per thread",
    "phases": phases,
    "harness_times": [{"test": key[0], "threads": key[1], "arguments": key[2], "executions_per_check": key[3],
                       "seconds": values, "best_seconds": min(values), "median_seconds": statistics.median(values)}
                      for key, values in samples.items()],
    "objects": objects,
    "object_identity_mutant": "fixed grain: only parallel.c object differs; script exits 1",
    "gates": {"native_seconds": 231.176, "uncached_oracle_seconds": 343.959,
              "vet": "passed", "gofmt": "no files", "final_fifty_run_seconds": 654.329},
}
print(json.dumps(result, indent=2))
