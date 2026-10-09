#!/usr/bin/env python3
"""The fast gate, run on the gate box against a checkout of the candidate.

It builds and vets the whole repository, tests every package affected by the candidate (reverse imports and declared compiler
dependencies, from its diff against the base), runs the oracle smoke set, and checks the skip census over the
run's log. It stops at the first failure: the first failing test or package build is printed
first, every other step is killed, and status.txt goes red naming it. Tests always run with
-count=1; only compiles are reused, from Go's build cache.

Usage: run.py --tree <checkout of the candidate> --sha <candidate> --base <main sha>
              --tools <tools checkout> --out <directory>

For pool work, --phase <name> runs one non-test stage, and --phases build,vet,smoke,census
writes one record of those stages. Pass --census <pool test.jsonl> to include the pool's tests in
that census. --list-units --tree <tree> lists the unconditional fast units; supply --base, --sha
and --tools (as for --select) to include this change's conditional executors.
Add --with-inputs for JSON lines with the exact unit line and its package/path inputs.
Package roots include test imports for commands that compile tests; paths are content inputs.
Unbounded runtime readers declare paths ["."] with an inputs.note explaining the limit.
Merged logs, Git history/branch state, toolchain and external spool identity still belong
in the pool's verdict key: repository content alone cannot identify those inputs.

With --full it is the whole uncached gate of main instead, run after landing: every package, the
oracle whole, TestWASI with the WASI SDK's clang first. The first failure turns status.txt red at
once (landings pause on it) and the rest keeps running for triage only; full.json is written when
the run ends.
"""

from concurrent.futures import ThreadPoolExecutor
import fcntl
import heapq
import math
import tempfile
import argparse
import base64
import fnmatch
import glob
import hashlib
import shlex
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
import datetime
import traceback

module = "github.com/system-inc/adamic"
# What the whole gate sets: no cached results, and the gate inputs' lanes on (see cloud/setup.sh --gate-inputs).
gateEnvironment = {"ADAMIC_GATE_UNCACHED": "1", "ADAMIC_TEST_WASI": "1", "ADAMIC_ORACLE_WASI": "1", "ADAMIC_GATE_COHERE": "1"}
fullPackageTimeout = "3h"
slowPackageSeconds = 3600
# Kirk's hard constraint (Oct 8): no test unit runs longer than this, so every unit can go to any of a hundred
# 4-CPU Codex instances. Each gate publishes its ledger of units over it (long-tests.tsv), and a unit over it
# that isn't on the burn-down (cloud/fast-gate/budget-burndown.tsv, main's units over it when the rule came,
# shrinking only) is red at 'budget' when it is new, its test absent from the base. An existing unit drifting over
# on a loaded gate box is named, not red: replayed over Oct 8's last 28 fast gates, red for every unit over would
# have failed 8, all on box load (TestInspectRequestRefusals took 147 to 265 s in 12-CPU slots). The gate measures
# on its own box until Loom's Codex tier measures every unit on the reference shape, and the verdict names it.
longTestSeconds = 30
unitKillSeconds = 90
# Products run to completion under a ten-minute ceiling; only test units get the 90 s kill, and a product over the
# budget is listed beside the verdict, never red (@system_adamic, Oct 9 00:41 MDT, live on Loom's pool first: the boxes'
# 90 s product kill at load 53 redded every candidate on main's own products).
productKillSeconds = 600
smokeTest = "TestNativeAgreesWithNode"
oracle = module + "/internal/oracle"
# A command run by literal name: exec.Command("x", exec.CommandContext(ctx, "x", exec.LookPath("x").
toolLiteral = re.compile(r'exec\.(?:Command\(|CommandContext\([A-Za-z_.()]+, |LookPath\()"([^"]+)"')
toolLiteralSearch = r'exec\.(Command\(|CommandContext\([A-Za-z_.()]+, |LookPath\()"[^"]+"'  # the same, for git grep -E
# The oracle's lanes: every test that runs each registered fixture as a subtest named by its path.
# A test found ranging over the fixtures that isn't listed here makes any oracle change whole.
oracleLanes = {"TestNativeAgreesWithNode": [], "TestWASIAgreesWithNode": [], "TestWASIEmission": [], "TestCountsAreRecorded": ["fixtures"]}
fixtureEntry = re.compile(r'^\s*\}?\{?"(internal/oracle/testdata/[^"]+)", (true|false), (true|false)\}?\)?,?\s*$')


# Variables that narrow what a gate's tests check; any of them set refuses the gate.
scopingEnvironment = ("ADAMIC_LINT_RULES",)


def boxLoad():
    """The box's 1-minute load average and core count, the whole box's and not the gate's slot: what the record
    keeps at a gate's start, its first failure and its end (@system_adamic, from the witness, Oct 9 00:30Z: about six
    of Oct 8's reds and stalls were load, not bugs, each diagnosed by hand)."""
    return {"load_1m": round(os.getloadavg()[0], 1), "cores": os.cpu_count() or 1}


def underLoad(reading):
    return reading is not None and reading["load_1m"] > reading["cores"]


def loadWords(reading):
    return "load %.1f on %d cores" % (reading["load_1m"], reading["cores"])


class TestSeconds:
    def __init__(self, path=None):
        self.path = path or os.path.expanduser("~/fast-gate/test-seconds.tsv")

    def read(self):
        try:
            rows = {}
            with open(self.path) as handle:
                for line in handle:
                    package, name, seconds, parallel = line.rstrip("\n").split("\t")
                    seconds = float(seconds)
                    if not math.isfinite(seconds) or seconds < 0 or parallel not in ("0", "1"):
                        return {}
                    rows[package, name] = (seconds, parallel == "1")
            return rows
        except (OSError, ValueError):
            return {}

    def update(self, observations):
        if not observations:
            return
        temporary = None
        try:
            os.makedirs(os.path.dirname(self.path), exist_ok=True)
            # Lock a stable sibling: locking the replaced inode would lose mutual exclusion.
            with open(self.path + ".lock", "a") as lock:
                fcntl.flock(lock, fcntl.LOCK_EX)
                rows = self.read()
                for key, (seconds, parallel) in observations.items():
                    rows[key] = (max(rows.get(key, (0, False))[0] * 0.8, seconds), parallel)
                with tempfile.NamedTemporaryFile(mode="w", dir=os.path.dirname(self.path), delete=False) as handle:
                    temporary = handle.name
                    for (package, name), (seconds, parallel) in sorted(rows.items()):
                        handle.write("%s\t%s\t%.9g\t%d\n" % (package, name, seconds, parallel))
                    handle.flush()
                    os.fsync(handle.fileno())
                os.replace(temporary, self.path)
                temporary = None
        except OSError:
            pass  # History is a hint, never a gate failure.
        finally:
            if temporary is not None:
                os.unlink(temporary)


class TestQueue:
    def __init__(self, slots):
        self.free = slots
        self.pending = []
        self.closed = False
        self.condition = threading.Condition()

    def add(self, tests):
        """(seconds, parallel, package, name, command) for each test of one package, at once."""
        with self.condition:
            for seconds, parallel, package, name, command in tests:
                heapq.heappush(self.pending, (-seconds, package, name, parallel, command))
            self.condition.notify_all()

    def close(self):
        """Every package is built and listed: nothing more will be added."""
        with self.condition:
            self.closed = True
            self.condition.notify_all()

    def take(self):
        with self.condition:
            while (not self.pending and not self.closed) or (self.pending and not self.free):
                self.condition.wait()
            if not self.pending:
                return None
            negative, package, name, parallel, command = heapq.heappop(self.pending)
            permits = 4 if parallel and -negative > 300 and self.free >= 4 else 1
            self.free -= permits
            return package, name, -negative, permits, command

    def release(self, permits):
        with self.condition:
            self.free += permits
            self.condition.notify_all()


# Input providers live with their stage implementations. Command inventories below are
# shared by execution and enumeration; an unaudited runtime reader fails closed.
phaseInputProviders = {}


def declaresInputs(phase, provider):
    def decorate(method):
        phaseInputProviders[phase] = provider
        return method
    return decorate


def treeInputs(note):
    return {"packages": [], "paths": ["."], "note": note}


def goPhaseCommand(phase):
    return ["go", phase, "./..."]


def goCommandInputs(tree, command):
    # Import paths, never patterns: the consumer computes each go list -deps closure.
    patterns = [arg for arg in command[2:] if arg.startswith(("./", module))]
    result = subprocess.run(["go", "list", "-json"] + patterns, cwd=tree,
                            capture_output=True, text=True, check=True, timeout=90)
    decoder, remaining, packages = json.JSONDecoder(), result.stdout.strip(), set()
    while remaining:
        package, end = decoder.raw_decode(remaining)
        remaining = remaining[end:].lstrip()
        packages.add(package["ImportPath"])
        if command[1] in ("test", "vet"):
            # go list -deps on a root alone omits the test imports vet/go test compile.
            packages.update(package.get("TestImports", []) + package.get("XTestImports", []))
    return {"packages": sorted(packages), "paths": []}


def productListingCommand():
    return ["go", "list", "-f", "{{.ImportPath}}\t{{.Dir}}", "./..."]


def productInputSources(tree):
    listing = subprocess.run(productListingCommand(), cwd=tree, capture_output=True,
                             text=True, check=True, timeout=90).stdout.splitlines()
    directories = dict(line.split("\t", 1) for line in listing)
    sources = [os.path.relpath(path, tree) for directory in directories.values()
               for path in glob.glob(os.path.join(directory, "*_test.go"))]
    return sources, any(productDeclarations(directories).values())


def catalogEntries(tree):
    with open(os.path.join(tree, "verify/catalog/catalog.json")) as handle:
        return json.load(handle)


def wasiCommands(names):
    return [["go", "test", "-count=1", "-json", "-timeout", fullPackageTimeout,
             "-p", "1", "-parallel", "1", "-skip", "^TestProduct_", "-run",
             "^%s$" % name if wasiUnit.fullmatch(name) else fixturePattern(["TestWASI"], [name]),
             "./internal/native"] for name in names]


def wasiInputs(tree, unit, full):
    names = wasiFixtures(tree)
    if unit and unit not in names:
        raise ValueError("unknown WASI unit " + unit)
    chosen = [unit] if unit else names
    inputs = goCommandInputs(tree, wasiCommands(chosen)[0])
    inputs["packages"].append(module + "/cmd/adamic")  # wasm_test.go builds the stage 0 executable.
    # The fixture may import siblings; the Node oracle reads its runtime and hooks,
    # and read_files.a runs from the oracle testdata directory. Include entire fixture
    # directories, not only their entry points. requests reads the wasm ABI and hosts.
    inputs["paths"] += ["oracle", "internal/native/wasm", "internal/oracle/testdata"]
    for name in chosen:
        if wasiUnit.fullmatch(name):
            return treeInputs("top-level WASI unit's fixture slice is not yet bounded")
        if "/" in name:
            inputs["paths"].append(os.path.dirname(name))
    return inputs


def catalogInputs(tree, unit, full):
    if not full:
        return {"packages": [], "paths": ["verify/catalog/check.sh"]}
    entries = [entry for entry in catalogEntries(tree) if not unit or str(entry["number"]) == unit]
    if not entries:
        raise ValueError("unknown catalog entry " + str(unit))
    inputs = treeInputs("catalog runs arbitrary named Go tests in a root working directory; their runtime reads are not yet bounded")
    for entry in entries:
        if entry.get("command"):
            inputs["packages"] += goCommandInputs(tree, shlex.split(entry["command"]))["packages"]
        if entry.get("patch"):
            inputs["paths"].append("verify/catalog/" + entry["patch"])
        if entry.get("fixture"):
            inputs["paths"].append(entry["fixture"])
    inputs["paths"].append("verify/catalog")
    return inputs


def phaseInputs(tree, line, full=False, products=None):
    phase, _, unit = line.partition(" ")
    provider = phaseInputProviders.get(phase)
    inputs = (provider(tree, unit, full) if provider else
              treeInputs("phase has no bounded input provider"))
    if "." not in inputs["paths"]:
        # Every phase enters runPhase's npm setup, and runs this runner/config.
        inputs["paths"] += ["cloud/fast-gate", "cloud/markdown-width/npm-bootstrap.json",
                            "stage3/api/package.json", "stage3/api/package-lock.json", "go.mod", "go.sum"]
        if full and phase in ("wasi", "stage3", "catalog", "products"):
            # Same source discovery as wholeProducts. A newly added declaration is
            # visible on the next enumeration. Existing products need a runtime-read audit.
            sources, declared = products if products is not None else productInputSources(tree)
            inputs["paths"] += sources
            if declared:
                inputs["paths"].append(".")
                inputs["note"] = "whole phase runs declared products first; product runtime reads are not yet bounded"
    inputs["packages"] = sorted(set(inputs["packages"]))
    inputs["paths"] = sorted(set(inputs["paths"]))
    return {"unit": line, "inputs": inputs}


def main():
    if "--list-units" in sys.argv:
        # The pool's planner reads the units from here, from a plain checkout, so it never parses what run.py owns.
        lister = argparse.ArgumentParser()
        lister.add_argument("--full", action="store_true")
        lister.add_argument("--list-units", action="store_true")
        lister.add_argument("--with-inputs", action="store_true")
        lister.add_argument("--tree", required=True)
        lister.add_argument("--base", help="include conditional fast phases selected by this base diff")
        lister.add_argument("--sha")
        lister.add_argument("--tools")
        arguments = lister.parse_args()
        units = wholeUnits(arguments.tree) if arguments.full else fastUnits(arguments.tree, arguments.base, arguments.sha, arguments.tools)
        if arguments.with_inputs:
            products = productInputSources(arguments.tree) if arguments.full else None
            for unit in units:
                print(json.dumps(phaseInputs(arguments.tree, unit, arguments.full, products)))
        else:
            print("\n".join(units))
        return
    parser = argparse.ArgumentParser()
    parser.add_argument("--tree", required=True)
    parser.add_argument("--sha", required=True)
    parser.add_argument("--base", required=True)
    # What the base sha is: main, or the area a worker's branch was cut from (gateBase in fast-gate-classify.sh).
    parser.add_argument("--base-name", default="main")
    parser.add_argument("--tools", required=True)
    parser.add_argument("--out", required=True)
    # The gate's CPUs come from its affinity (cloud/fast-gate.sh pins a slot, full-gate-main.sh the full
    # gate). Half of them run test processes, each at most two parallel subtests, since each subtest
    # also runs clang and Node: a slot that oversubscribed itself sent a 150 s node deadline red.
    parser.add_argument("--parallel", type=int, default=max(1, len(os.sched_getaffinity(0)) // 2))
    parser.add_argument("--full", action="store_true")
    # A whole gate that is also a parity proof runs to the end after its first failure, so every test can be
    # compared with the pool's run (@system_adamic, Oct 9 00:12Z). Its red still blocks landing.
    parser.add_argument("--run-to-end", action="store_true")
    # Landing and area gates (@system_adamic, Oct 8 09:43): run every test and fixture and report every
    # failure in one pass, failing at the end; area-next found its moved fixtures one per gate.
    parser.add_argument("--complete", action="store_true", help="fast gate that runs on after a failure and names every failed test")
    parser.add_argument("--branch", default="")
    parser.add_argument("--branch-source", default="")
    parser.add_argument("--session", default="")
    parser.add_argument("--session-source", default="")
    parser.add_argument("--weights", help="package seconds, longest first, to order the full gate's packages")
    # Side work's fast gates on Loom's pool (#xt96xyp): the pool runs this selection where the tree is whole, so one
    # selection decides what a fast gate tests wherever it runs. It writes select.json and runs nothing.
    parser.add_argument("--select", action="store_true", help="write the fast gate's package selection to <out>/select.json and stop")
    # Loom's pool runs the whole gate as units (Oct 9): one phase of it, optionally one unit of that phase, or the census
    # over the record the pool merged, each with this gate's own logic and record shape, so the pool never forks it.
    phases = parser.add_mutually_exclusive_group()
    phases.add_argument("--phase", help="run only this non-test phase of the selected gate")
    phases.add_argument("--phases", help="run comma-separated non-test phases into one gate record")
    parser.add_argument("--unit", help="with --phase wasi, stage3 or catalog: run only this unit (a fixture, a stage 3 command, a catalog entry number)")
    parser.add_argument("--census", metavar="TEST_JSONL", help="the census and deferred checks over a merged test.jsonl; may accompany --phases")
    arguments = parser.parse_args()
    os.makedirs(arguments.out, exist_ok=True)
    gate = Gate(arguments)
    if arguments.select:
        try:
            gate.run()
        except BaseException:
            gate.fail("runner", traceback.format_exc())
        sys.exit(0 if gate.failure is None and os.path.exists(os.path.join(arguments.out, "select.json")) else 1)
    signal.signal(signal.SIGTERM, gate.stop)
    try:
        gate.run()
    except BaseException:
        gate.fail("runner", traceback.format_exc())
    finally:
        gate.finish()
    sys.exit(0 if gate.failure is None and not gate.stopped else 1)


class Gate:
    def __init__(self, arguments):
        self.arguments = arguments
        # Queue ownership follows this run, not the user's shared cache. Keep these
        # beside the slot tree so publishing gate artifacts never publishes products.
        queueRoot = tempfile.mkdtemp(prefix="build-store-", dir=os.path.dirname(os.path.abspath(arguments.tree)))
        self.buildStoreEnvironment = {
            "ADAMIC_BUILD_STORE_SPOOL": os.path.join(queueRoot, "spool"),
            "ADAMIC_BUILD_STORE_AUDITS": os.path.join(queueRoot, "audits"),
        }
        for directory in self.buildStoreEnvironment.values():
            os.mkdir(directory)
        self.started = time.monotonic()
        self.failure = None
        self.stopped = None
        self.stopThread = None
        self.killedByStop = set()
        self.lock = threading.Lock()
        self.processes = []
        self.steps = {}
        self.counts = {"pass": 0, "fail": 0, "skip": 0}
        self.complete = vars(arguments).get("complete") is True
        self.failedTests = []
        self.skipped = []
        self.census = {"required_input": [], "unclassified": []}
        # Fail closed: green needs every planned stage to have recorded exit 0, not just no failure,
        # so a stage that died without reporting (an exception, a process that never started) is red.
        self.exits = {}
        self.watchers = []
        self.planned = (["coverage", "tools", "build", "vet", "tests", "wasi", "stage3", "catalog", "determinism", "census"] if arguments.full
                        else ["coverage", "tools", "build", "vet", "tests", "smoke", "determinism", "census"])
        if os.path.isfile(os.path.join(arguments.tree, "internal/buildcache/cmd/buildcache-publish/main.go")):
            index = self.planned.index("tests") + 1
            self.planned[index:index] = ["audit", "upload"]
        selected = vars(arguments).get("phases")
        phase = vars(arguments).get("phase")
        if selected is not None:
            self.planned = list(dict.fromkeys(selected.split(",")))
        elif phase is not None:
            self.planned = [phase]
        elif vars(arguments).get("census"):
            self.planned = ["census"]
        self.result = {
            "sha": arguments.sha,
            "branch": arguments.branch,
            "branch_source": arguments.branch_source,
            "session": arguments.session,
            "session_source": arguments.session_source,
            "base": arguments.base,
            "base_name": vars(arguments).get("base_name") or "main",
            "tools_sha": self.git(arguments.tools, "rev-parse", "HEAD"),
            "machine": {"hostname": socket.gethostname(), "nproc": os.cpu_count()},
            "uncached_tests": True,
            "build_ok": False,
            "vet_ok": False,
            "box_load": {"start": boxLoad()},
        }
        self.kind = "full" if arguments.full else "fast"
        self.status("running: %s gate of %s against %s" % (self.kind, arguments.sha, arguments.base))

    def stop(self, signum=None, frame=None):
        # The signal can interrupt code holding self.lock; drain in another thread.
        if self.stopped:
            return
        reason = os.environ.get("ADAMIC_FAST_GATE_STOP_REASON", "stopped")
        try:
            with open(self.arguments.out + ".stop-reason") as handle:
                reason = handle.read().strip() or reason
        except OSError:
            pass
        self.stopped = {"reason": reason, "at_seconds": round(time.monotonic() - self.started, 1)}
        def drain():
            with self.lock:
                self.killedByStop.update(id(p) for p in self.processes if p.poll() is None)
                self.killSessions()
        self.stopThread = threading.Thread(target=drain)
        self.stopThread.start()

    def run(self):
        # A scoped run skips corpus-wide tests by name, so a gate that can land anything on main refuses
        # to run scoped (#60hxabf, cohere's rule-scoped lint): push-main.sh lands only a clean scoped_env.
        self.result["scoped_env"] = sorted(name for name in scopingEnvironment if os.environ.get(name) is not None)
        if self.result["scoped_env"]:
            self.fail("environment", "refused: a gate that can land on main never runs scoped, and %s is set" % ", ".join(self.result["scoped_env"]))
            return
        tree = self.arguments.tree
        head = self.git(tree, "rev-parse", "HEAD")
        if head != self.arguments.sha:
            self.fail("setup", "the tree is at %s, not the candidate %s" % (head, self.arguments.sha))
            return
        selecting = vars(self.arguments).get("select") is True
        if vars(self.arguments).get("phase") is not None or vars(self.arguments).get("phases") is not None or vars(self.arguments).get("census"):
            self.runPhase()
            return
        if not selecting:
            self.npmCli()
            if not self.npmPackages():
                return
        if self.arguments.full:
            self.runFull()
            return
        smoke, smokeSource = self.smokeList()
        self.deferred = self.deferredList()
        self.runRequested()
        # Unquoted, so a path with non-ASCII bytes is itself and can match its package or a rule.
        changed = self.git(tree, "-c", "core.quotePath=false", "diff", "--name-only", "%s...%s" % (self.arguments.base, self.arguments.sha)).split("\n")
        changed = [path for path in changed if path]
        try:
            packages, unowned = self.touched(changed)
        except ValueError as error:
            if not str(error).startswith("compiler dependency census"):
                raise
            # A red the change can fix itself, never a gate-tool crash (developer tools, Oct 9).
            self.fail("census", "%s. Declare each in cloud/fast-gate/compiler-dependencies.json in this branch: its package, "
                      "then the compiler packages it exercises (entries there add to the gate tools' map)." % error)
            return
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
        # A requested deferred test in a package the change doesn't touch runs alone in it.
        self.onlyTests = {}
        for importPath, names in getattr(self, "requested", {}).items():
            if importPath not in packages:
                packages = sorted(set(packages) | {importPath})
                self.onlyTests[importPath] = set(names)
                self.packageDirectories.setdefault(importPath, os.path.join(tree, importPath[len(module) + 1:]))
        if "stage3" in executors:
            packages = sorted(set(packages) | set(self.goList("./stage3/...")))
            self.planned.append("stage3")
        if selecting:
            # What the pool can run is the Go tests; every other executor this change needs is named, so a pool
            # verdict says what it didn't cover.
            selection = {"sha": self.arguments.sha, "base": self.arguments.base, "packages": packages,
                         "env": dict(gateEnvironment, **self.sampling),
                         "deferred": {package: sorted(names) for package, names in sorted(self.deferred.items()) if package in packages},
                         "only_tests": {package: sorted(names) for package, names in sorted(self.onlyTests.items())},
                         "executors_beyond_go_tests": sorted(executors | ({"smoke"} if smoke else set()))}
            with open(os.path.join(self.arguments.out, "select.json"), "w") as handle:
                json.dump(selection, handle, indent=2)
                handle.write("\n")
            return
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
                   self.guarded("smoke", self.smoke, smoke, log),
                   self.guarded("determinism", self.determinism, smoke)]
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
        self.cacheDrains()
        if self.failure is not None:
            return
        # The skip census first, then the requested deferred tests: each records its own verdict, and a deferred red no
        # longer leaves the record without a census (trio 114a6439, Oct 9: its census never ran behind a deferred red).
        self.checkCensus()
        self.requestedRan()

    def fastPhaseInputs(self, recordCoverage=False):
        """The same base diff and executor map as a fast run, without running coverage as an extra phase."""
        tree = self.arguments.tree
        revision = "%s...%s" % (self.arguments.base, self.arguments.sha)
        changed = [path for path in self.git(tree, "-c", "core.quotePath=false", "diff", "--name-only", revision).split("\n") if path]
        _, unowned = self.touched(changed)
        pathSetChanged = self.git(tree, "-c", "core.quotePath=false", "diff", "--name-only", "--no-renames", "--diff-filter=AD", revision).split("\n")
        planned = list(self.planned)
        executors = self.cover(unowned, changed, pathSetChanged, record=recordCoverage)
        self.planned = planned
        if recordCoverage and "darwin" not in planned and "darwin" in self.exits:
            del self.exits["darwin"]
            self.exits["coverage"] = 1
        self.result["changed_files"] = changed
        self.result["unowned_files"] = unowned
        return changed, executors or set()

    def runPhase(self):
        """Pool phases share the gate's stage implementations and fail-closed record writer.

        A supplied pool log is copied before smoke appends to it; census runs last so it sees both.
        """
        census = vars(self.arguments).get("census")
        allowed = wholePhases + ["products", "census"] if self.arguments.full else fastPhases
        for phase in self.planned:
            if phase not in allowed:
                self.fail(phase or "phase", "%s gate has no phase %r" % (self.kind, phase))
                return
        unit = vars(self.arguments).get("unit")
        if unit and (len(self.planned) != 1 or self.planned[0] not in ("wasi", "stage3", "catalog") or
                     (not self.arguments.full and self.planned[0] == "catalog")):
            self.fail(self.planned[0], "--unit is unsupported for these phases")
            return
        self.npmCli()
        if not self.npmPackages():
            return
        logPath = os.path.join(self.arguments.out, "test.jsonl")
        if census:
            if "census" not in self.planned:
                self.fail("census", "--census requires census in --phases")
                return
            try:
                if os.path.abspath(census) != os.path.abspath(logPath):
                    shutil.copyfile(census, logPath)
            except OSError:
                self.fail("census", traceback.format_exc())
                return
        elif any(phase in self.planned for phase in ("smoke", "wasi")):
            # A reused output directory must not carry a previous run's test verdicts into this census.
            with open(logPath, "w"):
                pass
        ordered = [phase for phase in self.planned if phase != "census"] + (["census"] if "census" in self.planned else [])
        for phase in ordered:
            if self.failure is not None:
                break
            try:
                # A pool phase unit takes the products as given: Loom's planner runs every TestProduct_ as its own unit
                # first, and the phases fetch them by hash. Rebuilding them inside each phase cost every unit about 100 s
                # cold on 4 CPUs (main b5245943's phases, Oct 9 07:43Z). Only the products phase builds them.
                if self.arguments.full and phase == "products":
                    with open(logPath, "a") as log:
                        if not self.wholeProducts(log) and not getattr(self.arguments, "run_to_end", False):
                            break
                    if phase == "products":
                        self.steps.setdefault("products", 0.0)
                        self.exits.setdefault("products", 0)
                        continue
                self.executePhase(phase, logPath)
            except BaseException:
                self.fail(phase, traceback.format_exc())

    def executePhase(self, phase, logPath):
        if phase == "coverage":
            if self.arguments.full:
                parents = self.git(self.arguments.tree, "rev-list", "--parents", "-n", "1", self.arguments.sha).split()[1:]
                changed = [path for path in (self.git(self.arguments.tree, "-c", "core.quotePath=false", "diff", "--name-only", parents[0], self.arguments.sha).split("\n") if parents else []) if path]
                _, unowned = self.touched(changed)
                self.cover(unowned, changed)
            else:
                self.fastPhaseInputs(recordCoverage=True)
        elif phase == "tools":
            self.toolsDeclared()
        elif phase == "build":
            self.build()
        elif phase == "vet":
            self.vet()
        elif phase == "wasi":
            environment = None
            if os.environ.get("WASI_SYSROOT"):
                environment = {"PATH": os.path.join(os.path.dirname(os.path.dirname(os.environ["WASI_SYSROOT"])), "bin") + os.pathsep + os.environ["PATH"]}
            with open(logPath, "a") as log:
                self.wasiSplit(log, environment)
        elif phase == "stage3":
            self.stage3()
        elif phase == "catalog":
            if self.arguments.full:
                self.catalogFull()
            else:
                self.step("catalog", ["bash", "-n", "verify/catalog/check.sh"])
        elif phase in ("smoke", "determinism"):
            smoke, source = self.smokeList()
            if not smoke:
                self.fail(phase, "no smoke list in the gated tree or the tools checkout")
                return
            self.result.update({"smoke_list": "cloud/fast-gate/smoke.txt", "smoke_list_source": source,
                                "smoke_list_blob": self.git(source["root"], "hash-object", os.path.join(source["root"], "cloud/fast-gate/smoke.txt")),
                                "smoke_fixtures": ["%s %s" % entry for entry in smoke]})
            if phase == "smoke":
                with open(logPath, "a") as log:
                    self.smoke(smoke, log)
            else:
                self.determinism(smoke)
        elif phase == "census":
            if not os.path.isfile(logPath):
                self.fail("census", "no test.jsonl: supply the pool's merged log with --census")
                return
            if self.arguments.full:
                self.checkCensus()
                self.deferredRanWhole()
            else:
                self.deferred = self.deferredList()
                self.runRequested()
                self.checkCensus()
                self.requestedRan()
        elif phase in ("workers", "a-check", "catalog-apply"):
            changed, executors = self.fastPhaseInputs()
            if phase == "workers":
                if not executors & {"workers", "bench-workers"}:
                    raise ValueError("workers has no work for this change")
                self.workers(executors)
            elif phase == "a-check":
                paths = self.result.get("unchecked_a_files", [])
                if not paths:
                    raise ValueError("a-check has no work for this change")
                self.aCheck(paths)
            else:
                patches = self.catalogEntriesTouched(changed)
                if not patches:
                    raise ValueError("catalog-apply has no work for this change")
                self.catalogApply(patches)
        elif phase == "darwin":
            self.fail("darwin", "needs darwin leg: go test ./internal/apple/... on Kirk's Mac")

    def selectedUnits(self, phase, names):
        """The units of a phase this run takes: all of them, or the one --unit names (a catalog entry by its number).
        A --unit no unit answers to is an error, never an empty green."""
        unit = vars(self.arguments).get("unit")
        if not unit:
            return list(names)
        chosen = [name for name in names if name == unit or (phase == "catalog" and unit.isdigit() and name.startswith("%02d " % int(unit)))]
        if not chosen:
            raise ValueError("%s has no unit %r (units: %s)" % (phase, unit, ", ".join(names)))
        self.result["unit"] = unit
        return chosen

    @declaresInputs("products", lambda tree, unit, full: treeInputs("products run arbitrary declared tests; runtime reads are not yet bounded"))
    def wholeProducts(self, log):
        if getattr(self, "wholeProductsDone", False):
            return self.exits.get("products", 0) == 0
        self.wholeProductsDone = True
        listing = self.command(productListingCommand(),
                               cwd=self.arguments.tree, capture_output=True, text=True, check=True).stdout.splitlines()
        directories = {}
        for line in listing:
            package, directory = line.split("\t", 1)
            directories[package] = directory
        self.packageDirectories = directories
        # Only the packages whose source declares a product compile here: compiling every test binary first would
        # hold the whole gate's tests behind hundreds of builds. testSplit still checks each one's -test.list.
        declaring = [package for package, names in productDeclarations(directories, list(directories)).items() if names]
        if not declaring:
            return True
        return self.testSplit(declaring, log, productsOnly=True)

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
            self.result["build_ok"] = self.step("build", goPhaseCommand("build"))
        except BaseException:
            self.fail("build", traceback.format_exc())
        log = open(os.path.join(self.arguments.out, "test.jsonl"), "w")
        # A parity run (--run-to-end) keeps going past a failed product: its tests build what they need on a miss, so
        # the record still holds every red (Oct 9 07:23Z: library's area stopped at TestProduct_suite with no red list).
        if not self.wholeProducts(log) and not getattr(self.arguments, "run_to_end", False):
            log.close()
            return
        # Never more runnable threads than the box has cores (@system_adamic, Oct 9 00:11Z: -p of half the CPUs, each
        # test binary on every core, put Home at load 178 and the Threadripper at 81 to 145, and a wall-clock stall
        # guard failed V1 on it): a package's test binary gets GOMAXPROCS cores, and -p packages run at once, so the
        # two multiply to about the box. ADAMIC_FULL_GATE_PACKAGES overrides the package count. cores/4 packages (16 at
        # GOMAXPROCS 4 on 64 cores, @system_adamic, Oct 9 01:22Z): cores/8 peaked at load 29.2 on the Threadripper
        # (6b2c73f9, 50.5 min), half the box idle; back to cores/8 if a whole gate's peak crosses the core count.
        cpus = len(os.sched_getaffinity(0)) if hasattr(os, "sched_getaffinity") else (os.cpu_count() or 1)
        packagesAtOnce = int(os.environ.get("ADAMIC_FULL_GATE_PACKAGES", max(1, cpus // 4)))
        threadsEach = max(2, cpus // packagesAtOnce)
        self.result["parallelism"] = {"cpus": cpus, "packages_at_once": packagesAtOnce, "gomaxprocs_each": threadsEach}
        wasi = None
        if os.environ.get("WASI_SYSROOT"):
            wasi = {"PATH": os.path.join(os.path.dirname(os.path.dirname(os.environ["WASI_SYSROOT"])), "bin") + os.pathsep + os.environ["PATH"]}
        threads = [self.guarded("tools", self.toolsDeclared), self.guarded("vet", self.vet),
                   # Three hours a package (@system_adamic, Oct 8 00:08): the whole gate's 16 CPUs are slower
                   # on purpose, and a wall-clock package timeout there is the per-test fragility at a larger
                   # size. Hangs belong to stall guards that count from output; a package over an hour is
                   # named in the status line as slow, so it never reads as a quiet pass.
                   self.guarded("tests", self.test, "tests", ["go", "test", "-count=1", "-json", "-timeout", fullPackageTimeout, "-p", str(packagesAtOnce), "-skip", "^TestProduct_|" + wasiSkip] + packages, log, {"GOMAXPROCS": str(threadsEach)}),
                   self.guarded("wasi", self.wasiSplit, log, wasi),
                   self.guarded("stage3", self.stage3),
                   # The bug catalog: each catalogued bug reintroduced and caught, on every main that has it.
                   self.guarded("catalog", self.catalogFull),
                   self.guarded("determinism", self.determinism, self.smokeList()[0])]
        for thread in threads:
            thread.start()
        for thread in threads:
            thread.join()
        log.close()
        self.cacheDrains()
        try:
            self.checkCensus()
            self.deferredRanWhole()
        except BaseException:
            self.fail("census", traceback.format_exc())

    def deferredRanWhole(self):
        """The whole gate is where deferred tests run, so each must reach pass or fail here: a deferred test the
        whole gate skips or never runs is a void (@system_adamic, Oct 8, after TestProfileSnapshotsAgree, a
        measurement on the list, skipped in both gates). A row the census classes measurement or opt-in-lane is
        refused outright, since no gate gives it its input."""
        deferred = self.deferredList()
        outcomes = topLevelOutcomes(os.path.join(self.arguments.out, "test.jsonl"))
        results = {importPath + " " + name: outcomes.get(importPath + " " + name, "missing")
                   for importPath, names in deferred.items() for name in sorted(names)}
        self.result["deferred_whole_results"] = results
        refused = deferredClassedOut(self.arguments.tools, deferred)
        unrun = ["%s: %s" % (test, outcome) for test, outcome in sorted(results.items()) if outcome not in ("pass", "fail")]
        if refused or unrun:
            self.fail("census", "cloud/fast-gate/deferred.txt names tests the whole gate doesn't run to a verdict; take each off the list or give the gate its input:\n" +
                      "\n".join(["%s: census class %s, refused" % row for row in refused] + unrun))

    @declaresInputs("coverage", lambda tree, unit, full: treeInputs("coverage uses the base diff, package/source census and executor map across the tree"))
    def cover(self, unowned, changed=(), pathSetChanged=None, record=True):
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
        self.result["coverage_enforced"] = enforce
        if not record:
            return executors
        self.steps["coverage"] = 0.0
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

    @declaresInputs("stage3", lambda tree, unit, full: treeInputs("stage3 runs Python/shell lane and adaptation programs; their runtime reads are not yet bounded"))
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
        assert [name for name, _, _ in commands] == stage3Units
        chosen = self.selectedUnits("stage3", [name for name, _, _ in commands])
        commands = [command for command in commands if command[0] in chosen]
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
        # oracle/tests.log is tsc's own suite's output, each failure's title and message at its end: without
        # it a lane red named a test and nothing more (area/library ca016bab, extractFunction1, Oct 8).
        # The publish step gzips a log over 5 MB.
        for name in ("execution.json", "report.json", "verdict.json", "verdict.txt", "patch-set.md", "oracle/tests.log"):
            if os.path.isfile(os.path.join(results, name)):
                with open(os.path.join(results, name), "rb") as source, open(os.path.join(self.arguments.out, "stage3-lane-" + name.replace("/", "-")), "wb") as target:
                    target.write(source.read())
        self.steps["stage3"] = round(time.monotonic() - started, 1)
        self.result["stage3_exits"] = codes
        failed = [name for name, _, _ in commands if codes.get(name) != 0]
        self.exits["stage3"] = 1 if failed else 0
        if failed:
            # The red names what failed, not just an exit code (typescript, Oct 8: 'stage3-lane exit 1' said
            # nothing): the lane's own verdict line, then each failed command's last lines.
            detail = ["stage 3 tier 1 failed: %s (logs: %s)" % (", ".join("%s exit %s" % (name, codes.get(name)) for name in failed), ", ".join(name + ".log" for name in failed))]
            verdict = os.path.join(self.arguments.out, "stage3-lane-verdict.txt")
            if "stage3-lane" in failed and os.path.isfile(verdict):
                detail += [line for line in open(verdict, errors="replace").read().splitlines() if line.strip()][:2]
            for name in failed:
                try:
                    lines = [line for line in open(os.path.join(self.arguments.out, name + ".log"), errors="replace").read().splitlines() if line.strip()]
                except OSError:
                    continue
                detail += ["%s, last lines:" % name] + [line[:300] for line in lines[-6:]]
            self.fail("stage3", "\n".join(detail))

    @declaresInputs("workers", lambda tree, unit, full: treeInputs("conditional scripts and their runtime reads are not yet bounded"))
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

    @declaresInputs("a-check", lambda tree, unit, full: treeInputs("conditional source programs may import outside their directories"))
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
            # A change that deletes a .a has nothing left to check there (views slice 1, Oct 8, deleted five).
            if not os.path.isfile(os.path.join(self.arguments.tree, path)):
                results[path] = {"outcome": "deleted", "expected": "deleted", "first": ""}
                continue
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

    @declaresInputs("catalog-apply", lambda tree, unit, full: treeInputs("patch application reads candidate and base trees"))
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

    @declaresInputs("catalog", catalogInputs)
    def catalogFull(self):
        if not os.path.exists(os.path.join(self.arguments.tree, "verify/catalog/check.sh")):
            self.exits["catalog"] = 0
            self.steps["catalog"] = 0.0
            self.result["catalog"] = "not in this tree"
            return
        entries = catalogEntries(self.arguments.tree)
        names = ["%02d %s" % (entry["number"], entry["name"]) for entry in entries]
        if not names or len(set(names)) != len(names) or len({entry["number"] for entry in entries}) != len(entries):
            raise ValueError("empty or duplicate catalog entries")
        chosen = self.selectedUnits("catalog", names)
        entries = [entry for entry, name in zip(entries, names) if name in chosen]
        names = chosen
        commands = [["bash", "verify/catalog/check.sh", "--entry", str(entry["number"]),
                     "--jobs", "1", self.arguments.sha] for entry in entries]
        def execute(name, command, scratch):
            path = os.path.join(scratch, "catalog.log")
            with open(path, "w") as output:
                code = self.spawn(command, output, environment={"TMPDIR": scratch,
                    "GOMAXPROCS": "1", "GOFLAGS": "-p=1"}).wait()
            with open(path) as handle:
                lines = handle.read().splitlines()
            verdicts = [line for line in lines if line.startswith(name + ": ")]
            caught = any("applies-and-fails-as-recorded" in line for line in verdicts)
            stale = any("no-longer-applies" in line for line in verdicts)
            skipped = any(": skipped (" in line for line in verdicts)
            # A stale patch is upkeep; an error or a missing verdict is always red.
            ok = (caught and code == 0) or (stale and code == 1) or (skipped and code == 0)
            return {"exit": code, "ok": ok, "caught": caught, "stale": stale,
                    "skipped": skipped, "verdicts": verdicts, "log": path}
        self.phaseUnits("catalog", names, commands, execute)
        rows = self.result["catalog_units"]
        # At least one entry must catch its bug, over the whole catalog: a single unit (--unit) is judged by its own
        # verdict alone, since one entry may rightly be skipped or stale (Loom, Oct 9: entry 10 green alone, its phase red).
        if not vars(self.arguments).get("unit") and not any(row.get("caught") and row.get("ok") for row in rows):
            self.exits["catalog"] = 1
            self.fail("catalog", "no catalog entry caught its mutation (see catalog_units)")
        self.result["catalog"] = {"exit": self.exits["catalog"],
            "caught": sum(row.get("caught", False) for row in rows),
            "stale": [row["name"] for row in rows if row.get("stale")],
            "uncaught": [row["name"] for row in rows if not row.get("ok")]}
        with open(os.path.join(self.arguments.out, "catalog.log"), "w") as output:
            for row in rows:
                output.write("%s wall=%.3fs\n" % (row["name"], row["wall_seconds"]))
                if row.get("log"):
                    with open(row["log"]) as handle:
                        shutil.copyfileobj(handle, output)

    @declaresInputs("wasi", wasiInputs)
    def wasiSplit(self, log, environment=None):
        names = self.selectedUnits("wasi", wasiFixtures(self.arguments.tree))
        commands = wasiCommands(names)
        def execute(name, command, scratch):
            events = []
            class UnitLog:
                def write(self, line):
                    log.write(line)
                    try:
                        events.append(json.loads(line))
                    except ValueError:
                        pass
            variables = dict(environment or {}, TMPDIR=scratch, GOMAXPROCS="1", GOFLAGS="-p=1")
            code = self.stream("wasi/" + name, command, UnitLog(), environment=variables)
            wanted = name if wasiUnit.fullmatch(name) else "TestWASI/" + name
            passed = any(event.get("Test") == wanted and event.get("Action") == "pass" for event in events)
            unexpected = sorted({event["Test"] for event in events if event.get("Action") == "run"
                                 and (event.get("Test", "").startswith("TestWASI/") or wasiUnit.fullmatch(event.get("Test", "").split("/")[0]))
                                 and event["Test"] != wanted and not wanted.startswith(event["Test"] + "/")
                                 and not event["Test"].startswith(wanted + "/")})
            return {"exit": code, "ok": code == 0 and passed and not unexpected, "test": wanted,
                    "unexpected": unexpected,
                    "passed": passed}
        self.phaseUnits("wasi", names, commands, execute)

    def phaseUnits(self, phase, names, commands, execute):
        """One scratch and result per unit, with at most the slot's CPUs in flight.

        Wall includes process startup and compilation. A unit over the 30 s budget is recorded (over_budget, and
        the phase's slowest unit) and never disappears into an aggregate timing, but it isn't red: the budget's
        ruling makes only a new unit over it red, and these phases' units are over it until their builds are
        products by hash (#2q61kg9: the first catalog and WASI units pay the compiler's cold build).
        """
        started = time.monotonic()
        slots = max(1, min(self.arguments.parallel, slotCPUs()))
        rows = [{"name": name, "command": command, "status": "not run"}
                for name, command in zip(names, commands)]
        self.result[phase + "_units"] = rows
        self.result[phase + "_unit_count"] = len(rows)
        self.result[phase + "_unit_budget_seconds"] = 30
        root = os.path.join(os.path.abspath(self.arguments.out), phase + "-units")
        os.makedirs(root, exist_ok=True)
        def work(item):
            index, row = item
            before = time.monotonic()
            scratch = tempfile.mkdtemp(prefix="%02d-" % index, dir=root)
            row["scratch"] = scratch
            try:
                row.update(execute(row["name"], row["command"], scratch))
            except BaseException:
                row.update(ok=False, error=traceback.format_exc())
            elapsed = time.monotonic() - before
            row["wall_seconds"] = round(elapsed, 6)
            row["over_budget"] = elapsed > 30
            row["status"] = "passed" if row["ok"] else "failed"
            if row["status"] == "failed":
                self.fail(phase + "/" + row["name"], json.dumps(row))
        with ThreadPoolExecutor(max_workers=slots) as pool:
            list(pool.map(work, enumerate(rows)))
        self.result[phase + "_slowest_unit_seconds"] = max((row["wall_seconds"] for row in rows), default=0)
        self.steps[phase] = round(time.monotonic() - started, 3)
        self.exits[phase] = 0 if rows and all(row["status"] == "passed" for row in rows) else 1

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

    def runRequested(self):
        """A change may ask its fast gate to run a deferred test anyway (@system_adamic, Oct 8, for 04's
        split-build fix-forward): a "Gate-runs: <package> <Test>" trailer on any commit since the base
        takes that test off the deferred list for this gate, when its package is gated. Recorded in
        fast.json as deferred_run_by_request."""
        text = self.git(self.arguments.tree, "log", "--format=%(trailers:key=Gate-runs,valueonly)", "%s..%s" % (self.arguments.base, self.arguments.sha))
        self.requested = {}
        for line in text.splitlines():
            fields = line.split()
            if fields == ["deferred"]:
                # "Gate-runs: deferred" runs every deferred test, in its package whether or not the change
                # touches it (@system_adamic, Oct 8: a landing that changes emitted C or the runtime, or
                # carries more than one area's commits, proves the full gate's deferred tests before main).
                self.result["deferred_all_requested"] = True
                for importPath, names in self.deferred.items():
                    self.requested.setdefault(importPath, set()).update(names)
                    self.result.setdefault("deferred_run_by_request", []).extend(importPath + " " + name for name in sorted(names))
                self.deferred = {}
                continue
            if len(fields) != 2:
                continue
            importPath = module + "/" + fields[0].strip("/")
            if fields[1] in self.deferred.get(importPath, set()):
                self.deferred[importPath].discard(fields[1])
                self.requested.setdefault(importPath, set()).add(fields[1])
                self.result.setdefault("deferred_run_by_request", []).append(importPath + " " + fields[1])

    def requestedRan(self):
        """Every requested deferred test ran to a verdict rather than skipping (a skipped or missing one
        would let a void pass for a proof): fast.json's deferred_run_results names each one's outcome."""
        requested = getattr(self, "requested", {})
        if not requested:
            return
        outcomes = topLevelOutcomes(os.path.join(self.arguments.out, "test.jsonl"))
        results = {importPath + " " + name: familyOutcome(outcomes, importPath, name)
                   for importPath, names in requested.items() for name in sorted(names)}
        self.result["deferred_run_results"] = results
        unproven = ["%s: %s" % (test, outcome) for test, outcome in sorted(results.items()) if outcome not in ("pass", "fail")]
        if unproven:
            self.fail("deferred", "requested deferred tests that didn't run to a verdict:\n" + "\n".join(unproven))

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
        a package's tree is that package's to test). Reverse imports and compiler declarations then
        expand production owners transitively; test imports are followed only once at the end.
        Test source and data edits select their owners only. Anything unowned goes to cover()."""
        listing = self.command(["go", "list", "-deps", "-test", "-json", "./..."],
                               cwd=self.arguments.tree, capture_output=True, text=True)
        if listing.returncode != 0:
            raise SystemExit("go list failed: " + listing.stderr)
        directories, embedded, reverse, testReverse = {}, {}, {}, {}
        testEmbedded, realEmbedded = set(), set()
        self.packageDirectories = {}
        tree = os.path.realpath(self.arguments.tree)
        decoder = json.JSONDecoder()
        remaining = listing.stdout.strip()
        while remaining:
            package, end = decoder.raw_decode(remaining)
            remaining = remaining[end:].lstrip()
            # go list -test emits synthetic test binaries and annotated package variants.
            # A variant's Imports are test edges, never production edges.
            importPath = package.get("ForTest") or package["ImportPath"].split(" [", 1)[0]
            if package.get("Name") == "main" and importPath.endswith(".test"):
                continue
            variant = bool(package.get("ForTest")) or " [" in package["ImportPath"]
            for dependency in package.get("Imports", []):
                edges = testReverse if variant else reverse
                edges.setdefault(dependency.split(" [", 1)[0], set()).add(importPath)
            for dependency in set(package.get("TestImports", []) + package.get("XTestImports", [])):
                testReverse.setdefault(dependency.split(" [", 1)[0], set()).add(importPath)
            directory = package.get("Dir")
            if not directory or not (importPath == module or importPath.startswith(module + "/")):
                continue
            relative = os.path.relpath(os.path.realpath(directory), tree)
            directories[relative] = importPath
            self.packageDirectories[importPath] = directory
            for field in ("EmbedFiles", "TestEmbedFiles", "XTestEmbedFiles"):
                for name in package.get(field, []):
                    path = os.path.normpath(os.path.join(relative, name))
                    embedded[path] = importPath
                    if field == "EmbedFiles" and not variant:
                        realEmbedded.add(path)
                    else:
                        testEmbedded.add(path)
        dependencies = self.compilerDependencies(directories, changed)
        for package, inputs in dependencies.items():
            for dependency in inputs:
                reverse.setdefault(module + "/" + dependency, set()).add(module + "/" + package)
        packages, testsOnly, unowned = set(), set(), []
        # Another package's tests may read this package's testdata by path; test-reads.json names those readers. A tree
        # without it can't say, so its testdata edits keep seeding the whole closure. A _test.go edit is its own
        # package's alone either way: no other package can import a test file.
        readersDeclared = any(os.path.exists(os.path.join(root, "cloud/fast-gate/test-reads.json"))
                              for root in (self.arguments.tools, self.arguments.tree))
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
                if path in realEmbedded:
                    packages.add(owner)
                elif path.endswith("_test.go") or (readersDeclared and ("testdata" in path.split("/") or path in testEmbedded)):
                    testsOnly.add(owner)
                else:
                    packages.add(owner)
        # Optional explicit test-input readers; these name tests, not closure seeds.
        for root in dict.fromkeys((self.arguments.tools, self.arguments.tree)):
            name = os.path.join(root, "cloud/fast-gate/test-reads.json")
            if not os.path.exists(name):
                continue
            with open(name) as handle:
                readers = json.load(handle)
            for reader, inputs in readers.items():
                if any(path == prefix or path.startswith(prefix.rstrip("/") + "/") or prefix == "."
                       for path in changed for prefix in inputs):
                    testsOnly.add(module + "/" + reader)
        if "cloud/fast-gate/compiler-dependencies.json" in changed:
            packages.update(module + "/" + package for package in dependencies if package in directories)
            unowned = [path for path in unowned if path != "cloud/fast-gate/compiler-dependencies.json"]
        changedPackages = packages | testsOnly
        pending = list(packages)
        while pending:
            for dependent in reverse.get(pending.pop(), ()):
                if dependent not in packages:
                    packages.add(dependent)
                    pending.append(dependent)
        realClosure = set(packages)
        packages.update(testsOnly)
        for dependency in realClosure:
            packages.update(testReverse.get(dependency, ()))
        self.result["selected_compiler_drivers"] = sorted(
            driver for driver, declaration in getattr(self, "compilerDrivers", {}).items()
            if any(module + "/" + dependency in packages for dependency in declaration["dependencies"])
            or any(path.startswith(driver + "/") for path in changed)
            or "cloud/fast-gate/compiler-dependencies.json" in changed)
        packages.intersection_update(self.packageDirectories)
        compilerChanged = any(path.startswith(("internal/load/", "internal/lower/", "internal/native/", "internal/javascript/", "internal/ir/", "internal/flow/", "cmd/adamic/")) for path in changed)
        if compilerChanged:
            # No affected check is dropped to meet the fast gate's budget.
            self.deferred = {}
        if oracle in packages:
            self.oracleSelection = ({"whole": True, "reason": "compiler dependency changed"}
                                    if compilerChanged or oracle not in changedPackages else self.selectOracle())
            self.result["oracle_selection"] = self.oracleSelection
        self.result["direct_packages"] = sorted(changedPackages)
        return sorted(packages), unowned

    def compilerDependencies(self, directories, changed=()):
        """Package-level declarations cover helpers as well as their callers.

        The census scans all tracked Go package sources. A process-launching package must
        declare dependencies even when the executable name is computed: it may run the compiler.
        A missing package declaration fails selection before any tests start.
        """
        name = "cloud/fast-gate/compiler-dependencies.json"
        # The gate tools' map, with the candidate's own entries added over it: a branch that adds a compiler
        # consumer declares it itself (developer tools, Oct 9).
        declarations, root = None, None
        for source in (self.arguments.tools, self.arguments.tree):
            if not os.path.exists(os.path.join(source, name)):
                continue
            with open(os.path.join(source, name)) as handle:
                layer = json.load(handle)
            if declarations is None:
                declarations, root = layer, source
                continue
            if layer.get("version") != declarations.get("version"):
                raise ValueError("invalid compiler dependency map: " + name + " in " + source)
            for field in ("packages", "drivers"):
                if isinstance(layer.get(field), dict):
                    declarations.setdefault(field, {}).update(layer[field])
            root = root + " + " + source
        if declarations is None:
            raise ValueError("invalid compiler dependency map: no " + name)
        if declarations.get("version") != 1 or not isinstance(declarations.get("packages"), dict):
            raise ValueError("invalid compiler dependency map: " + name)
        packages = declarations["packages"]
        for package, inputs in packages.items():
            if not isinstance(inputs, list) or not inputs or any(not isinstance(value, str) or value not in directories for value in inputs):
                raise ValueError("invalid compiler dependencies for " + package)
        marker = re.compile(r'github\.com/system-inc/adamic/internal/(?:load|lower|native|javascript|ir|flow)|cmd/adamic|oracle/adamic|ADAMIC_(?:BIN|BINARY)|"adamic"|"os/exec"|os\.StartProcess|syscall\.Exec')
        self.compilerDrivers = declarations.get("drivers", {})
        if not isinstance(self.compilerDrivers, dict):
            raise ValueError("invalid compiler driver declarations")
        for driver, declaration in self.compilerDrivers.items():
            inputs = declaration.get("dependencies") if isinstance(declaration, dict) else None
            if not driver.startswith("stage3/drivers/") or not isinstance(inputs, list) or not inputs or any(value not in directories for value in inputs):
                raise ValueError("invalid compiler driver dependencies for " + driver)
        missing, touchedFiles = [], set(changed)
        for path in self.git(self.arguments.tree, "ls-files").splitlines():
            if path.startswith("stage3/drivers/") and path.endswith((".py", ".sh", ".cjs", ".mjs")):
                with open(os.path.join(self.arguments.tree, path)) as handle:
                    source = handle.read()
                if re.search(r"subprocess|child_process|\badamic\b|\bcompiler\b|ADAMIC", source) and not any(path.startswith(driver + "/") for driver in self.compilerDrivers):
                    missing.append(path)
                continue
            if not path.endswith(".go"):
                continue
            with open(os.path.join(self.arguments.tree, path)) as handle:
                source = handle.read()
            package = os.path.dirname(path)
            while package and package not in directories and package not in packages:
                package = os.path.dirname(package)
            if not package:
                if marker.search(source) and path.endswith("_test.go"):
                    missing.append(path)
                elif marker.search(source):
                    self.result.setdefault("compiler_consumers_without_test_package", []).append(path)
                continue
            if marker.search(source) and package not in packages:
                missing.append(path)
        # Only consumers this change adds or edits can turn its gate red; any already on the base are named, so one
        # undeclared package never reds every worker's gate.
        onBase = [path for path in missing if path not in touchedFiles]
        if onBase:
            self.result["compiler_consumers_undeclared_on_base"] = onBase
        missing = [path for path in missing if path in touchedFiles]
        if missing:
            raise ValueError("compiler dependency census: undeclared compiler consumers: " + ", ".join(missing))
        self.result["compiler_dependency_map"] = {"root": root, "path": name}
        return packages

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

    @declaresInputs("tools", lambda tree, unit, full: treeInputs("tool declaration census reads tracked Go sources across the repository"))
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

    @declaresInputs("build", lambda tree, unit, full: goCommandInputs(tree, goPhaseCommand("build")))
    def build(self):
        self.result["build_ok"] = self.step("build", goPhaseCommand("build"))

    @declaresInputs("vet", lambda tree, unit, full: goCommandInputs(tree, goPhaseCommand("vet")))
    def vet(self):
        self.result["vet_ok"] = self.step("vet", goPhaseCommand("vet"))

    def testSplit(self, packages, log, productsOnly=False):
        """Compile and list selected packages, build their products, then run ordinary test units.

        Product discovery precedes ordinary-test selection, so even a requested-only package
        builds every declared product. No tests start until all products have passed.
        """
        started = time.monotonic()
        # Beside the slot's tree, never in the out directory: everything there is published, and a test
        # binary is tens of megabytes. One slot runs one gate at a time, so each run overwrites its own.
        binaries = os.path.realpath(self.arguments.tree) + "-binaries"
        os.makedirs(binaries, exist_ok=True)
        record = TestSeconds()
        history = record.read()
        observations, paused = {}, set()

        def watch(line):
            try:
                event = json.loads(line)
            except ValueError:
                return
            name = event.get("Test") or ""
            if not name:
                return
            key = (event.get("Package"), name.split("/", 1)[0])
            with self.lock:
                if "=== PAUSE" in event.get("Output", "") or event.get("Action") == "pause":
                    paused.add(key)
                if "/" not in name and event.get("Action") in ("pass", "fail", "skip") and "Elapsed" in event:
                    observations[key] = (event["Elapsed"], key in paused)

        self.watchers.append(watch)
        slots = threading.Semaphore(self.arguments.parallel)
        queue = TestQueue(self.arguments.parallel)
        self.result["test_outcomes"] = []
        self.result["test_starts"] = []
        self.testStartDetails = {}
        self.testsStarted = started
        self.testLaunchLock = threading.Lock()
        # At most two test binaries build at once: each go test -c runs its own pool of compile processes as
        # wide as the slot, so a big gate building one per test slot ran hundreds of compiles in a 12-CPU
        # slot, and three such gates took Cloud to load 900 and Workshop to 10 GB free (Oct 8 11:2xZ).
        builds = threading.Semaphore(2)
        threads = []
        declared = productDeclarations(self.packageDirectories, packages)
        productsReady = threading.Event()
        productsFailed = threading.Event()
        buildsDone = threading.Event()
        productRows = [{"package": package, "test": name, "product": True, "status": "not run"}
                       for package in sorted(declared) for name in sorted(declared[package])]
        productsRemaining = len(productRows)
        productStarted = time.monotonic()
        if productRows:
            self.planned = getattr(self, "planned", ["tests"])
            if "products" not in self.planned:
                self.planned.insert(self.planned.index("tests") if "tests" in self.planned else 0, "products")
            self.result["product_units"] = productRows
        else:
            productsReady.set()
        productPool = ThreadPoolExecutor(max_workers=max(1, slotCPUs() // 4))

        def runProduct(row, binary):
            self.productUnit(row, binary, log)
            nonlocal productsRemaining
            with self.lock:
                if row["status"] == "passed":
                    productsRemaining -= 1
                    if productsRemaining == 0 and not productsFailed.is_set():
                        self.steps["products"] = round(time.monotonic() - productStarted, 3)
                        self.exits["products"] = 0
                        productsReady.set()
                else:
                    self.steps["products"] = round(time.monotonic() - productStarted, 3)
                    self.exits["products"] = 1
                    productsFailed.set()

        def checkProducts(package, names):
            listed = {name for name in names if name.startswith("TestProduct_")}
            if listed != declared[package]:
                productsFailed.set()
                self.exits["products"] = 1
                self.steps.setdefault("products", round(time.monotonic() - productStarted, 3))
                if "products" not in getattr(self, "planned", []):
                    self.planned = getattr(self, "planned", ["tests"])
                    self.planned.insert(self.planned.index("tests") if "tests" in self.planned else 0, "products")
                self.fail("products", "%s product scan/list mismatch: source=%s listed=%s" %
                          (package, sorted(declared[package]), sorted(listed)))
                return False
            return True
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
                    if not checkProducts(importPath, []):
                        return
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
            if not checkProducts(importPath, names):
                return
            for row in productRows:
                if row["package"] == importPath:
                    productPool.submit(runProduct, row, binary)
            names = [] if productsOnly else [name for name in names if not name.startswith("TestProduct_")]
            if importPath in getattr(self, "onlyTests", {}):
                names = [name for name in names if inFamily(name, self.onlyTests[importPath])]
                if not names:
                    # A package gated only for requested tests that selects none of them proves nothing: red, never a pass.
                    self.fail("requested", "requested tests ran: 0 in %s: %s match no test in its -test.list" % (importPath, ", ".join(sorted(self.onlyTests[importPath]))))
                    return
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
            deferred = sorted(set(names) & getattr(self, "deferred", {}).get(importPath, set()))
            if deferred:
                with self.lock:
                    self.result.setdefault("deferred_to_full_gate", []).extend(importPath + " " + name for name in deferred)
                names = [name for name in names if name not in deferred]
            self.result.setdefault("split_tests", {})[importPath] = len(names)
            with self.lock:
                tally["planned"] += len(names)
                self.result["test_outcomes"].extend({"package": importPath, "test": name, "status": "not run"} for name in names)
            ready = []
            for name in names:
                command = ["go", "tool", "test2json", "-t", "-p", importPath, binary, "-test.v=test2json", "-test.paniconexit0",
                           "-test.count=1"] + ([] if self.complete else ["-test.failfast"]) + ["-test.timeout=30m", "-test.parallel=2", "-test.skip", "^TestProduct_", "-test.run", patterns.get(name, "^%s$" % name)]
                seconds, parallel = history.get((importPath, name), (60, False))
                ready.append((seconds, parallel, importPath, name, command))
            # A package's tests arrive together, so a waiting worker can't take its short one before its long one.
            queue.add(ready)

        def packageWeight(package):
            known = [seconds for (path, _), (seconds, _) in history.items() if path == package]
            if not known:
                return (60, 1)
            return (max(known), 0)

        remaining = sorted(packages, key=lambda package: (not bool(declared[package]), -packageWeight(package)[0], -packageWeight(package)[1], package))

        def buildWorker():
            while True:
                with self.lock:
                    if not remaining or getattr(self, "stopped", None) or (self.failure is not None and not self.complete):
                        return
                    package = remaining.pop(0)
                try:
                    runPackage(package)
                except BaseException:
                    self.fail("tests", traceback.format_exc())

        def testWorker():
            while True:
                self.testLaunchLock.acquire()
                launched = threading.Event()
                item = queue.take()
                if item is None:
                    self.testLaunchLock.release()
                    return
                package, name, seconds, permits, command = item
                try:
                    while not productsReady.wait(0.05):
                        if productsFailed.is_set() or buildsDone.is_set() or getattr(self, "stopped", None) or (self.failure is not None and not self.complete):
                            return
                    if productsFailed.is_set() or getattr(self, "stopped", None) or (self.failure is not None and not self.complete):
                        return
                    if permits == 4:
                        command = ["-test.parallel=8" if part == "-test.parallel=2" else part for part in command]
                    with self.lock:
                        self.testStartDetails[tuple(command)] = {"package": package, "test": name,
                            "expected_seconds": seconds, "slots": permits, "launched": launched}
                    self.slotted(command, log, package, name, tally)
                except BaseException:
                    self.fail("tests", traceback.format_exc())
                finally:
                    if not launched.is_set():
                        self.testLaunchLock.release()
                    queue.release(permits)

        # Start tests immediately; only their launches wait for products, while builds stream in.
        testers = [] if productsOnly else [self.guarded("tests", testWorker) for _ in range(self.arguments.parallel)]
        for thread in testers:
            thread.start()
        for _ in range(min(2, self.arguments.parallel)):
            thread = self.guarded("tests", buildWorker)
            thread.start()
            threads.append(thread)
        for thread in threads:
            thread.join()
        productPool.shutdown(wait=True)
        if productRows:
            self.steps.setdefault("products", round(time.monotonic() - productStarted, 3))
            self.exits["products"] = 0 if productsReady.is_set() and not productsFailed.is_set() else 1
        buildsDone.set()
        queue.close()
        for thread in testers:
            thread.join()

        self.watchers.remove(watch)
        record.update(observations)
        if productsOnly:
            return tally["packages"] == len(packages) and (not productRows or self.exits.get("products") == 0) and self.failure is None
        self.steps["tests"] = round(time.monotonic() - started, 1)
        complete = tally["packages"] == len(packages) and tally["passed"] == tally["planned"]
        self.exits["tests"] = 0 if complete and self.failure is None else 1
        self.result["split_tally"] = dict(tally, touched=len(packages))

    def productUnit(self, row, binary, log):
        """A declared build, including setup, on the four-CPU reference shape."""
        if self.failure is not None or getattr(self, "stopped", None):
            return
        package, name = row["package"], row["test"]
        command = ["go", "tool", "test2json", "-t", "-p", package, binary,
                   "-test.v=test2json", "-test.paniconexit0", "-test.count=1", "-test.timeout=%ds" % productKillSeconds,
                   "-test.parallel=4", "-test.run", "^%s$" % re.escape(name)]
        before = time.monotonic()
        try:
            code = self.stream("products", command, log, self.packageDirectories[package], {"GOMAXPROCS": "4"})
        except BaseException:
            code = None
            self.fail("products", "%s %s\n%s" % (package, name, traceback.format_exc()))
        row.update(seconds=round(time.monotonic() - before, 6), status="passed" if code == 0 else "failed")
        if code != 0:
            self.fail("products", "%s %s failed or was killed" % (package, name))

    def slotted(self, command, log, importPath, name, tally):
        environment = None
        if (name == "TestWASI" or wasiUnit.fullmatch(name)) and importPath == module + "/internal/native" and os.environ.get("WASI_SYSROOT"):
            # As the whole gate runs it: the WASI SDK's clang first, so wasm-ld finds the wasm32 builtins.
            environment = {"PATH": os.path.join(os.path.dirname(os.path.dirname(os.environ["WASI_SYSROOT"])), "bin") + os.pathsep + os.environ["PATH"]}
        code = self.stream("tests", command, log, self.packageDirectories[importPath], environment)
        with self.lock:
            row = next(row for row in self.result["test_outcomes"] if row["package"] == importPath and row["test"] == name)
            row["status"] = "not run" if code is None else ("passed" if code == 0 else "failed")
        if code == 0:
            with self.lock:
                tally["passed"] += 1

    def capture(self, command, directory):
        process = self.spawn(command, subprocess.PIPE, subprocess.PIPE, directory)
        output, errors = process.communicate()
        if process.returncode != 0:
            self.fail("tests", "%s exited %d\n%s" % (" ".join(command), process.returncode, (output + errors)[-4000:]))
            return None
        return output

    def emitC(self, binary, path):
        """One emission of path's C by stage 0, in its own process: its exit code and stdout."""
        process = self.spawn([binary, "c", path], subprocess.PIPE, subprocess.DEVNULL)
        output, _ = process.communicate()
        return process.returncode, output

    @declaresInputs("determinism", lambda tree, unit, full: treeInputs("smoke source programs and their runtime/import reads are not yet bounded"))
    def determinism(self, smoke):
        """Emission is deterministic (@system_adamic, Oct 9: a 43,250-line main.c on the Threadripper against 5,470 lines
        on a Codex box for the same tree). Stage 0 emits each smoke fixture's C twice, each in its own process, so Go's
        per-process map order can differ, and the two must match byte for byte, exit code included. The record keeps
        each program's C sha256 (emission_sha256), so records of one sha on two boxes can be compared."""
        started = time.monotonic()
        binary = os.path.join(os.path.abspath(self.arguments.out), "determinism-adamic")
        with open(os.path.join(self.arguments.out, "determinism.log"), "w") as output:
            code = self.spawn(["go", "build", "-o", binary, "./cmd/adamic"], output).wait()
        if code != 0:
            self.exits["determinism"] = code
            self.fail("determinism", "go build ./cmd/adamic for the determinism check failed (determinism.log)")
            return
        programs = sorted({path for _, path in smoke if path.endswith((".a", ".ts")) and os.path.isfile(os.path.join(self.arguments.tree, path))})
        hashes, differing = {}, []

        def check(path):
            first, second = self.emitC(binary, path), self.emitC(binary, path)
            hashes[path] = hashlib.sha256((first[1] or "").encode()).hexdigest()
            if first != second:
                a, b = (first[1] or "").split("\n"), (second[1] or "").split("\n")
                line = next((index + 1 for index in range(max(len(a), len(b))) if index >= len(a) or index >= len(b) or a[index] != b[index]), 0)
                differing.append("%s (exit %d and %d, first difference at line %d)" % (path, first[0], second[0], line))

        threads = [threading.Thread(target=check, args=(path,)) for path in programs]
        for index in range(0, len(threads), 8):
            for thread in threads[index:index + 8]:
                thread.start()
            for thread in threads[index:index + 8]:
                thread.join()
        self.result["emission_sha256"] = dict(sorted(hashes.items()))
        self.steps["determinism"] = round(time.monotonic() - started, 1)
        if differing or len(hashes) != len(programs):
            self.exits["determinism"] = 1
            self.fail("determinism", "stage 0 emitted different C for the same program in two runs: %s" % "; ".join(sorted(differing) or ["a check didn't finish"]))
            return
        self.exits["determinism"] = 0

    @declaresInputs("smoke", lambda tree, unit, full: treeInputs("smoke source programs and their runtime/import reads are not yet bounded"))
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
        self.test("smoke", ["go", "test", "-count=1"] + ([] if self.complete else ["-failfast"]) + ["-json", "-timeout", "30m", "-run", smokePattern(smoke), "./internal/oracle"], log)
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

    def cacheDrains(self):
        # Both gates drain after their test threads join. The candidate owns the publisher;
        # older trees have neither stage. Pool scheduling is managed separately by Loom.
        if "audit" not in self.planned or self.stopped or getattr(self, "cancelled", False):
            return
        if self.failure is not None and not self.arguments.full and not self.complete:
            return
        self.cacheDrain("audit")
        if self.failure is None:
            self.cacheDrain("upload")

    @declaresInputs("upload", lambda tree, unit, full: treeInputs("publisher reads externally generated spool records; repository inputs alone do not identify the verdict"))
    @declaresInputs("audit", lambda tree, unit, full: treeInputs("publisher reads externally generated audit/spool records; repository inputs alone do not identify the verdict"))
    def cacheDrain(self, name):
        required = name == "audit"
        command = ["go", "run", "./internal/buildcache/cmd/buildcache-publish"]
        if required:
            command.append("-audit")
        started = time.monotonic()
        row = {"package": module + "/internal/buildcache/cmd/buildcache-publish",
               "test": name, "command": command, "required": required}
        self.result.setdefault("cache_drain_units", []).append(row)
        try:
            with open(os.path.join(self.arguments.out, name + ".log"), "w") as log:
                code = self.stream(name, command, log, fatal=required)
            code = 1 if code is None else code
        except BaseException:
            code = 1
            self.result.setdefault("cache_drain_output", {})[name] = traceback.format_exc()
            if required:
                self.fail(name, self.result["cache_drain_output"][name])
        seconds = round(time.monotonic() - started, 1)
        row.update(seconds=seconds, exit=code, status="passed" if code == 0 else "failed",
                   detail=self.result.get("cache_drain_output", {}).get(name, ""))
        # Publishing is the store's gain: record its actual exit in the unit, while the
        # optional stage completes successfully even when the store is unavailable.
        self.exits[name] = code if required else 0
        self.steps[name] = seconds

    def stream(self, name, command, log, directory=None, environment=None, fatal=True):
        """Runs one go test (or test2json) process, copying its events to the log; the first failing
        event fails the gate. Returns the exit code, or None if the gate failed."""
        stderrPath = os.path.join(self.arguments.out, "%s-%d.stderr" % (re.sub(r"[^A-Za-z0-9_.-]", "_", name), threading.get_ident()))
        stderr = open(stderrPath, "w")
        try:
            process = self.spawn(command, subprocess.PIPE, stderr, directory, environment)
        except BaseException:
            stderr.close()
            raise
        deadline = None
        drained = threading.Event()
        drain = name in ("audit", "upload")
        transcript = []
        if drain or (name == "products" and "test2json" in command):
            def expired():
                if not drained.is_set():
                    seconds = unitKillSeconds if drain else productKillSeconds
                    identity = name if drain else "%s %s" % (command[command.index("-p") + 1], command[command.index("-test.run") + 1])
                    transcript.append("%s killed at %d s\n" % (identity, seconds))
                    if fatal:
                        self.fail(name, "%s killed at %d s" % (identity, seconds))
                    with self.lock:
                        self.killSessions({process.pid})
            deadline = threading.Timer(unitKillSeconds if drain else productKillSeconds, expired)
            deadline.daemon = True
            deadline.start()
        output = {}
        for line in process.stdout:
            if drain:
                transcript.append(line)
            if log is not None:
                with self.lock:
                    log.write(line)
            if id(process) in getattr(self, "killedByStop", set()):
                continue
            if drain:
                continue
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
                    if event["Action"] == "fail":
                        self.failedTests.append("%s %s" % (event.get("Package"), event.get("Test")))
            if event.get("Action") == "fail" or event.get("Action") == "build-fail":
                text = "".join(output.get(key, [])) or "".join(output.get((event.get("Package"), None), []))
                self.fail(name, "%s %s\n%s" % (event.get("Package"), event.get("Test") or "(package)", text[-4000:]))
        code = process.wait()
        drained.set()
        if deadline is not None:
            deadline.cancel()
        process.stdout.close()
        stderr.close()
        if id(process) in getattr(self, "killedByStop", set()):
            return None
        if drain:
            with open(stderrPath) as handle:
                text = "".join(transcript) + handle.read()
            self.result.setdefault("cache_drain_output", {})[name] = text
            if code != 0 and fatal and self.failure is None:
                detail = text if len(text) <= 4000 else text[:2000] + "\n...\n" + text[-2000:]
                self.fail(name, "%s exited %d\n%s" % (" ".join(command), code, detail))
        if not drain and code != 0 and fatal and self.failure is None:
            with open(stderrPath) as handle:
                identity = " ".join(command[:4])
                if "test2json" in command and "-test.run" in command:
                    identity = "%s %s" % (command[command.index("-p") + 1], command[command.index("-test.run") + 1])
                self.fail(name, "%s exited %d\n%s" % (identity, code, handle.read()[-4000:]))
        return None if fatal and self.failure is not None and not self.arguments.full and not self.complete else code

    @declaresInputs("census", lambda tree, unit, full: treeInputs("skip census walks the whole repository and reads the supplied merged log and Git branch state"))
    def checkCensus(self):
        started = time.monotonic()
        tools = self.arguments.tools
        # census-extra.json: skips in main the tools tree doesn't have yet, classified, checked against the log only.
        # -git: a pending skip passes only while the branch it awaits is off main (asked of the candidate's origin).
        process = self.spawn(["go", "run", "./internal/skipcensus/cmd", "-root", tools, "-extra", os.path.join(tools, "cloud/fast-gate/census-extra.json"),
                              "-git", os.path.abspath(self.arguments.tree),
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
            if len(fields) == 5 and fields[0] in ("pending", "pending-landed", "pending-unknown"):
                self.census.setdefault("pending", []).append("%s %s (%s)" % (fields[1].rsplit("/", 1)[-1], fields[2], fields[4]))
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
            if getattr(self, "stopped", None) or getattr(self, "cancelled", False) or (self.failure is not None and not self.arguments.full and not self.complete):
                raise SystemExit(1)
            # Uncached: the oracle's result caches would otherwise answer a test without running it.
            variables = dict(os.environ, **gateEnvironment)
            if not self.arguments.full:
                variables.update(getattr(self, "sampling", {}))
            # The box has Node but no npm; stage 3's apply runs npm ci, so the gates' pinned npm is on PATH.
            variables["PATH"] = os.path.expanduser("~/fast-gate/npm/bin") + os.pathsep + variables["PATH"]
            variables.update(environment or {})
            variables.update(self.buildStoreEnvironment)
            process = subprocess.Popen(command, cwd=directory or self.arguments.tree, stdout=stdout, stderr=stderr, text=True, start_new_session=True, env=variables)
            self.processes.append(process)
            self.recordTestStart(command)
        return process

    def recordTestStart(self, command):
        # Called under self.lock immediately after Popen, so the artifact is launch order.
        details = getattr(self, "testStartDetails", {}).get(tuple(command))
        if details is not None:
            row = {key: value for key, value in details.items() if key != "launched"}
            self.result["test_starts"].append(dict(row,
                after_seconds=round(time.monotonic() - self.testsStarted, 3)))
            details["launched"].set()
            self.testLaunchLock.release()

    def fail(self, step, detail):
        with self.lock:
            if self.failure is not None or getattr(self, "stopped", None):
                return
            self.failure = {"step": step, "detail": detail, "after_seconds": round(time.monotonic() - self.started, 1)}
            reading = boxLoad()
            getattr(self, "result", {}).setdefault("box_load", {})["first_failure"] = reading
            self.failure["under_load"] = underLoad(reading)
            loaded = (", under load (%s at first failure)" % loadWords(reading)) if self.failure["under_load"] else ""
            detail = "%s\n(box %s at first failure)" % (detail, loadWords(reading))
            # A Python traceback is the gate tool's own exception, not the change's verdict: a complete or
            # whole run goes no further on broken tools, since it would teach nothing (@system_adamic, Oct 8,
            # after a-check crashed on views slice 1 and its complete run went on for minutes).
            crashed = detail.lstrip().startswith("Traceback (most recent call last)")
            if crashed:
                self.failure["tool_crash"] = True
            print("FIRST FAILURE (%s%s, at %.1f s):\n%s" % (step, ", the gate tool crashed" if crashed else "", self.failure["after_seconds"], detail), flush=True)
            if crashed:
                if self.arguments.full:
                    with open(os.path.join(self.arguments.out, "first-failure.txt"), "w") as handle:
                        handle.write(detail + "\n")
                self.killSessions()
                return
            if self.arguments.full and getattr(self.arguments, "run_to_end", False):
                with open(os.path.join(self.arguments.out, "first-failure.txt"), "w") as handle:
                    handle.write(detail + "\n")
                self.status("red: %s first failure at %s after %.1f s%s (parity: running to the end)" % (self.arguments.sha, step, self.failure["after_seconds"], loaded))
                return
            if self.arguments.full:
                # A whole gate fails fast and frees its box (Kirk, Oct 8: "all tests should fail fast and loud and
                # immediately give feedback and cancel themselves to free for next run"): the first failure is
                # published, the run is cancelled, and its record says so. Triage is the pool's job, where units run
                # in parallel and every red comes back at once.
                with open(os.path.join(self.arguments.out, "first-failure.txt"), "w") as handle:
                    handle.write(detail + "\n")
                self.status("red: %s first failure at %s after %.1f s%s (cancelled after first failure)" % (self.arguments.sha, step, self.failure["after_seconds"], loaded))
                self.cancelled = True
                self.killSessions()
                return
            if self.complete:
                # Nothing is stopped: every test and fixture still runs, and the verdict names them all.
                return
            self.killSessions()

    def killSessions(self, sessions=None):
        """Drain our sessions even after a leader exited and children were reparented.

        Setpgid cannot escape a session. Zombies cannot run or fork and must be
        reaped by their new parent; waiting for those would hang on a slow init.
        Call under self.lock so spawn cannot register a new session mid-drain.
        """
        if sessions is None:
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

    def budget(self, ledger, units):
        """Red at 'budget' for every new unit over longTestSeconds (its test isn't in the base) that the burn-down
        doesn't hold, naming each one, its seconds and the box that measured it. Existing units over it are named
        as drift, and burn-down units this run measured under the line too (budget_can_leave_burndown), so the
        record shrinks."""
        burndown = set()
        try:
            with open(os.path.join(self.arguments.tools, "cloud/fast-gate/budget-burndown.tsv")) as handle:
                for line in handle:
                    fields = line.rstrip("\n").split("\t")
                    if len(fields) >= 2 and not line.startswith("#"):
                        if not fields[1].startswith("TestProduct_"):
                            burndown.add((fields[0], fields[1]))
        except OSError:
            pass
        over = {(package, name) for package, name, _, _ in ledger}
        listed = sorted("%s %s" % key for key in over & burndown)
        offBurndown = [row for row in ledger if (row[0], row[1]) not in burndown]
        # New means added by this change: absent where the candidate forked from main, not from main's tip. A candidate
        # cut from an older main otherwise reads main's later renames as its own new tests (Oct 9 05:13Z: gocacheprog
        # 90b10e95 red at budget on 113 of main's own units, stage 3's fixture shards and TestOwnedWitnesses).
        found = subprocess.run(["git", "-C", self.arguments.tree, "merge-base", self.arguments.base, self.arguments.sha],
                               capture_output=True, text=True)
        fork = found.stdout.strip() if found.returncode == 0 and found.stdout.strip() else self.arguments.base
        products = [row for row in offBurndown if row[1].startswith("TestProduct_")]
        offBurndown = [row for row in offBurndown if row not in products]
        unlisted = [row for row in offBurndown if not testInBase(self.arguments.tree, fork, row[0], row[1])]
        drift = [row for row in offBurndown if row not in unlisted]
        def unitWords(package, name, seconds):
            return ("product " if name.startswith("TestProduct_") else "") + "%s %s %.1f s" % (package, name, seconds)
        self.result.update({
            "budget_seconds": longTestSeconds,
            "budget_instrument": {"box": os.uname().nodename, "cpus": os.cpu_count(), "load": [round(value, 1) for value in os.getloadavg()],
                                  "reference": "a 4-CPU Codex instance; this box until Loom's tier measures there"},
            "budget_burndown_units": listed,
            "budget_over": [unitWords(package, name, seconds) for package, name, seconds, _ in unlisted],
            "budget_drift": [unitWords(package, name, seconds) for package, name, seconds, _ in drift],
            "budget_can_leave_burndown": sorted("%s %s" % key for key in (burndown & set(units)) - over),
            "products_over_budget": [unitWords(package, name, seconds) for package, name, seconds, _ in products],
        })
        if unlisted and self.failure is None and not getattr(self, "stopped", None):
            instrument = self.result["budget_instrument"]
            self.fail("budget", "%d new test units over the %d s budget (measured on %s, %d CPUs, load %s; split each into units under %d s):\n%s" % (
                len(unlisted), longTestSeconds, instrument["box"], instrument["cpus"], "/".join(str(value) for value in instrument["load"]), longTestSeconds,
                "\n".join("  %s (%s)" % (unitWords(package, name, seconds), action) for package, name, seconds, action in unlisted)))

    def status(self, line):
        with open(os.path.join(self.arguments.out, "status.txt"), "w") as handle:
            handle.write(line + "\n")

    def finish(self):
        if getattr(self, "stopThread", None):
            self.stopThread.join()
        with self.lock:
            self.killSessions()
        wall = round(time.monotonic() - self.started, 1)
        self.result.setdefault("box_load", {})["end"] = boxLoad()
        units = testUnits(os.path.join(self.arguments.out, "test.jsonl"))
        for row in self.result.get("product_units", []):
            key = row["package"], row["test"]
            units = {unit: value for unit, value in units.items()
                     if not (unit[0] == key[0] and (unit[1] == key[1] or unit[1].startswith(key[1] + "/") or unit[1] == key[1] + " (setup)"))}
            if "seconds" in row:
                units[key] = (row["seconds"], "pass" if row["status"] == "passed" else "fail")
        # Drains have a 90 s wall deadline, not the tests' 30 s budget.
        self.budget(longTests(units), units)
        for row in self.result.get("cache_drain_units", []):
            units[row["package"], row["test"]] = (row["seconds"], "pass" if row["exit"] == 0 else "fail")
        self.result["units"] = [{"package": package, "test": name, "seconds": seconds, "action": action,
                                 "product": name.startswith("TestProduct_")}
                                for (package, name), (seconds, action) in sorted(units.items())]
        ledger = longTests(units)
        unfinished = [stage for stage in self.planned if self.exits.get(stage) != 0]
        if unfinished and self.failure is None and not getattr(self, "stopped", None):
            self.fail(unfinished[0], "planned stages without a recorded exit 0: %s (exits %s)" % (", ".join(unfinished), self.exits))
        green = self.failure is None and not unfinished and not getattr(self, "stopped", None)
        self.result.update({
            "wall_seconds": wall,
            "steps_seconds": self.steps,
            "pass": self.counts["pass"],
            "fail": self.counts["fail"],
            "failed_tests": self.failedTests,
            "complete": self.complete,
            "skip": self.counts["skip"],
            "required_input_skips": self.census["required_input"],
            "unclassified_skips": self.census["unclassified"],
            "pending_skips": self.census.get("pending", []),
            "failure": self.failure,
            "planned_stages": self.planned,
            "stages_exit": self.exits,
            "cancelled_after_first_failure": bool(getattr(self, "cancelled", False)),
            "finished": True,
        })
        if getattr(self, "stopped", None):
            self.result["stopped"] = self.stopped
        with open(os.path.join(self.arguments.out, "long-tests.tsv"), "w") as handle:
            for package, name, seconds, action in ledger:
                handle.write("%s\t%s\t%.2f\t%s\n" % (package, name, seconds, action))
        self.result.update({"long_test_threshold_seconds": longTestSeconds, "long_tests": len(ledger),
                            "long_test_seconds": round(sum(row[2] for row in ledger), 1)})
        outcomes = self.result.get("test_outcomes", [])
        self.result["planned_test_counts"] = {status: sum(row["status"] == status for row in outcomes)
                                               for status in ("passed", "failed", "not run")}
        self.result["gate_kind"] = self.kind
        with open(os.path.join(self.arguments.out, self.kind + ".json"), "w") as handle:
            json.dump(self.result, handle, indent=2)
            handle.write("\n")
        steps = " ".join("%s=%.1fs" % item for item in self.steps.items())
        if self.result.get("slow_packages"):
            steps += "; slow packages, over %d min: %s" % (slowPackageSeconds // 60, ", ".join("%s %.0fs" % (name.rsplit("/", 2)[-2] + "/" + name.rsplit("/", 1)[-1], seconds) for name, seconds in sorted(self.result["slow_packages"].items(), key=lambda item: -item[1])))
        steps += "; %d units over %d s (%.0f s, %d on the burn-down, %d drifted over)" % (len(ledger), longTestSeconds, self.result["long_test_seconds"], len(self.result.get("budget_burndown_units", [])), len(self.result.get("budget_drift", [])))
        if self.result.get("products_over_budget"):
            steps += "; products over %d s, listed not red: %s" % (longTestSeconds, "; ".join(self.result["products_over_budget"]))
        if self.census.get("pending"):
            steps += "; pending skips: %s" % "; ".join(self.census["pending"])
        steps += "; box " + ", ".join("%s at %s" % (loadWords(self.result["box_load"][moment]), moment.replace("_", " "))
                                      for moment in ("start", "first_failure", "end") if moment in self.result["box_load"])
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
        if getattr(self, "stopped", None) and self.failure is None:
            self.status("void: stopped before a verdict: " + self.stopped["reason"])
        elif green:
            self.status("green: %s %s gate in %.1f s (%s), %d packages, %d pass, %d skip, smoke %d fixtures" % (self.arguments.sha, self.kind, wall, steps, len(self.result.get("package_list", self.result.get("packages", []))), self.counts["pass"], self.counts["skip"], len(self.result.get("smoke_fixtures", []))))
        else:
            crash = " (the gate tool crashed, not the change)" if self.failure.get("tool_crash") else ""
            cancelled = ", cancelled after first failure" if getattr(self, "cancelled", False) else ""
            # A red whose first failure ran above the box's core count says so itself, so a page reads it as load first.
            loaded = ", under load (%s at first failure)" % loadWords(self.result["box_load"]["first_failure"]) if self.failure.get("under_load") else ""
            # A deferred red names the requested tests it lacked, so a page reads what didn't run (trio 114a6439, Oct 9
            # 10:50Z: "first failure at deferred" with nothing named, and five internal/native tests behind it).
            unproven = sorted(test.rsplit("/", 1)[-1] + " " + outcome for test, outcome in self.result.get("deferred_run_results", {}).items()
                              if outcome not in ("pass", "fail")) if self.failure["step"] == "deferred" else []
            if unproven:
                crash += " (%d requested deferred tests unproven: %s%s)" % (len(unproven), ", ".join(unproven[:5]), ", ..." if len(unproven) > 5 else "")
            self.status("red: %s %s gate, first failure at %s%s after %.1f s%s%s (%s), %d fail, %d pass" % (self.arguments.sha, self.kind, self.failure["step"], crash, self.failure["after_seconds"], loaded, cancelled, steps, self.counts["fail"], self.counts["pass"]))
        if getattr(self, "stopped", None) and self.failure is not None:
            with open(os.path.join(self.arguments.out, "status.txt"), "a") as handle:
                handle.write("stopped: " + self.stopped["reason"] + "\n")
        with open(os.path.join(self.arguments.out, "status.txt")) as handle:
            print(handle.read(), end="", flush=True)


def eventTime(value):
    """A test2json Time (RFC 3339, nanoseconds) as epoch seconds, or None."""
    if not isinstance(value, str):
        return None
    try:
        return datetime.datetime.fromisoformat(re.sub(r"(\.\d{6})\d+", r"\1", value.replace("Z", "+00:00"))).timestamp()
    except ValueError:
        return None


def testUnits(path):
    """The test units in a go test -json log that passed or failed, as {(package, unit): (seconds, action)}. A
    test without subtests is one unit. A test with subtests is split: each subtest is a unit (recursively), and
    so is the test's own time outside them, '<test> (setup)': from its start to its first subtest's start, plus
    from its last subtest's end to its own end, by the events' times (cohere, Oct 8: shards run in parallel, so a
    build before them would hide inside 'elapsed less the subtests''). Without times, its elapsed less its
    subtests'. A missing log has no units."""
    tests, started, ended = {}, {}, {}
    try:
        with open(path) as handle:
            for line in handle:
                try:
                    event = json.loads(line)
                except ValueError:
                    continue
                if not isinstance(event, dict):
                    continue
                name, seconds = event.get("Test"), event.get("Elapsed")
                if not isinstance(name, str) or not name:
                    continue
                key = (event.get("Package", ""), name)
                if event.get("Action") == "run" and eventTime(event.get("Time")) is not None:
                    started.setdefault(key, eventTime(event.get("Time")))
                if event.get("Action") not in ("pass", "fail") or not isinstance(seconds, (int, float)):
                    continue
                tests[key] = (seconds, event["Action"])
                if eventTime(event.get("Time")) is not None:
                    ended[key] = eventTime(event.get("Time"))
    except OSError:
        return {}
    children = {}
    for package, name in tests:
        if "/" in name:
            children.setdefault((package, name.rsplit("/", 1)[0]), []).append((package, name))
    units = {}
    for key, (seconds, action) in tests.items():
        if key not in children:
            units[key] = (seconds, action)
            continue
        kids = children[key]
        if key in started and key in ended and all(child in started and child in ended for child in kids):
            setup = (min(started[child] for child in kids) - started[key]) + (ended[key] - max(ended[child] for child in kids))
        else:
            setup = seconds - sum(tests[child][0] for child in kids)
        if setup > 0:
            units[key[0], key[1] + " (setup)"] = (round(setup, 2), action)
    return units


def familyMember(name, requested):
    """Whether a top-level test is a requested name or one of its split's generated shards: the name followed by
    Unit<n>, Points<n> or _<n> (@system_adamic, Oct 9 12:19Z). Never a bare prefix: TestWASIRefusesUnsupportedOptions
    is not TestWASI run, while TestNormalizeMatchesNodePoints00 is TestNormalizeMatchesNode."""
    return name == requested or re.fullmatch(re.escape(requested) + r"(Unit\d+|Points\d+|_\d+)", name) is not None


def inFamily(name, requested):
    """A requested test name selects itself and its split's shards: a family split into shards is requested by its
    name, and an exact anchor selected none of it, so a requested set ran nothing and read as passed (@system_adamic,
    Oct 9 11:01Z)."""
    return any(familyMember(name, prefix) for prefix in requested)


def familyOutcome(outcomes, package, name):
    """A requested name's outcome in topLevelOutcomes: the test's own when it ran under its name, else over its split's
    shards, fail if any failed, pass if any passed, and missing when none did (shards that all skipped proved nothing)."""
    exact = outcomes.get(package + " " + name)
    if exact:
        return exact
    members = [outcome for key, outcome in outcomes.items()
               if key.startswith(package + " ") and familyMember(key[len(package) + 1:], name)]
    for outcome in ("fail", "pass"):
        if outcome in members:
            return outcome
    return "missing"


def topLevelOutcomes(path):
    """Each top-level test's last pass, fail or skip in a go test -json log, by "<package> <test>"."""
    outcomes = {}
    try:
        with open(path) as handle:
            for line in handle:
                try:
                    event = json.loads(line)
                except ValueError:
                    continue
                if isinstance(event, dict) and event.get("Test") and "/" not in event["Test"] and event.get("Action") in ("pass", "fail", "skip"):
                    outcomes[event.get("Package", "") + " " + event["Test"]] = event["Action"]
    except OSError:
        pass
    return outcomes


def deferredClassedOut(tools, deferred):
    """Deferred tests the skip census classes measurement or opt-in-lane (internal/skipcensus/testdata/skips.json
    and cloud/fast-gate/census-extra.json), as ("<package> <test>", class): no gate gives them their input."""
    rows = []
    for name in ("internal/skipcensus/testdata/skips.json", "cloud/fast-gate/census-extra.json"):
        try:
            with open(os.path.join(tools, name)) as handle:
                rows += json.load(handle)
        except (OSError, ValueError):
            continue
    classed = {}
    for row in rows:
        if row.get("class") not in ("measurement", "opt-in-lane"):
            continue
        if not row.get("file"):
            continue
        for test in [row.get("test", "")] + list(row.get("callers") or []):
            classed[module + "/" + os.path.dirname(row["file"]), test] = row["class"]
    return sorted(("%s %s" % (importPath, name), classed[importPath, name])
                  for importPath, names in deferred.items() for name in names if (importPath, name) in classed)


def testInBase(tree, base, package, unit):
    """Whether the unit's top-level test function is defined in the package's tests at base. A unit is new
    when it isn't: added with this change."""
    test = unit.split("/", 1)[0].split(" ", 1)[0]
    directory = package[len(module) + 1:] if package.startswith(module + "/") else "."
    found = subprocess.run(["git", "-C", tree, "grep", "-q", "-E", r"^func %s\(" % re.escape(test), base, "--", ":(glob)" + directory + "/*_test.go"],
                           capture_output=True)
    return found.returncode == 0


def longTests(units):
    """The units (testUnits) over longTestSeconds, longest first: (package, unit, seconds, action)."""
    return sorted(((package, name, seconds, action) for (package, name), (seconds, action) in units.items() if seconds > longTestSeconds),
                  key=lambda row: (-row[2], row[0], row[1]))


# The whole gate's phases a pool unit can run alone (run.py --full --phase); the Go test set is the pool's own.
wholePhases = ["coverage", "tools", "build", "vet", "wasi", "stage3", "catalog", "determinism"]
# The unconditional non-test units; conditional executors are listed with --base.
fastBasePhases = ["coverage", "tools", "build", "vet", "smoke", "determinism", "census"]
fastPhases = fastBasePhases + ["stage3", "workers", "a-check", "catalog-apply", "catalog", "darwin"]
stage3Units = ["stage3-apply-tests", "stage3-lane-tests", "stage3-lane"]


def fastUnits(tree, base=None, sha=None, tools=None):
    """Fast non-test phases, adding change-dependent executors when --base is supplied.

    The planner supplies the same base, sha and tools as --select to enumerate the exact stage set.
    Enumeration reads the gate's executor map and catalog; it installs nothing and runs no stages.
    """
    units = list(fastBasePhases)
    if base is None:
        return units
    with tempfile.TemporaryDirectory() as out:
        gate = Gate(argparse.Namespace(tree=tree, base=base, sha=sha or git(tree, "rev-parse", "HEAD"),
                                       tools=tools or tree, out=out, full=False, branch="", branch_source="",
                                       session="", session_source=""))
        changed, executors = gate.fastPhaseInputs()
        if "stage3" in executors:
            units.append("stage3")
        if executors & {"workers", "bench-workers"}:
            units.append("workers")
        if "a-check" in executors and gate.result.get("unchecked_a_files"):
            units.append("a-check")
        if gate.catalogEntriesTouched(changed):
            units.append("catalog-apply")
        if "catalog" in executors:
            units.append("catalog")
        if "darwin" in executors:
            units.append("darwin")
    return units


def productDeclarations(directories, packages=None):
    """Source inventory, verified against the compiled binary before any product is run."""
    declared = {}
    for package in directories if packages is None else packages:
        names = set()
        for path in glob.glob(os.path.join(directories[package], "*_test.go")):
            with open(path) as source:
                names.update(re.findall(r"^func\s+(TestProduct_\w+)\s*\(\s*\w+\s+\*testing\.T\s*\)", source.read(), re.M))
        declared[package] = names
    return declared


def wholeUnits(tree):
    """Every pool unit of the whole gate, as '<phase>' or '<phase> <unit>' (run.py --full --list-units): each wasi
    fixture, stage 3's three commands, each catalog entry by number, and the other phases whole."""
    units = []
    for phase in wholePhases:
        if phase == "wasi":
            units += ["wasi " + name for name in wasiFixtures(tree)]
        elif phase == "stage3":
            units += ["stage3 " + name for name in stage3Units]
        elif phase == "catalog" and os.path.exists(os.path.join(tree, "verify/catalog/check.sh")):
            units += ["catalog %d" % entry["number"] for entry in catalogEntries(tree)]
        elif phase != "catalog":
            units.append(phase)
    return units


def slotCPUs():
    cpus = len(os.sched_getaffinity(0)) if hasattr(os, "sched_getaffinity") else (os.cpu_count() or 1)
    try:
        with open("/sys/fs/cgroup/cpu.max") as handle:
            quota, period = handle.read().split()
        if quota != "max":
            cpus = min(cpus, max(1, int(quota) // int(period)))
    except (OSError, ValueError):
        pass
    return cpus


def wasiFixtures(tree):
    """Read the compiler-owned fixture inventory; refuse changes we cannot enumerate. Since main 7e403e44 the
    fixtures are top-level units (TestWASIUnit00 and on, each a slice of the inventory): those are the phase's
    units, run by name, and the tests stage skips them (wasiSkip)."""
    with open(os.path.join(tree, "internal/native/wasm_test.go")) as handle:
        source = handle.read()
    if "func TestWASI(t *testing.T) {" not in source:
        units = re.findall(r"^func (TestWASIUnit[0-9]+)\(t \*testing\.T\) \{", source, re.M)
        if not units or len(set(units)) != len(units):
            raise ValueError("cannot enumerate TestWASI fixtures or TestWASIUnit units")
        return units
    body = source.split("func TestWASI(t *testing.T) {", 1)[1].split("\nfunc ", 1)[0]
    inventory = re.search(r'fixtures := \[\]string\{(.*?)\n\t\}', body, re.S)
    if inventory is None:
        raise ValueError("cannot enumerate TestWASI fixtures")
    fixtures = re.findall(r'"([^"\n]+)"', inventory[1])
    remainder = re.sub(r'"[^"\n]*"|\s|,', '', inventory[1])
    runs = re.findall(r't\.Run\(([^,]+),', body)
    if remainder or runs.count("fixture") != 1 or any(
            name != "fixture" and not re.fullmatch(r'"[^"\n]+"', name) for name in runs):
        raise ValueError("unrecognized TestWASI fixture registration")
    fixtures += [json.loads(name) for name in runs if name != "fixture"]
    if not fixtures or len(set(fixtures)) != len(fixtures):
        raise ValueError("empty or duplicate TestWASI fixtures")
    return fixtures


# TestWASI's fixtures as top-level units (main 7e403e44 on), and the tests stage's skip that leaves them to the wasi phase.
wasiUnit = re.compile(r"TestWASIUnit[0-9]+")
wasiSkip = "^TestWASI(Unit[0-9]+)?$"


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
