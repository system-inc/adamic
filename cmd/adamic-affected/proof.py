#!/usr/bin/env python3
"""Run the five scratch shapes against a saved, completed uncached reference.

All mutation and execution paths are outside the worker checkout. Tests always
run fresh. Checkpoints resume only this particular unfinished proof execution.
"""
import argparse
import concurrent.futures
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import time


def atomic_json(path, value):
    temporary = path.with_name(path.name + ".new")
    with temporary.open("w") as stream:
        json.dump(value, stream, indent=2, sort_keys=True)
        stream.flush()
        os.fsync(stream.fileno())
    temporary.replace(path)
    descriptor = os.open(path.parent, os.O_DIRECTORY)
    try:
        os.fsync(descriptor)
    finally:
        os.close(descriptor)


def output(arguments, root, environment):
    return subprocess.check_output(arguments, cwd=root, env=environment).decode()


def canonical_events(path):
    """Keep action/name/output, grouped by test; omit only harness timings.

    Interleaving different parallel tests is not an input-dependent event.
    Ordering within each test remains part of the comparison. No paths, counts,
    diagnostics, random values, or arbitrary numbers are normalized.
    """
    groups = {}
    if not path:
        return groups
    with open(path) as stream:
        for line in stream:
            event = json.loads(line)
            name = event.get("Test", "")
            action = event["Action"]
            text = event.get("Output", "")
            if event.get("OutputType") == "frame" and text.lstrip().startswith(("--- PASS:", "--- FAIL:", "--- SKIP:")):
                text = re.sub(r" \([0-9]+(?:\.[0-9]+)?s\)(\n?)$", r" (TIME)\1", text)
            groups.setdefault(name, []).append([action, text])
    return groups


def timing_metadata(root, environment, arguments):
    return {
        "commit": output(["git", "rev-parse", "HEAD"], root, environment).strip(),
        "nproc": output(["nproc"], root, environment).strip(),
        "cpu.max": Path("/sys/fs/cgroup/cpu.max").read_text().strip(),
        "go": output(["go", "version"], root, environment).strip(),
        "clang": output(["clang", "--version"], root, environment).splitlines()[0],
        "node": output(["node", "--version"], root, environment).strip(),
        "load_before": Path("/proc/loadavg").read_text().strip(),
        "uncached": True,
        "build_flags": "go test -c; -test.v=test2json -test.timeout=60m; four package workers",
        "instrument": arguments,
        "started": time.time(),
    }


def run_package(package, directory, root, destination, environment):
    stem = destination / package.replace("/", "_")
    binary = str(stem) + ".test"
    instrument = {"build": ["go", "test", "-c", "-o", binary, package],
                  "build_cwd": str(root), "test": [binary, "-test.v=test2json", "-test.timeout=60m"],
                  "test_cwd": str(directory), "converter": ["go", "tool", "test2json", "-t", "-p", package],
                  "json_output": str(stem) + ".jsonl", "uncached": True}
    with open(str(stem) + ".build.log", "w") as log:
        build = subprocess.run(["go", "test", "-c", "-o", binary, package],
                               cwd=root, env=environment, stdout=log, stderr=log)
    events = str(stem) + ".jsonl"
    if build.returncode:
        return {"exit": build.returncode, "build_failed": True, "events": "", "instrument": instrument}
    if not Path(binary).exists():
        return {"exit": 0, "events": "", "instrument": instrument}
    with open(events, "w") as log, open(str(stem) + ".stderr", "w") as errors:
        converter = subprocess.Popen(["go", "tool", "test2json", "-t", "-p", package],
                                     cwd=directory, env=environment, stdin=subprocess.PIPE,
                                     stdout=log, stderr=errors)
        try:
            test = subprocess.run([binary, "-test.v=test2json", "-test.timeout=60m"],
                                  cwd=directory, env=environment,
                                  stdout=converter.stdin, stderr=converter.stdin)
        finally:
            converter.stdin.close()
            conversion_exit = converter.wait()
        log.flush()
        os.fsync(log.fileno())
    Path(binary).unlink()
    return {"exit": test.returncode or conversion_exit, "events": events,
            "instrument": instrument,
            "event_hash": hashlib.sha256(Path(events).read_bytes()).hexdigest()}


def run_set(names, kind, state, state_path, packages, root, destination, environment, reference):
    destination.mkdir(parents=True, exist_ok=True)
    phase = state.setdefault(kind, {"packages": {}, "wall_seconds": 0.0})
    for package, result in phase["packages"].items():
        if result.get("events"):
            if hashlib.sha256(Path(result["events"]).read_bytes()).hexdigest() != result["event_hash"]:
                raise RuntimeError("saved proof events changed: " + package)
    pending = [name for name in names if name not in phase["packages"]]
    metadata = timing_metadata(root, environment, ["run_set", kind, *pending])
    phase.setdefault("sessions", []).append(metadata)
    prior = phase["wall_seconds"]
    started = time.monotonic()
    with concurrent.futures.ThreadPoolExecutor(max_workers=4) as workers:
        futures = {workers.submit(run_package, name, packages[name]["Dir"], root,
                                  destination, environment): name for name in pending}
        for future in concurrent.futures.as_completed(futures):
            name = futures[future]
            result = future.result()
            if kind == "skipped_run":
                expected = canonical_events(reference["Packages"][name]["Events"])
                actual = canonical_events(result["events"])
                result["identical"] = expected == actual
                if expected != actual:
                    atomic_json(destination / (name.replace("/", "_") + ".difference.json"),
                                {"reference": expected, "branch": actual})
            phase["packages"][name] = result
            phase["wall_seconds"] = prior + time.monotonic() - started
            atomic_json(state_path, state)
            print(kind, name, result["exit"], result.get("identical", ""), flush=True)
    phase["wall_seconds"] = prior + time.monotonic() - started
    metadata["load_after"] = Path("/proc/loadavg").read_text().strip()
    metadata["ended"] = time.time()
    phase["complete"] = True
    atomic_json(state_path, state)


def mutant_matrix(state, state_path, root, destination, environment, record, binary, tool_mutant):
    matrix = {}
    for kind in ("observed", "directories", "toolchain"):
        mutated = json.loads(json.dumps(record))
        if kind == "observed":
            for closure in mutated["Packages"].values():
                closure["Observed"] = {}
        elif kind == "directories":
            for closure in mutated["Packages"].values():
                closure["Observed"] = {
                    path: value for path, value in closure["Observed"].items()
                    if not (root / path).is_dir()
                }
        else:
            mutated["Toolchain"]["node --version"] = "v24.affected-mutant\n"
        path = destination / (kind + "-record.json")
        atomic_json(path, mutated)
        target = tool_mutant if kind == "toolchain" else binary
        selected = output([target, "select", "-record", str(path)], root, environment).splitlines()
        if kind == "toolchain":
            control = output([binary, "select", "-record", str(path)], root, environment).splitlines()
            all_packages = sorted(state["selected"] + state["skipped"])
            if sorted(control) != all_packages:
                raise RuntimeError("Node identity control failed to select everything")
            skipped = sorted(set(all_packages) - set(selected))
            caught = skipped
            instrument = "independent node --version requires all current packages"
        else:
            skipped = sorted(set(state["selected"]) - set(selected))
            caught = []
            for name in skipped:
                result = state["selected_run"]["packages"][name]
                if result["exit"] or canonical_events(result["events"]) != canonical_events(record["Packages"][name]["Events"]):
                    caught.append(name)
            instrument = "fresh branch package exit and name/action/output comparison with main"
        matrix[kind] = {"selected": selected, "newly_skipped": skipped,
                        "caught_by": caught, "instrument": instrument}
        print(state["shape"], "mutant", kind, "caught", caught, flush=True)
    if state["shape"] == "oracle":
        # Isolate new-directory-entry sensitivity from the primary fixture's
        # changed bytes. Otherwise its retained file hash masks this mutant.
        fixture = root / "internal/oracle/testdata/affected_proof.a"
        if fixture.exists():
            raise RuntimeError("supplemental fixture already exists")
        output(["git", "restore", "--source", record["Commit"], "--worktree", "--", state["path"]], root, environment)
        try:
            fixture.write_text("const affectedProofBox: { value: number } = { value: 1 };\naffectedProofBox.value = 2;\n")
            control = output([binary, "select", "-record", str(destination / "directories-record.json")], root, environment).splitlines()
            normal = output([binary, "select", "-record", str(destination / "original-record.json")], root, environment).splitlines()
            missed = sorted(set(normal) - set(control))
            supplementary = {}
            directory = destination / "directory-addition-proof"
            directory.mkdir(exist_ok=True)
            for name in missed:
                package_dir = root / name.removeprefix("github.com/system-inc/adamic/")
                result = run_package(name, package_dir, root, directory,
                                     dict(environment, ADAMIC_GATE_UNCACHED="1"))
                result["identical"] = canonical_events(result["events"]) == canonical_events(record["Packages"][name]["Events"])
                supplementary[name] = result
            caught = [name for name, result in supplementary.items() if result["exit"] or not result["identical"]]
            matrix["directories"]["new_fixture_probe"] = {
                "path": str(fixture.relative_to(root)), "normal_selected": normal,
                "mutant_selected": control, "packages": supplementary, "caught_by": caught}
            print("oracle directory-addition mutant caught", caught, flush=True)
        finally:
            fixture.unlink(missing_ok=True)
            output(["git", "restore", "--worktree", "--", state["path"]], root, environment)
    state["mutants"] = matrix
    atomic_json(state_path, state)


def build_toolchain_mutant(source, destination, environment):
    destination.mkdir(parents=True, exist_ok=True)
    for name in ("main.go", "trace.go", "record.go", "observer/notify.c"):
        target = destination / name
        target.parent.mkdir(parents=True, exist_ok=True)
        contents = (source / name).read_text()
        if name == "main.go":
            original = "!reflect.DeepEqual(record.Toolchain, tools)"
            if original not in contents:
                raise RuntimeError("toolchain mutant target missing")
            contents = contents.replace(original, "!reflect.DeepEqual(tools, tools)")
        target.write_text(contents)
    (destination / "go.mod").write_text("module affected-toolchain-mutant\n\ngo 1.27\n")
    binary = str(destination / "adamic-affected-mutant")
    with (destination / "build.log").open("w") as log:
        subprocess.run(["go", "build", "-o", binary, "."], cwd=destination,
                       env=environment, stdout=log, stderr=log, check=True)
    return binary


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", required=True, type=Path)
    parser.add_argument("--record", required=True, type=Path)
    parser.add_argument("--output", required=True, type=Path)
    parser.add_argument("--binary", default="/tmp/adamic-affected")
    args = parser.parse_args()
    environment = dict(os.environ, PWD=str(args.root), _=args.binary)
    test_environment = dict(environment, ADAMIC_GATE_UNCACHED="1")
    record = json.loads(args.record.read_text())
    if not record.get("Complete"):
        raise RuntimeError("proof requires a complete reference")
    record_hash = hashlib.sha256(args.record.read_bytes()).hexdigest()
    args.output.mkdir(parents=True, exist_ok=True)
    if output(["git", "status", "--porcelain", "--untracked-files=all"], args.root, environment):
        raise RuntimeError("scratch checkout changed; refusing to reuse proof evidence")
    for name, closure in record["Packages"].items():
        if closure["Events"]:
            actual = hashlib.sha256(Path(closure["Events"]).read_bytes()).hexdigest()
            if actual != record["EventHashes"][name]:
                raise RuntimeError("reference package log changed: " + name)
    output(["git", "checkout", "--detach", record["Commit"]], args.root, environment)
    baseline_selected = output([args.binary, "select", "-record", str(args.record)], args.root, environment).splitlines()
    expected = sorted(name for name, closure in record["Packages"].items() if closure["Uncertain"])
    if baseline_selected != expected:
        raise RuntimeError("baseline identity differs from record; refusing a vacuous all-selected proof")
    tool_mutant = build_toolchain_mutant(Path(__file__).resolve().parent,
                                       args.output / "toolchain-mutant", environment)
    shapes = {
        "stage3": ("stage3/affected-proof.a", "// A stage3-only input probe.\n"),
        "docs": ("docs/0.1.md", "\n<!-- A docs-only input probe. -->\n"),
        "slice": ("stage1/cohere/json/formatter.ts", "\n// A single-slice input probe.\n"),
        "runtime": ("internal/native/runtime/string.c", "\n/* A native runtime input probe. */\n"),
        "oracle": ("internal/oracle/testdata/numbers.a",
                   "const affectedProofBox: { value: number } = { value: 1 };\naffectedProofBox.value = 2;\n"),
    }
    reports = {}
    for shape, (relative, addition) in shapes.items():
        destination = args.output / shape
        destination.mkdir(exist_ok=True)
        state_path = destination / "state.json"
        state = json.loads(state_path.read_text()) if state_path.exists() else {}
        if state and state["record_hash"] != record_hash:
            raise RuntimeError("reference changed during proof")
        if not state:
            output(["git", "checkout", "--detach", record["Commit"]], args.root, environment)
            branch = "scratch/affected-proof-" + shape
            output(["git", "checkout", "-b", branch], args.root, environment)
            path = args.root / relative
            path.parent.mkdir(parents=True, exist_ok=True)
            with path.open("a") as stream:
                stream.write(addition)
            output(["git", "add", "--", relative], args.root, environment)
            output(["git", "commit", "-m", "Probe " + shape + " affected inputs"], args.root, environment)
            commit = output(["git", "rev-parse", "HEAD"], args.root, environment).strip()
            worktree = args.output / (shape + "-worktree")
            output(["git", "worktree", "add", "--detach", str(worktree), commit], args.root, environment)
            state = {"record_hash": record_hash, "commit": commit, "shape": shape,
                     "path": relative, "review_worktree": str(worktree)}
            atomic_json(state_path, state)
        output(["git", "checkout", "--detach", state["commit"]], args.root, environment)
        listed = output(["go", "list", "-json", "./..."], args.root, environment)
        decoder = json.JSONDecoder()
        packages = {}
        while listed.strip():
            package, end = decoder.raw_decode(listed.lstrip())
            packages[package["ImportPath"]] = package
            listed = listed.lstrip()[end:]
        selected = output([args.binary, "select", "-record", str(args.record)], args.root, environment).splitlines()
        skipped = sorted(set(packages) - set(selected))
        if "selected" in state and state["selected"] != selected:
            raise RuntimeError("selection changed during resumed proof")
        state.update(selected=selected, skipped=skipped)
        atomic_json(state_path, state)
        atomic_json(destination / "original-record.json", record)
        print(shape, "selected", len(selected), "skipped", len(skipped), flush=True)
        run_set(skipped, "skipped_run", state, state_path, packages, args.root,
                destination / "skipped", test_environment, record)
        run_set(selected, "selected_run", state, state_path, packages, args.root,
                destination / "selected", test_environment, record)
        mutant_matrix(state, state_path, args.root, destination, environment,
                      record, args.binary, tool_mutant)
        reports[shape] = {
            "commit": state["commit"], "selected": selected, "skipped": skipped,
            "skipped_pass_and_identical": all(result["exit"] == 0 and result["identical"]
                                              for result in state["skipped_run"]["packages"].values()),
            "selected_pass": all(result["exit"] == 0 for result in state["selected_run"]["packages"].values()),
            "selected_wall_seconds": state["selected_run"]["wall_seconds"],
            "reference_wall_seconds": record["WallSeconds"], "mutants": state["mutants"],
        }
        atomic_json(args.output / "summary.json", reports)
    output(["git", "checkout", "--detach", record["Commit"]], args.root, environment)
    if not all(report["skipped_pass_and_identical"] for report in reports.values()):
        raise RuntimeError("skipped-package proof failed; inspect saved differences")
    if not all(report["selected_pass"] for report in reports.values()):
        raise RuntimeError("selected packages failed; inspect saved logs")
    if not any(report["mutants"]["observed"]["caught_by"] for report in reports.values()):
        raise RuntimeError("observed-input mutant was not caught by a branch shape")
    if not reports["oracle"]["mutants"]["directories"].get("new_fixture_probe", {}).get("caught_by"):
        raise RuntimeError("directory-input mutant was not caught by the new oracle fixture")
    if not any(report["mutants"]["toolchain"]["caught_by"] for report in reports.values()):
        raise RuntimeError("toolchain-input mutant was not caught by a branch shape")


if __name__ == "__main__":
    main()
