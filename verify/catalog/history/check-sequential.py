"""Validate independent production-code mutants in a disposable, pinned worktree."""
import argparse
import json
import os
from pathlib import Path
import re
import shlex
import subprocess
import sys
import tempfile
import time


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
    if entry['number'] == 1 and destination.name.endswith('-control.log'):
        command.insert(2, '-x')
    started = time.monotonic()
    load_before = Path('/proc/loadavg').read_text().strip()
    with destination.open("w") as log:
        result = subprocess.run(command, cwd=worktree, env=dict(os.environ, ADAMIC_GATE_UNCACHED="1"),
                                stdout=log, stderr=subprocess.STDOUT)
    destination.with_suffix('.timing.json').write_text(json.dumps({'seconds':time.monotonic()-started,'load_before':load_before,'load_after':Path('/proc/loadavg').read_text().strip(),'command':command})+'\n')
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
    return result.returncode, "".join(output)


def check_entry(repository, worktree, catalog, entry, logs):
    number, name = entry["number"], entry["name"]
    prefix = f"{number:02} {name}"
    patch = catalog / entry["patch"]
    paths = git(worktree, "apply", "--numstat", str(patch), capture_output=True, text=True)
    changed = [line.split("\t", 2)[-1] for line in paths.stdout.splitlines()]
    if paths.returncode or not changed or any(
        not path.startswith("internal/") or "/oracle/" in path or "/testdata/" in path
        or path.endswith("_test.go") for path in changed
    ):
        print(prefix + ": invalid-patch (production code only)", flush=True)
        return False
    checked = git(worktree, "apply", "--check", str(patch), capture_output=True, text=True)
    if checked.returncode:
        (logs / f"{number:02}-apply.log").write_text(checked.stderr)
        print(prefix + ": no-longer-applies (needs a refresh)", flush=True)
        return False
    # The named tests must still pass on the supplied commit. An unrelated red baseline
    # must never be reported as successful detection of this mutation.
    baseline_code, baseline = run_test(worktree, entry, logs / f"{number:02}-control.log")
    if baseline_code or "no tests to run" in baseline:
        print(prefix + ": baseline-fails-or-test-missing (needs investigation)", flush=True)
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
        print(prefix + ": applies-and-fails-as-recorded", flush=True)
        return True
    print(prefix + ": applies-but-failure-not-as-recorded (needs investigation)", flush=True)
    return False


def main(commit, entry_number=None):
    catalog = Path(__file__).resolve().parent.parent
    repository = Path(subprocess.check_output(["git", "-C", str(catalog), "rev-parse", "--show-toplevel"], text=True).strip())
    resolved = subprocess.check_output(["git", "-C", str(repository), "rev-parse", "--verify", commit + "^{commit}"], text=True).strip()
    entries = json.loads((catalog / "catalog.json").read_text())
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
    cpu_max = Path("/sys/fs/cgroup/cpu.max").read_text().strip()
    print(f"build flags: commit={resolved}; nproc={processors}; cpu.max={cpu_max}; "
          + "; ".join(versions) + "; ADAMIC_GATE_UNCACHED=1; go test -count=1; "
          "native C11 strict warnings, -ffp-contract=off -fno-optimize-sibling-calls; "
          "release -O2; sanitizer -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all", flush=True)
    failure = False
    with tempfile.TemporaryDirectory(prefix="adamic-catalog-worktree-") as temporary:
        worktree = Path(temporary) / "source"
        with (logs / "setup.log").open("w") as setup:
            git(repository, "worktree", "add", "--detach", str(worktree), resolved,
                check=True, stdout=setup, stderr=subprocess.STDOUT)
        try:
            with (logs / "setup.log").open("a") as setup:
                git(worktree, "-c", "submodule.cohere.url=https://github.com/system-inc/cohere.git",
                    "submodule", "update", "--init", "--recursive", "--depth", "1",
                    check=True, stdout=setup, stderr=subprocess.STDOUT)
            for entry in entries:
                started = time.monotonic()
                load_before = Path("/proc/loadavg").read_text().strip()
                if entry["status"] == "skipped":
                    print(f"{entry['number']:02} {entry['name']}: skipped ({entry['reason']})", flush=True)
                    print(f"{entry['number']:02} wall={time.monotonic() - started:.3f}s load={load_before} -> {Path('/proc/loadavg').read_text().strip()}", flush=True)
                    continue
                # Every entry starts from the same clean tracked state, including submodules.
                status = git(worktree, "status", "--porcelain", capture_output=True, text=True, check=True)
                if status.stdout:
                    raise RuntimeError("worktree is not clean: " + status.stdout)
                try:
                    if not check_entry(repository, worktree, catalog, entry, logs):
                        failure = True
                finally:
                    print(f"{entry['number']:02} wall={time.monotonic() - started:.3f}s load={load_before} -> {Path('/proc/loadavg').read_text().strip()}", flush=True)
        finally:
            # This disposable directory and its worktree were created by this invocation.
            git(repository, "worktree", "remove", "--force", str(worktree), check=True)
    return 1 if failure else 0


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("commit")
    parser.add_argument("--entry", type=int, metavar="NN")
    arguments = parser.parse_args()
    sys.exit(main(arguments.commit, arguments.entry))
