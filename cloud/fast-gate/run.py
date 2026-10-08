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
import hashlib
import json
import re
import os
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
smokeTest = "TestNativeAgreesWithNode"
oracle = module + "/internal/oracle"
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
    # A slot owns half the box's CPUs (cloud/fast-gate.sh pins it); half of those again run test
    # processes, since each also runs parallel subtests, clang and Node: a timing-sensitive test's
    # verdict shouldn't depend on what else runs.
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
        self.planned = ["build", "vet", "tests", "wasi", "census"] if arguments.full else ["build", "vet", "tests", "smoke", "census"]
        self.result = {
            "sha": arguments.sha,
            "branch": arguments.branch,
            "branch_source": arguments.branch_source,
            "session": arguments.session,
            "session_source": arguments.session_source,
            "base": arguments.base,
            "tools_sha": git(arguments.tools, "rev-parse", "HEAD"),
            "machine": {"hostname": socket.gethostname(), "nproc": os.cpu_count()},
            "uncached_tests": True,
            "build_ok": False,
            "vet_ok": False,
        }
        self.kind = "full" if arguments.full else "fast"
        self.status("running: %s gate of %s against %s" % (self.kind, arguments.sha, arguments.base))

    def run(self):
        tree = self.arguments.tree
        head = git(tree, "rev-parse", "HEAD")
        if head != self.arguments.sha:
            self.fail("setup", "the tree is at %s, not the candidate %s" % (head, self.arguments.sha))
            return
        if not self.npmPackages():
            return
        if self.arguments.full:
            self.runFull()
            return
        smoke, smokeSource = self.smokeList()
        self.deferred = self.deferredList()
        changed = git(tree, "diff", "--name-only", "%s...%s" % (self.arguments.base, self.arguments.sha)).split("\n")
        changed = [path for path in changed if path]
        packages, unowned = self.touched(changed)
        self.result.update({
            "changed_files": changed,
            "packages": packages,
            "unowned_files": unowned,
            "smoke_list": "cloud/fast-gate/smoke.txt",
            "smoke_list_source": smokeSource,
            "smoke_list_blob": git(smokeSource["root"], "hash-object", os.path.join(smokeSource["root"], "cloud/fast-gate/smoke.txt")),
            "smoke_fixtures": smoke,
        })
        if not smoke:
            self.fail("smoke", "no smoke list in the gated tree or the tools checkout")
            return
        # Everything starts at once: the tests compile what they need through the same build cache,
        # and the first failure of any step still stops all of them.
        log = open(os.path.join(self.arguments.out, "test.jsonl"), "w")
        threads = [self.guarded("build", self.build), self.guarded("vet", self.vet),
                   self.guarded("tests", self.testSplit, packages, log),
                   self.guarded("smoke", self.test, "smoke", ["go", "test", "-count=1", "-failfast", "-json", "-timeout", "30m", "-run", smokePattern(smoke), "./internal/oracle"], log)]
        for thread in threads:
            thread.start()
        for thread in threads:
            thread.join()
        log.close()
        if self.failure is not None:
            return
        self.checkCensus()

    def runFull(self):
        listing = subprocess.run(["go", "list", "./..."], cwd=self.arguments.tree, capture_output=True, text=True, check=True).stdout.split()
        order = []
        if self.arguments.weights and os.path.exists(self.arguments.weights):
            with open(self.arguments.weights) as handle:
                weighed = [line.split() for line in handle if line.strip()]
            order = [name for _, name in sorted(weighed, key=lambda row: -float(row[0])) if name in listing]
        packages = order + [name for name in listing if name not in order]
        self.result.update({"packages": "all", "package_list": packages})
        try:
            self.result["build_ok"] = self.step("build", ["go", "build", "./..."])
        except BaseException:
            self.fail("build", traceback.format_exc())
        log = open(os.path.join(self.arguments.out, "test.jsonl"), "w")
        wasi = None
        if os.environ.get("WASI_SYSROOT"):
            wasi = {"PATH": os.path.join(os.path.dirname(os.path.dirname(os.environ["WASI_SYSROOT"])), "bin") + os.pathsep + os.environ["PATH"]}
        threads = [self.guarded("vet", self.vet),
                   self.guarded("tests", self.test, "tests", ["go", "test", "-count=1", "-json", "-timeout", "60m", "-p", str(self.arguments.parallel), "-skip", "^TestWASI$"] + packages, log),
                   self.guarded("wasi", self.test, "wasi", ["go", "test", "-count=1", "-json", "-timeout", "60m", "-run", "^TestWASI$", "./internal/native"], log, wasi)]
        for thread in threads:
            thread.start()
        for thread in threads:
            thread.join()
        log.close()
        try:
            self.checkCensus()
        except BaseException:
            self.fail("census", traceback.format_exc())

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
                process = subprocess.run(["node", self.npmCli(), "ci", "--ignore-scripts", "--no-audit", "--no-fund", "--install-strategy=hoisted",
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
                subprocess.run(["cp", "-al", os.path.join(cache, "node_modules"), target], check=True)
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
            subprocess.run(["curl", "-fsSL", pin["url"], "-o", archive], check=True)
            with open(archive, "rb") as handle:
                algorithm, expected = pin["integrity"].split("-", 1)
                if algorithm != "sha512" or base64.b64encode(hashlib.sha512(handle.read()).digest()).decode() != expected:
                    raise ValueError("npm bootstrap integrity mismatch for %s" % pin["url"])
            subprocess.run(["tar", "--no-same-owner", "-xzf", archive, "-C", staging], check=True)
            try:
                os.rename(staging, home)
            except OSError:
                pass  # another gate published the same pinned npm first
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
                self.result["deferred_list_blob"] = git(root, "hash-object", path)
                return deferred
        return {}

    def smokeList(self):
        for root, name in ((self.arguments.tree, "gated tree"), (self.arguments.tools, "tools checkout")):
            path = os.path.join(root, "cloud/fast-gate/smoke.txt")
            if os.path.exists(path):
                with open(path) as handle:
                    fixtures = [line.strip() for line in handle if line.strip() and not line.startswith("#")]
                return fixtures, {"root": root, "from": name}
        return [], {"root": self.arguments.tools, "from": "missing"}

    def touched(self, changed):
        """Each changed file's package: its own directory's, the package owning its testdata, or one
        that embeds it. Anything else (go.mod, docs, a submodule pointer, a file no package reads by
        name) is listed as unowned, for the full gate after landing."""
        listing = subprocess.run(["go", "list", "-f", "{{.Dir}}\t{{.ImportPath}}\t{{join .EmbedFiles \",\"}}\t{{join .TestEmbedFiles \",\"}}\t{{join .XTestEmbedFiles \",\"}}", "./..."],
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
            if owner is None and "/testdata/" in "/" + path:
                owner = directories.get(("/" + path).split("/testdata/")[0].lstrip("/") or ".")
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
        statuses = [line.split("\t") for line in git(tree, "diff", "--name-status", "%s...%s" % (base, sha), "--", "internal/oracle").splitlines() if line]
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
                for line in git(tree, "diff", "-U0", "%s...%s" % (base, sha), "--", path).splitlines():
                    if line.startswith(("+++", "---", "@@")) or not line.startswith(("+", "-")):
                        continue
                    entry = fixtureEntry.match(line[1:])
                    if not entry:
                        return whole("%s changes more than fixture entries" % path)
                    fixtures.add(entry.group(1))
            else:
                return whole("%s is not a fixture, a fixture registration or counts.md" % path)
        registered = set()
        for name in os.listdir(directory):
            if name.endswith("_test.go"):
                with open(os.path.join(directory, name)) as handle:
                    source = handle.read()
                registered.update(re.findall(r'"(internal/oracle/testdata/[^"]+\.a)"', source))
                for function in re.split(r"\nfunc ", source)[1:]:
                    lane = re.match(r"(Test\w+)\(", function)
                    if lane and "range fixtures" in function and lane.group(1) not in oracleLanes:
                        return whole("%s ranges over the fixtures and isn't a known lane" % lane.group(1))
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
        threads = []
        # Every package compiled and listed, and every test process it planned exited 0: counted, so a
        # process that never ran can't pass for one that did.
        tally = {"packages": 0, "planned": 0, "passed": 0}

        def runPackage(importPath):
            binary = os.path.join(binaries, importPath.replace("/", "_") + ".test")
            with slots:
                if self.stream("tests", ["go", "test", "-c", "-o", binary, importPath], None) is None or not os.path.exists(binary):
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
                           "-test.count=1", "-test.failfast", "-test.timeout=30m", "-test.run", patterns.get(name, "^%s$" % name)]
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
            try:
                event = json.loads(line)
            except ValueError:
                continue
            key = (event.get("Package"), event.get("Test"))
            if event.get("Action") == "output":
                output.setdefault(key, []).append(event.get("Output", ""))
                continue
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
        process = self.spawn(["go", "run", "./internal/skipcensus/cmd", "-root", tools, os.path.join(os.path.abspath(self.arguments.out), "test.jsonl")],
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

    def spawn(self, command, stdout, stderr=subprocess.STDOUT, directory=None, environment=None):
        with self.lock:
            if self.failure is not None and not self.arguments.full:
                raise SystemExit(1)
            # Uncached: the oracle's result caches would otherwise answer a test without running it.
            process = subprocess.Popen(command, cwd=directory or self.arguments.tree, stdout=stdout, stderr=stderr, text=True, start_new_session=True,
                                       env=dict(os.environ, **gateEnvironment, **(environment or {})))
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
            for process in self.processes:
                if process.poll() is None:
                    try:
                        os.killpg(process.pid, signal.SIGKILL)
                    except ProcessLookupError:
                        pass

    def status(self, line):
        with open(os.path.join(self.arguments.out, "status.txt"), "w") as handle:
            handle.write(line + "\n")

    def finish(self):
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
        deferred = self.result.get("deferred_to_full_gate", [])
        if not self.arguments.full:
            steps += "; deferred to full gate: %d tests%s" % (len(deferred), (" (" + ", ".join(name.split()[-1] for name in deferred) + ")") if deferred else "")
            steps += "; branch %s, session %s" % (self.arguments.branch or "none", self.arguments.session or "none")
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


def smokePattern(fixtures):
    """A -run pattern for exactly these fixtures. Go splits both the pattern and each subtest name
    on '/' and matches level by level, so alternation can't span a slash: each level gets its own
    alternation, and a name shorter than the pattern ignores the levels past its end."""
    levels = [[smokeTest]]
    for fixture in fixtures:
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
