#!/usr/bin/env python3
"""The fast gate, run on the gate box against a checkout of the candidate.

It builds and vets the whole repository, tests every package the candidate touches (its diff
against the base, by package), runs the oracle smoke set, and checks the skip census over the
run's log. It stops at the first failure: the first failing test or package build is printed
first, every other step is killed, and status.txt goes red naming it. Tests always run with
-count=1; only compiles are reused, from Go's build cache.

Usage: run.py --tree <checkout of the candidate> --sha <candidate> --base <main sha>
              --tools <tools checkout> --out <directory>
"""

import argparse
import json
import os
import signal
import socket
import subprocess
import sys
import threading
import time

module = "github.com/system-inc/adamic"
smokeTest = "TestNativeAgreesWithNode"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--tree", required=True)
    parser.add_argument("--sha", required=True)
    parser.add_argument("--base", required=True)
    parser.add_argument("--tools", required=True)
    parser.add_argument("--out", required=True)
    parser.add_argument("--parallel", type=int, default=os.cpu_count() or 8)
    arguments = parser.parse_args()
    os.makedirs(arguments.out, exist_ok=True)
    gate = Gate(arguments)
    try:
        gate.run()
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
        self.result = {
            "sha": arguments.sha,
            "base": arguments.base,
            "tools_sha": git(arguments.tools, "rev-parse", "HEAD"),
            "machine": {"hostname": socket.gethostname(), "nproc": os.cpu_count()},
            "uncached_tests": True,
            "build_ok": False,
            "vet_ok": False,
        }
        self.status("running: fast gate of %s against %s" % (arguments.sha, arguments.base))

    def run(self):
        tree = self.arguments.tree
        head = git(tree, "rev-parse", "HEAD")
        if head != self.arguments.sha:
            self.fail("setup", "the tree is at %s, not the candidate %s" % (head, self.arguments.sha))
            return
        smoke, smokeSource = self.smokeList()
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
        if not self.step("build", ["go", "build", "./..."]):
            return
        self.result["build_ok"] = True
        log = open(os.path.join(self.arguments.out, "test.jsonl"), "w")
        threads = [threading.Thread(target=self.vet)]
        parallel = str(self.arguments.parallel)
        if packages:
            threads.append(threading.Thread(target=self.test, args=("tests", ["go", "test", "-count=1", "-failfast", "-json", "-p", parallel, "-timeout", "30m"] + packages, log)))
        if smoke and module + "/internal/oracle" not in packages:
            threads.append(threading.Thread(target=self.test, args=("smoke", ["go", "test", "-count=1", "-failfast", "-json", "-timeout", "30m", "-run", smokePattern(smoke), "./internal/oracle"], log)))
        for thread in threads:
            thread.start()
        for thread in threads:
            thread.join()
        log.close()
        if self.failure is not None:
            return
        self.checkCensus()

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
        tree = os.path.realpath(self.arguments.tree)
        for line in listing.stdout.splitlines():
            directory, importPath, *embeds = line.split("\t")
            relative = os.path.relpath(os.path.realpath(directory), tree)
            directories[relative] = importPath
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
        return sorted(packages), unowned

    def step(self, name, command):
        started = time.monotonic()
        with open(os.path.join(self.arguments.out, name + ".log"), "w") as output:
            process = self.spawn(command, output)
            code = process.wait()
        self.steps[name] = round(time.monotonic() - started, 1)
        if code != 0 and self.failure is None:
            with open(os.path.join(self.arguments.out, name + ".log")) as handle:
                self.fail(name, handle.read()[-4000:])
        return code == 0

    def vet(self):
        self.result["vet_ok"] = self.step("vet", ["go", "vet", "./..."])

    def test(self, name, command, log):
        started = time.monotonic()
        stderr = open(os.path.join(self.arguments.out, name + ".stderr"), "w")
        process = self.spawn(command, subprocess.PIPE, stderr)
        output = {}
        for line in process.stdout:
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
        self.steps[name] = round(time.monotonic() - started, 1)
        if code != 0 and self.failure is None:
            with open(os.path.join(self.arguments.out, name + ".stderr")) as handle:
                self.fail(name, "go test exited %d\n%s" % (code, handle.read()[-4000:]))

    def checkCensus(self):
        started = time.monotonic()
        tools = self.arguments.tools
        process = subprocess.run(["go", "run", "./internal/skipcensus/cmd", "-root", tools, os.path.join(os.path.abspath(self.arguments.out), "test.jsonl")],
                                 cwd=tools, capture_output=True, text=True, env=dict(os.environ, GOWORK="off"))
        with open(os.path.join(self.arguments.out, "census.log"), "w") as handle:
            handle.write(process.stdout + process.stderr)
        self.steps["census"] = round(time.monotonic() - started, 1)
        for line in process.stdout.splitlines():
            fields = line.split("\t")
            if len(fields) == 3 and fields[0] == "required-input":
                self.census["required_input"].append(fields[1] + " " + fields[2])
            if len(fields) == 3 and fields[0] in ("unknown", "unclassified"):
                self.census["unclassified"].append(fields[1] + " " + fields[2])
        if process.returncode != 0:
            self.fail("census", (process.stdout + process.stderr)[-4000:])

    def spawn(self, command, stdout, stderr=subprocess.STDOUT):
        with self.lock:
            if self.failure is not None:
                raise SystemExit(1)
            # Uncached: the oracle's result caches would otherwise answer a test without running it.
            process = subprocess.Popen(command, cwd=self.arguments.tree, stdout=stdout, stderr=stderr, text=True, start_new_session=True,
                                       env=dict(os.environ, ADAMIC_GATE_UNCACHED="1"))
            self.processes.append(process)
        return process

    def fail(self, step, detail):
        with self.lock:
            if self.failure is not None:
                return
            self.failure = {"step": step, "detail": detail, "after_seconds": round(time.monotonic() - self.started, 1)}
            print("FIRST FAILURE (%s, at %.1f s):\n%s" % (step, self.failure["after_seconds"], detail), flush=True)
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
        green = self.failure is None
        self.result.update({
            "wall_seconds": wall,
            "steps_seconds": self.steps,
            "pass": self.counts["pass"],
            "fail": self.counts["fail"],
            "skip": self.counts["skip"],
            "required_input_skips": self.census["required_input"],
            "unclassified_skips": self.census["unclassified"],
            "failure": self.failure,
            "finished": True,
        })
        with open(os.path.join(self.arguments.out, "fast.json"), "w") as handle:
            json.dump(self.result, handle, indent=2)
            handle.write("\n")
        steps = " ".join("%s=%.1fs" % item for item in self.steps.items())
        if green:
            self.status("green: %s in %.1f s (%s), %d packages, %d pass, %d skip, smoke %d fixtures" % (self.arguments.sha, wall, steps, len(self.result.get("packages", [])), self.counts["pass"], self.counts["skip"], len(self.result.get("smoke_fixtures", []))))
        else:
            self.status("red: %s at %s after %.1f s (%s)" % (self.arguments.sha, self.failure["step"], self.failure["after_seconds"], steps))
        with open(os.path.join(self.arguments.out, "status.txt")) as handle:
            print(handle.read(), end="", flush=True)


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
