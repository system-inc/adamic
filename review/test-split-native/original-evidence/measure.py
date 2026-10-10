"""Measure exact Go test selectors; keep test observations out of build caches."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import time

from budget import check, leaves

ROOT = Path(__file__).resolve().parents[3]
PACKAGES = ["internal/native", "internal/javascript"]
CORPUS_SHARDS = {"TestDecodeASCII": 90, "TestRegExpBytecodeRandomNode": 40,
                 "TestNormalizeMatchesNode": 18, "TestWASI": 36,
                 "TestSplitTSGoAgrees": 2, "TestRecordMutants": 6}


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def run(package, name, selector, output, cache, go_cache, shard=None, unit=False):
    environment = os.environ.copy()
    environment.update(GOMAXPROCS="4", ADAMIC_GATE_UNCACHED="1",
                       ADAMIC_TEST_WASI="1", ADAMIC_ORACLE_WASI="1",
                       GOCACHE=go_cache, XDG_CACHE_HOME=str(cache))
    if name.split("/")[0] == "TestWASI":
        sysroot = environment.get("WASI_SYSROOT")
        if not sysroot:
            raise ValueError("WASI_SYSROOT is required")
        sdk = Path(sysroot).parent.parent / "bin"
        environment["PATH"] = str(sdk) + os.pathsep + environment["PATH"]
    if shard is not None:
        environment["ADAMIC_TEST_SHARD"] = shard
    command = ["go", "test", "./" + package, "-run", selector,
               "-count=1", "-timeout", "30m", "-json"]
    label = name if shard is None else name + "__shard_" + shard.replace("/", "_")
    log = output / (label.replace("/", "__") + ".jsonl")
    started = time.monotonic()
    with log.open("w") as stream:
        result = subprocess.run(command, cwd=ROOT, env=environment,
                                stdout=stream, stderr=stream)
    events = []
    for line in log.read_text().splitlines():
        try:
            events.append(json.loads(line))
        except ValueError:
            continue
    terminal = [event for event in events if event.get("Test") == name and
                event.get("Action") in {"pass", "fail", "skip"}]
    row = dict(package=package, test=name, shard=shard, unit_selector=unit, log=log.name, command=command,
               exit=result.returncode, wall=time.monotonic() - started,
               result=terminal[-1] if terminal else None)
    with (output / "results.jsonl").open("a") as stream:
        stream.write(json.dumps(row) + "\n")
    print(name, result.returncode, row["result"], flush=True)
    return log, result.returncode


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("output", type=Path)
    parser.add_argument("--shards-from", type=Path, help="Run each corpus shard from prepared group logs")
    parser.add_argument("--units-from", type=Path,
                        help="Select every passing leaf from prepared group logs")
    parser.add_argument("--exclude-corpus", action="store_true",
                        help="With --units-from, omit corpus pieces already measured by shard")
    args = parser.parse_args()
    args.output = args.output.resolve()
    if args.units_from:
        args.units_from = args.units_from.resolve()
    if args.shards_from:
        args.shards_from = args.shards_from.resolve()
    if args.output.exists():
        raise ValueError("Use a new output directory for each measurement")
    args.output.mkdir(parents=True)
    inputs = {str(path.relative_to(ROOT)): digest(path) for package in PACKAGES
              for path in (ROOT / package).rglob("*") if path.is_file()}
    (args.output / "inputs.json").write_text(json.dumps(inputs, indent=2) + "\n")
    go_cache = subprocess.check_output(["go", "env", "GOCACHE"], text=True).strip()
    if not os.environ.get("ADAMIC_CLANG_TSGO_ARCHIVE"):
        raise ValueError("Build and set ADAMIC_CLANG_TSGO_ARCHIVE first")
    if args.shards_from:
        counts = CORPUS_SHARDS
        for line in (args.shards_from / "results.jsonl").read_text().splitlines():
            group = json.loads(line)
            name = group["test"]
            if name not in counts:
                continue
            assert group["exit"] == 0, group
            units = list(leaves(args.shards_from / (name + ".jsonl")))
            if name == "TestNormalizeMatchesNode":
                units.sort(key=lambda unit: (unit.endswith("/contexts"), int(unit.split("cases_")[1].split("_")[0]) if "cases_" in unit else 0))
            assert len(units) == counts[name], (name, len(units))
            observed = []
            for index, expected in enumerate(units):
                log, exit_code = run(group["package"], name, "^" + name + "$", args.output,
                                     args.shards_from / (name + "-cache"), go_cache,
                                     f"{index}/{counts[name]}")
                events = leaves(log)
                assert exit_code == 0 and list(events) == [expected], (name, index, events)
                check(events)
                observed.extend(events)
            assert set(observed) == set(units) and len(observed) == len(units), name
    elif args.units_from:
        for line in (args.units_from / "results.jsonl").read_text().splitlines():
            group = json.loads(line)
            assert group["exit"] == 0, group
            if args.exclude_corpus and group["test"] in CORPUS_SHARDS:
                continue
            cache = args.units_from / (group["test"] + "-cache")
            for name, event in sorted(leaves(args.units_from / (group["test"] + ".jsonl")).items()):
                if event["Action"] == "skip":
                    continue
                selector = "/".join("^" + re.escape(part) + "$" for part in name.split("/"))
                log, exit_code = run(group["package"], name, selector, args.output, cache, go_cache, unit=True)
                observed = leaves(log)
                assert exit_code == 0 and list(observed) == [name], (name, observed)
                check(observed)
    else:
        for package in PACKAGES:
            names = sorted({name for path in (ROOT / package).glob("*_test.go")
                            for name in re.findall(r"^func (Test\w+)\(t \*testing.T\)", path.read_text(), re.M)})
            for name in names:
                run(package, name, "^" + name + "$", args.output,
                    args.output / (name + "-cache"), go_cache)


if __name__ == "__main__":
    main()
