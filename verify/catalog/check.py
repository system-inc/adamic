"""Validate independent production-code mutants in a disposable, pinned worktree."""
import json
import os
from pathlib import Path
import re
import shlex
import subprocess
import sys
import tempfile


def git(repository, *arguments, **options):
    return subprocess.run(["git", "-C", str(repository), *arguments], **options)


def run_test(worktree, entry, destination):
    # Parse the recorded command without shell evaluation. No result cache is added here.
    command = shlex.split(entry["test_command"])
    if command.pop(0) != "ADAMIC_GATE_UNCACHED=1" or command != entry["test_args"]:
        raise ValueError("test_command must match test_args and request uncached execution")
    with destination.open("w") as log:
        result = subprocess.run(command, cwd=worktree, env=dict(os.environ, ADAMIC_GATE_UNCACHED="1"),
                                stdout=log, stderr=subprocess.STDOUT)
    return result.returncode, destination.read_text()


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
    # The fix's fixture must still pass on the supplied commit. An unrelated red baseline
    # must never be reported as successful detection of this mutation.
    baseline_code, baseline = run_test(worktree, entry, logs / f"{number:02}-control.log")
    if baseline_code or "no tests to run" in baseline:
        print(prefix + ": baseline-fails-or-fixture-missing (needs investigation)", flush=True)
        return False
    git(worktree, "apply", str(patch), check=True)
    try:
        mutant_code, mutant = run_test(worktree, entry, logs / f"{number:02}-mutant.log")
    finally:
        git(worktree, "apply", "-R", str(patch), check=True)
    test_name = "TestNativeAgreesWithNode/" + entry["fixture"]
    failed_fixture = re.search(r"--- FAIL: " + re.escape(test_name) + r" \(", mutant)
    # Ignore source line-number drift, but require the entire diagnostic message as recorded.
    expected = re.sub(r"^[^:]+\.go:\d+: ", "", entry["expected_failure_line"])
    diagnostics = [re.sub(r"^\s*[^:]+\.go:\d+: ", "", line) for line in mutant.splitlines()]
    if mutant_code == 1 and failed_fixture and expected in diagnostics:
        print(prefix + ": applies-and-fails-as-recorded", flush=True)
        return True
    print(prefix + ": applies-but-failure-not-as-recorded (needs investigation)", flush=True)
    return False


def main(commit):
    catalog = Path(__file__).resolve().parent
    repository = Path(subprocess.check_output(["git", "-C", str(catalog), "rev-parse", "--show-toplevel"], text=True).strip())
    resolved = subprocess.check_output(["git", "-C", str(repository), "rev-parse", "--verify", commit + "^{commit}"], text=True).strip()
    entries = json.loads((catalog / "catalog.json").read_text())
    # Leave test and setup logs available after removing the worktree.
    logs = Path(tempfile.mkdtemp(prefix="adamic-catalog-logs-"))
    print("target " + resolved, flush=True)
    print("logs " + str(logs), flush=True)
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
                if entry["status"] == "skipped":
                    print(f"{entry['number']:02} {entry['name']}: skipped ({entry['reason']})", flush=True)
                    continue
                # Every entry starts from the same clean tracked state, including submodules.
                status = git(worktree, "status", "--porcelain", capture_output=True, text=True, check=True)
                if status.stdout:
                    raise RuntimeError("worktree is not clean: " + status.stdout)
                if not check_entry(repository, worktree, catalog, entry, logs):
                    failure = True
        finally:
            # This disposable directory and its worktree were created by this invocation.
            git(repository, "worktree", "remove", "--force", str(worktree), check=True)
    return 1 if failure else 0


if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.exit("usage: check.py <commit>")
    sys.exit(main(sys.argv[1]))
