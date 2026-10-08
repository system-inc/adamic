"""Validate independent production-code mutants in a disposable, pinned worktree."""
import argparse
from concurrent.futures import ThreadPoolExecutor
import json
import os
from pathlib import Path
import re
import shlex
import subprocess
import sys
import tempfile
import time
import threading


def git(repository, *arguments, **options):
    return subprocess.run(["git", "-C", str(repository), *arguments], **options)


def run_components(pattern):
    # Go splits subtest patterns only at slashes outside groups and character classes.
    components, start, group_depth, in_class, escaped = [], 0, 0, False, False
    for index, character in enumerate(pattern):
        if escaped:
            escaped = False
            continue
        if character == "\\":
            escaped = True
        elif in_class:
            if character == "]":
                in_class = False
        elif character == "[":
            in_class = True
        elif character == "(":
            group_depth += 1
        elif character == ")":
            group_depth -= 1
        elif character == "/" and group_depth == 0:
            components.append(pattern[start:index])
            start = index + 1
    return components + [pattern[start:]]


def run_test(worktree, entry, destination):
    # Parse the recorded command without shell evaluation. No result cache is added here.
    command = shlex.split(entry["command"])
    patterns = [command[index + 1] for index, argument in enumerate(command[:-1]) if argument == "-run"]
    patterns += [argument[5:] for argument in command if argument.startswith("-run=")]
    if command[:2] != ["go", "test"] or len(patterns) != 1:
        raise ValueError("command must be a named go test invocation")
    pattern = patterns[0]
    if any(not part.startswith("^") or not part.endswith("$") for part in run_components(pattern)):
        raise ValueError("every -run component must be anchored")
    # Prevent Go's successful-test cache as well as the oracle's result cache.
    counts = [arg for arg in command if arg.startswith("-count=")]
    if counts and counts != ["-count=1"]:
        raise ValueError("command must use -count=1")
    if not counts:
        command.insert(2, "-count=1")
    if not any(argument.startswith("-trimpath") for argument in command):
        command.insert(2, "-trimpath")
    started = time.monotonic()
    load_before = Path("/proc/loadavg").read_text().strip()
    with destination.open("w") as log:
        result = subprocess.run(command, cwd=worktree, env=dict(os.environ, ADAMIC_GATE_UNCACHED="1"),
                                stdout=log, stderr=subprocess.STDOUT)
    elapsed = time.monotonic() - started
    output = []
    for line in destination.read_text().splitlines(keepends=True):
        try:
            event = json.loads(line)
        except ValueError:
            output.append(line)
            continue
        if isinstance(event, dict) and "Action" in event:
            output.append(event.get("Output", ""))
        else:
            output.append(line)
    decoded = "".join(output)
    execution = re.search(r"^(?:ok|FAIL)\s+\S+\s+([0-9.]+)s", decoded, re.MULTILINE)
    destination.with_suffix(".timing.json").write_text(json.dumps({
        "command": command, "seconds": elapsed, "exit_code": result.returncode,
        "go_reported_test_seconds": float(execution[1]) if execution else None,
        "load_before": load_before, "load_after": Path("/proc/loadavg").read_text().strip()
    }, indent=2) + "\n")
    return result.returncode, decoded


def check_entry(repository, worktree, catalog, entry, logs, report=print):
    number, name = entry["number"], entry["name"]
    prefix = f"{number:02} {name}"
    patch = catalog / entry["patch"]
    paths = git(worktree, "apply", "--numstat", str(patch), capture_output=True, text=True)
    changed = [line.split("\t", 2)[-1] for line in paths.stdout.splitlines()]
    if paths.returncode or not changed or any(
        not path.startswith("internal/") or "/oracle/" in path or "/testdata/" in path
        or path.endswith("_test.go") for path in changed
    ):
        report(prefix + ": invalid-patch (production code only)")
        return False
    checked = git(worktree, "apply", "--check", str(patch), capture_output=True, text=True)
    if checked.returncode:
        (logs / f"{number:02}-apply.log").write_text(checked.stderr)
        report(prefix + ": no-longer-applies (needs a refresh)")
        return False
    # The named tests must still pass on the supplied commit. An unrelated red baseline
    # must never be reported as successful detection of this mutation.
    baseline_code, baseline = run_test(worktree, entry, logs / f"{number:02}-control.log")
    if baseline_code or "no tests to run" in baseline:
        report(prefix + ": baseline-fails-or-test-missing (needs investigation)")
        return False
    git(worktree, "apply", str(patch), check=True)
    try:
        mutant_code, mutant = run_test(worktree, entry, logs / f"{number:02}-mutant.log")
    finally:
        git(worktree, "apply", "-R", str(patch), check=True)
    failed_tests = re.findall(r"--- FAIL: (\S+) \(", mutant)
    expected_tests = entry["expected_tests"]
    # Permit children of a selected test and their parent reporting frames, but no other failures.
    matches = lambda actual, wanted: actual == wanted or actual.startswith(wanted + "/")
    selected_failed = all(any(matches(actual, wanted) for actual in failed_tests)
                          for wanted in expected_tests)
    unexpected = [actual for actual in failed_tests if not any(
        matches(actual, wanted) or wanted.startswith(actual + "/") for wanted in expected_tests)]
    # Ignore source line-number drift, but require the entire diagnostic message as recorded.
    expected = re.sub(r"^[^:]+\.go:\d+: ", "", entry["expected_failure_line"])
    diagnostics = [re.sub(r"^\s*[^:]+\.go:\d+: ", "", line) for line in mutant.splitlines()]
    if mutant_code == 1 and selected_failed and not unexpected and expected in diagnostics:
        report(prefix + ": applies-and-fails-as-recorded")
        return True
    report(prefix + ": applies-but-failure-not-as-recorded (needs investigation)")
    return False


def check_one(repository, catalog, entry, resolved, parent, logs, git_lock):
    messages = []
    started = time.monotonic()
    load_before = Path("/proc/loadavg").read_text().strip()
    number = entry["number"]
    directory = logs / f"{number:02}"
    directory.mkdir()
    result = {"number": number, "made_against": entry["made_against"], "target": resolved}
    if entry["status"] == "skipped":
        messages.append(f"{number:02} {entry['name']}: skipped ({entry['reason']})")
        result.update(ok=True, setup_seconds=0, check_seconds=0, cleanup_seconds=0)
    else:
        worktree = parent / f"entry-{number:02}"
        result["worktree"] = str(worktree)
        added = False
        setup_started = time.monotonic()
        try:
            with (directory / "setup.log").open("w") as setup:
                # Git registration is brief; each checkout/index and submodule belongs to one entry.
                with git_lock:
                    git(repository, "worktree", "add", "--detach", str(worktree), resolved,
                        check=True, stdout=setup, stderr=subprocess.STDOUT)
                added = True
                git(worktree, "-c", "submodule.cohere.url=https://github.com/system-inc/cohere.git",
                    "submodule", "update", "--init", "--recursive", "--depth", "1",
                    check=True, stdout=setup, stderr=subprocess.STDOUT)
            result["setup_seconds"] = time.monotonic() - setup_started
            status = git(worktree, "status", "--porcelain", capture_output=True, text=True, check=True)
            if status.stdout:
                raise RuntimeError("worktree is not clean: " + status.stdout)
            check_started = time.monotonic()
            result["ok"] = check_entry(repository, worktree, catalog, entry, directory, messages.append)
            result["check_seconds"] = time.monotonic() - check_started
            restored = git(worktree, "status", "--porcelain", capture_output=True, text=True, check=True)
            if restored.stdout:
                raise RuntimeError("worktree was not restored: " + restored.stdout)
        except Exception as error:
            messages.append(f"{number:02} {entry['name']}: error ({error})")
            result["ok"] = False
        finally:
            cleanup_started = time.monotonic()
            if added:
                try:
                    with git_lock:
                        git(repository, "worktree", "remove", "--force", str(worktree), check=True,
                            stdout=subprocess.DEVNULL, stderr=subprocess.PIPE)
                except subprocess.CalledProcessError as error:
                    messages.append(f"{number:02}: cleanup-error ({error})")
                    result["ok"] = False
            result["cleanup_seconds"] = time.monotonic() - cleanup_started
    result.update(wall_seconds=time.monotonic() - started, load_before=load_before,
                  load_after=Path("/proc/loadavg").read_text().strip())
    messages.append(f"{number:02} made against={entry['made_against']}")
    messages.append(f"{number:02} wall={result['wall_seconds']:.3f}s "
                    f"setup={result.get('setup_seconds', 0):.3f}s "
                    f"check={result.get('check_seconds', 0):.3f}s "
                    f"cleanup={result['cleanup_seconds']:.3f}s "
                    f"load={load_before} -> {result['load_after']}")
    result["messages"] = messages
    (directory / "result.json").write_text(json.dumps(result, indent=2) + "\n")
    return result


def main(commit, entry_number=None, jobs=None):
    total_started = time.monotonic()
    total_load_before = Path("/proc/loadavg").read_text().strip()
    catalog = Path(__file__).resolve().parent
    repository = Path(subprocess.check_output(["git", "-C", str(catalog), "rev-parse", "--show-toplevel"], text=True).strip())
    resolved = subprocess.check_output(["git", "-C", str(repository), "rev-parse", "--verify", commit + "^{commit}"], text=True).strip()
    entries = sorted(json.loads((catalog / "catalog.json").read_text()), key=lambda entry: entry["number"])
    if entry_number is not None:
        entries = [entry for entry in entries if entry["number"] == entry_number]
        if not entries:
            raise ValueError("unknown entry: " + str(entry_number))
    # Leave test and setup logs available after removing the worktree.
    logs = Path(tempfile.mkdtemp(prefix="adamic-catalog-logs-"))
    print("target " + resolved, flush=True)
    print("logs " + str(logs), flush=True)
    versions = [subprocess.check_output(command, text=True).splitlines()[0]
                for command in (["go", "version"], ["clang", "--version"], ["node", "--version"])]
    processors = subprocess.check_output(["nproc"], text=True).strip()
    # Not every machine mounts cgroup v2; cloud/setup.sh reports the same absence as unknown.
    cpu_max_path = Path("/sys/fs/cgroup/cpu.max")
    cpu_max = cpu_max_path.read_text().strip() if cpu_max_path.exists() else "unknown"
    print(f"build flags: commit={resolved}; nproc={processors}; cpu.max={cpu_max}; "
          + "; ".join(versions) + "; ADAMIC_GATE_UNCACHED=1; go test -trimpath -count=1; "
          "native C11 strict warnings, -ffp-contract=off -fno-optimize-sibling-calls; "
          "release -O2; sanitizer -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all", flush=True)
    if jobs is None:
        jobs = int(processors)
    if jobs < 1:
        raise ValueError("jobs must be positive")
    print(f"jobs {jobs}; Go -trimpath shares only content-keyed build artifacts; test results uncached", flush=True)
    failure = False
    results = []
    git_lock = threading.Lock()
    with tempfile.TemporaryDirectory(prefix="adamic-catalog-worktree-") as temporary:
        parent = Path(temporary)
        with ThreadPoolExecutor(max_workers=jobs) as executor:
            # map yields in catalog order; workers never write to shared stdout.
            def run(entry):
                return check_one(repository, catalog, entry, resolved, parent, logs, git_lock)
            for result in executor.map(run, entries):
                results.append(result)
                for message in result["messages"]:
                    print(message, flush=True)
                failure = failure or not result["ok"]
    total = time.monotonic() - total_started
    total_load_after = Path("/proc/loadavg").read_text().strip()
    (logs / "results.json").write_text(json.dumps({"target": resolved, "jobs": jobs,
        "total_wall_seconds": total, "load_before": total_load_before,
        "load_after": total_load_after, "entries": results}, indent=2) + "\n")
    print(f"total wall={total:.3f}s load={total_load_before} -> {total_load_after}", flush=True)
    return 1 if failure else 0


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("commit")
    parser.add_argument("--entry", type=int, metavar="NN")
    parser.add_argument("-jobs", "--jobs", type=int, metavar="N")
    arguments = parser.parse_args()
    sys.exit(main(arguments.commit, arguments.entry, arguments.jobs))
