#!/usr/bin/env python3
"""Summarize preserved unit evidence without treating a partial gate as green."""
import collections, datetime, json, pathlib, re, statistics, sys
root = pathlib.Path(sys.argv[1] if len(sys.argv) > 1 else "/tmp")
def events(name):
    result=[]
    for line in (root / name).read_text().splitlines():
        try: result.append(json.loads(line))
        except ValueError: continue
    return result
cases=events("wave28-owned-cases.jsonl")
recovered=events("wave28-recovered-cases.jsonl")
mutants=events("wave28-owned-mutants-witnesses.jsonl")
if (root / "wave28-jsx-fragments-mutant-rerun.jsonl").exists():
    mutants += events("wave28-jsx-fragments-mutant-rerun.jsonl")
bridge=events("wave28-bridge-release-sanitizer.jsonl")
normal=collections.Counter()
refused=collections.Counter()
for event in cases:
    test=event.get("Test", "")
    if "/case-" in test and event["Action"] in ("pass", "fail"):
        name=test.split("/",1)[1].rsplit("/case-",1)[0]
        (normal if event["Action"]=="pass" else refused)[name]+=1
recovery=collections.Counter()
for event in recovered:
    test=event.get("Test", "")
    if "/row-" in test and event["Action"]=="pass":
        recovery[test.split("/",1)[1].rsplit("/row-",1)[0]]+=1
costs=collections.defaultdict(lambda: {"Go":[], "native":[]})
def seconds(value):
    factor = {"ns":1e-9,"µs":1e-6,"ms":1e-3,"s":1.0}
    number, unit=re.fullmatch(r"([0-9.]+)(ns|µs|ms|s)",value).groups()
    return float(number)*factor[unit]
for event in cases:
    match=re.search(r"Go program and lint=(\S+) native program, lint and recording=(\S+)",event.get("Output", ""))
    if match:
        name=event["Test"].split("/",1)[1].rsplit("/case-",1)[0]
        costs[name]["Go"].append(seconds(match[1]))
        costs[name]["native"].append(seconds(match[2]))
summary={
 "normal_case_passes": dict(normal), "normal_mode_refusals":dict(refused),
 "recovered_case_passes":dict(recovery),
 "unique_cases_compared":sum(normal.values())+sum(recovery.values()),
 "caught_rule_mutants":[event["Test"] for event in mutants if event["Action"]=="pass" and event.get("Test","").startswith("TestMutants/")],
 "owned_witnesses_passed":any(event["Action"]=="pass" and event.get("Test")=="TestOwnedWitnesses" for event in mutants),
 "capture_count_mutant_caught":any("captured 22 cases, want 23" in event.get("Output", "") for event in events("wave28-owned-count-mutant.jsonl")),
 "bridge_passed":any(event["Action"]=="pass" and event.get("Test")=="TestBridge" for event in bridge),
 "case_runtime_medians_seconds":{name:{side:statistics.median(samples) for side,samples in values.items()} for name,values in costs.items()},
 "performance_scope":"Per-case whole Go program+lint versus ASan/UBSan native program+lint+recording; concurrent compilation, not an isolated throughput benchmark.",
 "full_package":json.loads((root/"wave28-checker-full-stopped.json").read_text()),
}
print(json.dumps(summary,indent=2))
