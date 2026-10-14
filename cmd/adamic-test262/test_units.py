#!/usr/bin/env python3
"""Measure independently selectable Go tests with cold runtime and result caches."""
import argparse
import collections
import hashlib
import json
import os
import pathlib
import re
import shutil
import signal
import subprocess
import sys
import time

LIMIT_SECONDS = 30
PRODUCT_LIMIT_SECONDS = 60


def coverage(expected, shards):
    counts = collections.Counter(unit for shard in shards for unit in shard)
    wanted = set(expected)
    if len(wanted) != len(expected):
        raise ValueError("duplicate unit in test inventory")
    missing = sorted(wanted - counts.keys())
    unexpected = sorted(counts.keys() - wanted)
    duplicate = sorted(unit for unit, count in counts.items() if count != 1)
    if missing or unexpected or duplicate:
        raise ValueError(f"invalid coverage: missing={missing}, unexpected={unexpected}, duplicate={duplicate}")


def plan(units):
    shards = [[unit] for unit in units]
    coverage(units, shards)
    return shards


def observations(unit, events, wall_seconds, limit_seconds=LIMIT_SECONDS):
    rows = []
    seen = set()
    for event in events:
        if event.get("Action") not in ("pass", "fail", "skip") or "Test" not in event:
            continue
        name = event["Test"]
        if name != unit and not name.startswith(unit + "/"):
            raise ValueError(f"unexpected test {name} in shard {unit}")
        if name in seen:
            raise ValueError(f"duplicate completion of {name}")
        seen.add(name)
        rows.append({"test": name, "seconds": event.get("Elapsed", 0), "result": event["Action"]})
    if unit not in seen:
        raise ValueError(f"shard {unit} did not complete its selected test")
    valid = wall_seconds < limit_seconds and all(
        row["seconds"] < limit_seconds and row["result"] == "pass" for row in rows
    )
    return {"unit": unit, "wall_seconds": wall_seconds, "valid": valid, "tests": rows}


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def inputs(binary, root):
    paths = sorted((root / "cmd/adamic-test262/testdata").rglob("*"))
    paths += sorted((root / "cmd/adamic-test262").glob("*.go"))
    paths += sorted((root / "cmd/adamic-test262").glob("*.cjs"))
    paths += sorted((root / "cmd/adamic-test262").glob("*.py"))
    paths += sorted((root / "internal/native/runtime").glob("*.[ch]"))
    files = {str(path.relative_to(root)): digest(path) for path in paths if path.is_file()}
    tools = {name: subprocess.check_output(command, text=True, timeout=15).strip() for name, command in {
        "go": ["go", "version"], "node": ["node", "--version"],
        "clang": ["clang", "--version"],
    }.items()}
    record = {"binary_sha256": digest(binary), "files": files, "tools": tools}
    record["sha256"] = hashlib.sha256(json.dumps(record, sort_keys=True).encode()).hexdigest()
    return record


def measure(binary, unit, directory, env, root, product=False):
    directory.mkdir(parents=True, exist_ok=True)
    budget = PRODUCT_LIMIT_SECONDS if product else LIMIT_SECONDS
    kill = 90 if product else LIMIT_SECONDS
    command = ["go", "tool", "test2json", "-t", "-p", "cmd/adamic-test262", str(binary),
               "-test.v=test2json", "-test.run=^" + unit + "$", "-test.count=1",
               "-test.parallel=4", "-test.timeout=" + str(kill) + "s"]
    start = time.monotonic()
    with (directory / "events.jsonl").open("w") as log:
        process = subprocess.Popen(command, cwd=root / "cmd/adamic-test262", env=env,
                                   stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
        try:
            status = process.wait(timeout=kill + 5)
        except subprocess.TimeoutExpired:
            try:
                os.killpg(process.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            status = process.wait(timeout=5)
    elapsed = time.monotonic() - start
    events = [json.loads(line) for line in (directory / "events.jsonl").read_text().splitlines()]
    try:
        result = observations(unit, events, elapsed, budget)
    except ValueError as error:
        result = {"unit": unit, "wall_seconds": elapsed, "valid": False, "tests": [], "error": str(error)}
    result.update(exit=status, product=product, limit_seconds=budget)
    result["valid"] = result["valid"] and status == 0
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", required=True, type=pathlib.Path)
    parser.add_argument("--output", type=pathlib.Path)
    parser.add_argument("--unit", help="Run one exact top-level test; default: all shards")
    parser.add_argument("--list", action="store_true", help="Print the complete hashed shard plan")
    args = parser.parse_args()
    root = pathlib.Path(__file__).resolve().parents[2]
    binary = args.binary.resolve()
    units = subprocess.check_output([str(binary), "-test.list=^Test"], text=True, timeout=15).splitlines()
    if not units or any(not re.fullmatch(r"Test\w+", unit) for unit in units):
        raise ValueError("empty or invalid test inventory")
    shards = plan(units)
    products = [unit for unit in units if unit.startswith("TestProduct_")]
    manifest = {"inputs": inputs(binary, root), "shards": shards, "products": products, "limit_seconds": LIMIT_SECONDS,
                "product_limit_seconds": PRODUCT_LIMIT_SECONDS,
                "cold": "fresh native runtime and observation caches; declared Go products are built once and shared"}
    if args.list:
        print(json.dumps(manifest, indent=2))
        return 0
    if sys.platform != "linux":
        parser.error("cold measurements require Linux XDG cache isolation; the listed Go selectors are platform independent")
    if args.output is None:
        parser.error("--output is required when measuring")
    if args.unit is not None and args.unit not in units:
        parser.error("--unit must name an exact test in the inventory")
    selected = [args.unit] if args.unit else units
    output = args.output.resolve()
    output.mkdir(parents=True, exist_ok=False)
    (output / "inputs.json").write_text(json.dumps(manifest, indent=2) + "\n")
    # Changing XDG_CACHE_HOME must not accidentally rebuild Go's standard library.
    go_cache = subprocess.check_output(["go", "env", "GOCACHE"], text=True, timeout=15).strip()
    user_cache = pathlib.Path(os.environ.get("XDG_CACHE_HOME") or pathlib.Path.home() / ".cache")
    helper_cache = os.environ.get("ADAMIC_GOCACHE_DIR") or str(user_cache / "adamic-gocache")
    results = []
    product_cache = output / "products" / "cache" / "build"
    # Main declares its Go compiler as a product. Prepare it once with the normal
    # product budget; seed only Go products, never a native archive or observation.
    for unit in products:
        directory = output / "products" / unit
        env = dict(os.environ, GOCACHE=go_cache, ADAMIC_GOCACHE_DIR=helper_cache, XDG_CACHE_HOME=str(output / "products" / "cache"),
                   ADAMIC_BUILD_CACHE_DIR=str(product_cache), ADAMIC_BUILD_CACHE="",
                   ADAMIC_GATE_UNCACHED="1", ADAMIC_TEST262_MEASURE="1", GOMAXPROCS="4")
        measured = measure(binary, unit, directory, env, root, product=True)
        if unit in selected:
            results.append(measured)
        (output / "products.json").write_text(json.dumps(measured, indent=2) + "\n")
        if not measured["valid"]:
            return 1
    for unit in selected:
        if unit in products:
            continue
        directory = output / unit
        cache = directory / "cache" / "build"
        if (product_cache / "local").exists():
            shutil.copytree(product_cache / "local", cache / "local")
        env = dict(os.environ, GOCACHE=go_cache, ADAMIC_GOCACHE_DIR=helper_cache, XDG_CACHE_HOME=str(directory / "cache"),
                   ADAMIC_BUILD_CACHE_DIR=str(cache), ADAMIC_BUILD_CACHE="",
                   ADAMIC_GATE_UNCACHED="1", ADAMIC_TEST262_MEASURE="1", GOMAXPROCS="4")
        measured = measure(binary, unit, directory, env, root)
        results.append(measured)
        (output / "results.json").write_text(json.dumps(results, indent=2) + "\n")
    coverage(selected, [[result["unit"]] for result in results])
    (output / "results.json").write_text(json.dumps(results, indent=2) + "\n")
    table = ["| Test or subtest | Seconds | Result |", "|---|---:|---|"]
    for result in results:
        for row in result["tests"]:
            table.append(f"| {row['test']} | {row['seconds']:.2f} | {row['result']} |")
    (output / "table.md").write_text("\n".join(table) + "\n")
    return 0 if all(result["valid"] for result in results) else 1


if __name__ == "__main__":
    raise SystemExit(main())
