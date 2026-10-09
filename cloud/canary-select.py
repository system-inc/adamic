#!/usr/bin/env python3
"""What main's tip gated against main~N selects, for N from 10 down to 1, and what each selected package costs, so the
fast-gate watcher can size its half-hourly canary to a slot (#psh61tb, @system_adamic, Oct 9 19:01Z):

    cloud/canary-select.py --repo <checkout> --tools <tools checkout> --sha <main's tip> --cache <directory> [--box <box>]...

The selection is run.py's own (cloud/fast-gate/run.py --select, from the tools that will run the canary), over a
checkout of the tip kept at <cache>/tree with its submodules, listed for linux/amd64 as the box lists it; one go list
serves every depth. A package's cost is its top-level tests' seconds from the boxes' own history (each box's
~/fast-gate/test-seconds.tsv, the table run.py's queue orders tests by, the highest of the boxes' readings), less the
deferred tests the gate skips, and a test with no recorded time counts as run.py counts it: 60 s. Compiles aren't
counted. Prints one line per selected package, "<N>\\t<package>\\t<seconds>\\t<longest test's seconds>", a
"<N>\\t.\\t0\\t0" line for every depth it sized, and "<N>\\t!\\t<why>" for a depth it couldn't.
"""
import argparse
import contextlib
import glob
import importlib.util
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile

module = "github.com/system-inc/adamic/"
depths = range(10, 0, -1)
# run.py's default for a test with no history (testSplit's history.get).
unknownSeconds = 60
topLevel = re.compile(r"^func\s+((?:Test|Example|Fuzz)\w*)\s*\(", re.M)


def git(directory, *arguments, quiet=False):
    return subprocess.run(["git", "-C", directory] + list(arguments), capture_output=quiet, text=True, check=True).stdout


def placeTree(repo, tree, sha):
    """The tip, submodules included (go.work names cohere/TypeScript/tsc): about six minutes the first time, then a
    switch and a no-op submodule update while the pin holds."""
    git(repo, "fetch", "-q", "origin", sha, quiet=True)
    if not os.path.isdir(tree):
        git(repo, "worktree", "add", "-q", "--detach", tree, sha, quiet=True)
    else:
        git(tree, "switch", "-q", "--detach", sha, quiet=True)
    git(tree, "submodule", "update", "-q", "--init", "--recursive", quiet=True)


def boxSeconds(boxes, cache, run):
    """Each box's test-seconds.tsv, fetched now or the last copy fetched; per test, the highest reading."""
    seconds = {}
    for box in boxes:
        path = os.path.join(cache, "test-seconds-%s.tsv" % box)
        fetched = subprocess.run(["ssh", "-o", "BatchMode=yes", "-o", "ConnectTimeout=10", box, "cat ~/fast-gate/test-seconds.tsv"],
                                 capture_output=True, text=True, timeout=120)
        if fetched.returncode == 0 and fetched.stdout:
            with open(path + ".partial", "w") as handle:
                handle.write(fetched.stdout)
            os.replace(path + ".partial", path)
        for key, (value, _) in run.TestSeconds(path).read().items():
            seconds[key] = max(seconds.get(key, 0), value)
    return seconds


def load(tools):
    spec = importlib.util.spec_from_file_location("fastGateRun", os.path.join(tools, "cloud/fast-gate/run.py"))
    run = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(run)
    return run


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo", required=True)
    parser.add_argument("--tools", required=True)
    parser.add_argument("--sha", required=True)
    parser.add_argument("--cache", required=True)
    parser.add_argument("--box", action="append", default=[])
    parser.add_argument("--tree", help="the checkout of the tip to keep (default <cache>/tree)")
    arguments = parser.parse_args()
    os.makedirs(arguments.cache, exist_ok=True)
    tree = arguments.tree or os.path.join(arguments.cache, "tree")
    placeTree(arguments.repo, tree, arguments.sha)
    run = load(arguments.tools)
    seconds = boxSeconds(arguments.box, arguments.cache, run)
    # The box's package set, not the Mac's: files behind a linux build tag are in it.
    os.environ.update(GOOS="linux", GOARCH="amd64", CGO_ENABLED="1")
    listings = {}

    class Selecting(run.Gate):
        # The tree is the same for every depth, so its go list runs once.
        def command(self, command, cwd=None, **options):
            if command[:2] != ["go", "list"]:
                return super().command(command, cwd, **options)
            key = (tuple(command), cwd)
            if key not in listings:
                listings[key] = super().command(command, cwd, **options)
            return listings[key]

    lines = []
    for depth in depths:
        out = tempfile.mkdtemp(prefix="select-", dir=arguments.cache)
        gate = None
        try:
            base = git(tree, "rev-parse", "%s~%d" % (arguments.sha, depth), quiet=True).strip()
            gate = Selecting(argparse.Namespace(tree=tree, sha=arguments.sha, base=base, base_name="main~%d" % depth,
                                                tools=arguments.tools, out=out, full=False, complete=False, select=True,
                                                branch="canary/main", branch_source="given", session="", session_source="none",
                                                phase=None, phases=None, census=None, unit=None, weights=None, run_to_end=False))
            # run.py prints a failure to stdout, which is this script's answer.
            with contextlib.redirect_stdout(sys.stderr):
                gate.run()
            if gate.failure is not None or not os.path.exists(os.path.join(out, "select.json")):
                why = (gate.failure or {}).get("detail", "no select.json").strip().splitlines()[0]
                lines.append("%d\t!\t%s at %s" % (depth, why[:200], (gate.failure or {}).get("step", "select")))
                continue
            lines.append("%d\t.\t0\t0" % depth)
            with open(os.path.join(out, "select.json")) as handle:
                selection = json.load(handle)
            for package in sorted(selection["packages"]):
                directory = gate.packageDirectories.get(package, os.path.join(tree, package[len(module):]))
                names = set()
                for path in glob.glob(os.path.join(directory, "*_test.go")):
                    with open(path, errors="replace") as handle:
                        names.update(topLevel.findall(handle.read()))
                names -= {"TestMain"} | set(selection.get("deferred", {}).get(package, []))
                costs = [seconds.get((package, name), unknownSeconds) for name in names]
                lines.append("%d\t%s\t%.1f\t%.1f" % (depth, package[len(module):] if package.startswith(module) else package,
                                                     sum(costs), max(costs, default=0)))
        except Exception as error:  # A depth it can't size is named, never a crash that sizes none.
            lines.append("%d\t!\t%s" % (depth, (str(error).strip().splitlines() or [type(error).__name__])[0][:200]))
        finally:
            # Gate makes a build-store directory beside the tree; this run's own, so it goes with the run.
            if gate is not None:
                shutil.rmtree(os.path.dirname(gate.buildStoreEnvironment["ADAMIC_BUILD_STORE_SPOOL"]), ignore_errors=True)
            shutil.rmtree(out, ignore_errors=True)
    print("\n".join(lines))


if __name__ == "__main__":
    main()
