"""Plant one program failure and omissions through Go overlays, without editing sources."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import sys

from budget import check, leaves

ROOT = Path(__file__).resolve().parents[3]
CASES = [
    ("missing-range", "internal/native/split_units_test.go", "first < total;", "first < total-width;", "^TestSplitUnitCoverage$", {"TestSplitUnitCoverage/random_regex", "TestSplitUnitCoverage/decode_prefixes", "TestSplitUnitCoverage/normalize_points"}, "units, want"),
    ("missing-wasi-fixture", "internal/native/wasm_test.go", '"internal/oracle/testdata/write_stderr_order.a",', "", "^TestSplitUnitCoverage$/^WASI$", {"TestSplitUnitCoverage/WASI"}, "WASI fixtures, want 35"),
    ("changed-wasi-membership", "internal/native/wasm_test.go", '"internal/oracle/testdata/write_stderr_order.a",', '"internal/oracle/testdata/write_stdout_order.a",', "^TestSplitUnitCoverage$/^WASI$", {"TestSplitUnitCoverage/WASI"}, "WASI fixture membership changed"),
    ("missing-cache-mode", "internal/native/split_units_test.go", 'var splitCacheModes = []string{"0", "1"}', 'var splitCacheModes = []string{"0"}', "^TestSplitUnitCoverage$/^split_cache$", {"TestSplitUnitCoverage/split_cache"}, "cached and bypass builds must both run"),
    ("duplicate-record-mutant", "internal/native/record_test.go", '{"uint32-max-as-index", "record.c",', '{"indices-in-insertion-order", "record.c",', "^TestSplitUnitCoverage$/^record_mutants$", {"TestSplitUnitCoverage/record_mutants"}, "duplicate record mutant"),
    ("missing-decode-target", "internal/native/split_units_test.go", 'var decodeTargets = []string{"native", "wasi"}', 'var decodeTargets = []string{"native"}', "^TestSplitUnitCoverage$/^decode_targets$", {"TestSplitUnitCoverage/decode_targets"}, "decoder targets must both run"),
    ("duplicate-shard-owner", "internal/native/split_units_test.go", "piece%s.count == s.index", "piece%s.count == 0", "^TestSplitUnitCoverage$/^shard_union$", {"TestSplitUnitCoverage/shard_union"}, "owners across"),
    ("invalid-shard-accepted", "internal/native/split_units_test.go", "index >= count", "false", "^TestShardSelection$", {"TestShardSelection"}, "invalid shard accepted"),
    ("artifact-input-ignored", "internal/native/products_test.go", "sha256.Sum256(input)", "sha256.Sum256(input[:0])", "^TestArtifactInputsAndReuse$", {"TestArtifactInputsAndReuse"}, "artifact inputs/reuse:"),
    ("artifact-reuse-disabled", "internal/native/products_test.go", "in.Name = label", 'in.Name = label; c.inputs.Flags[0] += "."', "^TestArtifactInputsAndReuse$", {"TestArtifactInputsAndReuse"}, "artifact inputs/reuse:"),
]


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("output", type=Path)
    args = parser.parse_args()
    output = args.output.resolve()
    output.mkdir(parents=True, exist_ok=False)
    rows = []
    go_cache = subprocess.check_output(["go", "env", "GOCACHE"], text=True).strip()
    for name, file, before, after, selector, expected, witness in CASES:
        package = "./internal/native"
        directory = output / name
        directory.mkdir()
        source = ROOT / file
        text = source.read_text()
        assert text.count(before) == 1, (name, "overlay target must be unique")
        replacement = directory / source.name
        replacement.write_text(text.replace(before, after, 1))
        overlay = directory / "overlay.json"
        overlay.write_text(json.dumps({"Replace": {str(source): str(replacement)}}))
        command = ["go", "test", "-overlay", str(overlay), package,
                   "-run", selector, "-count=1", "-json", "-timeout", "10m"]
        environment = os.environ.copy()
        environment.update(GOMAXPROCS="4", ADAMIC_GATE_UNCACHED="1",
                           XDG_CACHE_HOME=str(directory / "cache"), GOCACHE=go_cache)
        log = directory / "test.jsonl"
        with log.open("w") as stream:
            result = subprocess.run(command, cwd=ROOT, env=environment,
                                    stdout=stream, stderr=stream)
        events = leaves(log)
        failed = {unit for unit, event in events.items() if event["Action"] == "fail"}
        passed = sum(event["Action"] == "pass" for event in events.values())
        assert result.returncode != 0 and failed == expected, (name, result.returncode, failed)
        assert witness in log.read_text(), (name, "required assertion did not run")
        row = dict(mutant=name, command=command, exit=result.returncode,
                   failed=sorted(failed), passed=passed)
        rows.append(row)
        print(json.dumps(row), flush=True)
    directory = output / "regex-fixture"
    directory.mkdir()
    source = ROOT / "internal/native/split_units_test.go"
    text = source.read_text()
    before = "shard := currentTestShard(t)\n\tfor index, piece := range unitRanges(len(cases), randomRegexUnitSize) {"
    assert text.count(before) == 1
    after = 'if len(cases[500].Expected.Captures) == 0 { cases[500].Pattern = "(?:)" } else { cases[500].Pattern = "(?!)" }; cases[500].Flags = ""; cases[500].PatternUnits = nil\n' + before
    replacement = directory / source.name
    replacement.write_text(text.replace(before, after, 1))
    overlay = directory / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {str(source): str(replacement)}}))
    failed_shards = []
    for index in range(40):
        environment = os.environ.copy()
        environment.update(GOMAXPROCS="4", ADAMIC_GATE_UNCACHED="1", GOCACHE=go_cache,
                           XDG_CACHE_HOME=str(directory / "cache"), ADAMIC_TEST_SHARD=f"{index}/40")
        command = ["go", "test", "-overlay", str(overlay), "./internal/native", "-run", "^TestRegExpBytecodeRandomNode$", "-count=1", "-json", "-timeout", "10m"]
        log = directory / f"shard-{index}.jsonl"
        with log.open("w") as stream:
            result = subprocess.run(command, cwd=ROOT, env=environment, stdout=stream, stderr=stream)
        events = leaves(log)
        expected = f"TestRegExpBytecodeRandomNode/cases_{index*250:06d}_{(index+1)*250:06d}"
        assert list(events) == [expected], (index, events)
        if result.returncode:
            assert index == 2 and events[expected]["Action"] == "fail" and "DISAGREEMENT case=0" in log.read_text(), (index, events)
            failed_shards.append(index)
        else:
            assert events[expected]["Action"] == "pass", (index, events)
        row = dict(mutant="regex-fixture", shard=f"{index}/40", command=command, exit=result.returncode, unit=expected)
        rows.append(row)
        print(json.dumps(row), flush=True)
    assert failed_shards == [2], failed_shards
    os.environ["ADAMIC_UNIT_BUDGET"] = "1"
    check({"boundary": {"Action": "pass", "Elapsed": 30.0}})
    try:
        check({"planted": {"Action": "pass", "Elapsed": 30.001}})
    except ValueError as error:
        rows.append(dict(mutant="elapsed-over-budget", caught=str(error)))
    else:
        raise AssertionError("budget mutant escaped")
    directory = output / "invocation-budget"
    directory.mkdir()
    event = dict(Action="pass", Test="probe", Elapsed=1.0)
    (directory / "probe.jsonl").write_text(json.dumps(event) + "\n")
    row = dict(package="internal/native", test="probe", exit=0, wall=30.0,
               unit_selector=True, result=event)
    command = [sys.executable, str(Path(__file__).with_name("budget.py")), str(directory)]
    for wall, expected in [(30.0, 0), (30.001, 1)]:
        row["wall"] = wall
        (directory / "results.jsonl").write_text(json.dumps(row) + "\n")
        log = directory / f"budget-{wall}.log"
        with log.open("w") as stream:
            result = subprocess.run(command, stdout=stream, stderr=stream)
        assert (result.returncode != 0) == bool(expected), (wall, result.returncode)
        if expected:
            assert "invocation took 30.001 seconds" in log.read_text()
    rows.append(dict(mutant="invocation-over-budget", caught="selector probe invocation took 30.001 seconds"))
    (output / "results.json").write_text(json.dumps(rows, indent=2) + "\n")
    print(json.dumps(rows[-1]), flush=True)


if __name__ == "__main__":
    main()
