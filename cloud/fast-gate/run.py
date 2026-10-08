#!/usr/bin/env python3
"""The fast gate, run on the gate box against a checkout of the candidate.

It builds and vets the whole repository, tests every package the candidate touches (its diff
against the base, by package), runs the oracle smoke set, and checks the skip census over the
run's log. It stops at the first failure: the first failing test or package build is printed
first, every other step is killed, and status.txt goes red naming it. Tests always run with
-count=1; only compiles are reused, from Go's build cache.

Usage: run.py --tree <checkout of the candidate> --sha <candidate> --base <main sha>
              --tools <tools checkout> --out <directory>

With --full it is the whole uncached gate of main instead, run after landing: every package, the
oracle whole, TestWASI with the WASI SDK's clang first. The first failure turns status.txt red at
once (landings pause on it) and the rest keeps running for triage only; full.json is written when
the run ends.
"""

import argparse
import base64
import fnmatch
import glob
import hashlib
import json
import re
import os
import shutil
import signal
import socket
import subprocess
import sys
import threading
import time
import traceback

module = "github.com/system-inc/adamic"
# What the whole gate sets: no cached results, and the gate inputs' lanes on (see cloud/setup.sh --gate-inputs).
gateEnvironment = {"ADAMIC_GATE_UNCACHED": "1", "ADAMIC_TEST_WASI": "1", "ADAMIC_ORACLE_WASI": "1", "ADAMIC_GATE_COHERE": "1"}
fullPackageTimeout = "3h"
slowPackageSeconds = 3600
smokeTest = "TestNativeAgreesWithNode"
oracle = module + "/internal/oracle"
# A command run by literal name: exec.Command("x", exec.CommandContext(ctx, "x", exec.LookPath("x").
toolLiteral = re.compile(r'exec\.(?:Command\(|CommandContext\([A-Za-z_.()]+, |LookPath\()"([^"]+)"')
toolLiteralSearch = r'exec\.(Command\(|CommandContext\([A-Za-z_.()]+, |LookPath\()"[^"]+"'  # the same, for git grep -E
# The oracle's lanes: every test that runs each registered fixture as a subtest named by its path.
# A test found ranging over the fixtures that isn't listed here makes any oracle change whole.
oracleLanes = {"TestNativeAgreesWithNode": [], "TestWASIAgreesWithNode": [], "TestWASIEmission": [], "TestCountsAreRecorded": ["fixtures"]}
fixtureEntry = re.compile(r'^\s*\}?\{?"(internal/oracle/testdata/[^"]+)", (true|false), (true|false)\}?\)?,?\s*$')


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--tree", required=True)
    parser.add_argument("--sha", required=True)
    parser.add_argument("--base", required=True)
    parser.add_argument("--tools", required=True)
    parser.add_argument("--out", required=True)
    # The gate's CPUs come from its affinity (cloud/fast-gate.sh pins a slot, full-gate-main.sh the full
    # gate). Half of them run test processes, each at most two parallel subtests, since each subtest
    # also runs clang and Node: a slot that oversubscribed itself sent a 150 s node deadline red.
    parser.add_argument("--parallel", type=int, default=max(1, len(os.sched_getaffinity(0)) // 2))
    parser.add_argument("--full", action="store_true")
    parser.add_argument("--branch", default="")
    parser.add_argument("--branch-source", default="")
    parser.add_argument("--session", default="")
    parser.add_argument("--session-source", default="")
    parser.add_argument("--weights", help="package seconds, longest first, to order the full gate's packages")
    arguments = parser.parse_args()
    os.makedirs(arguments.out, exist_ok=True)
    gate = Gate(arguments)
    try:
        gate.run()
    except BaseException:
        gate.fail("runner", traceback.format_exc())
    finally:
        gate.finish()
    sys.exit(0 if gate.failure is None else 1)


class Gate:
    def __init__(self, arguments):
        self.arguments = arguments
        self.started = time.monotonic()
        self.failure = None
        self.lock = threading.Lock()
        self.processes = []
        self.steps = {}
        self.counts = {"pass": 0, "fail": 0, "skip": 0}
        self.skipped = []
        self.census = {"required_input": [], "unclassified": []}
        # Fail closed: green needs every planned stage to have recorded exit 0, not just no failure,
        # so a stage that died without reporting (an exception, a process that never started) is red.
        self.exits = {}
        self.watchers = []
        self.planned = ["coverage", "tools", "build", "vet", "tests", "wasi", "stage3", "catalog", "census"] if arguments.full else ["coverage", "tools", "build", "vet", "tests", "smoke", "census"]
        self.result = {
            "sha": arguments.sha,
            "branch": arguments.branch,
            "branch_source": arguments.branch_source,
            "session": arguments.session,
            "session_source": arguments.session_source,
            "base": arguments.base,
            "tools_sha": self.git(arguments.tools, "rev-parse", "HEAD"),
            "machine": {"hostname": socket.gethostname(), "nproc": os.cpu_count()},
            "uncached_tests": True,
            "build_ok": False,
            "vet_ok": False,
        }
        self.kind = "full" if arguments.full else "fast"
        self.status("running: %s gate of %s against %s" % (self.kind, arguments.sha, arguments.base))

    def run(self):
        tree = self.arguments.tree
        head = self.git(tree, "rev-parse", "HEAD")
        if head != self.arguments.sha:
            self.fail("setup", "the tree is at %s, not the candidate %s" % (head, self.arguments.sha))
            return
        self.npmCli()
        if not self.npmPackages():
            return
        if self.arguments.full:
            self.runFull()
            return
        smoke, smokeSource = self.smokeList()
        self.deferred = self.deferredList()
        # Unquoted, so a path with non-ASCII bytes is itself and can match its package or a rule.
        changed = self.git(tree, "-c", "core.quotePath=false", "diff", "--name-only", "%s...%s" % (self.arguments.base, self.arguments.sha)).split("\n")
        changed = [path for path in changed if path]
        packages, unowned = self.touched(changed)
        # Stage 1's corpus tests sample in the landing gate (@system_adamic's ruling; the interface agreed
        # with @system_cohere_adamic): the main sha the gate diffs against sets the stride's offset, and the
        # changed paths name the corpus files that always run. The full gate sets neither, so it runs all.
        changedList = os.path.join(os.path.abspath(self.arguments.out), "changed-paths.txt")
        with open(changedList, "w") as handle:
            handle.write("".join(path + "\n" for path in changed))
        self.sampling = {"ADAMIC_GATE_SAMPLE": self.arguments.base, "ADAMIC_GATE_CHANGED": changedList}
        # Paths the change adds or deletes (renames as both): a "reads ... paths" line fires only on these.
        pathSetChanged = self.git(tree, "-c", "core.quotePath=false", "diff", "--name-only", "--no-renames", "--diff-filter=AD", "%s...%s" % (self.arguments.base, self.arguments.sha)).split("\n")
        executors = self.cover(unowned, changed, [path for path in pathSetChanged if path])
        if executors is None:
            return
        if "cohere" in executors:
            packages = sorted(set(packages) | set(self.goList("./stage1/cohere/...")))
        packages = sorted(set(packages) | set(self.extraPackages))
        if "stage3" in executors:
            packages = sorted(set(packages) | set(self.goList("./stage3/...")))
            self.planned.append("stage3")
        self.result.update({
            "changed_files": changed,
            "packages": packages,
            "unowned_files": unowned,
            "smoke_list": "cloud/fast-gate/smoke.txt",
            "smoke_list_source": smokeSource,
            "smoke_list_blob": self.git(smokeSource["root"], "hash-object", os.path.join(smokeSource["root"], "cloud/fast-gate/smoke.txt")),
            "smoke_fixtures": ["%s %s" % entry for entry in smoke],
        })
        if not smoke:
            self.fail("smoke", "no smoke list in the gated tree or the tools checkout")
            return
        # Everything starts at once: the tests compile what they need through the same build cache,
        # and the first failure of any step still stops all of them.
        log = open(os.path.join(self.arguments.out, "test.jsonl"), "w")
        threads = [self.guarded("tools", self.toolsDeclared), self.guarded("build", self.build), self.guarded("vet", self.vet),
                   self.guarded("tests", self.testSplit, packages, log),
                   self.guarded("smoke", self.smoke, smoke, log)]
        if "stage3" in executors:
            threads.append(self.guarded("stage3", self.stage3))
        if executors & {"workers", "bench-workers"}:
            self.planned.append("workers")
            threads.append(self.guarded("workers", self.workers, executors))
        if "a-check" in executors and self.result.get("unchecked_a_files"):
            self.planned.append("a-check")
            threads.append(self.guarded("a-check", self.aCheck, self.result["unchecked_a_files"]))
        staling = self.catalogEntriesTouched(changed)
        if staling:
            self.planned.append("catalog-apply")
            threads.append(self.guarded("catalog-apply", self.catalogApply, staling))
        if "catalog" in executors:
            self.planned.append("catalog")
            threads.append(self.guarded("catalog", self.step, "catalog", ["bash", "-n", "verify/catalog/check.sh"]))
        for thread in threads:
            thread.start()
        for thread in threads:
            thread.join()
        log.close()
        if self.failure is not None:
            return
        self.checkCensus()

    def runFull(self):
        listing = self.command(["go", "list", "./..."], cwd=self.arguments.tree, capture_output=True, text=True, check=True).stdout.split()
        order = []
        if self.arguments.weights and os.path.exists(self.arguments.weights):
            with open(self.arguments.weights) as handle:
                weighed = [line.split() for line in handle if line.strip()]
            order = [name for _, name in sorted(weighed, key=lambda row: -float(row[0])) if name in listing]
        packages = order + [name for name in listing if name not in order]
        self.result.update({"packages": "all", "package_list": packages})
        # The no-executor check holds on main too: what this main changed against its first parent.
        parents = self.git(self.arguments.tree, "rev-list", "--parents", "-n", "1", self.arguments.sha).split()[1:]
        changed = self.git(self.arguments.tree, "-c", "core.quotePath=false", "diff", "--name-only", parents[0], self.arguments.sha).split("\n") if parents else []
        _, unowned = self.touched([path for path in changed if path])
        self.cover(unowned, [path for path in changed if path])
        try:
            self.result["build_ok"] = self.step("build", ["go", "build", "./..."])
        except BaseException:
            self.fail("build", traceback.format_exc())
        log = open(os.path.join(self.arguments.out, "test.jsonl"), "w")
        wasi = None
        if os.environ.get("WASI_SYSROOT"):
            wasi = {"PATH": os.path.join(os.path.dirname(os.path.dirname(os.environ["WASI_SYSROOT"])), "bin") + os.pathsep + os.environ["PATH"]}
        threads = [self.guarded("tools", self.toolsDeclared), self.guarded("vet", self.vet),
                   # Three hours a package (@system_adamic, Oct 8 00:08): the whole gate's 16 CPUs are slower
                   # on purpose, and a wall-clock package timeout there is the per-test fragility at a larger
                   # size. Hangs belong to stall guards that count from output; a package over an hour is
                   # named in the status line as slow, so it never reads as a quiet pass.
                   self.guarded("tests", self.test, "tests", ["go", "test", "-count=1", "-json", "-timeout", fullPackageTimeout, "-p", str(self.arguments.parallel), "-skip", "^TestWASI$"] + packages, log),
                   self.guarded("wasi", self.test, "wasi", ["go", "test", "-count=1", "-json", "-timeout", fullPackageTimeout, "-run", "^TestWASI$", "./internal/native"], log, wasi),
                   self.guarded("stage3", self.stage3),
                   # The bug catalog: each catalogued bug reintroduced and caught, on every main that has it.
                   self.guarded("catalog", self.catalogFull)]
        for thread in threads:
            thread.start()
        for thread in threads:
            thread.join()
        log.close()
        try:
            self.checkCensus()
        except BaseException:
            self.fail("census", traceback.format_exc())

    def cover(self, unowned, changed=(), pathSetChanged=None):
        """Each changed path outside a Go package to its executor (cloud/fast-gate/executors.txt from the
        tools checkout, since what counts as inert is a ruling, not the candidate's to change). Any path
        with none turns the gate red before anything runs. Reads rules independently add tests for
        matching changed paths, including package-owned paths. Returns executors, or None."""
        rules, readers, enforce = [], [], True
        with open(os.path.join(self.arguments.tools, "cloud/fast-gate/executors.txt")) as handle:
            for number, line in enumerate(handle, 1):
                fields = line.split()
                if fields[:2] == ["mode", "report"]:
                    enforce = False  # record uncovered paths without failing, until the rulings are in
                elif fields and fields[0] == "reads":
                    # "reads <package> <glob> paths": the test reads which paths exist, not what they hold
                    # (gitignore walks the whole tree), so only an added or deleted path changes its input.
                    if len(fields) not in (3, 4) or (len(fields) == 4 and fields[3] != "paths"):
                        raise ValueError("invalid reads line %d: %s" % (number, line.strip()))
                    readers.append((number, fields[1], fields[2], len(fields) == 4))
                elif fields and not fields[0].startswith("#") and fields[0] != "mode":
                    rules.append((fields[0], fields[1]))
        # "a-check-exempt <glob>" lines aren't executors: they name .a files that aren't Adamic programs.
        exempt = [glob for name, glob in rules if name == "a-check-exempt"]
        rules = [rule for rule in rules if rule[0] != "a-check-exempt"]
        executors, covered, uncovered, unchecked = set(), {}, [], []
        self.result["reads"] = []
        readerPackages = set()
        mapChanged = "cloud/fast-gate/executors.txt" in changed
        for number, package, pattern, pathsOnly in readers:
            # Without the added and deleted list (a caller that didn't compute it), every match fires.
            paths = [path for path in changed if fnmatch.fnmatchcase(path, pattern) and (not pathsOnly or pathSetChanged is None or path in pathSetChanged)]
            if paths or mapChanged:
                readerPackages.add(module + "/" + package)
                self.result["reads"].append({"line": number, "package": package, "glob": pattern,
                                             "paths": paths, "map_changed": mapChanged})
        if "cloud/fast-gate/executors.txt" in changed:
            # The map itself changed: every executor it names runs, so it can't quietly lose a path.
            executors.update(name for name, _ in rules if name not in ("inert", "build"))
        # Script executors also run for paths a package owns (bench/workers lives in bench's tree).
        scripted = ("stage3", "cohere", "workers", "bench-workers", "catalog", "darwin")
        self.scriptedPaths = {}
        for path in changed:
            executor = next((name for name, glob in rules if fnmatch.fnmatchcase(path, glob)), None)
            if executor in scripted:
                executors.add(executor)
                self.scriptedPaths.setdefault(executor, []).append(path)
        for path in unowned:
            executor = next((name for name, glob in rules if fnmatch.fnmatchcase(path, glob)), None)
            if executor is None:
                uncovered.append(path)
            else:
                executors.add(executor)
                covered[executor] = covered.get(executor, 0) + 1

                if executor == "darwin" and path not in self.result.get("needs_darwin", []):
                    self.result.setdefault("needs_darwin", []).append(path)
        # Every changed .a outside a package gets a-check, beside whatever else runs for it (stage3's
        # probes too): the extension promises proven types wherever the file sits.
        # A ruled exemption is named in the result, never dropped silently.
        unchecked = [path for path in unowned if path.endswith(".a") and not any(fnmatch.fnmatchcase(path, glob) for glob in exempt)]
        if unchecked:
            executors.add("a-check")
        self.result["unchecked_a_files"] = unchecked
        self.result["a_check_exempt"] = [path for path in unowned if path.endswith(".a") and path not in unchecked]
        # "package:<path>" runs that package's tests, for files it reads by path from outside its tree.
        self.extraPackages = sorted(readerPackages | {module + "/" + name.split(":", 1)[1] for name in executors if name.startswith("package:")})
        self.result["executors"] = covered
        self.result["uncovered_files"] = uncovered
        self.steps["coverage"] = 0.0
        self.result["coverage_enforced"] = enforce
        if uncovered and enforce:
            self.exits["coverage"] = 1
            self.fail("coverage", "no gate covers %d changed path%s:\n%s" % (len(uncovered), "" if len(uncovered) == 1 else "s", "\n".join(uncovered[:200])))
            return None
        self.exits["coverage"] = 0
        for path in self.scriptedPaths.get("darwin", []):
            if path not in self.result.get("needs_darwin", []):
                self.result.setdefault("needs_darwin", []).append(path)
        if "darwin" in executors:
            # Never green on this box alone: the Mac's darwin leg is attached by integration.
            self.planned.append("darwin")
            self.exits["darwin"] = 1
            self.fail("darwin", "needs darwin leg: go test ./internal/apple/... on Kirk's Mac, for %s" % ", ".join(self.result.get("needs_darwin", []) or ["the executor map's change"]))
            return None
        return executors

    def goList(self, pattern):
        listing = self.command(["go", "list", pattern], cwd=self.arguments.tree, capture_output=True, text=True)
        return listing.stdout.split() if listing.returncode == 0 else []

    def stage3(self):
        """Stage 3's tier-1 lane: apply's tests, the lane's own tests, and stage3/lane/run.sh, which
        applies the adaptations to the pinned TypeScript and runs the upstream suite and the sanctioned
        API check. Its results (an 81,500-file adapted tree among them) stay beside the tree; only the
        verdict and the logs under 5 MB are copied into the published out directory."""
        started = time.monotonic()
        if not os.path.exists(os.path.join(self.arguments.tree, "stage3/lane/run.sh")):
            self.steps["stage3"] = 0.0
            self.exits["stage3"] = 1
            self.fail("stage3", "this change touches stage3/, but its tree has no stage3/lane/run.sh to verify it with: rebase it onto main, which has the lane")
            return
        results = os.path.realpath(self.arguments.tree) + "-stage3-lane"
        if os.path.lexists(results):
            os.rename(results, "%s-%d" % (results, time.time()))
        # Each run's adapted tree is about 700 MB: keep the last two runs' for triage, drop the older.
        previous = sorted(glob.glob(glob.escape(results) + "-[0-9]*"), key=lambda path: int(path.rsplit("-", 1)[1]))
        for path in previous[:-2]:
            shutil.rmtree(path, ignore_errors=True)
        # The lane's own tests import check and run from their directory (from the root both fail with
        # ModuleNotFoundError), and read the pinned TypeScript API from STAGE3_CACHE/api, so they get
        # the same npm ci of stage3/api that apply.py makes first: 32 of 32 on the box.
        install = 'mkdir -p "$STAGE3_CACHE/api" && cp ../api/package.json ../api/package-lock.json "$STAGE3_CACHE/api/" && npm ci --prefix "$STAGE3_CACHE/api" --ignore-scripts --no-audit --no-fund'
        commands = [("stage3-apply-tests", ["python3", "stage3/test_apply.py"], None),
                    ("stage3-lane-tests", ["bash", "-c", install + " && python3 -m unittest test_check test_table"], os.path.join(self.arguments.tree, "stage3/lane")),
                    ("stage3-lane", ["bash", "stage3/lane/run.sh", results], None)]
        codes = {}
        # Each command gets its own STAGE3_CACHE: apply runs npm ci into the cache's api/, which deletes
        # node_modules first, so two applies sharing one (apply's test and the lane here, or another
        # slot's gate) take typescript away from each other's adaptations mid-run (MODULE_NOT_FOUND,
        # 5a5d4436's red). Only the bare TypeScript mirror is shared, and clone only reads it.
        mirror = os.path.expanduser("~/.cache/adamic-stage3/typescript.git")

        def one(name, command, directory):
            cache = os.path.realpath(self.arguments.tree) + "-stage3-cache/" + name
            if os.path.isdir(cache):
                shutil.rmtree(cache)
            os.makedirs(cache)
            if os.path.isdir(mirror):
                os.symlink(mirror, os.path.join(cache, "typescript.git"))
            with open(os.path.join(self.arguments.out, name + ".log"), "w") as output:
                codes[name] = self.spawn(command, output, subprocess.STDOUT, directory, {"STAGE3_CACHE": cache}).wait()

        threads = [self.guarded("stage3", one, name, command, directory) for name, command, directory in commands]
        for thread in threads:
            thread.start()
        for thread in threads:
            thread.join()
        for name in ("execution.json", "report.json", "verdict.json", "verdict.txt", "patch-set.md"):
            if os.path.isfile(os.path.join(results, name)):
                with open(os.path.join(results, name), "rb") as source, open(os.path.join(self.arguments.out, "stage3-lane-" + name), "wb") as target:
                    target.write(source.read())
        self.steps["stage3"] = round(time.monotonic() - started, 1)
        self.result["stage3_exits"] = codes
        failed = [name for name, _, _ in commands if codes.get(name) != 0]
        self.exits["stage3"] = 1 if failed else 0
        if failed:
            self.fail("stage3", "stage 3 tier 1 failed: %s (logs: %s)" % (", ".join("%s exit %s" % (name, codes.get(name)) for name in failed), ", ".join(name + ".log" for name in failed)))

    def workers(self, executors):
        """Platforms' executors for Cloudflare Workers (their inventory, Oct 7 20:59), each from the
        repository root on Node 24, and node --check for the workerd-only measurement harnesses."""
        started = time.monotonic()
        node = ["node", "--disable-warning=ExperimentalWarning"]
        commands = []
        if "workers" in executors:
            adamic = os.path.realpath(self.arguments.tree) + "-binaries/adamic"
            commands += [("workers-build-adamic", ["go", "build", "-o", adamic, "./cmd/adamic"]),
                         ("workers-verify", node + ["workers/compute/verify.mjs"]),
                         ("workers-replay-all", node + ["workers/replay-all.mjs"]),
                         ("workers-verify-handler", node + ["workers/compute/verify-handler.mjs", adamic])]
            if set(self.scriptedPaths.get("workers", [])) & {"workers/replay.mjs", "workers/compute/verify-url.mjs"}:
                commands.append(("workers-verify-url", node + ["workers/compute/verify-url.mjs"]))
        for path in self.scriptedPaths.get("bench-workers", []):
            if os.path.exists(os.path.join(self.arguments.tree, path)):
                commands.append(("bench-workers-check-" + path.replace("/", "_"), ["node", "--check", path]))
        codes = {}
        # The handler check needs the adamic it builds first; the rest run in order too, to keep it simple.
        for name, command in commands:
            with open(os.path.join(self.arguments.out, name + ".log"), "w") as output:
                codes[name] = self.spawn(command, output).wait()
            if codes[name] != 0:
                break
        self.steps["workers"] = round(time.monotonic() - started, 1)
        self.result["workers_exits"] = codes
        failed = [name for name, _ in commands if codes.get(name) != 0]
        self.exits["workers"] = 1 if failed else 0
        if failed:
            self.fail("workers", "workers executor failed: %s (log %s.log)" % (failed[0], failed[0]))

    def aCheck(self, paths):
        """Each changed .a outside a package through stage 0's front end (adamic c: the checker's proven
        types, then the refusal pass; it stops before C): a clean result or a "can't lower ... yet" stop
        passes, since nothing was refused. A file whose first line is "// a-check: refused <rule>" or
        "// a-check: type error <code>" must fail exactly that way instead."""
        started = time.monotonic()
        adamic = os.path.realpath(self.arguments.tree) + "-binaries/adamic-acheck"
        if self.step("a-check-build", ["go", "build", "-o", adamic, "./cmd/adamic"]) is False:
            self.exits["a-check"] = 1
            return
        results, failed = {}, []
        for path in paths:
            with open(os.path.join(self.arguments.tree, path), errors="replace") as handle:
                header = handle.readline().strip()
            expect = header[len("// a-check:"):].strip() if header.startswith("// a-check:") else ""
            process = self.spawn([adamic, "c", path], subprocess.PIPE, subprocess.PIPE)
            _, errors = process.communicate()
            if process.returncode == 0 or ("can't lower" in errors and " yet" in errors):
                outcome = "checked"
            elif "Adamic 0.1 refuses" in errors:
                outcome = "refused"
            elif " error TS" in errors:
                outcome = "type error"
            else:
                outcome = "failed"
            first = errors.strip().splitlines()[0] if errors.strip() else ""
            if expect.startswith("refused"):
                ok = outcome == "refused" and expect[len("refused"):].strip() in errors
            elif expect.startswith("type error"):
                code = expect[len("type error"):].strip()
                ok = outcome == "type error" and (not code or ("error " + code) in errors)
            else:
                ok = outcome == "checked"
            results[path] = {"outcome": outcome, "expected": expect or "checked", "first": first[:300]}
            if not ok:
                failed.append("%s: %s, expected %s (%s)" % (path, outcome, expect or "checked", first[:200]))
        self.result["a_check"] = results
        self.steps["a-check"] = round(time.monotonic() - started, 1)
        self.exits["a-check"] = 1 if failed else 0
        if failed:
            self.fail("a-check", "a-check failed for %d file%s:\n%s" % (len(failed), "" if len(failed) == 1 else "s", "\n".join(failed[:50])))

    def catalogEntriesTouched(self, changed):
        """The bug catalog's entries whose undo patch edits a file this change touches."""
        directory = os.path.join(self.arguments.tree, "verify/catalog")
        if not os.path.isdir(directory):
            return []
        touched = []
        for name in sorted(os.listdir(directory)):
            if name.endswith(".patch"):
                with open(os.path.join(directory, name), errors="replace") as handle:
                    files = set(re.findall(r"^\+\+\+ b/(\S+)", handle.read(), re.M))
                if files & set(changed):
                    touched.append(name)
        return touched

    def catalogApply(self, patches):
        """Staling moves to where it's cheap to fix (@system_adamic, Oct 7 21:28): an entry's undo patch
        that applied to main and no longer applies with this change makes the change red, naming the
        entry, so the landing that stales it refreshes it. One already stale on main stays a warning."""
        started = time.monotonic()
        stale, staled = [], []
        for name in patches:
            patch = os.path.join(self.arguments.tree, "verify/catalog", name)
            onCandidate = self.spawn(["git", "apply", "--check", patch], subprocess.DEVNULL, subprocess.DEVNULL).wait() == 0
            index = os.path.join(self.arguments.out, "catalog-base.index")
            environment = {"GIT_INDEX_FILE": index}
            self.spawn(["git", "read-tree", self.arguments.base], subprocess.DEVNULL, subprocess.DEVNULL, None, environment).wait()
            onBase = self.spawn(["git", "apply", "--check", "--cached", patch], subprocess.DEVNULL, subprocess.DEVNULL, None, environment).wait() == 0
            if onBase and not onCandidate:
                staled.append(name)
            elif not onCandidate:
                stale.append(name)
        self.result["catalog_apply"] = {"checked": patches, "staled_by_this_change": staled, "already_stale": stale}
        self.steps["catalog-apply"] = round(time.monotonic() - started, 1)
        self.exits["catalog-apply"] = 1 if staled else 0
        if staled:
            self.fail("catalog-apply", "this change stops the bug catalog's undo patch from applying: %s. Refresh it in this landing (verify/catalog/check.sh --entry NN <sha> must say applies-and-fails-as-recorded)" % ", ".join(staled))

    def catalogFull(self):
        if not os.path.exists(os.path.join(self.arguments.tree, "verify/catalog/check.sh")):
            self.exits["catalog"] = 0
            self.steps["catalog"] = 0.0
            self.result["catalog"] = "not in this tree"
            return
        # check.sh takes the commit whose bugs it reintroduces and proves caught. Its verdict per entry:
        # applies-and-fails-as-recorded (caught), skipped (with its reason), no-longer-applies (the undo
        # patch went stale as the code moved: the catalog's upkeep, recorded, not main's red), or
        # applies-but-failure-not-as-recorded (a reintroduced bug no longer caught as recorded: red).
        started = time.monotonic()
        with open(os.path.join(self.arguments.out, "catalog.log"), "w") as output:
            code = self.spawn(["bash", "verify/catalog/check.sh", self.arguments.sha], output).wait()
        self.steps["catalog"] = round(time.monotonic() - started, 1)
        with open(os.path.join(self.arguments.out, "catalog.log")) as handle:
            lines = handle.read().splitlines()
        uncaught = [line for line in lines if "applies-but-failure-not-as-recorded" in line]
        stale = [line.split(":")[0] for line in lines if "no-longer-applies" in line]
        caught = [line for line in lines if "applies-and-fails-as-recorded" in line]
        self.result["catalog"] = {"exit": code, "caught": len(caught), "stale": stale, "uncaught": uncaught}
        if uncaught or (code != 0 and not stale) or not caught:
            self.exits["catalog"] = 1
            self.fail("catalog", "the bug catalog: %s" % ("; ".join(uncaught) if uncaught else "check.sh exited %d with no entry caught or stale (see catalog.log)" % code))
        else:
            self.exits["catalog"] = 0

    def npmPackages(self):
        """The pinned npm packages a tree's tests read (stage3/api's @types/node for node:* imports),
        installed with npm ci from its lockfile, once per lockfile: the install lives in a cache keyed
        by the lockfile's hash and hardlinked into the tree, so an unchanged lockfile costs nothing and a
        changed one gets a fresh install."""
        for directory in ("stage3/api",):
            lockfile = os.path.join(self.arguments.tree, directory, "package-lock.json")
            if not os.path.exists(lockfile):
                continue
            started = time.monotonic()
            with open(lockfile, "rb") as handle:
                key = hashlib.sha256(handle.read()).hexdigest()
            cache = os.path.join(os.path.expanduser("~/fast-gate/npm"), key)
            if not os.path.isdir(os.path.join(cache, "node_modules")):
                staging = cache + ".staging-%d" % os.getpid()
                os.makedirs(staging)
                for name in ("package.json", "package-lock.json"):
                    with open(os.path.join(self.arguments.tree, directory, name), "rb") as source, open(os.path.join(staging, name), "wb") as target:
                        target.write(source.read())
                process = self.command(["node", self.npmCli(), "ci", "--ignore-scripts", "--no-audit", "--no-fund", "--install-strategy=hoisted",
                                          "--registry=https://registry.npmjs.org", "--cache", os.path.join(staging, ".npm-cache")],
                                         cwd=staging, capture_output=True, text=True, env=dict(os.environ, npm_config_update_notifier="false"))
                if process.returncode != 0:
                    self.fail("setup", "npm ci for %s failed:\n%s" % (directory, (process.stdout + process.stderr)[-4000:]))
                    return False
                try:
                    os.rename(staging, cache)
                except OSError:
                    pass  # another gate published the same lockfile's install first; its bytes are the same
            # A real directory, as npm would leave it (git ignores node_modules/ only as a directory, and
            # tests that list files must not see a stray link): a hardlink copy of the cached install,
            # kept while its marker names this lockfile. Anything else there moves aside, never deleted.
            target = os.path.join(self.arguments.tree, directory, "node_modules")
            marker = os.path.join(target, ".fast-gate-lockfile-sha256")
            current = open(marker).read().strip() if os.path.isfile(marker) and not os.path.islink(target) else ""
            if current != key:
                if os.path.lexists(target):
                    os.rename(target, os.path.join(os.path.expanduser("~/fast-gate/npm"), "replaced-%d-%d" % (time.time(), os.getpid())))
                self.command(["cp", "-al", os.path.join(cache, "node_modules"), target], check=True)
                with open(marker, "w") as handle:
                    handle.write(key + "\n")
            self.result.setdefault("npm", {})[directory] = {"lockfile_sha256": key, "seconds": round(time.monotonic() - started, 1)}
        return True

    def npmCli(self):
        """The box has Node but no npm, so the npm that setup pins (cloud/markdown-width/npm-bootstrap.json,
        from the tools checkout) is fetched once, its sha512 integrity checked, and kept by that hash."""
        with open(os.path.join(self.arguments.tools, "cloud/markdown-width/npm-bootstrap.json")) as handle:
            pin = json.load(handle)
        home = os.path.join(os.path.expanduser("~/fast-gate/npm"), "npm-" + hashlib.sha256(pin["integrity"].encode()).hexdigest()[:16])
        cli = os.path.join(home, "package/bin/npm-cli.js")
        if not os.path.exists(cli):
            staging = home + ".staging-%d" % os.getpid()
            os.makedirs(staging)
            archive = os.path.join(staging, "npm.tgz")
            self.command(["curl", "-fsSL", pin["url"], "-o", archive], check=True)
            with open(archive, "rb") as handle:
                algorithm, expected = pin["integrity"].split("-", 1)
                if algorithm != "sha512" or base64.b64encode(hashlib.sha512(handle.read()).digest()).decode() != expected:
                    raise ValueError("npm bootstrap integrity mismatch for %s" % pin["url"])
            self.command(["tar", "--no-same-owner", "-xzf", archive, "-C", staging], check=True)
            try:
                os.rename(staging, home)
            except OSError:
                pass  # another gate published the same pinned npm first
        shim = os.path.expanduser("~/fast-gate/npm/bin/npm")
        text = "#!/bin/sh\n# The gates' pinned npm (cloud/markdown-width/npm-bootstrap.json), integrity-checked when fetched.\nexec node %s \"$@\"\n" % cli
        if not os.path.exists(shim) or open(shim).read() != text:
            os.makedirs(os.path.dirname(shim), exist_ok=True)
            with open(shim + ".%d" % os.getpid(), "w") as handle:
                handle.write(text)
            os.chmod(shim + ".%d" % os.getpid(), 0o755)
            os.replace(shim + ".%d" % os.getpid(), shim)
        return cli

    def guarded(self, stage, target, *arguments):
        """A thread whose any exception (a process that can't start for want of file descriptors, a
        missing binary, a bug here) fails the gate at that stage instead of vanishing with the thread."""
        def body():
            try:
                target(*arguments)
            except BaseException:
                self.fail(stage, traceback.format_exc())
        return threading.Thread(target=body)

    def deferredList(self):
        """Tests the landing gate leaves to the full gate on main (cloud/fast-gate/deferred.txt, by
        @system_adamic's ruling): each over 15 s warm and a mutant test, a randomized or differential
        sweep, or a whole external suite. Read like the smoke list, from the gated tree first."""
        for root in (self.arguments.tree, self.arguments.tools):
            path = os.path.join(root, "cloud/fast-gate/deferred.txt")
            if os.path.exists(path):
                deferred = {}
                with open(path) as handle:
                    for line in handle:
                        fields = line.split()
                        if fields and not fields[0].startswith("#"):
                            deferred.setdefault(module + "/" + fields[0], set()).add(fields[1])
                self.result["deferred_list_blob"] = self.git(root, "hash-object", path)
                return deferred
        return {}

    def smokeList(self):
        for root, name in ((self.arguments.tree, "gated tree"), (self.arguments.tools, "tools checkout")):
            path = os.path.join(root, "cloud/fast-gate/smoke.txt")
            if os.path.exists(path):
                with open(path) as handle:
                    # "fixture" runs in TestNativeAgreesWithNode; "Lane fixture" in that lane.
                    fixtures = [(line.split() if len(line.split()) == 2 else [smokeTest, line.strip()]) for line in handle if line.strip() and not line.startswith("#")]
                return [tuple(entry) for entry in fixtures], {"root": root, "from": name}
        return [], {"root": self.arguments.tools, "from": "missing"}

    def touched(self, changed):
        """Each changed file's package: its own directory's, one that embeds it, or the nearest package
        directory above it (its testdata, a lint rule's directory, a fixture subdirectory: what lives in
        a package's tree is that package's to test). Anything else is unowned and goes to cover()."""
        listing = self.command(["go", "list", "-f", "{{.Dir}}\t{{.ImportPath}}\t{{join .EmbedFiles \",\"}}\t{{join .TestEmbedFiles \",\"}}\t{{join .XTestEmbedFiles \",\"}}", "./..."],
                                 cwd=self.arguments.tree, capture_output=True, text=True)
        if listing.returncode != 0:
            raise SystemExit("go list failed: " + listing.stderr)
        directories, embedded = {}, {}
        self.packageDirectories = {}
        tree = os.path.realpath(self.arguments.tree)
        for line in listing.stdout.splitlines():
            directory, importPath, *embeds = line.split("\t")
            relative = os.path.relpath(os.path.realpath(directory), tree)
            directories[relative] = importPath
            self.packageDirectories[importPath] = directory
            for files in embeds:
                for name in filter(None, files.split(",")):
                    embedded[os.path.normpath(os.path.join(relative, name))] = importPath
        packages, unowned = set(), []
        for path in changed:
            directory = os.path.dirname(path) or "."
            owner = directories.get(directory) or embedded.get(path)
            ancestor = directory
            while owner is None and ancestor not in (".", ""):
                ancestor = os.path.dirname(ancestor) or "."
                owner = directories.get(ancestor) if ancestor != "." else None
            if owner is None:
                unowned.append(path)
            else:
                packages.add(owner)
        if oracle in packages:
            self.oracleSelection = self.selectOracle()
            self.result["oracle_selection"] = self.oracleSelection
        return sorted(packages), unowned

    def selectOracle(self):
        """What a change to internal/oracle has to run, when it only adds or edits fixtures: each changed
        fixture in every lane, the Test functions of test files it adds (their init registers the
        fixture), TestCountsAreRecorded whole when counts.md changed, and the smoke set as always. The
        same coverage as the whole package for such a change, since nothing else in the package moved.
        Anything it can't classify (a changed helper, a deleted file, any other file) runs it whole."""
        tree, base, sha = self.arguments.tree, self.arguments.base, self.arguments.sha
        whole = lambda reason: {"whole": True, "reason": reason}
        statuses = [line.split("\t") for line in self.git(tree, "diff", "--name-status", "%s...%s" % (base, sha), "--", "internal/oracle").splitlines() if line]
        fixtures, tests = set(), set()
        directory = os.path.join(tree, "internal/oracle")
        for status, *paths in statuses:
            path = paths[-1]
            name = path[len("internal/oracle/"):]
            if status[0] not in "AM":
                return whole("%s %s" % (status, path))
            if name.startswith("testdata/") and path.endswith(".a"):
                fixtures.add(path)
            elif name == "counts.md":
                tests.add("TestCountsAreRecorded")
            elif name.endswith("_test.go") and "/" not in name and status[0] == "A":
                with open(os.path.join(tree, path)) as handle:
                    source = handle.read()
                tests.update(re.findall(r"^func (Test\w+)\(t \*testing\.T\)", source, re.M))
                fixtures.update(re.findall(r'"(internal/oracle/testdata/[^"]+\.a)"', source))
            elif name.endswith("_test.go") and "/" not in name:
                for line in self.git(tree, "diff", "-U0", "%s...%s" % (base, sha), "--", path).splitlines():
                    if line.startswith(("+++", "---", "@@")) or not line.startswith(("+", "-")):
                        continue
                    entry = fixtureEntry.match(line[1:])
                    if not entry:
                        return whole("%s changes more than fixture entries" % path)
                    fixtures.add(entry.group(1))
            else:
                return whole("%s is not a fixture, a fixture registration or counts.md" % path)
        registered, inputs, tables, sources = set(), set(), {}, []
        for name in os.listdir(directory):
            if name.endswith("_test.go"):
                with open(os.path.join(directory, name)) as handle:
                    source = handle.read()
                sources.append(source)
                literals = re.findall(r'"(internal/oracle/testdata/[^"]+\.a)"', source)
                registered.update(literals)
                if "inputFixtures" in source:
                    # Input fixtures have lanes of their own (TestInputAgreesWithNode and kin); a changed
                    # one runs the package whole rather than be held to the fixtures' lanes alone.
                    inputs.update(literals)
                for function in re.split(r"\nfunc ", source)[1:]:
                    lane = re.match(r"(Test\w+)\(", function)
                    for table in re.findall(r"range (\w*[Ff]ixtures)\b", function):
                        if lane and table == "fixtures" and lane.group(1) not in oracleLanes:
                            return whole("%s ranges over the fixtures and isn't a known lane" % lane.group(1))
                        if lane and table not in ("fixtures", "inputFixtures"):
                            tables.setdefault(table, set()).add(lane.group(1))
        # A test over a table of its own (checkedCastFixtures, typeofNullFixtures) runs whole when a changed
        # fixture is one of its entries, named by path or by bare name, and not at all otherwise: nothing
        # else it reads changed. A table filled anywhere but its own literal runs the package whole.
        source = "\n".join(sources)
        for table, lanes in sorted(tables.items()):
            declaration = re.search(r"^var %s = (.*?)^\}" % table, source, re.M | re.S)
            if not declaration or re.search(r"\b%s\s*(=|\[[^]]*\]\s*=)" % table, source[:declaration.start()] + source[declaration.end():]):
                return whole("%s ranges over %s, a fixture table the selection can't read whole" % (sorted(lanes)[0], table))
            entries = set(re.findall(r'"([^"]+)"', declaration.group(1)))
            if any(path in entries or path[len("internal/oracle/"):] in entries or os.path.basename(path)[:-len(".a")] in entries for path in fixtures):
                tests.update(lanes)
        if fixtures & inputs:
            return whole("an input fixture changed: %s" % ", ".join(sorted(fixtures & inputs)))
        unregistered = sorted(fixtures - registered)
        if unregistered:
            # A changed .a no table names (a module a fixture imports, say): which fixture reads it isn't known here.
            return whole("not a registered fixture: %s" % ", ".join(unregistered))
        return {"whole": False, "fixtures": sorted(fixtures), "tests": sorted(tests - set(oracleLanes) | (tests & {"TestCountsAreRecorded"})), "lanes": sorted(oracleLanes)}

    def step(self, name, command):
        started = time.monotonic()
        with open(os.path.join(self.arguments.out, name + ".log"), "w") as output:
            process = self.spawn(command, output)
            code = process.wait()
        self.steps[name] = round(time.monotonic() - started, 1)
        self.exits[name] = code
        if code != 0 and self.failure is None:
            with open(os.path.join(self.arguments.out, name + ".log")) as handle:
                self.fail(name, handle.read()[-4000:])
        return code == 0

    def toolsDeclared(self):
        """Every command the repository's Go code runs by literal name is declared in the tools checkout's
        cloud/fast-gate/tools.txt, which the box checks before the gate (cloud/fast-gate/tools-check.sh).
        A change that reaches for a new tool without declaring it is red here, naming the file and line,
        instead of red on every box that lacks it (library's stock tsc, Oct 8 11:2xZ)."""
        started = time.monotonic()
        with open(os.path.join(self.arguments.tools, "cloud/fast-gate/tools.txt")) as handle:
            declared = {line.split("\t")[0] for line in handle if line.strip() and not line.startswith("#")}
        found = self.command(["git", "-C", self.arguments.tree, "grep", "-nE", toolLiteralSearch, "--", "*.go", ":!cohere", ":!stage3/upstream"],
                               capture_output=True, text=True)
        if found.returncode not in (0, 1):
            raise RuntimeError("git grep for tool literals failed: %s" % found.stderr)
        undeclared = []
        for line in found.stdout.splitlines():
            path, number, text = line.split(":", 2)
            for name in toolLiteral.findall(text):
                if name not in declared:
                    undeclared.append("%s:%s runs %s, which cloud/fast-gate/tools.txt doesn't declare" % (path, number, name))
        self.steps["tools"] = round(time.monotonic() - started, 1)
        self.result["undeclared_tools"] = undeclared
        self.exits["tools"] = 1 if undeclared else 0
        if undeclared:
            self.fail("tools", "\n".join(undeclared) + "\nDeclare each in cloud/fast-gate/tools.txt (devtools/fast-gate), with a check, and have setup provide it.")

    def build(self):
        self.result["build_ok"] = self.step("build", ["go", "build", "./..."])

    def vet(self):
        self.result["vet_ok"] = self.step("vet", ["go", "vet", "./..."])

    def testSplit(self, packages, log):
        """Each touched package's tests: its test binary built once, then every top-level test run as
        its own process, all of them side by side on the box's threads. One package's tests no longer
        wait on each other in one process (round 59 proved the verdicts identical this way)."""
        started = time.monotonic()
        # Beside the slot's tree, never in the out directory: everything there is published, and a test
        # binary is tens of megabytes. One slot runs one gate at a time, so each run overwrites its own.
        binaries = os.path.realpath(self.arguments.tree) + "-binaries"
        os.makedirs(binaries, exist_ok=True)
        slots = threading.Semaphore(self.arguments.parallel)
        # At most two test binaries build at once: each go test -c runs its own pool of compile processes as
        # wide as the slot, so a big gate building one per test slot ran hundreds of compiles in a 12-CPU
        # slot, and three such gates took Cloud to load 900 and Workshop to 10 GB free (Oct 8 11:2xZ).
        builds = threading.Semaphore(2)
        threads = []
        # Every package compiled and listed, and every test process it planned exited 0: counted, so a
        # process that never ran can't pass for one that did.
        tally = {"packages": 0, "planned": 0, "passed": 0}

        def runPackage(importPath):
            binary = os.path.join(binaries, importPath.replace("/", "_") + ".test")
            with slots:
                if os.path.exists(binary):
                    os.rename(binary, binary + ".previous")
                with builds:
                    built = self.stream("tests", ["go", "test", "-c", "-o", binary, importPath], None)
                if built != 0:
                    return
                if not os.path.exists(binary):
                    # go test -c succeeded and wrote nothing: the package has no test files.
                    with self.lock:
                        tally["packages"] += 1
                    self.result.setdefault("split_tests", {})[importPath] = 0
                    return
                listing = self.capture([binary, "-test.list", "."], self.packageDirectories[importPath])
            if listing is None:
                return
            with self.lock:
                tally["packages"] += 1
            names = [line for line in listing.splitlines() if line.startswith(("Test", "Example", "Fuzz"))]
            selection = getattr(self, "oracleSelection", {"whole": True}) if importPath == oracle else {"whole": True}
            patterns = {}
            if not selection["whole"]:
                # The lanes run only the changed fixtures; the selected tests run whole.
                for lane in oracleLanes:
                    if lane in names and selection["fixtures"] and lane not in selection["tests"]:
                        patterns[lane] = fixturePattern([lane] + oracleLanes[lane], selection["fixtures"])
                for test in selection["tests"]:
                    if test in names:
                        patterns[test] = "^%s$" % test
                names = sorted(patterns)
            deferred = sorted(set(names) & self.deferred.get(importPath, set()))
            if deferred:
                with self.lock:
                    self.result.setdefault("deferred_to_full_gate", []).extend(importPath + " " + name for name in deferred)
                names = [name for name in names if name not in deferred]
            self.result.setdefault("split_tests", {})[importPath] = len(names)
            with self.lock:
                tally["planned"] += len(names)
            tests = []
            for name in names:
                command = ["go", "tool", "test2json", "-t", "-p", importPath, binary, "-test.v=test2json", "-test.paniconexit0",
                           "-test.count=1", "-test.failfast", "-test.timeout=30m", "-test.parallel=2", "-test.run", patterns.get(name, "^%s$" % name)]
                thread = self.guarded("tests", self.slotted, slots, command, log, importPath, name, tally)
                thread.start()
                tests.append(thread)
            for thread in tests:
                thread.join()

        for importPath in packages:
            thread = self.guarded("tests", runPackage, importPath)
            thread.start()
            threads.append(thread)
        for thread in threads:
            thread.join()
        self.steps["tests"] = round(time.monotonic() - started, 1)
        complete = tally["packages"] == len(packages) and tally["passed"] == tally["planned"]
        self.exits["tests"] = 0 if complete and self.failure is None else 1
        self.result["split_tally"] = dict(tally, touched=len(packages))

    def slotted(self, slots, command, log, importPath, name, tally):
        with slots:
            if self.failure is not None:
                return
            environment = None
            if name == "TestWASI" and os.environ.get("WASI_SYSROOT"):
                # As the whole gate runs it: the WASI SDK's clang first, so wasm-ld finds the wasm32 builtins.
                environment = {"PATH": os.path.join(os.path.dirname(os.path.dirname(os.environ["WASI_SYSROOT"])), "bin") + os.pathsep + os.environ["PATH"]}
            if self.stream("tests", command, log, self.packageDirectories[importPath], environment) == 0:
                with self.lock:
                    tally["passed"] += 1

    def capture(self, command, directory):
        process = self.spawn(command, subprocess.PIPE, subprocess.PIPE, directory)
        output, errors = process.communicate()
        if process.returncode != 0:
            self.fail("tests", "%s exited %d\n%s" % (" ".join(command), process.returncode, (output + errors)[-4000:]))
            return None
        return output

    def smoke(self, smoke, log):
        """The pinned smoke set in one process, then a check that every entry passed: an entry whose
        lane exists but whose fixture didn't run is red, never a silent pass. An entry whose lane the
        gated tree doesn't have yet (it grew from a red on a stack that hasn't landed) is listed as
        absent."""
        lanes = sorted({lane for lane, _ in smoke})
        passed, seen = set(), set()

        def watch(line):
            try:
                event = json.loads(line)
            except ValueError:
                return
            test = event.get("Test") or ""
            lane = test.split("/", 1)[0]
            if lane in lanes:
                seen.add(lane)
                if event.get("Action") == "pass" and "/" in test:
                    passed.add((lane, test.split("/", 1)[1]))

        self.watchers.append(watch)
        self.test("smoke", ["go", "test", "-count=1", "-failfast", "-json", "-timeout", "30m", "-run", smokePattern(smoke), "./internal/oracle"], log)
        if self.exits.get("smoke") != 0:
            return
        absent = sorted("%s %s" % entry for entry in smoke if entry[0] not in seen)
        missing = sorted("%s %s" % entry for entry in smoke if entry[0] in seen and entry not in passed)
        self.result["smoke_absent_lanes"] = absent
        if missing:
            self.exits["smoke"] = 1
            self.fail("smoke", "smoke entries that didn't run: %s" % ", ".join(missing))

    def test(self, name, command, log, environment=None):
        started = time.monotonic()
        code = self.stream(name, command, log, None, environment)
        self.exits[name] = 1 if code is None else code
        self.steps[name] = round(time.monotonic() - started, 1)

    def stream(self, name, command, log, directory=None, environment=None):
        """Runs one go test (or test2json) process, copying its events to the log; the first failing
        event fails the gate. Returns the exit code, or None if the gate failed."""
        stderrPath = os.path.join(self.arguments.out, "%s-%d.stderr" % (name, threading.get_ident()))
        stderr = open(stderrPath, "w")
        process = self.spawn(command, subprocess.PIPE, stderr, directory, environment)
        output = {}
        for line in process.stdout:
            if log is not None:
                with self.lock:
                    log.write(line)
            for watch in self.watchers:
                watch(line)
            if "gate-sample: " in line:
                # A corpus test's sample, for status.txt and fast.json: what this landing didn't check.
                try:
                    sampled = json.loads(line).get("Output", "")
                except ValueError:
                    sampled = line
                with self.lock:
                    self.result.setdefault("gate_samples", []).append(sampled.split("gate-sample: ", 1)[1].strip())
            try:
                event = json.loads(line)
            except ValueError:
                continue
            key = (event.get("Package"), event.get("Test"))
            if event.get("Action") == "output":
                output.setdefault(key, []).append(event.get("Output", ""))
                continue
            if event.get("Test") is None and event.get("Action") in ("pass", "fail") and event.get("Elapsed", 0) > slowPackageSeconds:
                with self.lock:
                    self.result.setdefault("slow_packages", {})[event.get("Package")] = event["Elapsed"]
            if event.get("Test") is not None and event.get("Action") in self.counts:
                with self.lock:
                    self.counts[event["Action"]] += 1
                    if event["Action"] == "skip":
                        self.skipped.append(key)
            if event.get("Action") == "fail" or event.get("Action") == "build-fail":
                text = "".join(output.get(key, [])) or "".join(output.get((event.get("Package"), None), []))
                self.fail(name, "%s %s\n%s" % (event.get("Package"), event.get("Test") or "(package)", text[-4000:]))
        code = process.wait()
        stderr.close()
        if code != 0 and self.failure is None:
            with open(stderrPath) as handle:
                self.fail(name, "%s exited %d\n%s" % (" ".join(command[:4]), code, handle.read()[-4000:]))
        return None if self.failure is not None and not self.arguments.full else code

    def checkCensus(self):
        started = time.monotonic()
        tools = self.arguments.tools
        # census-extra.json: skips in main the tools tree doesn't have yet, classified, checked against the log only.
        process = self.spawn(["go", "run", "./internal/skipcensus/cmd", "-root", tools, "-extra", os.path.join(tools, "cloud/fast-gate/census-extra.json"),
                              os.path.join(os.path.abspath(self.arguments.out), "test.jsonl")],
                             subprocess.PIPE, subprocess.PIPE, tools, {"GOWORK": "off"})
        stdout, stderr = process.communicate()
        with open(os.path.join(self.arguments.out, "census.log"), "w") as handle:
            handle.write(stdout + stderr)
        self.steps["census"] = round(time.monotonic() - started, 1)
        self.exits["census"] = process.returncode
        for line in stdout.splitlines():
            fields = line.split("\t")
            if len(fields) == 3 and fields[0] == "required-input":
                self.census["required_input"].append(fields[1] + " " + fields[2])
            if len(fields) == 3 and fields[0] in ("unknown", "unclassified"):
                self.census["unclassified"].append(fields[1] + " " + fields[2])
        if process.returncode != 0:
            self.fail("census", (stdout + stderr)[-4000:])

    def git(self, directory, *arguments):
        return self.command(["git", "-C", directory] + list(arguments),
                            capture_output=True, text=True, check=True).stdout.strip()

    def command(self, command, cwd=None, capture_output=False, text=True, check=False, env=None):
        process = self.spawn(command, subprocess.PIPE if capture_output else None,
                             subprocess.PIPE if capture_output else None, cwd, env)
        stdout, stderr = process.communicate()
        result = subprocess.CompletedProcess(command, process.returncode, stdout, stderr)
        if check:
            result.check_returncode()
        return result

    def spawn(self, command, stdout, stderr=subprocess.STDOUT, directory=None, environment=None):
        with self.lock:
            if self.failure is not None and not self.arguments.full:
                raise SystemExit(1)
            # Uncached: the oracle's result caches would otherwise answer a test without running it.
            variables = dict(os.environ, **gateEnvironment)
            if not self.arguments.full:
                variables.update(getattr(self, "sampling", {}))
            # The box has Node but no npm; stage 3's apply runs npm ci, so the gates' pinned npm is on PATH.
            variables["PATH"] = os.path.expanduser("~/fast-gate/npm/bin") + os.pathsep + variables["PATH"]
            variables.update(environment or {})
            process = subprocess.Popen(command, cwd=directory or self.arguments.tree, stdout=stdout, stderr=stderr, text=True, start_new_session=True, env=variables)
            self.processes.append(process)
        return process

    def fail(self, step, detail):
        with self.lock:
            if self.failure is not None:
                return
            self.failure = {"step": step, "detail": detail, "after_seconds": round(time.monotonic() - self.started, 1)}
            print("FIRST FAILURE (%s, at %.1f s):\n%s" % (step, self.failure["after_seconds"], detail), flush=True)
            if self.arguments.full:
                # Landings pause on this line now; the rest of the run goes on for triage only.
                with open(os.path.join(self.arguments.out, "first-failure.txt"), "w") as handle:
                    handle.write(detail + "\n")
                self.status("red: %s first failure at %s after %.1f s (still running for triage)" % (self.arguments.sha, step, self.failure["after_seconds"]))
                return
            self.killSessions()

    def killSessions(self):
        """Drain our sessions even after a leader exited and children were reparented.

        Setpgid cannot escape a session. Zombies cannot run or fork and must be
        reaped by their new parent; waiting for those would hang on a slow init.
        Call under self.lock so spawn cannot register a new session mid-drain.
        """
        sessions = {process.pid for process in self.processes if process.pid > 0}
        while sessions:
            live = []
            for path in glob.glob("/proc/[0-9]*/stat"):
                try:
                    with open(path) as handle:
                        # comm may contain spaces and ')'; fields after its final ')' start at 3.
                        fields = handle.read().rsplit(")", 1)[1].split()
                    if int(fields[3]) in sessions and fields[0] not in ("Z", "X"):
                        live.append(int(path.split("/")[2]))
                except (FileNotFoundError, ProcessLookupError):
                    continue
            if not live:
                return
            for pid in live:
                try:
                    os.kill(pid, signal.SIGKILL)
                except ProcessLookupError:
                    pass
            time.sleep(0.01)

    def status(self, line):
        with open(os.path.join(self.arguments.out, "status.txt"), "w") as handle:
            handle.write(line + "\n")

    def finish(self):
        with self.lock:
            self.killSessions()
        wall = round(time.monotonic() - self.started, 1)
        unfinished = [stage for stage in self.planned if self.exits.get(stage) != 0]
        if unfinished and self.failure is None:
            self.fail(unfinished[0], "planned stages without a recorded exit 0: %s (exits %s)" % (", ".join(unfinished), self.exits))
        green = self.failure is None and not unfinished
        self.result.update({
            "wall_seconds": wall,
            "steps_seconds": self.steps,
            "pass": self.counts["pass"],
            "fail": self.counts["fail"],
            "skip": self.counts["skip"],
            "required_input_skips": self.census["required_input"],
            "unclassified_skips": self.census["unclassified"],
            "failure": self.failure,
            "planned_stages": self.planned,
            "stages_exit": self.exits,
            "finished": True,
        })
        with open(os.path.join(self.arguments.out, self.kind + ".json"), "w") as handle:
            json.dump(self.result, handle, indent=2)
            handle.write("\n")
        steps = " ".join("%s=%.1fs" % item for item in self.steps.items())
        if self.result.get("slow_packages"):
            steps += "; slow packages, over %d min: %s" % (slowPackageSeconds // 60, ", ".join("%s %.0fs" % (name.rsplit("/", 2)[-2] + "/" + name.rsplit("/", 1)[-1], seconds) for name, seconds in sorted(self.result["slow_packages"].items(), key=lambda item: -item[1])))
        deferred = self.result.get("deferred_to_full_gate", [])
        if not self.arguments.full:
            steps += "; deferred to full gate: %d tests%s" % (len(deferred), (" (" + ", ".join(name.split()[-1] for name in deferred) + ")") if deferred else "")
            steps += "; branch %s, session %s" % (self.arguments.branch or "none", self.arguments.session or "none")
            if self.result.get("gate_samples"):
                steps += "; sampled: %s" % "; ".join(sorted(set(self.result["gate_samples"]))[:10])
            if self.result.get("a_check_exempt"):
                steps += "; a-check exempt by ruling: %d .a files (fast.json a_check_exempt)" % len(self.result["a_check_exempt"])
            if self.result.get("unchecked_a_files") and "a_check" not in self.result:
                steps += "; unchecked .a: %s" % ", ".join(self.result["unchecked_a_files"][:20])
        if green:
            self.status("green: %s %s gate in %.1f s (%s), %d packages, %d pass, %d skip, smoke %d fixtures" % (self.arguments.sha, self.kind, wall, steps, len(self.result.get("package_list", self.result.get("packages", []))), self.counts["pass"], self.counts["skip"], len(self.result.get("smoke_fixtures", []))))
        else:
            self.status("red: %s %s gate, first failure at %s after %.1f s (%s), %d fail, %d pass" % (self.arguments.sha, self.kind, self.failure["step"], self.failure["after_seconds"], steps, self.counts["fail"], self.counts["pass"]))
        with open(os.path.join(self.arguments.out, "status.txt")) as handle:
            print(handle.read(), end="", flush=True)


def fixturePattern(prefix, fixtures):
    """A -run pattern for these fixtures' subtests under the test and parent subtests in prefix,
    level by level, as smokePattern builds it."""
    levels = [[part] for part in prefix]
    for fixture in fixtures:
        for depth, part in enumerate(fixture.split("/"), start=len(prefix)):
            if len(levels) <= depth:
                levels.append([])
            if part not in levels[depth]:
                levels[depth].append(part)
    return "/".join("^(%s)$" % "|".join(re.escape(part) for part in level) for level in levels)


def smokePattern(entries):
    """A -run pattern for exactly these (lane, fixture) entries. Go splits both the pattern and each
    subtest name on '/' and matches level by level, so alternation can't span a slash: each level gets
    its own alternation, and a name shorter than the pattern ignores the levels past its end. A lane
    may also run another lane's fixture that it has; the check after the run reads only its own."""
    levels = [sorted({lane for lane, _ in entries})]
    for _, fixture in entries:
        for depth, part in enumerate(fixture.split("/"), start=1):
            if len(levels) <= depth:
                levels.append([])
            if part not in levels[depth]:
                levels[depth].append(part)
    return "/".join("^(%s)$" % "|".join(part.replace(".", "\\.") for part in level) for level in levels)


def git(directory, *arguments):
    return subprocess.run(["git", "-C", directory] + list(arguments), capture_output=True, text=True, check=True).stdout.strip()


if __name__ == "__main__":
    main()
