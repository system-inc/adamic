#!/usr/bin/env python3
"""The fast gate fails closed: a stage that dies without reporting is red, never green.

Every process the gate starts goes through subprocess.Popen, so these tests replace it with fake
processes that pass, then make one stage at a time raise what a loaded box raises (too many open
files) or quietly do nothing, and require a red status, a nonzero exit and the stage named. The
all-pass case proves the harness can go green at all. Run: python3 -m unittest cloud/fast-gate/run_test.py
"""

import glob
import io
import re
import shutil
import contextlib
import hashlib
import json
import os
import subprocess
import sys
import tempfile
import types
import signal
import threading
import time
import unittest
from unittest import mock

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import run  # noqa: E402

realRun, realPopen = subprocess.run, subprocess.Popen


def stage(command):
    if command[:2] == ["go", "build"]:
        return "build"
    if command[:2] == ["go", "vet"]:
        return "vet"
    if command[:3] == ["go", "run", "./internal/buildcache/cmd/buildcache-publish"]:
        return "audit" if "-audit" in command else "upload"
    if command[:2] == ["go", "run"]:
        return "census"
    if command[:2] == ["go", "test"] and "-c" not in command:
        return "wasi" if any("^" in part and "TestWASI" in part for part in command) and "-run" in command else "tests"
    return "tests"


def writeUnitNeeds(root, floors=None):
    """root's cloud/fast-gate/unit-needs.json, giving each package of floors its memoryMegabytes: the memory floor
    census reads the gate tools' table on every selection, and a tools checkout without one is a bad table."""
    path = os.path.join(root, "cloud/fast-gate/unit-needs.json")
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as handle:
        json.dump({"version": 1, "units": [{"package": package, "memoryMegabytes": floor, "cpus": 8} for package, floor in (floors or {}).items()]}, handle)


def landingProblems(sha, parts):
    """Fixture of push-main's --also-gate union and fast VERDICT checks (969092f5)."""
    problems = []
    first = dict(parts[0][1])
    for status, other in parts:
        if not status.startswith("green") or other.get("sha") != sha or other.get("finished") is not True:
            problems.append("record isn't green, finished and of the same sha")
    for _, other in parts[1:]:
        for key in ("steps_seconds", "stages_exit"):
            first[key] = dict(first.get(key) or {}, **(other.get(key) or {}))
        first["planned_stages"] = sorted(set(first.get("planned_stages") or []) | set(other.get("planned_stages") or []))
        first["fail"] = (first.get("fail") or 0) + (other.get("fail") or 0)
        for key in ("build_ok", "vet_ok", "uncached_tests"):
            first[key] = bool(first.get(key) or other.get(key))
        for key in ("unclassified_skips", "required_input_skips"):
            first[key] = (first.get(key) or []) + (other.get(key) or [])
    fast, status = first, parts[0][0]
    if fast.get("sha") != sha:
        problems.append("it gated %s, not %s" % (fast.get("sha"), sha))
    if not status.startswith("green"):
        problems.append("its status is %r" % status)
    if fast.get("fail") != 0:
        problems.append("%s failures" % fast.get("fail"))
    if not fast.get("build_ok") or not fast.get("vet_ok"):
        problems.append("go build or go vet failed")
    for kind in ("unclassified_skips", "required_input_skips"):
        if fast.get(kind):
            problems.append("%s: %s" % (kind.replace("_", " "), " ".join(fast[kind])))
    if fast.get("uncached_tests") is not True:
        problems.append("its tests weren't run uncached")
    planned = ("build", "vet", "tests", "census")
    ran = fast.get("steps_seconds") or {}
    missing = [stage for stage in planned if stage not in ran]
    if missing:
        problems.append("stages with no completion recorded: %s" % " ".join(missing))
    exits = fast.get("stages_exit") or fast.get("exit_codes") or {}
    nonzero = ["%s=%s" % (stage, code) for stage, code in exits.items() if code != 0]
    if nonzero:
        problems.append("stages that exited nonzero: %s" % " ".join(nonzero))
    unrecorded = [stage for stage in fast.get("planned_stages") or [] if exits.get(stage) != 0]
    if unrecorded:
        problems.append("planned stages without a recorded exit of 0: %s" % " ".join(unrecorded))
    if fast.get("finished") is not True:
        problems.append("the gate didn't record that it finished")
    if fast.get("scoped_env"):
        problems.append("it ran scoped (%s set)" % " ".join(fast["scoped_env"]))
    return problems


class FakeProcess:
    def __init__(self, command, stdout):
        self.pid = 0
        self.returncode = 0
        lines = []
        if "-c" in command and "-o" in command:
            with open(command[command.index("-o") + 1], "w") as handle:
                handle.write("binary")
        elif "-test.list" in command:
            lines = ["TestOne\n"]
        elif "test2json" in command or command[:2] == ["go", "test"]:
            name = "TestOne"
            if "-run" in command and "TestWASI" in command[command.index("-run") + 1]:
                name = "TestWASI/" + ("requests" if "requests" in command[command.index("-run") + 1] else "a.a")
            package = command[command.index("-p") + 1] if "test2json" in command and "-p" in command else "p"
            lines = [json.dumps({"Action": "pass", "Package": package, "Test": name}) + "\n"]
        self.lines = lines
        self.stdout = io.StringIO("".join(lines)) if stdout == subprocess.PIPE else None

    def wait(self):
        return 0

    def poll(self):
        return 0

    def communicate(self):
        return "".join(self.lines), ""


class MainRunsLast(unittest.TestCase):
    """Run as a script, unittest.main() runs only the classes defined above it, so a main left mid-file skips every class
    below it in silence: until Nov 15, 2026 eleven classes (WASIUnits through CensusMemoryFloors) never ran under
    python3 cloud/fast-gate/run_test.py. This class sits above every other test class, so it runs wherever main stands."""

    def test_no_test_class_is_defined_after_unittest_main(self):
        import ast
        with open(os.path.abspath(__file__)) as handle:
            module = ast.parse(handle.read())
        mains = [index for index, node in enumerate(module.body)
                 if isinstance(node, ast.If) and "__main__" in ast.unparse(node.test)]
        self.assertEqual(len(mains), 1, "run_test.py has exactly one if __name__ == \"__main__\" block")
        after = [node.name for node in module.body[mains[0] + 1:] if isinstance(node, ast.ClassDef)]
        self.assertEqual(after, [], "these classes are defined after unittest.main(), so a script run never runs them")


class FailClosed(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.mkdtemp()
        self.tree = os.path.join(self.directory, "tree")
        os.makedirs(os.path.join(self.tree, "cloud/fast-gate"))
        with open(os.path.join(self.tree, "cloud/fast-gate/executors.txt"), "w") as handle:
            handle.write("inert *.md\nstage3 stage3/*\n")
        with open(os.path.join(self.tree, "cloud/fast-gate/tools.txt"), "w") as handle:
            handle.write("# declared tools\ngo\tall\tcommand -v go\n")
        for command in (["init", "-q"], ["add", "."], ["-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "t"]):
            realRun(["git", "-C", self.tree] + command, check=True)
        self.sha = run.git(self.tree, "rev-parse", "HEAD")
        os.makedirs(os.path.join(self.tree, "internal/native"), exist_ok=True)
        with open(os.path.join(self.tree, "internal/native/wasm_test.go"), "w") as handle:
            handle.write('func TestWASI(t *testing.T) {\n\tfixtures := []string{\n\t\t"a.a",\n\t}\n\tt.Run(fixture, f)\n\tt.Run("requests", f)\n}\n')

    def gate(self, full=False, broken=None, silent=None, unowned=(), base=None, failing=None, runToEnd=False, extra=None):
        out = tempfile.mkdtemp(dir=self.directory)
        arguments = mock.Mock(tree=self.tree, sha=self.sha, base=base or self.sha, tools=self.tree, out=out, parallel=4, full=full,
                              branch="", branch_source="", session="", session_source="", weights=None, run_to_end=runToEnd, **(extra or {}))

        def popen(command, **options):
            if command[0] == "git":
                return realPopen(command, **options)
            if stage(command) == broken:
                raise OSError(24, "Too many open files")
            process = FakeProcess(command, options.get("stdout"))
            if stage(command) == silent:
                process.returncode = 0
            return process

        listing = mock.Mock(stdout="example.com/p\n")
        with mock.patch.object(run.subprocess, "Popen", side_effect=popen), \
                mock.patch.object(run.Gate, "touched", lambda gate, changed: (gate.packageDirectories.update({"p": self.tree, run.module + "/stage1/cohere/tsprinter": self.tree}) or ["p"], list(unowned))), \
                mock.patch.object(run.Gate, "command", side_effect=lambda command, **options: (mock.Mock(stdout="p\t" + self.tree + "\n") if "-f" in command else listing) if command[:2] == ["go", "list"] else realRun(command, **options)), \
                mock.patch.object(run.Gate, silent, lambda *arguments: None) if silent else contextlib.nullcontext(), \
                mock.patch.object(run.Gate, failing, lambda gate, *arguments: gate.fail(failing, "planted failure")) if failing else mock.patch.object(run, "longTestSeconds", run.longTestSeconds), \
                mock.patch.object(run.Gate, "npmCli", lambda gate: "npm-cli.js"), \
                mock.patch.object(sys, "argv", ["run.py"]), mock.patch("builtins.print"):
            gate = run.Gate(arguments)
            gate.packageDirectories = {}
            try:
                gate.run()
            except BaseException as exception:  # main() does exactly this
                gate.fail("runner", repr(exception))
            finally:
                gate.finish()
        with open(os.path.join(out, "status.txt")) as handle:
            status = handle.read()
        with open(os.path.join(out, ("full" if full else "fast") + ".json")) as handle:
            result = json.load(handle)
        return gate, status, result

    def test_an_undeclared_tool_is_red_at_tools_naming_it(self):
        with open(os.path.join(self.tree, "probe.go"), "w") as handle:
            handle.write('package probe\n\nimport "os/exec"\n\nvar a = exec.Command("go", "version")\nvar b, _ = exec.LookPath("wasmtime")\n')
        for command in (["add", "."], ["-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "probe"]):
            realRun(["git", "-C", self.tree] + command, check=True)
        self.sha = run.git(self.tree, "rev-parse", "HEAD")
        gate, status, result = self.gate()
        self.assertIn("first failure at tools", status)
        self.assertEqual(result["undeclared_tools"], ["probe.go:6 runs wasmtime, which cloud/fast-gate/tools.txt doesn't declare"])

    def test_all_pass_is_green(self):
        for full in (False, True):
            gate, status, result = self.gate(full=full)
            self.assertTrue(status.startswith("green:"), status)
            self.assertIsNone(gate.failure)
            self.assertEqual(result["gate_kind"], "full" if full else "fast")
            if full:
                # One record per kind: a whole gate writes full.json alone.
                self.assertFalse(os.path.exists(os.path.join(gate.arguments.out, "fast.json")))
                self.assertEqual(result["wasi_unit_count"], 2)
                self.assertEqual([row["name"] for row in result["wasi_units"]], ["a.a", "requests"])
            self.assertEqual(set(result["stages_exit"]), set(result["planned_stages"]))
            # Every verdict carries its long-test ledger, empty or not.
            self.assertEqual((result["long_tests"], result["long_test_threshold_seconds"]), (0, run.longTestSeconds))
            self.assertIn("0 units over %d s" % (run.longTestSeconds), status)
            self.assertEqual(result["budget_over"], [])

    def test_every_stage_raising_is_red(self):
        for full, stages in ((False, ["build", "vet", "tests", "census"]), (True, ["build", "vet", "tests", "wasi", "census"])):
            for broken in stages:
                with self.subTest(full=full, stage=broken):
                    gate, status, result = self.gate(full=full, broken=broken)
                    self.assertTrue(status.startswith("red:"), status)
                    self.assertIsNotNone(gate.failure)
                    self.assertNotEqual(result["stages_exit"].get(broken), 0)

    def test_a_whole_gate_never_runs_more_threads_than_the_box_has_cores(self):
        gate, status, result = self.gate(full=True)
        shape = result["parallelism"]
        self.assertLessEqual(shape["packages_at_once"] * shape["gomaxprocs_each"], max(shape["cpus"], 2 * shape["packages_at_once"]))
        with mock.patch.dict(os.environ, {"ADAMIC_FULL_GATE_PACKAGES": "2"}):
            gate, status, result = self.gate(full=True)
        self.assertEqual(result["parallelism"]["packages_at_once"], 2)

    def test_a_whole_gate_cancels_itself_at_its_first_failure(self):
        # Kirk, Oct 8: fail fast and loud, and free the box. The record is red and says it was cancelled.
        gate, status, result = self.gate(full=True, failing="vet")
        self.assertTrue(status.startswith("red:"), status)
        self.assertIn("first failure at vet", status)
        self.assertIn("cancelled after first failure", status)
        self.assertTrue(result["cancelled_after_first_failure"])
        self.assertTrue(result["finished"])
        # Nothing after the failure ran: the census never reported.
        self.assertNotEqual(result["stages_exit"].get("census"), 0)
        # A parity proof runs to the end: red, not cancelled, and its census still ran.
        gate, status, result = self.gate(full=True, failing="vet", runToEnd=True)
        self.assertTrue(status.startswith("red:"), status)
        self.assertNotIn("cancelled", status)
        self.assertFalse(result["cancelled_after_first_failure"])
        self.assertEqual(result["stages_exit"].get("census"), 0)
        # A fast gate is not a whole gate: it is never marked cancelled.
        gate, status, result = self.gate(full=False, failing="vet")
        self.assertNotIn("cancelled", status)

    def test_the_record_keeps_the_box_load_and_a_red_above_its_cores_says_under_load(self):
        # The witness (Oct 9): six of Oct 8's reds were load, each diagnosed by hand. The record names it itself.
        for full, runToEnd in ((False, False), (True, False), (True, True)):
            with self.subTest(full=full, runToEnd=runToEnd):
                lines = []
                realStatus = run.Gate.status
                with mock.patch.object(run, "boxLoad", return_value={"load_1m": 178.0, "cores": 64}), \
                        mock.patch.object(run.Gate, "status", lambda gate, line: lines.append(line) or realStatus(gate, line)):
                    gate, status, result = self.gate(full=full, failing="vet", runToEnd=runToEnd)
                self.assertTrue(status.startswith("red:"), status)
                self.assertIn("under load (load 178.0 on 64 cores at first failure)", status)
                self.assertTrue(gate.failure["under_load"])
                self.assertEqual(result["box_load"], {moment: {"load_1m": 178.0, "cores": 64} for moment in ("start", "first_failure", "end")})
                # A whole gate's red line goes out at the first failure, before the run ends, and carries it too.
                if full:
                    self.assertIn("under load", next(line for line in lines if "first failure at vet" in line))
                    with open(os.path.join(gate.arguments.out, "first-failure.txt")) as handle:
                        self.assertIn("box load 178.0 on 64 cores at first failure", handle.read())
        # At or under the core count it is a red like any other, the load still on the record.
        with mock.patch.object(run, "boxLoad", return_value={"load_1m": 64.0, "cores": 64}):
            gate, status, result = self.gate(failing="vet")
        self.assertNotIn("under load", status)
        self.assertFalse(gate.failure["under_load"])
        self.assertIn("box load 64.0 on 64 cores at start, load 64.0 on 64 cores at first failure, load 64.0 on 64 cores at end", status)
        # A green under load is green; its record still carries the readings, with no first failure.
        with mock.patch.object(run, "boxLoad", return_value={"load_1m": 178.0, "cores": 64}):
            gate, status, result = self.gate()
        self.assertTrue(status.startswith("green:"), status)
        self.assertNotIn("under load", status)
        self.assertEqual(sorted(result["box_load"]), ["end", "start"])

    def test_select_writes_the_fast_gate_s_selection_and_runs_nothing(self):
        # Side work's fast gates on Loom's pool (#xt96xyp): the same selection, run where the tree is whole.
        out = tempfile.mkdtemp(dir=self.directory)
        arguments = mock.Mock(tree=self.tree, sha=self.sha, base=self.sha, tools=self.tree, out=out, parallel=4, full=False, select=True,
                              branch="", branch_source="", session="", session_source="", weights=None, run_to_end=False)
        launched = []
        with mock.patch.object(run.subprocess, "Popen", side_effect=lambda command, **options: launched.append(command) or realPopen(command, **options)), \
                mock.patch.object(run.Gate, "touched", lambda gate, changed: (gate.packageDirectories.update({"example.com/p": self.tree}) or ["example.com/p"], [])), \
                mock.patch.object(run.Gate, "npmCli", lambda gate: self.fail("select installs nothing")), \
                mock.patch("builtins.print"):
            gate = run.Gate(arguments)
            gate.packageDirectories = {}
            gate.run()
        self.assertIsNone(gate.failure)
        with open(os.path.join(out, "select.json")) as handle:
            selection = json.load(handle)
        self.assertEqual(selection["packages"], ["example.com/p"])
        self.assertEqual((selection["sha"], selection["base"]), (self.sha, self.sha))
        # The tests' environment is the gate's own: uncached, with the sampling the landing gate uses.
        self.assertEqual(selection["env"]["ADAMIC_GATE_UNCACHED"], "1")
        self.assertEqual(selection["env"]["ADAMIC_GATE_SAMPLE"], self.sha)
        # Nothing beyond the Go tests: the smoke phase retired (#r8ghgkt), and this change maps to no executor.
        self.assertEqual(selection["executors_beyond_go_tests"], [])
        # Nothing but git ran: no build, vet or test.
        self.assertEqual([command[0] for command in launched if command[0] != "git"], [])

    def test_a_phase_of_the_whole_gate_runs_alone_as_a_pool_unit(self):
        # Loom's pool runs the whole gate as units (Oct 9): one phase, one unit of it, with this gate's own record.
        gate, status, result = self.gate(full=True, extra={"phase": "wasi", "unit": "requests"})
        self.assertTrue(status.startswith("green:"), status)
        self.assertEqual(result["planned_stages"], ["wasi"])
        self.assertEqual(set(result["stages_exit"]), {"wasi"})
        self.assertEqual([row["name"] for row in result["wasi_units"]], ["requests"])
        self.assertEqual(result["unit"], "requests")
        # A unit the phase doesn't have is red, never an empty green.
        gate, status, result = self.gate(full=True, extra={"phase": "wasi", "unit": "no-such-fixture"})
        self.assertIn("first failure at wasi", status)
        self.assertIn("has no unit 'no-such-fixture'", gate.failure["detail"])
        # A phase runs nothing of the others: no tests, no build.
        gate, status, result = self.gate(full=True, extra={"phase": "tools"})
        self.assertTrue(status.startswith("green:"), status)
        self.assertEqual(result["planned_stages"], ["tools"])
        self.assertNotIn("tests", result["stages_exit"])
        self.assertNotIn("build", result["stages_exit"])

    def poolLog(self):
        path = os.path.join(self.directory, "pool.jsonl")
        with open(path, "w") as handle:
            handle.write(json.dumps({"Action": "pass", "Package": "p", "Test": "TestPool"}) + "\n")
        return path

    def poolPart(self):
        return ("green: pool Go tests", {"sha": self.sha, "base": self.sha, "finished": True, "fail": 0,
                                        "uncached_tests": True, "planned_stages": ["tests"],
                                        "steps_seconds": {"tests": 1}, "stages_exit": {"tests": 0}})

    def test_fast_phases_plan_only_themselves(self):
        for phase in ("coverage", "tools", "build", "vet", "darwin-vet", "census", "catalog"):
            with self.subTest(phase=phase):
                extra = {"phase": phase}
                if phase == "census":
                    extra["census"] = self.poolLog()
                if phase == "darwin-vet":
                    # The darwin vet reads the compiler map for the compilers the leg builds, and has none here.
                    os.makedirs(os.path.join(self.tree, os.path.dirname(run.compilerDependenciesPath)), exist_ok=True)
                    with open(os.path.join(self.tree, run.compilerDependenciesPath), "w") as handle:
                        handle.write(json.dumps({"version": 1, "packages": {}}))
                gate, status, result = self.gate(extra=extra)
                self.assertTrue(status.startswith("green:"), status)
                self.assertEqual(result["planned_stages"], [phase])
                self.assertEqual(result["stages_exit"], {phase: 0})
                self.assertEqual(set(result["steps_seconds"]), {phase})
                self.assertEqual(result["build_ok"], phase == "build")
                self.assertEqual(result["vet_ok"], phase == "vet")
                self.assertTrue(result["finished"])

    def test_conditional_fast_phases_dispatch_only_their_selected_work(self):
        with open(os.path.join(self.tree, "cloud/fast-gate/executors.txt"), "a") as handle:
            handle.write("workers workers/*\na-check probes/*\n")
        for phase, path, method in (("workers", "workers/probe.mjs", "workers"),
                                    ("a-check", "probes/probe.a", "aCheck"),
                                    ("catalog-apply", "probe.md", "catalogApply")):
            calls = []
            def execute(gate, inputs):
                calls.append(inputs)
                gate.steps[phase] = 0.1
                gate.exits[phase] = 0
            with self.subTest(phase=phase), mock.patch.object(run.Gate, method, execute), \
                    mock.patch.object(run.Gate, "catalogEntriesTouched", return_value=["01.patch"]):
                gate, status, result = self.gate(unowned=[path], extra={"phase": phase})
            self.assertTrue(status.startswith("green:"), status)
            self.assertEqual(result["planned_stages"], [phase])
            self.assertEqual(result["stages_exit"], {phase: 0})
            self.assertEqual(set(result["steps_seconds"]), {phase})
            self.assertEqual(len(calls), 1)
            self.assertIn("workers" if phase == "workers" else path if phase == "a-check" else "01.patch", calls[0])

    def test_fast_inventory_uses_the_base_diff_for_conditional_phases(self):
        base = self.sha
        for path in ("stage3/probe.md", "workers/probe.mjs", "probes/probe.a"):
            os.makedirs(os.path.dirname(os.path.join(self.tree, path)), exist_ok=True)
            with open(os.path.join(self.tree, path), "w") as handle:
                handle.write("probe\n")
        with open(os.path.join(self.tree, "cloud/fast-gate/executors.txt"), "a") as handle:
            handle.write("workers workers/*\na-check probes/*\ninert cloud/*\n")
        for command in (["add", "."], ["-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "conditional"]):
            realRun(["git", "-C", self.tree] + command, check=True)
        self.sha = run.git(self.tree, "rev-parse", "HEAD")
        with mock.patch.object(run.Gate, "touched", lambda gate, paths: ([], paths)), \
                mock.patch.object(run.Gate, "npmCli", side_effect=AssertionError("listing installs nothing")):
            self.assertEqual(run.fastUnits(self.tree, base, self.sha), run.fastBasePhases + ["workers", "a-check"])
            self.assertEqual(run.fastUnits(self.tree, self.sha, self.sha), run.fastBasePhases)
        gate, status, result = self.gate(base=base, extra={"phase": "coverage"})
        self.assertTrue(status.startswith("green:"), status)
        self.assertIn("workers/probe.mjs", result["changed_files"])
        self.assertEqual(result["planned_stages"], ["coverage"])

    def test_unknown_and_whole_only_fast_phases_are_red(self):
        # smoke retired (#r8ghgkt): a plan that still names it is refused, never run.
        for phase in ("no-such-phase", "smoke", "wasi", "tests", "workers", "a-check", "catalog-apply", "darwin"):
            with self.subTest(phase=phase):
                gate, status, result = self.gate(extra={"phase": phase})
                self.assertTrue(status.startswith("red:"), status)
                self.assertEqual(result["planned_stages"], [phase])
                self.assertIn(phase, result["failure"]["step"])

    def test_fast_census_requires_a_log_and_checks_requested_tests(self):
        gate, status, result = self.gate(extra={"phase": "census"})
        self.assertIn("first failure at census", status)
        self.assertIn("no test.jsonl", result["failure"]["detail"])
        with mock.patch.object(run.Gate, "runRequested", lambda gate: setattr(gate, "requested", {"p": {"TestMissing"}})):
            gate, status, result = self.gate(extra={"phase": "census", "census": self.poolLog()})
        self.assertTrue(status.startswith("red:"), status)
        self.assertEqual(result["deferred_run_results"], {"p TestMissing": "missing"})
        # The red line names what didn't run, not just "first failure at deferred" (trio 114a6439, Oct 9 10:50Z).
        self.assertIn("first failure at deferred (1 requested deferred tests unproven: p TestMissing missing)", status)
        # The skip census still ran and recorded its verdict (integration, Oct 9 12:14Z: a deferred red skipped it).
        self.assertEqual(result["stages_exit"].get("census"), 0)
        self.assertTrue(os.path.exists(os.path.join(gate.arguments.out, "census.log")))

    def test_a_package_gated_for_requested_tests_runs_their_family_and_reds_when_it_selects_none(self):
        # A requested name selects itself and its split's shards; one that selects nothing is red, never an empty pass
        # (@system_adamic, Oct 9 11:01Z, after Loom's unit ran "no tests to run" and passed).
        with mock.patch.object(run.Gate, "runRequested", lambda gate: setattr(gate, "requested", {"q": {"TestOne"}})):
            gate, status, result = self.gate()
        self.assertIsNone(gate.failure, status)
        self.assertEqual(result["split_tests"].get("q"), 1)
        with mock.patch.object(run.Gate, "runRequested", lambda gate: setattr(gate, "requested", {"q": {"TestNoSuchFamily"}})):
            gate, status, result = self.gate()
        self.assertIn("first failure at requested", status)
        self.assertIn("requested tests ran: 0 in q: TestNoSuchFamily", gate.failure["detail"])

    def test_several_fast_phases_complete_the_pool_landing_record(self):
        gate, status, result = self.gate(extra={"phases": "build,vet,census", "census": self.poolLog()})
        self.assertTrue(status.startswith("green:"), status)
        self.assertEqual(result["planned_stages"], ["build", "vet", "census"])
        self.assertEqual(result["stages_exit"], {phase: 0 for phase in result["planned_stages"]})
        self.assertEqual(landingProblems(self.sha, [self.poolPart(), (status, result)]), [])
        with open(os.path.join(gate.arguments.out, "test.jsonl")) as handle:
            log = handle.read()
        self.assertIn("TestPool", log)  # the census read the pool's log

    def test_a_missing_stage_or_red_vet_cannot_complete_the_pool_record(self):
        gate, status, result = self.gate(extra={"phases": "build,census", "census": self.poolLog()})
        self.assertTrue(status.startswith("green:"), status)
        self.assertIn("stages with no completion recorded: vet", landingProblems(self.sha, [self.poolPart(), (status, result)]))
        gate, status, result = self.gate(failing="vet", extra={"phases": "build,vet,census", "census": self.poolLog()})
        self.assertIn("first failure at vet", status)
        self.assertFalse(result["vet_ok"])
        self.assertTrue(landingProblems(self.sha, [self.poolPart(), (status, result)]))
        gate, status, result = self.gate(extra={"phases": "build,vet,census", "census": self.poolLog()})
        result["sha"] = "b" * 40
        self.assertTrue(landingProblems(self.sha, [self.poolPart(), (status, result)]))

    def test_a_silent_fast_phase_is_red_even_in_a_combined_record(self):
        for method, phase in (("build", "build"), ("vet", "vet"), ("checkCensus", "census")):
            with self.subTest(phase=phase):
                gate, status, result = self.gate(silent=method, extra={"phases": "build,vet,census", "census": self.poolLog()})
                self.assertTrue(status.startswith("red:"), status)
                self.assertIn(phase, result["failure"]["step"])

    def test_the_census_runs_over_a_merged_record(self):
        # The pool merges its units' test.jsonl; the gate's census and deferred checks run over it, nothing else.
        merged = os.path.join(self.directory, "merged.jsonl")
        with open(merged, "w") as handle:
            handle.write(json.dumps({"Action": "pass", "Package": run.module + "/p", "Test": "TestOne"}) + "\n")
        gate, status, result = self.gate(full=True, extra={"census": merged})
        self.assertTrue(status.startswith("green:"), status)
        self.assertEqual(result["planned_stages"], ["census"])
        with open(os.path.join(gate.arguments.out, "test.jsonl")) as handle:
            self.assertIn("TestOne", handle.read())

    def test_a_stage_that_reports_nothing_is_red(self):
        for silent, stageName in (("build", "build"), ("vet", "vet"), ("testSplit", "tests"), ("checkCensus", "census")):
            with self.subTest(stage=stageName):
                gate, status, result = self.gate(silent=silent)
                self.assertTrue(status.startswith("red:"), status)
                self.assertIn(stageName, status)


class BuildStore(unittest.TestCase):
    """A gate's store root, beside its tree (#tfp4gtm): a run removes its own when it finishes, and a listing makes none,
    since it runs no stage. Each plan's listing left one beside Eve's planner checkout, 2,299 of them (27 GB, Nov 21).
    Mutants, each red here: finish() leaving the root; a listing making one."""
    gate = FailClosed.gate

    def setUp(self):
        FailClosed.setUp(self)
        self.addCleanup(shutil.rmtree, self.directory)

    def roots(self):
        return sorted(glob.glob(os.path.join(os.path.dirname(os.path.abspath(self.tree)), "build-store-*")))

    def test_a_gate_makes_its_store_root_and_removes_it_when_it_finishes(self):
        with mock.patch.object(sys, "argv", ["run.py"]):
            gate = run.Gate(mock.Mock(tree=self.tree, sha=self.sha, base=self.sha, tools=self.tree, out=tempfile.mkdtemp(dir=self.directory),
                                      parallel=4, full=False, branch="", branch_source="", session="", session_source="", weights=None))
        self.assertEqual(self.roots(), [gate.buildStoreRoot])
        self.assertTrue(os.path.isdir(gate.buildStoreEnvironment["ADAMIC_BUILD_STORE_SPOOL"]))
        gate.removeBuildStore()
        self.assertEqual(self.roots(), [])
        self.assertNotIn("build_store_unremoved", gate.result)

    def test_a_finished_run_leaves_no_store_root_beside_its_tree(self):
        gate, status, result = self.gate()
        self.assertTrue(status.startswith("green:"), status)
        self.assertEqual(self.roots(), [])

    def test_a_listing_makes_no_store_root(self):
        with mock.patch.object(run.Gate, "fastPhaseInputs", lambda gate, recordCoverage=False: (set(), set())), \
                mock.patch.object(run.Gate, "catalogEntriesTouched", lambda gate, changed: []):
            units = run.fastUnits(self.tree, base=self.sha, sha=self.sha, tools=self.tree)
        self.assertEqual(units, run.fastBasePhases)
        self.assertEqual(self.roots(), [])


class Products(unittest.TestCase):
    setUp = FailClosed.setUp
    gate = FailClosed.gate

    def probe(self, names=("TestProduct_A", "TestX"), failing=False, complete=False, sourceNames=None, full=False, extra=None, runToEnd=False):
        commands, order = [], []
        with open(os.path.join(self.tree, "product_test.go"), "w") as source:
            source.write("\n".join("func %s(t *testing.T) {}" % name for name in (names if sourceNames is None else sourceNames) if name.startswith("TestProduct_")))
        original = FakeProcess.__init__
        def product(process, command, stdout):
            original(process, command, stdout)
            if "-test.list" in command:
                process.lines = [name + "\n" for name in names]
            elif "test2json" in command or (full and command[:2] == ["go", "test"] and "-run" not in command and "-skip" in command and "^TestProduct_" in command[command.index("-skip") + 1]):
                commands.append(command)
                name = command[command.index("-test.run") + 1].strip("^$") if "-test.run" in command else "TestX"
                order.append(name)
                if name == "TestX" and "-test.skip" not in command and "-skip" not in command and "TestProduct_A" in names:
                    order.append("TestProduct_A")
                action = "fail" if failing and name.startswith("TestProduct_") else "pass"
                process.returncode = int(action == "fail")
                process.lines = [json.dumps({"Action": action, "Package": "p", "Test": name, "Elapsed": 0.01}) + "\n"]
                process.wait = lambda: process.returncode
            process.stdout = io.StringIO("".join(process.lines)) if stdout == subprocess.PIPE else None
        with mock.patch.object(FakeProcess, "__init__", product):
            gate, status, result = self.gate(full=full, extra=dict(extra or {}, complete=complete), runToEnd=runToEnd)
        return gate, status, result, commands, order

    def test_products_before_tests_and_skip(self):
        gate, status, result, commands, order = self.probe()
        self.assertEqual(order, ["TestProduct_A", "TestX"])
        self.assertTrue(status.startswith("green:"), status)
        self.assertLess(result["planned_stages"].index("products"), result["planned_stages"].index("tests"))
        self.assertEqual(result["stages_exit"]["products"], 0)
        self.assertIn("products", result["steps_seconds"])
        self.assertEqual(commands[0][commands[0].index("-test.run") + 1], "^TestProduct_A$")
        # The binary's own timeout matches the product ceiling: a 90 s -test.timeout panicked products at 90 s anyway
        # (Oct 9 07:36Z, typeaware TestProduct_corpus_record on gocacheprog 41ff2dbe).
        self.assertIn("-test.timeout=%ds" % run.productKillSeconds, commands[0])
        self.assertEqual(commands[1][commands[1].index("-test.skip") + 1], "^TestProduct_")
        self.assertTrue(result["product_units"][0]["product"])
        unit = next(row for row in result["units"] if row["test"] == "TestProduct_A")
        self.assertTrue(unit["product"])
        self.assertGreater(unit["seconds"], 0)

    def test_failed_product_blocks_tests_even_in_complete_mode(self):
        for complete in (False, True):
            gate, status, result, commands, order = self.probe(failing=True, complete=complete)
            self.assertEqual(result["failure"]["step"], "products")
            self.assertIn("p TestProduct_A", result["failure"]["detail"])
            self.assertEqual(result["stages_exit"]["products"], 1)
            self.assertEqual(order, ["TestProduct_A"])
            self.assertEqual(result["test_outcomes"][0]["status"], "not run")

    def test_other_packages_wait_for_products_while_builds_continue(self):
        out = tempfile.mkdtemp(dir=self.directory)
        p, q = os.path.join(self.tree, "p"), os.path.join(self.tree, "q")
        os.makedirs(p)
        os.makedirs(q)
        with open(os.path.join(q, "product_test.go"), "w") as source:
            source.write("func TestProduct_A(t *testing.T) {}\n")
        arguments = types.SimpleNamespace(tree=self.tree, tools=self.tree, out=out, sha=self.sha,
            base=self.sha, parallel=4, full=False, complete=False, branch="", branch_source="", session="", session_source="")
        gate = run.Gate(arguments)
        gate.packageDirectories, gate.deferred = {"p": p, "q": q}, {}
        productStarted, releaseProduct, pListed, testLaunched = [threading.Event() for _ in range(4)]
        order, shapes = [], []
        def stream(stage, command, log, *args):
            if "-c" in command:
                if command[-1] == "p":
                    if not productStarted.wait(2):
                        gate.fail("tests", "p's binary did not overlap q's product")
                        return 1
                with open(command[command.index("-o") + 1], "w") as binary:
                    binary.write("binary")
                return 0
            if stage == "products":
                shapes.append(args[-1])
                productStarted.set()
                if not releaseProduct.wait(2):
                    return 1
                order.append("product passed")
            else:
                with gate.lock:
                    gate.recordTestStart(command)
                order.append("TestX")
                testLaunched.set()
            return 0
        def capture(command, directory):
            if directory == p:
                pListed.set()
                return "TestX\n"
            return "TestProduct_A\n"
        gate.stream, gate.capture = stream, capture
        record = run.TestSeconds(os.path.join(out, "record.tsv"))
        with mock.patch.object(run, "TestSeconds", return_value=record), mock.patch.object(run, "slotCPUs", return_value=8):
            thread = gate.guarded("tests", gate.testSplit, ["p", "q"], io.StringIO())
            thread.start()
            try:
                listedDuringProduct = pListed.wait(2)
                premature = testLaunched.wait(0.1)
            finally:
                releaseProduct.set()
                thread.join(3)
        self.assertFalse(thread.is_alive(), "test workers stuck at products barrier")
        self.assertTrue(listedDuringProduct, "other binaries stopped compiling during products")
        self.assertFalse(premature, "TestX launched before q's product passed")
        self.assertEqual(order, ["product passed", "TestX"])
        self.assertEqual(shapes, [{"GOMAXPROCS": "4"}])
        self.assertEqual(gate.exits, {"products": 0, "tests": 0})

    def test_scan_list_mismatch_is_red_at_products(self):
        for source, listed in ((["TestProduct_A"], ["TestProduct_B"]),
                               (["TestProduct_A"], []), ([], ["TestProduct_A"])):
            with self.subTest(source=source, listed=listed):
                _, status, result, _, order = self.probe(names=tuple(listed + ["TestX"]), sourceNames=source)
                self.assertEqual(result["failure"]["step"], "products")
                self.assertIn("p product scan/list mismatch", result["failure"]["detail"])
                self.assertIn("source=%s listed=%s" % (source, listed), result["failure"]["detail"])
                self.assertEqual(result["stages_exit"]["products"], 1)
                self.assertEqual(order, [])

    def test_whole_gate_products_before_tests_and_skip(self):
        _, status, result, commands, order = self.probe(full=True)
        self.assertTrue(status.startswith("green:"), status)
        self.assertEqual(order, ["TestProduct_A", "TestX"])
        self.assertEqual(result["stages_exit"]["products"], 0)
        self.assertIn("^TestProduct_", commands[1][commands[1].index("-skip") + 1])

    def test_whole_gate_phases_build_products_only_in_the_products_phase(self):
        for extra in ({"phases": "products,wasi"}, {"phase": "products"}):
            _, status, result, _, order = self.probe(full=True, extra=extra)
            self.assertTrue(status.startswith("green:"), status)
            self.assertEqual(order, ["TestProduct_A"])
            self.assertEqual(result["stages_exit"]["products"], 0)
            self.assertEqual(result["planned_stages"].count("products"), 1)
        # Every other phase takes the products as given: Loom ran them as their own units first.
        for extra in ({"phase": "wasi"}, {"phases": "vet,wasi"}):
            _, status, result, _, order = self.probe(full=True, extra=extra)
            self.assertTrue(status.startswith("green:"), status)
            self.assertEqual(order, [])

    def test_whole_gate_failed_product_blocks_test_units(self):
        _, status, result, _, order = self.probe(full=True, failing=True)
        self.assertEqual(result["failure"]["step"], "products")
        self.assertEqual(order, ["TestProduct_A"])
        self.assertNotIn("tests", result["steps_seconds"])

    def test_whole_gate_run_to_end_runs_tests_past_a_failed_product(self):
        # Oct 9 07:23Z: library's area parity run stopped at TestProduct_suite and gave no red list.
        _, status, result, _, order = self.probe(full=True, failing=True, runToEnd=True)
        self.assertEqual(result["failure"]["step"], "products")
        self.assertEqual(result["stages_exit"]["products"], 1)
        self.assertIn("TestX", order)

    def test_wall_deadline_kills_and_fails_in_complete_mode(self):
        for phase in ("products",):
            with self.subTest(phase=phase), tempfile.TemporaryDirectory() as out:
                arguments = types.SimpleNamespace(tree=self.tree, tools=self.tree, out=out, sha=self.sha,
                    base=self.sha, full=False, complete=True, branch="", branch_source="", session="", session_source="")
                gate = run.Gate(arguments)
                original = gate.spawn
                command = ["go", "tool", "test2json", "-p", "p", "-test.run", "^TestProduct_A$"]
                gate.spawn = lambda command, *args: original([sys.executable, "-c", "import time; time.sleep(600)"], *args)
                with mock.patch.object(run, "productKillSeconds", 0.05), mock.patch("builtins.print"):
                    code = gate.stream(phase, command, io.StringIO())
                self.assertNotEqual(code, 0)
                self.assertEqual(gate.failure["step"], phase)
                self.assertIn("p ^TestProduct_A$ killed at 0 s", gate.failure["detail"])
                self.assertTrue(all(process.poll() is not None for process in gate.processes))

    def test_product_pool_uses_four_cpu_slots(self):
        for cpus, expected in ((2, 1), (8, 2), (20, 5)):
            with self.subTest(cpus=cpus), mock.patch.object(run, "slotCPUs", return_value=cpus), \
                    mock.patch.object(run, "ThreadPoolExecutor", wraps=run.ThreadPoolExecutor) as pool:
                _, status, _, commands, _ = self.probe()
                self.assertTrue(status.startswith("green:"), status)
                pool.assert_called_once_with(max_workers=expected)
                self.assertIn("-test.timeout=30m", commands[1])

    def test_all_fast_stages_start_concurrently(self):
        barrier = threading.Barrier(4)
        stages = {"toolsDeclared": "tools", "build": "build", "vet": "vet", "testSplit": "tests"}
        def target(stage):
            def concurrent(gate, *args):
                barrier.wait(2)
                gate.exits[stage], gate.steps[stage] = 0, 0.0
                if stage in ("build", "vet"):
                    gate.result[stage + "_ok"] = True
            return concurrent
        from contextlib import ExitStack
        with ExitStack() as stack:
            for method, stage in stages.items():
                stack.enter_context(mock.patch.object(run.Gate, method, target(stage)))
            _, status, _ = self.gate(unowned=("stage3/probe.py",))
        self.assertTrue(status.startswith("green:"), status)

    def test_no_products_plan_no_stage(self):
        _, status, result, _, order = self.probe(names=("TestX",))
        self.assertTrue(status.startswith("green:"), status)
        self.assertNotIn("products", result["planned_stages"])
        self.assertNotIn("products", result["stages_exit"])
        self.assertEqual(order, ["TestX"])

    def test_product_only_package_is_fine(self):
        _, status, result, _, order = self.probe(names=("TestProduct_A",))
        self.assertTrue(status.startswith("green:"), status)
        self.assertEqual(order, ["TestProduct_A"])
        self.assertEqual(result["split_tally"]["planned"], 0)
        self.assertEqual(result["stages_exit"]["tests"], 0)


class ProductMutants(unittest.TestCase):
    def test_order_and_skip_mutants_are_killed(self):
        with open(run.__file__) as handle:
            original = handle.read()
        delayed = original.replace('productPool.submit(runProduct, row, binary)',
                                   'productPool.submit(lambda: (time.sleep(0.1), runProduct(row, binary)))')
        delayed = delayed.replace('        productsRemaining = len(productRows)',
                                  '        productsReady.set()\n        productsRemaining = len(productRows)')
        mutants = [("products after tests", delayed),
                   ("skip dropped", original.replace('"-test.skip", "^TestProduct_", ', '')),
                   ("Event set before products pass", original.replace('        productsRemaining = len(productRows)',
                        '        productsReady.set()\n        productsRemaining = len(productRows)'))]
        for name, source in mutants:
            with self.subTest(mutant=name):
                namespace = dict(run.__dict__)
                exec(compile(source, run.__file__, "exec"), namespace)
                with mock.patch.object(run.Gate, "testSplit", namespace["Gate"].testSplit):
                    result = unittest.TestResult()
                    test = "test_other_packages_wait_for_products_while_builds_continue" if name == "Event set before products pass" else "test_products_before_tests_and_skip"
                    Products(test).run(result)
                self.assertEqual(result.errors, [], result.errors)
                self.assertEqual(len(result.failures), 1, (name, result.failures))


class WholeProductMutants(unittest.TestCase):
    def test_whole_gate_dropped_barrier_is_killed(self):
        with open(run.__file__) as source:
            original = source.read()
        namespace = dict(run.__dict__)
        exec(compile(original.replace('if not self.wholeProducts(log) and not getattr(self.arguments, "run_to_end", False):\n            log.close()', 'if False:\n            log.close()'), run.__file__, "exec"), namespace)
        with mock.patch.object(run.Gate, "runFull", namespace["Gate"].runFull):
            result = unittest.TestResult()
            Products("test_whole_gate_products_before_tests_and_skip").run(result)
        self.assertEqual(result.errors, [], result.errors)
        self.assertEqual(len(result.failures), 1, result.failures)


class DeferredFamiliesInTheWholeGate(unittest.TestCase):
    """Main 20d538c0's whole log (#6g4zmzt): five deferred families read missing with every shard passing, since the whole
    gate matched exact names, and the record had no deferred verdict at all, since the census red had cancelled the run
    and the list's hash-object spawn refused."""

    def check(self, events):
        directory = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, directory, True)
        os.makedirs(os.path.join(directory, "cloud/fast-gate"))
        with open(os.path.join(directory, "cloud/fast-gate/deferred.txt"), "w") as handle:
            handle.write("# package test seconds kind\ninternal/native    TestFamily    40.0    differential-sweep\n")
        realRun(["git", "-C", directory, "init", "-q"], check=True)
        with open(os.path.join(directory, "test.jsonl"), "w") as handle:
            for action, test in events:
                handle.write(json.dumps({"Action": action, "Package": run.module + "/internal/native", "Test": test}) + "\n")
        failures = []

        def refuse(*arguments):
            raise SystemExit(1)

        gate = types.SimpleNamespace(arguments=types.SimpleNamespace(tree=directory, tools=directory, out=directory), result={},
                                     cancelled=True, git=refuse, fail=lambda step, detail: failures.append((step, detail)))
        gate.deferredList = lambda: run.Gate.deferredList(gate)
        with mock.patch.object(run, "deferredClassedOut", lambda tools, deferred: []):
            run.Gate.deferredRanWhole(gate)
        return gate.result["deferred_whole_results"][run.module + "/internal/native TestFamily"], failures

    def test_a_split_family_is_judged_over_its_shards_even_after_the_run_was_cancelled(self):
        self.assertEqual(self.check([("pass", "TestFamilyUnit00"), ("pass", "TestFamilyUnit01")]), ("pass", []))
        outcome, failures = self.check([("pass", "TestFamilyUnit00"), ("fail", "TestFamilyUnit01")])
        self.assertEqual((outcome, failures), ("fail", []))
        # A bare prefix is no shard, and shards that all skipped prove nothing.
        for events in ([("pass", "TestFamilyOther")], [("skip", "TestFamilyUnit00")]):
            outcome, failures = self.check(events)
            self.assertEqual(outcome, "missing")
            self.assertEqual(failures[0][0], "census")


class DeferredInTheWholeGate(FailClosed):
    def test_a_deferred_test_the_whole_gate_never_runs_is_red_at_census(self):
        with open(os.path.join(self.tree, "cloud/fast-gate/deferred.txt"), "w") as handle:
            handle.write("# package test seconds kind\ninternal/gone    TestGone    40.0    mutant\n")
        for command in (["add", "."], ["-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "deferred"]):
            realRun(["git", "-C", self.tree] + command, check=True)
        self.sha = run.git(self.tree, "rev-parse", "HEAD")
        gate, status, result = self.gate(full=True)
        self.assertIn("first failure at census", status)
        self.assertIn(run.module + "/internal/gone TestGone: missing", gate.failure["detail"])
        self.assertEqual(result["deferred_whole_results"], {run.module + "/internal/gone TestGone": "missing"})

    def test_the_tools_deferred_list_holds_no_measurement_or_opt_in_test(self):
        tools = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(run.__file__))))
        deferred = {}
        with open(os.path.join(tools, "cloud/fast-gate/deferred.txt")) as handle:
            for line in handle:
                fields = line.split()
                if fields and not fields[0].startswith("#"):
                    deferred.setdefault(run.module + "/" + fields[0], set()).add(fields[1])
        self.assertEqual(run.deferredClassedOut(tools, deferred), [])
        # The row that turned the star red on Oct 8 is refused.
        deferred[run.module + "/stage1/cohere/lint"] = {"TestProfileSnapshotsAgree"}
        self.assertEqual(run.deferredClassedOut(tools, deferred), [(run.module + "/stage1/cohere/lint TestProfileSnapshotsAgree", "measurement")])


class Coverage(FailClosed):
    def test_reads_add_package_for_matching_changed_paths(self):
        with open(os.path.join(self.tree, "cloud/fast-gate/executors.txt"), "w") as handle:
            handle.write("reads stage1/cohere/tsprinter **/*.ts\ninert *\n")
        for path, matches, unowned in (("stage3/new.ts", True, True), ("stage3/owned.ts", True, False), ("stage3/new.txt", False, True)):
            with self.subTest(path=path):
                # A real diff, including a newly added file, drives the additive selection.
                os.makedirs(os.path.dirname(os.path.join(self.tree, path)), exist_ok=True)
                with open(os.path.join(self.tree, path), "w") as handle:
                    handle.write("const x = 1\n")
                realRun(["git", "-C", self.tree, "add", path], check=True)
                realRun(["git", "-C", self.tree, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", path], check=True)
                base = self.sha
                self.sha = run.git(self.tree, "rev-parse", "HEAD")
                gate, status, result = self.gate(unowned=[path] if unowned else [], base=base)
                self.assertTrue(status.startswith("green:"), status)
                self.assertEqual(run.module + "/stage1/cohere/tsprinter" in result["packages"], matches)
                self.assertEqual(result["executors"], {"inert": 1} if unowned else {})
                self.assertEqual(result["reads"], [{"line": 1, "package": "stage1/cohere/tsprinter", "glob": "**/*.ts", "paths": [path], "map_changed": False}] if matches else [])

    def test_a_paths_reader_fires_only_when_paths_come_or_go(self):
        with open(os.path.join(self.tree, "cloud/fast-gate/executors.txt"), "w") as handle:
            handle.write("reads stage1/cohere/gitignore * paths\nreads stage1/cohere/gitignore *.gitignore\ninert *\n")
        gate = run.Gate.__new__(run.Gate)
        gate.arguments = mock.Mock(tools=self.tree)
        gate.result, gate.steps, gate.exits = {}, {}, {}
        gate.cover(["notes/edited.md"], ["notes/edited.md"], [])
        self.assertEqual(gate.extraPackages, [])
        gate.cover(["notes/added.md"], ["notes/added.md"], ["notes/added.md"])
        self.assertEqual(gate.extraPackages, [run.module + "/stage1/cohere/gitignore"])
        gate.cover(["stage3/.gitignore"], ["stage3/.gitignore"], [])
        self.assertEqual(gate.extraPackages, [run.module + "/stage1/cohere/gitignore"])

    def test_a_generated_reader_fires_on_what_its_list_bounds(self):
        # The one fixture loom's planner keys the same line against (planner/testdata/generated-reads.json, byte for
        # byte), so the two readers of a generated line can't mean two things (Node's condition 2 on #q11k6k6).
        with open(os.path.join(os.path.dirname(os.path.abspath(run.__file__)), "testdata/generated-reads.json")) as handle:
            fixture = json.load(handle)

        def treeOf(files):
            tree = tempfile.mkdtemp()
            self.addCleanup(shutil.rmtree, tree)
            for name, text in files.items():
                os.makedirs(os.path.dirname(os.path.join(tree, name)), exist_ok=True)
                with open(os.path.join(tree, name), "w") as handle:
                    handle.write(text)
            with open(os.path.join(tree, "cloud/fast-gate/executors.txt"), "w") as handle:
                handle.write("reads %s %s generated\ninert *\n" % (fixture["package"], fixture["list"]))
            return tree

        def fires(tree, changed, comeOrGo):
            gate = run.Gate.__new__(run.Gate)
            gate.arguments = mock.Mock(tools=tree, tree=tree)
            gate.result, gate.steps, gate.exits = {}, {}, {}
            gate.cover(changed, changed, comeOrGo)
            return run.module + "/" + fixture["package"] in gate.extraPackages

        for change in fixture["changes"]:
            with self.subTest(change=change["why"]):
                files = dict(fixture["tree"])
                for name in change.get("edit", []):
                    files[name] = files[name] + "\n"
                for name in change.get("add", []):
                    files[name] = "added %s\n" % name
                for name in change.get("delete", []):
                    del files[name]
                comeOrGo = change.get("add", []) + change.get("delete", [])
                changed = change.get("edit", []) + comeOrGo
                self.assertEqual(fires(treeOf(files), changed, comeOrGo), change["moves"])
        for bad in fixture["badLists"]:
            with self.subTest(bad=bad["why"]):
                files = dict(fixture["tree"])
                files[fixture["list"]] = bad["list"]
                tree = treeOf(files)
                self.assertIsNone(run.generatedReach(tree, fixture["list"]))
                self.assertTrue(fires(tree, ["README.md"], []))
        reach = run.generatedReach(treeOf(fixture["tree"]), fixture["list"])
        bounded = sorted(name for name in fixture["tree"] if name in reach[0] or name.startswith(reach[1]))
        self.assertEqual(bounded, fixture["reads"])

    def test_reads_do_not_cover_an_unowned_path(self):
        with open(os.path.join(self.tree, "cloud/fast-gate/executors.txt"), "w") as handle:
            handle.write("reads stage1/cohere/tsprinter **/*.ts\n")
        gate, status, result = self.gate(unowned=["stage3/new.ts"])
        self.assertEqual(gate.failure["step"], "coverage")
        self.assertEqual(result["uncovered_files"], ["stage3/new.ts"])

    def test_all_matching_reads_fire_and_name_their_paths(self):
        with open(os.path.join(self.tree, "cloud/fast-gate/executors.txt"), "w") as handle:
            handle.write("inert *\nreads stage1/cohere/tsprinter **/*.ts\nreads internal/flow stage3/*\n")
        gate = run.Gate.__new__(run.Gate)
        gate.arguments = mock.Mock(tools=self.tree)
        gate.result, gate.steps, gate.exits = {}, {}, {}
        paths = ["stage3/new.ts", "stage3/second.ts", "elsewhere/notes.txt"]
        self.assertEqual(gate.cover(paths, paths), {"inert"})
        self.assertEqual(gate.extraPackages, [run.module + "/internal/flow", run.module + "/stage1/cohere/tsprinter"])
        self.assertEqual([row["line"] for row in gate.result["reads"]], [2, 3])
        self.assertEqual([row["paths"] for row in gate.result["reads"]], [paths[:2], paths[:2]])

    def test_map_change_runs_every_reader(self):
        with open(os.path.join(self.tree, "cloud/fast-gate/executors.txt"), "w") as handle:
            handle.write("reads stage1/cohere/tsprinter **/*.ts\nreads internal/flow dedication/*.a\ninert *\n")
        arguments = mock.Mock(tools=self.tree)
        gate = run.Gate.__new__(run.Gate)
        gate.arguments, gate.result, gate.steps, gate.exits = arguments, {}, {}, {}
        gate.cover([], ["cloud/fast-gate/executors.txt"])
        self.assertEqual(gate.extraPackages, [run.module + "/internal/flow", run.module + "/stage1/cohere/tsprinter"])
        self.assertEqual(len(gate.result["reads"]), 2)
        self.assertTrue(all(row["map_changed"] for row in gate.result["reads"]))

    def test_a_path_with_no_executor_is_red_and_named(self):
        for full in (False, True):
            gate, status, result = self.gate(full=full, unowned=["tools/stray.py"])
            self.assertTrue(status.startswith("red:"), status)
            self.assertEqual(gate.failure["step"], "coverage")
            self.assertIn("tools/stray.py", gate.failure["detail"])
            self.assertEqual(result["uncovered_files"], ["tools/stray.py"])

    def test_an_inert_path_is_covered(self):
        gate, status, result = self.gate(unowned=["notes/README.md"])
        self.assertTrue(status.startswith("green:"), status)
        self.assertEqual(result["executors"], {"inert": 1})

    def test_a_ruled_corpus_is_exempt_from_a_check_and_named(self):
        # The exemption line comes first, so a path it matches would take it as its executor if the
        # map read it as one; the path must still be covered by stage3.
        with open(os.path.join(self.tree, "cloud/fast-gate/executors.txt"), "w") as handle:
            handle.write("a-check-exempt stage3/corpus/*\ninert *.md\nstage3 stage3/*\n")
        os.makedirs(os.path.join(self.tree, "stage3/corpus"))
        for path in ("stage3/corpus/upstream.a", "stage3/probe.a"):
            with open(os.path.join(self.tree, path), "w") as handle:
                handle.write("const x = 1\n")
        gate, status, result = self.gate(unowned=["stage3/corpus/upstream.a", "stage3/probe.a"])
        self.assertEqual(result["unchecked_a_files"], ["stage3/probe.a"])
        self.assertEqual(result["a_check_exempt"], ["stage3/corpus/upstream.a"])
        self.assertEqual(list(result["a_check"]), ["stage3/probe.a"])
        self.assertEqual(result["executors"], {"stage3": 2})


class ReadsGuard(unittest.TestCase):
    """The reads guard (#r2a9vd4): a reads line a change's delete leaves matching no tracked path, and a // Reads: comment
    naming a path its package's key doesn't hold (its own files, its content reads lines, its Go closure), are red, each
    naming its fix; a line dead on the base isn't the change's. go list is stubbed at its two seams, run.packageGraph (the
    package directories, self.packages: a, b and d) and run.readsClosureListing (the closure: a and b, whatever it's
    asked, recorded in self.listed), so no test here needs go. Mutants, each red here: the guard always passing; the
    allow list read as allowing everything; a bare allow accepted; go list's failure read as covered, or as the change's
    red; the closure ignored; a line dead on the base blamed on the change; a comment naming a deleted path unchecked; a
    string read as a comment; a wildcard comment matching nothing read as naming nothing; a paths line read as covering a
    comment; generated lines skipped by the Performance.jsonc check, an untracked list read, or a list that can't be read
    read as covering; a glob the gate and the planner read apart passed, in a line or a comment; an indirect shrink
    leaving unchanged files' comments unchecked, a modified .go file's among them; a dead line a change adds to the
    tree's executors.txt passed, or a tools' line charged to the change; a /* */ comment hiding a // Reads: after it; a
    carriage return read as a line; two keys folding to one field of a generated list read as one, or an integer past
    Go's int64 read as one; a crashed census exiting as findings do; the root package owning files below it."""

    executors = "mode guard-reads\nreads a b/data.txt\nreads a b/names/* paths\n"
    uncovered = ("a/a_test.go:%d reads b/lonely.txt, which a's key doesn't hold: add `reads a %s` to the tools line's "
                 "cloud/fast-gate/executors.txt (Keys or Judge adds it there; it isn't the change's to add)")

    def setUp(self):
        self.directory = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, self.directory)
        self.tree = os.path.join(self.directory, "tree")
        os.makedirs(self.tree)
        realRun(["git", "-C", self.tree, "init", "-q"], check=True)
        self.packages = ["a", "b", "d"]
        self.listed = []
        self.closure = types.SimpleNamespace(returncode=0, stderr="", stdout="".join(json.dumps(record) + "\n" for record in (
            {"ImportPath": "fmt", "Standard": True, "Dir": "/usr/lib/go/src/fmt", "GoFiles": ["print.go"]},
            {"ImportPath": run.module + "/b", "Dir": os.path.join(self.tree, "b"), "GoFiles": ["b.go"]},
            {"ImportPath": run.module + "/a", "Dir": os.path.join(self.tree, "a"), "GoFiles": ["a.go"]})))
        for name, replacement in (("packageGraph", self.graph), ("readsClosureListing", self.listing)):
            patcher = mock.patch.object(run, name, side_effect=replacement)
            patcher.start()
            self.addCleanup(patcher.stop)
        # a declares b as a compiler it builds, so its closure is listed with b's.
        self.base = self.write({"a/a.go": "package a\n", "a/testdata/own.txt": "own\n", "b/b.go": "package b\n", "b/data.txt": "data\n",
                                "b/names/one.txt": "one\n", "b/lonely.txt": "lonely\n", "b/old.txt": "old\n", "d/d.go": "package d\n",
                                run.compilerDependenciesPath: json.dumps({"version": 1, "packages": {"a": ["b"]}}),
                                "cloud/fast-gate/executors.txt": self.executors})

    def graph(self, tree):
        return types.SimpleNamespace(returncode=0, stderr="", stdout="".join(
            json.dumps({"ImportPath": run.module + "/" + package, "Dir": os.path.join(self.tree, package)}) + "\n" for package in self.packages))

    def listing(self, tree, command):
        self.listed.append(command)
        return self.closure

    def write(self, files=None, removed=()):
        """files written and removed paths deleted, as one commit of the tree: its sha."""
        for name, text in (files or {}).items():
            os.makedirs(os.path.dirname(os.path.join(self.tree, name)), exist_ok=True)
            with open(os.path.join(self.tree, name), "w") as handle:
                handle.write(text)
        for name in removed:
            realRun(["git", "-C", self.tree, "rm", "-q", name], check=True)
        for command in (["add", "-A"], ["-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "t"]):
            realRun(["git", "-C", self.tree] + command, check=True)
        return run.git(self.tree, "rev-parse", "HEAD")

    def gate(self, base):
        """A fast Gate of the tree's HEAD against base, its tools the tree."""
        return run.Gate(types.SimpleNamespace(tree=self.tree, sha=run.git(self.tree, "rev-parse", "HEAD"), base=base, tools=self.tree,
                                              out=tempfile.mkdtemp(dir=self.directory), full=False, branch="", branch_source="",
                                              session="", session_source=""))

    def census(self):
        """run.py --reads-census over the tree, its tools the tree: (exit code, stdout)."""
        printed = io.StringIO()
        with mock.patch.object(sys, "argv", ["run.py", "--reads-census", "--tree", self.tree, "--tools", self.tree]), \
                contextlib.redirect_stdout(printed), self.assertRaises(SystemExit) as exited:
            run.main()
        return exited.exception.code, printed.getvalue()

    def test_a_dead_line_made_by_the_changes_delete_is_red_and_one_dead_at_the_base_isnt(self):
        self.write({"cloud/fast-gate/executors.txt": self.executors + "reads a b/old.txt\nreads a b/never.txt\n"})
        self.write(removed=["b/old.txt"])
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["b/old.txt"], ["b/old.txt"]),
                         ["`reads a b/old.txt` (executors.txt:4) matches no tracked path: point it at where the file lives now, or delete it"])
        # The census, which isn't a change's, names the line dead all along too.
        self.assertIn("`reads a b/never.txt` (executors.txt:5) matches no tracked path: point it at where the file lives now, or delete it",
                      run.readsGuardFindings(self.tree, self.tree))

    def test_an_uncovered_comment_in_a_changed_file_is_red_naming_the_line_to_add(self):
        self.write({"a/a_test.go": "package a\n\n// Reads: a/testdata/own.txt\n// Reads: b/data.txt\n// Reads: b/b.go\n"
                                   "var quoted = \"// Reads: b/lonely.txt\"\n// Reads: b/lonely.txt\n"})
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["a/a_test.go"], []), [self.uncovered % (7, "b/lonely.txt")])
        # One go list of a's closure, its declared compiler b beside it.
        self.assertEqual([command[-2:] for command in self.listed], [["./a", "./b"]])

    def test_a_package_rebuilt_for_anothers_test_never_makes_its_files_the_others_own(self):
        # go list -test lists b rebuilt for a's test with ForTest a, which goPackages files under a's import path, so b's
        # directory maps to a there. Read by import path, b's lonely.txt would be a's own file and the comment covered.
        def withVariant(tree):
            listed = self.graph(tree)
            listed.stdout += json.dumps({"ImportPath": run.module + "/b [" + run.module + "/a.test]", "ForTest": run.module + "/a",
                                         "Dir": os.path.join(self.tree, "b")}) + "\n"
            return listed
        self.write({"a/a_test.go": "package a\n\n// Reads: b/lonely.txt\n"})
        with mock.patch.object(run, "packageGraph", side_effect=withVariant):
            self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["a/a_test.go"], []), [self.uncovered % (3, "b/lonely.txt")])

    def test_a_content_reads_line_matching_the_root_performance_file_is_red_and_a_paths_line_isnt(self):
        # Performance.jsonc at the root is rewritten by every landing (#hed1mkb), so a content line whose glob matches its
        # name is red, a bare * among them, and an allow line doesn't exempt it. The same * as a paths line passes, as
        # gitignore's does, since names alone move only at the seed; a glob under a directory, as docs/*.jsonc, passes.
        lines = "reads a Performance*\nreads a *\nreads a * paths\nreads a docs/*.jsonc\n# allow: a Performance*: no allow exempts it\n"
        self.write({"cloud/fast-gate/executors.txt": self.executors + lines, "Performance.jsonc": "{}\n", "docs/x.jsonc": "{}\n"})
        expected = ["`reads a Performance*` (executors.txt:4) matches Performance.jsonc, which every landing rewrites at adamic's root, "
                    "so no unit may read it (#hed1mkb): narrow the glob so it can't match the root's Performance.jsonc",
                    "`reads a *` (executors.txt:5) matches Performance.jsonc, which every landing rewrites at adamic's root, "
                    "so no unit may read it (#hed1mkb): narrow the glob so it can't match the root's Performance.jsonc"]
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["docs/x.jsonc"], []), expected)
        self.assertEqual(self.census(), (1, "\n".join(expected) + "\n"))

    def test_a_performance_file_anywhere_but_the_root_is_red(self):
        self.write({"Performance.jsonc": "{}\n", "b/Performance.jsonc": "{}\n"})
        expected = ["b/Performance.jsonc is a Performance.jsonc outside adamic's root, where only the root's is the landing's "
                    "measurements (#hed1mkb): move or rename it"]
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["b/Performance.jsonc", "Performance.jsonc"], ["b/Performance.jsonc"]), expected)
        # A change that doesn't touch the copy isn't charged with it; the census names it.
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["a/a.go"], []), [])
        self.assertEqual(self.census(), (1, expected[0] + "\n"))

    def test_a_comment_covered_by_own_files_a_line_or_the_closure_passes(self):
        for glob in ("a/testdata/own.txt", "b/data.txt", "b/b.go"):
            with self.subTest(glob=glob):
                self.write({"a/a_test.go": "package a\n\n// Reads: %s\n" % glob})
                self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["a/a_test.go"], []), [])
        # Only the comment that own files and lines don't cover lists the closure.
        self.assertEqual(len(self.listed), 1)

    def test_the_fast_gate_is_red_at_reads_naming_the_finding(self):
        self.write({"a/a_test.go": "package a\n\n// Reads: b/lonely.txt\n"})
        gate = self.gate(self.base)
        with mock.patch("builtins.print"):
            self.assertFalse(gate.readsGuard())
        self.assertEqual((gate.failure["step"], gate.exits["reads"], gate.planned[-1]), ("reads", 1, "reads"))
        self.assertFalse(gate.failure.get("tool_crash"))
        self.assertIn(self.uncovered % (3, "b/lonely.txt"), gate.failure["detail"])
        self.assertEqual(gate.result["reads_guard"], [self.uncovered % (3, "b/lonely.txt")])
        # The change adds a .go file, so every comment was checked, and the record says why and how long it took.
        self.assertEqual(gate.result["reads_comment_census"]["paths"], ["a/a_test.go"])
        self.assertIsInstance(gate.result["reads_comment_census"]["seconds"], float)

    def test_a_comment_naming_a_path_the_change_deletes_is_red(self):
        self.write({"a/a_test.go": "package a\n\n// Reads: b/data.txt\n", "d/d_test.go": "package d\n\n// Reads: b/lonely.txt\n"})
        self.write(removed=["b/data.txt"])
        # d's uncovered comment sits in a file the change doesn't touch and names nothing it deletes: the census's, never this change's.
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["b/data.txt"], ["b/data.txt"]), [
            "`reads a b/data.txt` (executors.txt:2) matches no tracked path: point it at where the file lives now, or delete it",
            "a/a_test.go:3 reads b/data.txt, which isn't a tracked path: point the comment at where the file lives now, or delete it"])

    def test_go_list_failing_is_the_gates_error_not_the_changes_red(self):
        self.write({"a/a_test.go": "package a\n\n// Reads: b/lonely.txt\n"})
        self.closure = types.SimpleNamespace(returncode=1, stdout="", stderr="go: cannot load the module")
        with self.assertRaises(ValueError):
            run.readsGuardFindings(self.tree, self.tree, ["a/a_test.go"], [])
        gate = self.gate(self.base)
        with mock.patch("builtins.print"):
            self.assertFalse(gate.readsGuard())
        # Charged as a stage that couldn't start: its traceback is the failure, the gate tool's crash, never a finding.
        self.assertEqual(gate.failure["step"], "reads")
        self.assertTrue(gate.failure.get("tool_crash"), gate.failure)
        self.assertIn("go: cannot load the module", gate.failure["detail"])
        self.assertNotIn("reads_guard", gate.result)
        self.assertNotIn("reads", gate.exits)

    def test_an_allow_exempts_exactly_its_package_and_glob_and_a_bare_allow_is_red(self):
        self.write({"cloud/fast-gate/executors.txt": self.executors +
                    "# allow: a b/lonely.txt: a's test reads it through a helper its key holds\n# allow: a b/gone.txt: dead on purpose\n"
                    "reads a b/gone.txt\nreads d b/gone.txt\n# allow: a b/x.txt:\n# allow: a: no glob\n",
                    "a/a_test.go": "package a\n\n// Reads: b/lonely.txt\n// Reads: b/lonely*\n", "d/d_test.go": "package d\n\n// Reads: b/lonely.txt\n"})
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree), [
            "executors.txt:8 is an allow without a package, a glob and why: write it `# allow: <package> <glob>: <why>`, or delete it",
            "executors.txt:9 is an allow without a package, a glob and why: write it `# allow: <package> <glob>: <why>`, or delete it",
            "`reads d b/gone.txt` (executors.txt:7) matches no tracked path: point it at where the file lives now, or delete it",
            self.uncovered % (4, "b/lonely*"),
            "d/d_test.go:3 reads b/lonely.txt, which d's key doesn't hold: add `reads d b/lonely.txt` to the tools line's "
            "cloud/fast-gate/executors.txt (Keys or Judge adds it there; it isn't the change's to add)"])

    def test_off_the_guard_does_nothing_and_plans_nothing(self):
        self.write({"cloud/fast-gate/executors.txt": self.executors.replace("mode guard-reads\n", ""),
                    "a/a_test.go": "package a\n\n// Reads: b/lonely.txt\n"})
        gate = self.gate(self.base)
        with mock.patch.object(run, "readsGuardFindings", side_effect=AssertionError("the guard is off")):
            self.assertTrue(gate.readsGuard())
        self.assertNotIn("reads", gate.planned)
        self.assertIsNone(gate.failure)

    def test_the_census_names_every_finding_over_the_tree_and_exits_0_on_none(self):
        # Off in the tools, the census runs all the same: it's what turns the guard on.
        self.write({"cloud/fast-gate/executors.txt": "reads a b/data.txt\nreads a b/never.txt\n", "a/a_test.go": "package a\n\n// Reads: b/lonely.txt\n"})
        self.assertEqual(self.census(), (1, "`reads a b/never.txt` (executors.txt:2) matches no tracked path: point it at where the file lives now, "
                                            "or delete it\n" + self.uncovered % (3, "b/lonely.txt") + "\n"))
        self.write({"cloud/fast-gate/executors.txt": "reads a b/data.txt\nreads a b/lonely.txt\n"})
        self.assertEqual(self.census(), (0, ""))

    def test_a_wildcard_comment_matching_nothing_is_red_as_a_missing_path(self):
        # b's own test reads its fixtures by a glob; a change that moves every one away leaves the glob matching nothing,
        # which is a missing path, never a comment naming nothing.
        self.write({"b/b_test.go": "package b\n\n// Reads: b/fixtures/*.golden\n", "b/fixtures/one.golden": "one\n"})
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["b/b_test.go"], []), [])
        self.write({"b/moved/one.golden": "one\n"}, removed=["b/fixtures/one.golden"])
        moved = ["b/fixtures/one.golden", "b/moved/one.golden"]
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, moved, moved), [
            "b/b_test.go:3 reads b/fixtures/*.golden, which isn't a tracked path: point the comment at where the file lives now, or delete it"])
        # A glob inside a gitlink matches nothing the superproject tracks, ever: git ls-files holds cohere alone.
        os.makedirs(os.path.join(self.tree, "cohere"))
        realRun(["git", "-C", self.tree, "update-index", "--add", "--cacheinfo", "160000,%s,cohere" % self.base], check=True)
        self.write({"a/a_test.go": "package a\n\n// Reads: cohere/*.ts\n"})
        # A .go file the change touches checks every comment, b's still matching nothing among them.
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["a/a_test.go"], []), [
            "a/a_test.go:3 reads cohere/*.ts, which isn't a tracked path: point the comment at where the file lives now, or delete it",
            "b/b_test.go:3 reads b/fixtures/*.golden, which isn't a tracked path: point the comment at where the file lives now, or delete it"])

    def test_a_comment_only_a_paths_line_matches_is_red(self):
        # reads a b/names/* paths keys its matches by name alone, so an edit to b/names/one.txt reruns nothing, and a
        # comment says the test reads what the file holds.
        self.write({"a/a_test.go": "package a\n\n// Reads: b/names/*\n"})
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["a/a_test.go"], []),
                         [self.uncovered.replace("b/lonely.txt", "b/names/one.txt") % (3, "b/names/*")])

    def test_a_generated_line_reaching_the_performance_file_or_whose_list_cant_be_read_is_red_and_covers_nothing(self):
        # The planner keys a generated line's reach (planner/reads.go generatedReads): the whole tree, Performance.jsonc
        # among it, when the list won't decode strictly or isn't tracked, and any path the list names. Each such line is
        # red; one naming b/data.txt alone passes. A list that can't be read covers no comment, since the guard can't
        # tell what the planner keys.
        lines = ("reads a b/broken.json generated\nreads a b/root.json generated\nreads a b/fine.json generated\n"
                 "reads a b/untracked.json generated\n")
        self.write({"cloud/fast-gate/executors.txt": self.executors + lines, "Performance.jsonc": "{}\n", "b/broken.json": "{\n",
                    "b/root.json": json.dumps({"files": [{"path": "Performance.jsonc"}]}),
                    "b/fine.json": json.dumps({"files": [{"path": "b/data.txt"}]}), "a/a_test.go": "package a\n\n// Reads: b/lonely.txt\n"})
        # On disk, readable, naming the comment's file, and never added: the planner reads only a tracked list.
        with open(os.path.join(self.tree, "b/untracked.json"), "w") as handle:
            json.dump({"files": [{"path": "b/lonely.txt"}]}, handle)
        unreadable = ("`reads a %s generated` (executors.txt:%d) names a list loom's planner can't read as cloud/generated_test.go "
                      "decodes it (untracked, or not strict JSON of its shape), so it keys the whole tree, Performance.jsonc among it, "
                      "which every landing rewrites at adamic's root (#hed1mkb): make the list tracked and strict")
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["a/a_test.go"], []), [
            unreadable % ("b/broken.json", 4),
            "`reads a b/root.json generated` (executors.txt:5) names a list whose entries reach Performance.jsonc, which every "
            "landing rewrites at adamic's root, so no unit may read it (#hed1mkb): drop the entry that reaches it",
            unreadable % ("b/untracked.json", 7),
            self.uncovered % (3, "b/lonely.txt")])

    def test_a_glob_the_gate_and_the_planner_read_apart_is_red_in_a_line_and_a_comment_whatever_allows_it(self):
        # Python's fnmatch negates [!x] and Go's path.Match [^x]; Go escapes with a backslash, which Python reads as
        # itself; and Go matches nothing with a [ that opens no class, which Python matches as itself. Each is red in a
        # line and in a comment, its allow line notwithstanding. A class both read alike, [dl], passes.
        negates = "a class opening [! or [^, which Python's fnmatch and Go's path.Match negate on different characters"
        escapes = "a backslash, which Python's fnmatch reads as itself and Go's path.Match as an escape"
        unclosed = "a [ opening no class Go's path.Match can read, which Python's fnmatch matches as itself and Go as nothing"
        forms = [("b/[!x]*.txt", negates), ("b/[^x]*.txt", negates), ("b/\\data.txt", escapes), ("b/[data.txt", unclosed),
                 ("b/[a-].txt", unclosed)]
        lines = "".join("reads a %s\n# allow: a %s: allowed all the same\n" % (glob, glob) for glob, _ in forms)
        comments = "".join("// Reads: %s\n" % glob for glob, _ in forms)
        self.write({"cloud/fast-gate/executors.txt": self.executors + lines + "reads a b/[dl]ata.txt\n",
                    "a/a_test.go": "package a\n\n" + comments + "// Reads: b/[dl]ata.txt\n"})
        for reds in (run.readsGuardFindings(self.tree, self.tree), run.readsGuardFindings(self.tree, self.tree, ["a/a_test.go"], [])):
            for index, (glob, why) in enumerate(forms):
                with self.subTest(glob=glob):
                    self.assertIn("`reads a %s` (executors.txt:%d) holds %s: write the glob so both read it alike, since no allow "
                                  "exempts it" % (glob, 4 + 2 * index, why), reds)
                    self.assertIn("a/a_test.go:%d reads %s, which holds %s: write the glob so both read it alike, since no allow "
                                  "exempts it" % (3 + index, glob, why), reds)
            self.assertFalse([red for red in reds if "[dl]" in red], reds)

    def test_a_change_dropping_the_covering_line_from_the_trees_executors_reds_an_unchanged_files_comment(self):
        # The comment's file is untouched, but the line that covered it is gone: every comment is checked, as the census
        # checks them, and the record says which paths made it and its seconds. A change that can't shrink a key elsewhere
        # checks only its own files' comments.
        self.write({"cloud/fast-gate/executors.txt": self.executors + "reads a b/lonely.txt\n", "a/a_test.go": "package a\n\n// Reads: b/lonely.txt\n"})
        self.write({"cloud/fast-gate/executors.txt": self.executors})
        record = {}
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["cloud/fast-gate/executors.txt"], [], record),
                         [self.uncovered % (3, "b/lonely.txt")])
        self.assertEqual(record["reads_comment_census"]["paths"], ["cloud/fast-gate/executors.txt"])
        self.assertIsInstance(record["reads_comment_census"]["seconds"], float)
        record = {}
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["b/data.txt"], [], record), [])
        self.assertEqual(record, {})

    def test_a_modified_go_file_dropping_an_import_reds_an_unchanged_files_comment(self):
        # a's test reads b/b.go, which only a's Go closure holds, through a.go's import of b (no compiler a declares). A
        # change that modifies a.go alone to drop the import shrinks the closure, adding and deleting no path and leaving
        # the comment's file untouched: every comment is checked, as the census checks them, the record naming a.go. A
        # change editing only a non-Go file outside executors.txt and the other shrinking paths checks its own Go files'
        # comments alone, and records no census.
        self.write({run.compilerDependenciesPath: json.dumps({"version": 1, "packages": {}}),
                    "a/a.go": "package a\n\nimport _ \"%s/b\"\n" % run.module, "a/a_test.go": "package a\n\n// Reads: b/b.go\n"})
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["a/a_test.go"], []), [])
        self.write({"a/a.go": "package a\n"})
        # a's closure as go list answers it without the import: a alone.
        self.closure = types.SimpleNamespace(returncode=0, stderr="", stdout=json.dumps(
            {"ImportPath": run.module + "/a", "Dir": os.path.join(self.tree, "a"), "GoFiles": ["a.go"]}) + "\n")
        record = {}
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["a/a.go"], [], record),
                         [self.uncovered.replace("b/lonely.txt", "b/b.go") % (3, "b/b.go")])
        self.assertEqual(record["reads_comment_census"]["paths"], ["a/a.go"])
        record = {}
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["b/data.txt"], [], record), [])
        self.assertEqual(record, {})

    def test_a_dead_line_a_change_adds_to_the_trees_executors_is_red_and_one_the_tools_hold_isnt(self):
        # The gate tools are another checkout, so the tree's executors.txt is the change's: a line it adds matching no
        # tracked path is red in change mode, though it matched no path the change deleted. The same line held by the
        # tools' executors.txt is the tools line's, never the change's, and only the census names it.
        tools = os.path.join(self.directory, "tools")
        os.makedirs(os.path.join(tools, "cloud/fast-gate"))
        with open(os.path.join(tools, "cloud/fast-gate/executors.txt"), "w") as handle:
            handle.write(self.executors)
        self.write({"cloud/fast-gate/executors.txt": self.executors + "reads a nothing/here.txt\n"})
        changed = ["cloud/fast-gate/executors.txt"]
        dead = "`reads a nothing/here.txt` (%s:4) matches no tracked path: point it at where the file lives now, or delete it"
        self.assertEqual(run.readsGuardFindings(self.tree, tools, changed, []), [dead % "the tree's executors.txt"])
        with open(os.path.join(tools, "cloud/fast-gate/executors.txt"), "w") as handle:
            handle.write(self.executors + "reads a nothing/here.txt\n")
        self.assertEqual(run.readsGuardFindings(self.tree, tools, changed, []), [])
        self.assertEqual(run.readsGuardFindings(self.tree, tools), [dead % "executors.txt"])

    def test_a_dead_line_a_change_adds_where_the_tools_are_the_tree_is_red(self):
        # A gate whose tools are its own tree reads every line as the tools', and a change to the tree's executors.txt
        # answers for each of them: a dead line it adds is red in change mode there too (Adamic Loom's question on
        # #r2a9vd4). Mutant: sameTree dropped from the dead-line rule, green.
        self.write({"cloud/fast-gate/executors.txt": self.executors + "reads a nothing/here.txt\n"})
        dead = "`reads a nothing/here.txt` (executors.txt:4) matches no tracked path: point it at where the file lives now, or delete it"
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["cloud/fast-gate/executors.txt"], []), [dead])

    def test_a_go_mod_renamed_away_checks_every_comment(self):
        # git mv of a nested go.mod folds its module into its parent's, so a package's files can enter another unit's key
        # without touching the comment's file. With rename detection the old path is only in the added and deleted set,
        # so names are read from both (Adamic Loom's should 2 on #r2a9vd4). Mutant: names read from changed alone, no
        # census recorded.
        self.write({"c/go.mod": "module c\n", "c/c.go": "package c\n"})
        record = {}
        run.readsGuardFindings(self.tree, self.tree, ["c/go.mod.off"], ["c/go.mod", "c/go.mod.off"], record)
        self.assertIn("c/go.mod", record["reads_comment_census"]["paths"])

    def test_a_reads_comment_is_read_as_repowalk_reads_one(self):
        # repowalk takes every // comment go/scanner finds (repowalk.go readsLines): one after a /* */ comment on its line
        # counts, as does one after a /* */ comment that ends there; go/scanner drops a // comment's carriage returns, and
        # counts a line only at a newline; Go's unicode.IsSpace isn't a space at \x1f. A // Reads: inside a /* */ comment, or
        # after other words in a // comment, declares nothing.
        source = ("package a\n\n/* a note */ // Reads: b/lonely.txt\n/* a note\nover lines */ // Reads: b/old.txt\n"
                  "// Reads: b/lonely\r.txt\r\n// one\r// Reads: b/hidden.txt\n// Reads: b/x\x1fy\n/* // Reads: b/inside.txt */\n"
                  "// see // Reads: b/quoted.txt\n")
        self.assertEqual(run.readsComments(source), [(3, "b/lonely.txt"), (5, "b/old.txt"), (6, "b/lonely.txt"), (8, "b/x\x1fy")])
        # Read from the file, a lone carriage return is no line: the comment is on line 3, as go/scanner counts it.
        self.write({"a/a_test.go": "package a\n// one\r// two\n/* a note */ // Reads: b/lonely.txt\n"})
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["a/a_test.go"], []), [self.uncovered % (3, "b/lonely.txt")])

    def test_a_generated_list_is_decoded_as_gos_decoder_reads_it(self):
        # encoding/json matches a key to a field case folded, so {"Files": ...} is the list's files, as the planner reads it;
        # two keys folding to one field are decoded into it one over the other, an integer past int64 is an error, and a
        # key that isn't ASCII folds by rules this reader doesn't copy: each of those is the whole tree, None.
        lists = {"b/folded.json": '{"MaxFiles": 1, "Files": [{"Path": "b/data.txt", "UNDER": null, "MaxBytes": 2}]}',
                 "b/twice.json": '{"files": [{"under": "b/names/"}], "Files": [{"path": "b/data.txt"}]}',
                 "b/same.json": '{"files": [{"path": "b/data.txt", "path": "b/old.txt"}]}',
                 "b/wide.json": '{"maxFiles": 9223372036854775808, "files": [{"path": "b/data.txt"}]}',
                 "b/long.json": '{"files": [{"path": "b/data.txt", "\\u017fuffix": ".txt"}]}'}
        self.write(lists)
        self.assertEqual(run.generatedReach(self.tree, "b/folded.json"), ({"b/folded.json", "b/data.txt"}, ()))
        for name in ("b/twice.json", "b/same.json", "b/wide.json", "b/long.json"):
            with self.subTest(name=name):
                self.assertIsNone(run.generatedReach(self.tree, name))

    def test_a_census_that_crashes_exits_2_never_1_as_findings_do(self):
        printed, failed = io.StringIO(), io.StringIO()
        with mock.patch.object(run, "readsGuardFindings", side_effect=ValueError("go list of the tree failed")), \
                mock.patch.object(sys, "argv", ["run.py", "--reads-census", "--tree", self.tree, "--tools", self.tree]), \
                contextlib.redirect_stdout(printed), contextlib.redirect_stderr(failed), self.assertRaises(SystemExit) as exited:
            run.main()
        self.assertEqual((exited.exception.code, printed.getvalue()), (2, ""))
        self.assertIn("ValueError: go list of the tree failed", failed.getvalue())

    def test_the_root_package_owns_only_the_files_directly_in_it(self):
        # As planner/ownfiles.go's nearestPackage walks: notes.txt beside the root's test is its own, and e/notes.txt,
        # under a directory that's no package, is no package's own, never the root's.
        self.packages = self.packages + ["."]
        self.write({"root_test.go": "package adamic\n\n// Reads: notes.txt\n// Reads: e/notes.txt\n", "notes.txt": "notes\n", "e/notes.txt": "notes\n"})
        self.assertEqual(run.readsGuardFindings(self.tree, self.tree, ["root_test.go"], []), [
            "root_test.go:4 reads e/notes.txt, which .'s key doesn't hold: add `reads . e/notes.txt` to the tools line's "
            "cloud/fast-gate/executors.txt (Keys or Judge adds it there; it isn't the change's to add)"])


class ReadsGuardInTheGate(unittest.TestCase):
    """The reads guard in a whole fast gate run, after cover() and before any test: off by default, the gate runs exactly
    as before it, the guard never asked; on ("mode guard-reads" in the tools' executors.txt), a change with an uncovered
    comment is red at reads. Mutant, red here: the switch read as on by default."""
    gate = FailClosed.gate

    def setUp(self):
        FailClosed.setUp(self)
        self.addCleanup(shutil.rmtree, self.directory)

    def change(self, executors):
        """The tools' map as executors on the base, then a change adding a's test with a comment no line covers: the base."""
        files = {"cloud/fast-gate/executors.txt": executors, "a/a.go": "package a\n", "b/lonely.txt": "lonely\n",
                 run.compilerDependenciesPath: json.dumps({"version": 1, "packages": {}})}
        shas = []
        for batch in (files, {"a/a_test.go": "package a\n\n// Reads: b/lonely.txt\n"}):
            for name, text in batch.items():
                os.makedirs(os.path.dirname(os.path.join(self.tree, name)), exist_ok=True)
                with open(os.path.join(self.tree, name), "w") as handle:
                    handle.write(text)
            for command in (["add"] + list(batch), ["-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "t"]):
                realRun(["git", "-C", self.tree] + command, check=True)
            shas.append(run.git(self.tree, "rev-parse", "HEAD"))
        self.sha = shas[1]
        return shas[0]

    def stubbed(self):
        """go list stubbed at the guard's two seams: the package graph (a and b) and a's closure (a alone)."""
        graph = types.SimpleNamespace(returncode=0, stderr="", stdout="".join(
            json.dumps({"ImportPath": run.module + "/" + package, "Dir": os.path.join(self.tree, package)}) + "\n" for package in ("a", "b")))
        closure = types.SimpleNamespace(returncode=0, stderr="", stdout=json.dumps({"ImportPath": run.module + "/a", "Dir": os.path.join(self.tree, "a"), "GoFiles": ["a.go"]}))
        return mock.patch.object(run, "packageGraph", return_value=graph), mock.patch.object(run, "readsClosureListing", return_value=closure)

    def test_off_by_default_the_fast_gate_runs_no_guard(self):
        base = self.change("inert *.md\nstage3 stage3/*\n")
        graph, closure = self.stubbed()
        with graph, closure, mock.patch.object(run, "readsGuardFindings", side_effect=AssertionError("the guard is off")):
            gate, status, result = self.gate(base=base)
        self.assertTrue(status.startswith("green:"), status)
        self.assertNotIn("reads", result["planned_stages"])
        self.assertNotIn("reads_guard", result)

    def test_on_a_change_with_an_uncovered_comment_is_red_at_reads(self):
        base = self.change("mode guard-reads\ninert *.md\nstage3 stage3/*\n")
        graph, closure = self.stubbed()
        with graph, closure:
            gate, status, result = self.gate(base=base)
        self.assertIn("first failure at reads", status)
        self.assertEqual(result["reads_guard"], ["a/a_test.go:3 reads b/lonely.txt, which a's key doesn't hold: add `reads a b/lonely.txt` to the "
                                                 "tools line's cloud/fast-gate/executors.txt (Keys or Judge adds it there; it isn't the change's to add)"])
        self.assertEqual(result["stages_exit"]["reads"], 1)


class DarwinLegs(unittest.TestCase):
    """The darwin executor passes on a pool Mac's recorded leg, cloud/fast-gate/darwin-legs/<tip>.json in the tools
    checkout, whose darwinPaths digests the candidate's leg inputs, and is red without one, naming the pool Mac and the
    digest it looked for (no box builds on Kirk's Mac, so no wording names it). The candidate is the change rebased onto
    main, so the leg's tip is never the gate's sha: a main that moved outside the leg's inputs still matches, one that
    moved inside them, the darwin paths, the in-module closure go list names, or a compiler the leg's tests build, doesn't.
    go list is stubbed at its one seam (run.darwinClosureListing, self.listing), which records what it was asked and
    answers by its patterns, so no test here needs go. Mutants, each red here: cover() failing darwin unread; coverage
    reding on a passed darwin exit; the darwin phase failing unread; any one green condition dropped; the digest over
    every path; the dir lines dropped, or a directory outside the tree or a sibling of it kept; a gitlink keyed by
    anything but its commit; the checkout check dropped; the declared compilers left out of the listing, read from the
    tools' map alone, or taken from a package outside internal/apple; a map that can't be read listed without them; a
    failed or empty go list read as a digest, or crashing the gate rather than reding darwin; a record's tools unchecked,
    checked a key short, or against a tools.json that isn't exactly four strings; a-check passing a darwin .a on any matching leg rather than on the
    run of it the leg recorded, or ignoring its header there; a-check reading the leg for a live refusal, an untyped stop
    or a path that isn't darwin's."""
    gate = FailClosed.gate
    edited = "package apple\n\nvar Edited = 1\n"
    # The leg's closure as the stubbed go list names it for ./internal/apple/..., tree-relative: internal/apple and the
    # in-module package it imports. internal/unrelated sits in the tree outside it, and outside the darwin globs.
    closure = ("internal/apple", "internal/lower")
    # The listing the setUp map makes: ./internal/apple/..., then the compilers it declares for internal/apple, sorted.
    declaredCommand = ["go", "list", "-deps", "-test", "-f", "{{.Dir}}", "./internal/apple/...", "./cmd/adamic", "./internal/lower"]
    # The pool Mac's tools as the gate tools' darwin-legs/tools.json declares them, the spellings the records carry.
    declaredTools = {"xcode": "Xcode 27.2 (27B52)", "sdk": "macOS 27.2 (26B118)", "go": "go1.27.1 darwin/arm64", "node": "v24.19.0"}

    def setUp(self):
        FailClosed.setUp(self)
        self.addCleanup(shutil.rmtree, self.directory)
        sources = {"cloud/fast-gate/executors.txt": "a-check-exempt examples/apple/exempt/*\ndarwin internal/apple/*\ndarwin examples/apple/*\n"
                                                    "inert *.md\nstage3 stage3/*\ninert notes/*\n",
                   "go.mod": "module " + run.module + "\n\ngo 1.24\n", "go.sum": "",
                   # internal/apple's tests build cmd/adamic and run it, so the map declares it, with lower, as its compilers.
                   run.compilerDependenciesPath: self.mapOf({"internal/apple": ["cmd/adamic", "internal/lower"]}),
                   "cmd/adamic/main.go": "package main\n", "internal/ir/ir.go": "package ir\n",
                   "internal/apple/apple.go": "package apple\n", "internal/lower/lower.go": "package lower\n",
                   "internal/unrelated/unrelated.go": "package unrelated\n", "examples/apple/app.a": "const x = 1\n",
                   "examples/apple/refused.a": "// a-check: refused no-any\nconst x: any = 1\n", "examples/apple/exempt/vendored.a": "vendored\n",
                   "notes/main.md": "# main\n", "probes/probe.a": "const y = 1\n"}
        for name, source in sources.items():
            self.write(name, source)
        realRun(["git", "-C", self.tree, "add", "."], check=True)
        self.main = self.commit("the darwin map")
        self.legs = os.path.join(self.tree, run.darwinLegsDirectory)
        # The gate tools' tools.json beside the records, untracked as they are, its keys in another order than a record's.
        self.declareTools(json.dumps(self.declaredTools))
        self.listed, self.commands = None, []
        stub = mock.patch.object(run, "darwinClosureListing", side_effect=self.listing)
        stub.start()
        self.addCleanup(stub.stop)

    def listing(self, tree, command):
        """The stubbed go list of the leg's closure, recording command in self.commands and answering by its patterns:
        for ./internal/apple/..., self.closure's directories inside tree, one twice as -test's variants repeat a
        package's; for each pattern after it, that directory; and three the digest drops, the standard library's, a
        module cache's and a sibling whose name tree's is a prefix of. self.listed, when set, is the answer instead."""
        self.commands.append(list(command))
        if self.listed is not None:
            return self.listed
        named = []
        if "./internal/apple/..." in command:
            named = list(self.closure + self.closure[:1]) + [pattern[len("./"):] for pattern in command[command.index("./internal/apple/...") + 1:]]
        outside = ["/usr/local/go/src/fmt", os.path.join(self.directory, "gopath/pkg/mod/example.com/x@v1.0.0"), self.tree + "-binaries/x"]
        directories = [os.path.join(tree, directory) for directory in named] + outside
        return subprocess.CompletedProcess(command, 0, "".join(directory + "\n" for directory in directories), "")

    @staticmethod
    def mapOf(packages, version=1):
        return json.dumps({"version": version, "packages": packages})

    def mapCommit(self, source):
        """A commit on self.main whose compiler-dependencies.json is source (None deletes it), left checked out."""
        realRun(["git", "-C", self.tree, "checkout", "-q", "--detach", self.main], check=True)
        if source is None:
            realRun(["git", "-C", self.tree, "rm", "-q", "--", run.compilerDependenciesPath], check=True)
        else:
            self.write(run.compilerDependenciesPath, source)
            realRun(["git", "-C", self.tree, "add", "--", run.compilerDependenciesPath], check=True)
        return self.commit("the compiler map moved")

    def toolsWith(self, source):
        """A gate tools checkout apart from the tree: the tree's executors.txt, and compiler-dependencies.json as source
        (None for none)."""
        tools = tempfile.mkdtemp(dir=self.directory)
        os.makedirs(os.path.join(tools, "cloud/fast-gate"))
        shutil.copyfile(os.path.join(self.tree, "cloud/fast-gate/executors.txt"), os.path.join(tools, "cloud/fast-gate/executors.txt"))
        if source is not None:
            with open(os.path.join(tools, run.compilerDependenciesPath), "w") as handle:
                handle.write(source)
        return tools

    @contextlib.contextmanager
    def checkedOut(self, commit):
        """The tree checked out at commit for the block, as a hand computing a record's darwinPaths stands at its tip
        (darwinPathsDigest refuses any other checkout), then back where it was."""
        was = run.git(self.tree, "rev-parse", "HEAD")
        realRun(["git", "-C", self.tree, "checkout", "-q", "--detach", commit], check=True)
        try:
            yield
        finally:
            realRun(["git", "-C", self.tree, "checkout", "-q", "--detach", was], check=True)

    def digestAt(self, commit):
        """darwinPathsDigest at commit, computed in a checkout at it."""
        with self.checkedOut(commit):
            return run.darwinPathsDigest(self.tree, commit, self.tree)

    def objectAt(self, commit, path):
        return run.git(self.tree, "rev-parse", "%s:%s" % (commit, path))

    def write(self, name, source):
        os.makedirs(os.path.dirname(os.path.join(self.tree, name)), exist_ok=True)
        with open(os.path.join(self.tree, name), "w") as handle:
            handle.write(source)

    def commit(self, message):
        realRun(["git", "-C", self.tree, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", message], check=True)
        self.sha = run.git(self.tree, "rev-parse", "HEAD")
        return self.sha

    def on(self, parent, path, source, message):
        # Only the edited path is added: the records are the tools checkout's, untracked here, never the candidate's.
        realRun(["git", "-C", self.tree, "checkout", "-q", "--detach", parent], check=True)
        self.write(path, source)
        realRun(["git", "-C", self.tree, "add", "--", path], check=True)
        return self.commit(message)

    def candidate(self, mainEdit=None):
        """(base, tip): the change's tip, cut from self.main, and the main the candidate is rebased onto, which took
        mainEdit (a path and its source) after the cut. The tree is left at the candidate, self.sha."""
        tip = self.on(self.main, "internal/apple/apple.go", self.edited, "the change")
        base = self.main if mainEdit is None else self.on(self.main, mainEdit[0], mainEdit[1], "main moved")
        self.on(base, "internal/apple/apple.go", self.edited, "the change, rebased")
        return base, tip

    def record(self, at, name=None, **fields):
        """A green record of the leg at commit at, at its darwinPaths (computed in a checkout at it), with fields over
        it (None drops one), as darwin-legs/name (<at>.json unless given)."""
        record = {"tip": at, "host": "Shackler", "claim": "C-probe", "log": {"sha256": "a" * 64, "kept": "Shackler:~/legs/probe.log"},
                  "command": "go test ./internal/apple/...", "tools": dict(reversed(list(self.declaredTools.items()))),
                  "counts": {"passed": 12, "failed": 0, "skipped": 1}, "exit": 0, "darwinPaths": self.digestAt(at)}
        record = {key: value for key, value in dict(record, **fields).items() if value is not None}
        os.makedirs(self.legs, exist_ok=True)
        name = name or at + ".json"
        with open(os.path.join(self.legs, name), "w") as handle:
            json.dump(record, handle)
        return name

    def clearLegs(self):
        """Every record gone; the gate tools' tools.json stays."""
        for name in os.listdir(self.legs) if os.path.isdir(self.legs) else []:
            if name != run.darwinToolsName:
                os.remove(os.path.join(self.legs, name))

    def declareTools(self, source):
        """The gate tools' darwin-legs/tools.json as source, or none at all for None."""
        path = os.path.join(self.tree, run.darwinToolsPath)
        os.makedirs(os.path.dirname(path), exist_ok=True)
        if source is None:
            os.remove(path)
        else:
            with open(path, "w") as handle:
                handle.write(source)

    def assertNeedsLeg(self, gate, status, digest, why):
        self.assertTrue(status.startswith("red:"), status)
        self.assertEqual(gate.failure["step"], "darwin")
        self.assertEqual(gate.failure["detail"],
                         "needs a darwin leg on a pool Mac (Shackler), recorded in cloud/fast-gate/darwin-legs/<tip>.json at darwinPaths %s, "
                         "for internal/apple/apple.go: %s" % (digest, why))
        self.assertNotIn("Kirk", gate.failure["detail"])

    def test_a_recorded_leg_at_the_candidate_s_darwin_paths_passes_the_gate(self):
        # Every line alike: the candidate's tree is its tip's, so each dir, file and darwin entry line is the same.
        base, tip = self.candidate()
        self.record(tip)
        summary = {"record": tip + ".json", "tip": tip, "host": "Shackler", "claim": "C-probe", "log": "a" * 64,
                   "darwinPaths": self.digestAt(tip)}
        for full in (False, True):
            with self.subTest(full=full):
                gate, status, result = self.gate(base=base, full=full)
                self.assertTrue(status.startswith("green:"), status)
                self.assertIn("darwin", result["planned_stages"])
                self.assertEqual((result["stages_exit"]["coverage"], result["stages_exit"]["darwin"]), (0, 0))
                self.assertEqual(result["darwin_leg"], summary)
        # The pool's units: coverage alone keeps darwin's passed exit off its record and stays green, and the darwin unit
        # passes on the same record.
        for phase in ("coverage", "darwin"):
            with self.subTest(phase=phase):
                gate, status, result = self.gate(base=base, extra={"phase": phase})
                self.assertTrue(status.startswith("green:"), status)
                self.assertEqual(result["stages_exit"], {phase: 0})
                self.assertEqual(set(result["steps_seconds"]), {phase})
                self.assertEqual(result["darwin_leg"], summary)

    def test_main_moving_inside_the_darwin_paths_after_the_leg_is_red_naming_the_pool_mac_and_the_digest(self):
        # The darwin entry lines, and internal/apple's dir line with them.
        base, tip = self.candidate(("internal/apple/moved.go", "package apple\n\nvar Moved = 1\n"))
        self.record(tip)
        digest = run.darwinPathsDigest(self.tree, self.sha, self.tree)
        self.assertNotEqual(digest, self.digestAt(tip))
        for extra in ({}, {"phase": "coverage"}, {"phase": "darwin"}):
            with self.subTest(extra=extra):
                gate, status, result = self.gate(base=base, extra=extra)
                self.assertNeedsLeg(gate, status, digest, "no darwin leg recorded for darwinPaths " + digest)
                self.assertNotIn("darwin_leg", result)
                if extra.get("phase") == "coverage":
                    self.assertEqual(result["stages_exit"], {"coverage": 1})
                else:
                    self.assertEqual(result["stages_exit"]["darwin"], 1)

    def test_main_moving_elsewhere_still_matches(self):
        # Every line names a leg input alone: a note, and a package go list doesn't name outside the darwin globs (the
        # dir lines' listed directories only), keep the digest and the record.
        for why, path, source in (("a note", "notes/main.md", "# main, moved\n"),
                                  ("a package outside the closure", "internal/unrelated/unrelated.go", "package unrelated\n\nvar Moved = 1\n")):
            with self.subTest(why):
                self.clearLegs()
                base, tip = self.candidate((path, source))
                self.assertNotEqual(run.git(self.tree, "rev-parse", tip + "^{tree}"), run.git(self.tree, "rev-parse", self.sha + "^{tree}"))
                self.assertEqual(run.darwinPathsDigest(self.tree, self.sha, self.tree), self.digestAt(tip))
                self.record(tip)
                gate, status, result = self.gate(base=base)
                self.assertTrue(status.startswith("green:"), status)
                self.assertEqual(result["darwin_leg"]["record"], tip + ".json")

    def test_a_record_that_isnt_green_never_passes(self):
        base, tip = self.candidate()
        digest = run.darwinPathsDigest(self.tree, self.sha, self.tree)
        for why, fields, reason in (
                ("exit 1", {"exit": 1}, "exit is 1, not 0"),
                ("exit false, which Python calls 0", {"exit": False}, "exit is False, not 0"),
                ("exit as a string", {"exit": "0"}, "exit is '0', not 0"),
                ("a failed test", {"counts": {"passed": 12, "failed": 1, "skipped": 0}}, "counts.failed is 1, not 0"),
                ("no test passed", {"counts": {"passed": 0, "failed": 0, "skipped": 12}}, "counts.passed is 0, not above 0"),
                ("no counts", {"counts": None}, "counts isn't an object"),
                ("both claim and job", {"job": "J-probe"}, "names claim and job, not exactly one of claim and job"),
                ("neither claim nor job", {"claim": None}, "names neither, not exactly one of claim and job"),
                ("an empty claim", {"claim": " "}, "claim isn't a non-empty string"),
                ("a short log digest", {"log": {"sha256": "a" * 63}}, "log.sha256 isn't 64 lowercase hex"),
                ("an uppercase log digest", {"log": {"sha256": "A" * 64}}, "log.sha256 isn't 64 lowercase hex"),
                ("no log", {"log": None}, "log.sha256 isn't 64 lowercase hex"),
                ("an uppercase tip", {"tip": tip.upper(), "name": tip + ".json"}, "tip isn't 40 lowercase hex"),
                ("a tip that isn't its file name", {"name": "b" * 40 + ".json"}, "tip %s isn't its file name" % tip),
                # The tools check against tools.json (darwinTools): equal values, key by key, every key present and no other.
                ("an sdk the gate tools don't declare", {"tools": dict(self.declaredTools, sdk="macOS 27.1 (26A1)")},
                 "tools.sdk is 'macOS 27.1 (26A1)', the gate tools declare 'macOS 27.2 (26B118)'"),
                ("an xcode the gate tools don't declare", {"tools": dict(self.declaredTools, xcode="Xcode 27.1 (27A1)")},
                 "tools.xcode is 'Xcode 27.1 (27A1)', the gate tools declare 'Xcode 27.2 (27B52)'"),
                ("no tools", {"tools": None}, "tools isn't an object"),
                ("tools as a list", {"tools": list(self.declaredTools.values())}, "tools isn't an object"),
                ("a tool missing", {"tools": {key: value for key, value in self.declaredTools.items() if key != "node"}},
                 "tools.node is missing, the gate tools declare 'v24.19.0'"),
                ("a tool the gate tools don't name", {"tools": dict(self.declaredTools, cpu="M4")}, "tools.cpu is 'M4', which the gate tools don't declare"),
                ("aCheck as a list", {"aCheck": [{"exit": 0, "errors": ""}]}, "aCheck isn't an object"),
                ("an aCheck entry without exit", {"aCheck": {"examples/apple/app.a": {"errors": ""}}},
                 "aCheck examples/apple/app.a isn't an object of a whole number exit and a string errors"),
                ("an aCheck exit as a bool", {"aCheck": {"examples/apple/app.a": {"exit": False, "errors": ""}}},
                 "aCheck examples/apple/app.a isn't an object of a whole number exit and a string errors"),
                ("aCheck errors that aren't a string", {"aCheck": {"examples/apple/app.a": {"exit": 0, "errors": None}}},
                 "aCheck examples/apple/app.a isn't an object of a whole number exit and a string errors"),
                ("an aCheck entry that isn't an object", {"aCheck": {"examples/apple/app.a": 0}},
                 "aCheck examples/apple/app.a isn't an object of a whole number exit and a string errors")):
            with self.subTest(why):
                self.clearLegs()
                name = self.record(tip, **fields)
                self.assertEqual(run.darwinLegFor(self.tree, self.sha, self.tree), (None, "%s isn't green: %s" % (name, reason)))
                gate, status, result = self.gate(base=base)
                self.assertNeedsLeg(gate, status, digest, "%s isn't green: %s" % (name, reason))
        # What can't be read holds no digest: a file that isn't JSON, a JSON list, a record with no darwinPaths.
        self.clearLegs()
        with open(os.path.join(self.legs, tip + ".json"), "w") as handle:
            handle.write('{"tip": "' + tip + '", ')
        with open(os.path.join(self.legs, "c" * 40 + ".json"), "w") as handle:
            handle.write(json.dumps([digest]))
        self.record(tip, name="d" * 40 + ".json", darwinPaths=None)
        self.assertEqual(run.darwinLegFor(self.tree, self.sha, self.tree), (None, "no darwin leg recorded for darwinPaths " + digest))
        # The control: the same record whole is green, and a red record sorted before it doesn't hide it.
        self.clearLegs()
        self.record(tip, name="0" * 40 + ".json", tip="0" * 40, exit=1)
        self.assertEqual(run.darwinLegFor(self.tree, self.sha, self.tree)[0], None)
        self.record(tip)
        self.assertEqual(run.darwinLegFor(self.tree, self.sha, self.tree)[0], tip + ".json")

    def test_no_record_is_red_naming_the_pool_mac(self):
        base, tip = self.candidate()
        digest = run.darwinPathsDigest(self.tree, self.sha, self.tree)
        self.assertEqual(os.listdir(self.legs), [run.darwinToolsName])
        for full in (False, True):
            with self.subTest(full=full):
                gate, status, result = self.gate(base=base, full=full)
                self.assertNeedsLeg(gate, status, digest, "no darwin leg recorded for darwinPaths " + digest)
                self.assertEqual(result["stages_exit"]["darwin"], 1)
                self.assertEqual(result["needs_darwin"], ["internal/apple/apple.go"])

    def test_a_missing_record_exits_four_and_a_failed_one_stays_red(self):
        # #ycqnr4b, Adamic Loom's ruling: a record missing for the candidate is the evidence missing, never the change's
        # red, so the gate exits DarwinLegMissingExit, which the judge reads from a phase as void darwinRecordMissing. A
        # record at the candidate's digest whose tests failed on the Mac stays exit 1. Mutants, each red here: every miss
        # read as missing (darwinLegFailed dropped); every miss read as failed; a later failure beside a missing record
        # not counted; a-check's paths that await a record counted as faulted, or its faulted ones not.
        base, tip = self.candidate()
        digest = run.darwinPathsDigest(self.tree, self.sha, self.tree)
        self.assertEqual(run.DarwinLegMissingExit, 4)
        retooled = dict(self.declaredTools, xcode="Xcode 27.1 (27A11)")
        for why, fields, exit, named in (
                ("no record", None, run.DarwinLegMissingExit, "no darwin leg recorded for darwinPaths " + digest),
                ("a record whose tools differ", {"tools": retooled}, run.DarwinLegMissingExit, "tools.xcode is 'Xcode 27.1 (27A11)'"),
                ("a record out of shape", {"log": {"sha256": "short"}}, run.DarwinLegMissingExit, "log.sha256 isn't 64 lowercase hex"),
                ("a record whose leg exited 1", {"exit": 1}, 1, "exit is 1, not 0"),
                ("a record whose tests failed", {"counts": {"passed": 11, "failed": 1, "skipped": 1}}, 1, "counts.failed is 1, not 0"),
                ("a failed record under other tools", {"exit": 1, "tools": retooled}, 1, "exit is 1, not 0")):
            self.clearLegs()
            if fields is not None:
                self.record(tip, **fields)
            for extra in ({}, {"phase": "darwin"}, {"phase": "coverage"}):
                with self.subTest(why, extra=extra):
                    gate, status, result = self.gate(base=base, extra=extra)
                    self.assertTrue(status.startswith("red:"), status)
                    self.assertIn(digest, gate.failure["detail"])
                    self.assertIn(named, gate.failure["detail"])
                    self.assertEqual(gate.exitCode(), exit)
        # Any other failure beside a missing record is the change's red, whichever came first: a later one on the gate
        # that missed, and one before the miss on a gate of its own.
        self.clearLegs()
        gate, status, result = self.gate(base=base)
        self.assertEqual(gate.exitCode(), run.DarwinLegMissingExit)
        with mock.patch("builtins.print"):
            gate.fail("tests", "a planted failure after the miss")
        self.assertEqual(gate.exitCode(), 1)
        bare = run.Gate.__new__(run.Gate)
        bare.arguments = types.SimpleNamespace(tree=self.tree, out=self.directory, sha=self.sha, tools=self.tree, full=False, run_to_end=False)
        bare.result, bare.failure, bare.stopped, bare.complete = {}, None, None, False
        bare.lock, bare.started, bare.killSessions, bare.status = threading.Lock(), 0.0, lambda: None, lambda *arguments: None
        with mock.patch("builtins.print"):
            bare.fail("tests", "a planted failure before the miss")
            bare.fail("darwin", "needs a darwin leg", evidenceMissing=True)
        self.assertEqual(bare.exitCode(), 1)
        # A record at another digest is no record: missing, never failed.
        self.clearLegs()
        self.record(tip, darwinPaths="f" * 64, exit=1)
        self.assertEqual(run.darwinLegVerdict(self.tree, self.sha, self.tree)[2], False)
        # The control: a green record is exit 0.
        self.clearLegs()
        self.record(tip)
        gate, status, result = self.gate(base=base)
        self.assertEqual(gate.exitCode(), 0)

    def test_a_check_waits_on_a_missing_record_and_reds_on_a_failed_one(self):
        # A darwin path stopped on the platform with no green leg: the record missing is evidence missing, so a-check's
        # failure says so (failedBeyondEvidence unset); a record whose tests failed, or no digest, is red as today.
        base, tip = self.candidate()
        self.clearLegs()
        gate = self.aCheck("examples/apple/app.a")
        self.assertEqual(gate.exits["a-check"], 1)
        self.assertFalse(getattr(gate, "failedBeyondEvidence", False))
        self.record(tip, exit=1)
        gate = self.aCheck("examples/apple/app.a")
        self.assertEqual(gate.exits["a-check"], 1)
        self.assertTrue(getattr(gate, "failedBeyondEvidence", False))
        # A path red on its own, beside the platform stop, is faulted however the leg reads.
        self.clearLegs()
        gate = self.aCheck("probes/probe.a", returncode=1, errors="type error TS2322\n")
        self.assertEqual(gate.exits["a-check"], 1)
        self.assertTrue(getattr(gate, "failedBeyondEvidence", False))

    def digestMode(self, sha, at=None):
        """What run.py --darwin-paths-digest at sha says, run in this process so the stubbed go list is the one it reads,
        from a checkout at at (sha unless given): its exit's code (None when it returns) and what it printed."""
        argv = ["run.py", "--darwin-paths-digest", "--tree", self.tree, "--sha", sha, "--tools", self.tree]
        printed, code = io.StringIO(), None
        with self.checkedOut(at or sha), mock.patch.object(sys, "argv", argv), contextlib.redirect_stdout(printed):
            try:
                run.darwinMain()
            except SystemExit as exited:
                code = exited.code
        return code, printed.getvalue()

    def test_the_digest_mode_prints_the_digest_the_gate_matches(self):
        # Pins the definition whole, every line rebuilt by hand as the README gives it: the dir lines of the listed
        # directories inside the tree, deduplicated, by their tree objects; the file lines of the module files present
        # (go.work and go.work.sum aren't); the darwin entry lines.
        base, tip = self.candidate()
        code, printed = self.digestMode(tip)
        self.assertIsNone(code)
        self.assertRegex(printed, r"^[0-9a-f]{64}\n$")
        # The listing named ./internal/apple/... and then internal/apple's declared compilers, so cmd/adamic is a dir line.
        self.assertEqual(self.commands[-1], self.declaredCommand)
        digest = printed.strip()
        lines = sorted(["100644 %s %s\n" % (self.objectAt(tip, path), path)
                        for path in ("internal/apple/apple.go", "examples/apple/app.a", "examples/apple/refused.a", "examples/apple/exempt/vendored.a")]
                       + ["dir %s %s\n" % (directory, self.objectAt(tip, directory)) for directory in ("internal/apple", "internal/lower", "cmd/adamic")]
                       + ["file %s %s\n" % (path, self.objectAt(tip, path)) for path in ("go.mod", "go.sum")])
        self.assertEqual(digest, hashlib.sha256("".join(lines).encode()).hexdigest())
        gate, status, result = self.gate(base=base)
        self.assertNeedsLeg(gate, status, digest, "no darwin leg recorded for darwinPaths " + digest)
        self.record(tip, darwinPaths=digest)
        gate, status, result = self.gate(base=base)
        self.assertTrue(status.startswith("green:"), status)
        self.assertEqual(result["darwin_leg"]["darwinPaths"], digest)

    def test_the_a_paths_mode_prints_every_darwin_a_a_check_could_be_handed(self):
        base, tip = self.candidate()
        printed = realRun([sys.executable, run.__file__, "--darwin-a-paths", "--tree", self.tree, "--sha", tip, "--tools", self.tree],
                          capture_output=True, text=True, timeout=60)
        self.assertEqual((printed.returncode, printed.stderr), (0, ""))
        # Not the Go file, not a .a outside the darwin globs, not one an a-check-exempt line names.
        self.assertEqual(printed.stdout.splitlines(), ["examples/apple/app.a", "examples/apple/refused.a"])
        self.assertEqual(run.darwinAPaths(self.tree, tip, self.tree), ["examples/apple/app.a", "examples/apple/refused.a"])

    def test_main_moving_inside_the_closure_is_red(self):
        # The dir lines: internal/lower is no darwin path, so the darwin entry lines alone match, but go list names it,
        # and its tree object, subdirectories and all, moves with a source of it or a file below it.
        rules = run.executorMap(self.tree)[0]
        for why, path, source in (("a source of a listed package", "internal/lower/lower.go", "package lower\n\nvar Moved = 1\n"),
                                  ("a file below a listed directory", "internal/lower/testdata/case.txt", "moved\n")):
            with self.subTest(why):
                self.clearLegs()
                base, tip = self.candidate((path, source))
                self.assertEqual(run.darwinEntries(run.commitEntries(self.tree, self.sha), rules), run.darwinEntries(run.commitEntries(self.tree, tip), rules))
                self.record(tip)
                digest = run.darwinPathsDigest(self.tree, self.sha, self.tree)
                self.assertNotEqual(digest, self.digestAt(tip))
                self.assertEqual(run.darwinLegFor(self.tree, self.sha, self.tree), (None, "no darwin leg recorded for darwinPaths " + digest))
                for extra in ({}, {"phase": "coverage"}, {"phase": "darwin"}):
                    with self.subTest(extra=extra):
                        gate, status, result = self.gate(base=base, extra=extra)
                        self.assertNeedsLeg(gate, status, digest, "no darwin leg recorded for darwinPaths " + digest)
                        self.assertNotIn("darwin_leg", result)

    def assertNoDigest(self, gate, status, why):
        self.assertTrue(status.startswith("red:"), status)
        self.assertEqual(gate.failure["step"], "darwin")
        self.assertEqual(gate.failure["detail"], "no darwin leg can match %s, since its darwinPaths can't be computed, for "
                                                 "internal/apple/apple.go: %s" % (self.sha, why))

    def test_go_list_failing_is_red_naming_it_never_green(self):
        # The raise before any line: a listing that fails, names nothing, or names nothing inside the tree is never a
        # digest, so a record that matched the moment before passes nothing, and no reader of the digest crashes on it.
        base, tip = self.candidate()
        self.record(tip)
        self.assertEqual(run.darwinLegFor(self.tree, self.sha, self.tree)[0], tip + ".json")
        # A refusal names the whole command that ran, the declared compilers among its patterns.
        words = "GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 GOFLAGS=-mod=readonly go list -deps -test -f {{.Dir}} ./internal/apple/... ./cmd/adamic ./internal/lower"
        stderr = "go: updates to go.mod needed\n\tgo mod tidy\nthird\nfourth\nfifth\nsixth, past the first lines\n"
        for why, listed, refusal in (
                ("go list failing", subprocess.CompletedProcess([], 1, "", stderr),
                 "%s failed in %s, exit 1: go: updates to go.mod needed; go mod tidy; third; fourth; fifth" % (words, self.tree)),
                ("an empty listing", subprocess.CompletedProcess([], 0, "", ""), "%s named no directory inside %s: no stderr" % (words, self.tree)),
                ("a listing of the standard library alone", subprocess.CompletedProcess([], 0, "/usr/local/go/src/fmt\n", ""),
                 "%s named no directory inside %s: no stderr" % (words, self.tree))):
            with self.subTest(why):
                self.listed = listed
                with self.assertRaises(run.DarwinPathsRefusal) as raised:
                    run.darwinPathsDigest(self.tree, self.sha, self.tree)
                self.assertEqual(str(raised.exception), refusal)
                for full in (False, True):
                    for extra in ({}, {"phase": "coverage"}, {"phase": "darwin"}) if not full else ({},):
                        with self.subTest(full=full, extra=extra):
                            gate, status, result = self.gate(base=base, full=full, extra=extra)
                            self.assertNoDigest(gate, status, refusal)
                            self.assertNotIn("darwin_leg", result)
                            if extra.get("phase") == "coverage":
                                self.assertEqual(result["stages_exit"], {"coverage": 1})
                            else:
                                self.assertEqual(result["stages_exit"]["darwin"], 1)
                # a-check's darwin branch reads the refusal as no leg: the path keeps its live red, as with no record.
                gate = self.aCheck("examples/apple/app.a")
                self.assertEqual(gate.exits["a-check"], 1)
                self.assertEqual(gate.result["a_check"]["examples/apple/app.a"], {"outcome": "failed", "expected": "checked", "first": self.missing.strip()})
                # The digest mode prints nothing and exits nonzero, saying why.
                self.assertEqual(self.digestMode(tip), ("no darwinPaths: " + refusal, ""))

    def test_a_tree_not_at_the_commit_is_refused(self):
        # The checkout check, before go list runs: go list reads the files on disk, so a digest of another checkout
        # would be that checkout's, never commit's.
        base, tip = self.candidate()
        self.assertNotEqual(tip, self.sha)
        with self.assertRaises(run.DarwinPathsRefusal) as raised:
            run.darwinPathsDigest(self.tree, tip, self.tree)
        self.assertEqual(str(raised.exception), "%s is checked out at %s, not at %s, and go list reads the checkout's files, so its "
                                                "darwinPaths would be another commit's" % (self.tree, self.sha, tip))
        with self.assertRaises(run.DarwinPathsRefusal) as raised:
            run.darwinPathsDigest(self.tree, "f" * 40, self.tree)
        self.assertTrue(str(raised.exception).startswith("%s isn't a commit %s can name: " % ("f" * 40, self.tree)), raised.exception)
        # The digest mode refuses the same, printing nothing, where a hand runs it outside a checkout at the tip.
        self.assertEqual(self.digestMode(tip, at=self.sha), ("no darwinPaths: %s is checked out at %s, not at %s, and go list reads the "
                                                             "checkout's files, so its darwinPaths would be another commit's" % (self.tree, self.sha, tip), ""))
        # The same commit, from a checkout at it, is a digest.
        self.assertRegex(self.digestAt(tip), r"^[0-9a-f]{64}$")

    def test_a_gitlink_in_the_listing_contributes_its_commit_object(self):
        # The dir lines' gitlink case: a submodule's directory, and one inside it, which commit's tree holds nothing of,
        # are one line, the gitlink's commit object, and moving the submodule moves the digest.
        realRun(["git", "-C", self.tree, "checkout", "-q", "--detach", self.main], check=True)
        realRun(["git", "-C", self.tree, "update-index", "--add", "--cacheinfo", "160000,%s,internal/vendored" % self.main], check=True)
        linked = self.commit("a submodule")
        os.makedirs(os.path.join(self.tree, "internal/vendored/inner"), exist_ok=True)
        self.closure = ("internal/apple", "internal/lower", "internal/vendored", "internal/vendored/inner")
        lines = sorted(["100644 %s %s\n" % (self.objectAt(linked, path), path)
                        for path in ("internal/apple/apple.go", "examples/apple/app.a", "examples/apple/refused.a", "examples/apple/exempt/vendored.a")]
                       + ["dir %s %s\n" % (directory, self.objectAt(linked, directory)) for directory in ("internal/apple", "internal/lower", "cmd/adamic")]
                       + ["dir internal/vendored %s\n" % self.main]
                       + ["file %s %s\n" % (path, self.objectAt(linked, path)) for path in ("go.mod", "go.sum")])
        self.assertEqual(run.darwinPathsDigest(self.tree, linked, self.tree), hashlib.sha256("".join(lines).encode()).hexdigest())
        realRun(["git", "-C", self.tree, "update-index", "--cacheinfo", "160000,%s,internal/vendored" % linked], check=True)
        moved = self.commit("the submodule moved")
        self.assertEqual(self.objectAt(moved, "internal/vendored"), linked)
        self.assertNotEqual(run.darwinPathsDigest(self.tree, moved, self.tree), self.digestAt(linked))

    def test_main_moving_the_declared_compiler_is_red(self):
        # The dir lines of the declared compilers: internal/apple's tests build cmd/adamic and run it, a main package go
        # list -deps of ./internal/apple/... never names, so only the map's declaration puts it in the listing.
        rules = run.executorMap(self.tree)[0]
        base, tip = self.candidate(("cmd/adamic/main.go", "package main\n\nvar Moved = 1\n"))
        self.assertEqual(run.darwinEntries(run.commitEntries(self.tree, self.sha), rules), run.darwinEntries(run.commitEntries(self.tree, tip), rules))
        self.record(tip)
        self.assertEqual(self.commands[-1], self.declaredCommand)
        digest = run.darwinPathsDigest(self.tree, self.sha, self.tree)
        self.assertNotEqual(digest, self.digestAt(tip))
        for extra in ({}, {"phase": "coverage"}, {"phase": "darwin"}):
            with self.subTest(extra=extra):
                gate, status, result = self.gate(base=base, extra=extra)
                self.assertNeedsLeg(gate, status, digest, "no darwin leg recorded for darwinPaths " + digest)
                self.assertNotIn("darwin_leg", result)

    def test_the_listing_names_internal_apple_s_compilers_from_the_map_the_gate_reads(self):
        # The listing's patterns: the tools' map with the tree's laid over it, as the census reads it (compilerDeclarations),
        # every package that is internal/apple or below it, and no other.
        base = run.darwinClosureCommand()
        self.assertEqual(base, ["go", "list", "-deps", "-test", "-f", "{{.Dir}}", "./internal/apple/..."])
        for why, toolsMap, treeMap, command in (
                ("the tools' declaration alone", self.mapOf({"internal/apple": ["cmd/adamic", "internal/lower"]}),
                 self.mapOf({"bench": ["internal/ir"]}), self.declaredCommand),
                ("a declaration the tree adds joins it", self.mapOf({"internal/apple": ["cmd/adamic", "internal/lower"]}),
                 self.mapOf({"internal/apple/generate": ["internal/ir", "cmd/adamic"], "internal/applesauce": ["internal/unrelated"]}),
                 base + ["./cmd/adamic", "./internal/ir", "./internal/lower"]),
                ("the tree's declaration over the tools'", self.mapOf({"internal/apple": ["cmd/adamic", "internal/lower"]}),
                 self.mapOf({"internal/apple": ["internal/ir"]}), base + ["./internal/ir"]),
                ("the tree's map alone", None, self.mapOf({"internal/apple": ["cmd/adamic"]}), base + ["./cmd/adamic"]),
                # No internal/apple package declared: the listing is ./internal/apple/... alone, the map-less shape.
                ("no internal/apple package declared", self.mapOf({"bench": ["cmd/adamic"]}), self.mapOf({"internal/applesauce": ["internal/ir"]}), base)):
            with self.subTest(why):
                tools, sha = self.toolsWith(toolsMap), self.mapCommit(treeMap)
                run.darwinPathsDigest(self.tree, sha, tools)
                self.assertEqual(self.commands[-1], command)

    def test_a_compiler_map_that_can_t_be_read_is_refused_and_the_darwin_phase_red(self):
        # The raise before the listing: a map that can't be read never yields a digest computed without the compilers.
        cannot = "the compiler dependency map can't be read, so the compilers the darwin leg builds can't be listed: "
        good = self.mapOf({"internal/apple": ["cmd/adamic"]})
        for why, toolsMap, treeMap, refusal in (
                ("the tree's map isn't JSON", good, "{", cannot),
                ("the tree's map isn't the tools' version", good, self.mapOf({"internal/apple": ["cmd/adamic"]}, version=2),
                 cannot + "invalid compiler dependency map: %s in %s" % (run.compilerDependenciesPath, self.tree)),
                ("a map that isn't version 1", self.mapOf({}, version=2), None, cannot + "invalid compiler dependency map: " + run.compilerDependenciesPath),
                ("no map in either", None, None, cannot + "invalid compiler dependency map: no " + run.compilerDependenciesPath),
                ("internal/apple's compilers aren't a list", good, self.mapOf({"internal/apple": "cmd/adamic"}),
                 "the compiler dependency map's internal/apple isn't a list of directories: 'cmd/adamic'")):
            with self.subTest(why):
                tools, sha = self.toolsWith(toolsMap), self.mapCommit(treeMap)
                del self.commands[:]
                with self.assertRaises(run.DarwinPathsRefusal) as raised:
                    run.darwinPathsDigest(self.tree, sha, tools)
                self.assertTrue(str(raised.exception).startswith(refusal), raised.exception)
                self.assertEqual(self.commands, [])
        # The gate's tools are the tree here, so a main whose map isn't JSON reds the darwin phase, naming why.
        base, tip = self.candidate((run.compilerDependenciesPath, "{"))
        with self.assertRaises(run.DarwinPathsRefusal) as raised:
            run.darwinPathsDigest(self.tree, self.sha, self.tree)
        self.assertTrue(str(raised.exception).startswith(cannot), raised.exception)
        for extra in ({}, {"phase": "coverage"}, {"phase": "darwin"}):
            with self.subTest(extra=extra):
                del self.commands[:]
                gate, status, result = self.gate(base=base, extra=extra)
                self.assertNoDigest(gate, status, str(raised.exception))
                self.assertEqual(self.commands, [])

    def test_a_record_matches_the_gate_tools_tools_json_in_any_key_order(self):
        # The tools check, order aside: a record's tools equal to tools.json pass whichever order either file writes.
        base, tip = self.candidate()
        orders = (list(self.declaredTools), list(reversed(list(self.declaredTools))), ["go", "xcode", "node", "sdk"])
        for declaredOrder in orders:
            for recordOrder in orders:
                with self.subTest(declared=declaredOrder, record=recordOrder):
                    self.clearLegs()
                    self.declareTools(json.dumps({key: self.declaredTools[key] for key in declaredOrder}))
                    self.record(tip, tools={key: self.declaredTools[key] for key in recordOrder})
                    self.assertEqual(run.darwinLegFor(self.tree, self.sha, self.tree)[0], tip + ".json")
        gate, status, result = self.gate(base=base, extra={"phase": "darwin"})
        self.assertTrue(status.startswith("green:"), status)
        self.assertEqual(result["darwin_leg"]["record"], tip + ".json")

    def test_tools_json_that_doesn_t_declare_four_strings_leaves_no_record_green(self):
        # darwinTools, read once a darwinLegFor: when the gate tools can't say what tools a record must match, no record
        # is green, the darwin phase is red naming why, and a-check keeps its live red.
        base, tip = self.candidate()
        self.record(tip)
        digest = run.darwinPathsDigest(self.tree, self.sha, self.tree)
        self.assertEqual(run.darwinLegFor(self.tree, self.sha, self.tree)[0], tip + ".json")
        path = run.darwinToolsPath
        for why, source, refusal in (
                ("tools.json missing", None, "%s can't be read: No such file or directory" % path),
                ("an extra key", json.dumps(dict(self.declaredTools, cpu="M4")), "%s names cpu, go, node, sdk, xcode, not exactly xcode, sdk, go and node" % path),
                ("a key missing", json.dumps({key: value for key, value in self.declaredTools.items() if key != "sdk"}),
                 "%s names go, node, xcode, not exactly xcode, sdk, go and node" % path),
                ("a value that isn't a string", json.dumps(dict(self.declaredTools, go=1.27)), "%s's go isn't a non-empty string" % path),
                ("an empty value", json.dumps(dict(self.declaredTools, node=" ")), "%s's node isn't a non-empty string" % path),
                ("a list", json.dumps(list(self.declaredTools.values())), "%s isn't a JSON object" % path),
                ("not JSON", "{", "%s isn't JSON: " % path)):
            with self.subTest(why):
                self.declareTools(source)
                with self.assertRaises(ValueError) as raised:
                    run.darwinTools(self.tree)
                self.assertTrue(str(raised.exception).startswith(refusal), raised.exception)
                expected = "no darwin leg record can be green: %s" % raised.exception
                self.assertEqual(run.darwinLegFor(self.tree, self.sha, self.tree), (None, expected))
                for extra in ({}, {"phase": "coverage"}, {"phase": "darwin"}):
                    with self.subTest(extra=extra):
                        gate, status, result = self.gate(base=base, extra=extra)
                        self.assertNeedsLeg(gate, status, digest, expected)
                        self.assertNotIn("darwin_leg", result)
                gate = self.aCheck("examples/apple/app.a")
                self.assertEqual(gate.exits["a-check"], 1)
                self.assertEqual(gate.result["a_check"]["examples/apple/app.a"], {"outcome": "failed", "expected": "checked", "first": self.missing.strip()})
                self.declareTools(json.dumps(self.declaredTools))
        # The control: the same record against the declared tools again is green.
        self.assertEqual(run.darwinLegFor(self.tree, self.sha, self.tree)[0], tip + ".json")

    missing = "adamic: load: darwin platform missing: Xcode SDK\n"

    def aCheck(self, path, returncode=run.platformMissingExit, errors=missing):
        gate = run.Gate.__new__(run.Gate)
        gate.arguments = types.SimpleNamespace(tree=self.tree, out=self.directory, sha=self.sha, tools=self.tree, full=False)
        gate.result, gate.steps, gate.exits, gate.failure = {}, {}, {}, None
        gate.lock, gate.started, gate.complete, gate.killSessions = threading.Lock(), 0.0, False, lambda: None

        def spawn(command, stdout, stderr=None, directory=None, environment=None):
            return types.SimpleNamespace(communicate=lambda: ("", errors), returncode=returncode)

        with mock.patch.object(gate, "step", lambda *arguments: True), mock.patch.object(gate, "spawn", spawn), mock.patch("builtins.print"):
            gate.aCheck([path])
        return gate

    def test_a_check_judges_a_darwin_program_stopped_on_the_platform_on_what_the_mac_recorded(self):
        base, tip = self.candidate()
        # Without a leg, and with one that isn't green, exactly today's red.
        for why, fields in (("no leg", None), ("a leg whose aCheck is malformed", {"aCheck": []})):
            with self.subTest(why):
                self.clearLegs()
                if fields is not None:
                    self.record(tip, **fields)
                gate = self.aCheck("examples/apple/app.a")
                self.assertEqual(gate.exits["a-check"], 1)
                self.assertEqual(gate.result["a_check"]["examples/apple/app.a"], {"outcome": "failed", "expected": "checked", "first": self.missing.strip()})
                self.assertIn("examples/apple/app.a: failed, expected checked (%s)" % self.missing.strip(), gate.failure["detail"])
        typeError = "examples/apple/app.a(1,1): error TS2322: Type 'string' is not assignable to type 'number'.\n"
        refusal = "examples/apple/refused.a:2:7: Adamic 0.1 refuses no-any: any isn't a type\n"
        unrecorded = "the darwin leg %s.json recorded no adamic c for this path" % tip
        # The recorded run, judged by a-check's own rules and the file's header, exactly as a run on this box.
        for why, path, aCheck, outcome, first, passes in (
                ("recorded clean", "examples/apple/app.a", {"examples/apple/app.a": {"exit": 0, "errors": ""}}, "checked", "", True),
                ("recorded a type error", "examples/apple/app.a", {"examples/apple/app.a": {"exit": 1, "errors": typeError}}, "type error", typeError.strip(), False),
                ("recorded the refusal its header expects", "examples/apple/refused.a", {"examples/apple/refused.a": {"exit": 1, "errors": refusal}}, "refused", refusal.strip(), True),
                ("recorded clean where its header expects a refusal", "examples/apple/refused.a", {"examples/apple/refused.a": {"exit": 0, "errors": ""}}, "checked", "", False),
                ("recorded another rule's refusal", "examples/apple/refused.a", {"examples/apple/refused.a": {"exit": 1, "errors": refusal.replace("no-any", "no-eval")}}, "refused",
                 refusal.replace("no-any", "no-eval").strip(), False),
                ("recorded the platform stop too", "examples/apple/app.a", {"examples/apple/app.a": {"exit": run.platformMissingExit, "errors": self.missing}}, "failed", self.missing.strip(), False),
                ("recorded nothing for the path", "examples/apple/app.a", {"examples/apple/refused.a": {"exit": 0, "errors": ""}}, "failed", unrecorded, False),
                ("recorded no aCheck at all", "examples/apple/app.a", None, "failed", unrecorded, False)):
            with self.subTest(why):
                self.clearLegs()
                self.record(tip, aCheck=aCheck)
                gate = self.aCheck(path)
                self.assertEqual(gate.result["a_check"][path], {"outcome": outcome, "expected": "refused no-any" if path.endswith("refused.a") else "checked",
                                                                "first": first, "source": "darwin leg %s.json" % tip})
                self.assertEqual(gate.exits["a-check"], 0 if passes else 1)
                if passes:
                    self.assertIsNone(gate.failure)
                else:
                    self.assertIn("%s: %s, expected %s (%s)" % (path, outcome, gate.result["a_check"][path]["expected"], first), gate.failure["detail"])

    def test_a_check_keeps_every_other_verdict_beside_a_leg(self):
        base, tip = self.candidate()
        self.record(tip, aCheck={path: {"exit": 0, "errors": ""} for path in ("examples/apple/app.a", "probes/probe.a")})
        # Only the typed stop on a darwin path reads the leg: a live refusal, a live type error, the words without the
        # stop's exit, and a path that isn't darwin's each keep the live verdict, red, even where the leg recorded clean.
        for why, path, returncode, errors, outcome in (
                ("a refusal", "examples/apple/app.a", 1, "examples/apple/app.a:1:1: Adamic 0.1 refuses x\n" + self.missing, "refused"),
                ("a type error", "examples/apple/app.a", 1, "examples/apple/app.a(1,1): error TS2322: x\n", "type error"),
                ("the words without the typed stop", "examples/apple/app.a", 1, self.missing, "failed"),
                ("a path that isn't darwin's", "probes/probe.a", run.platformMissingExit, self.missing, "failed")):
            with self.subTest(why):
                gate = self.aCheck(path, returncode, errors)
                self.assertEqual(gate.exits["a-check"], 1)
                self.assertEqual(gate.result["a_check"][path]["outcome"], outcome)
                self.assertNotIn("source", gate.result["a_check"][path])
        # The live classification and the header rules are the functions the recorded path judges by.
        self.assertEqual([run.aCheckOutcome(code, errors) for code, errors in ((0, ""), (1, "can't lower x yet"), (1, "Adamic 0.1 refuses r"), (1, "a error TS1"), (1, "?"))],
                         ["checked", "checked", "refused", "type error", "failed"])
        self.assertEqual([run.aCheckMeets(expect, outcome, errors) for expect, outcome, errors in (
            ("", "checked", ""), ("refused no-any", "refused", "refuses no-any"), ("refused no-any", "refused", "refuses no-eval"),
            ("type error TS1", "type error", "error TS1"), ("type error TS1", "type error", "error TS2"), ("type error", "type error", ""))],
            [True, True, False, True, False, True])


class GateRuns(unittest.TestCase):
    def test_a_trailer_runs_a_deferred_test_in_this_gate(self):
        with tempfile.TemporaryDirectory() as tree:
            commit = lambda message: realRun(["git", "-C", tree, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", message], check=True)
            realRun(["git", "init", "-q", tree], check=True)
            commit("base\n\nGate-runs: internal/native TestRecordMutants")
            base = run.git(tree, "rev-parse", "HEAD")
            commit("fix the split build\n\nGate-runs: internal/native TestSplitTSGoAgrees\nGate-runs: internal/native TestNotDeferred")
            commit("unrelated")
            gate = run.Gate.__new__(run.Gate)
            gate.arguments = types.SimpleNamespace(tree=tree, base=base, sha=run.git(tree, "rev-parse", "HEAD"))
            gate.result = {}
            gate.spawn = lambda command, stdout, stderr=None, directory=None, environment=None: realPopen(command, stdout=stdout, stderr=stderr, cwd=directory, text=True)
            native = run.module + "/internal/native"
            gate.deferred = {native: {"TestSplitTSGoAgrees", "TestRecordMutants", "TestWASI"}}
            gate.runRequested()
            # Only trailers since the base count, and only for deferred tests.
            self.assertEqual(gate.deferred[native], {"TestRecordMutants", "TestWASI"})
            self.assertEqual(gate.result["deferred_run_by_request"], [native + " TestSplitTSGoAgrees"])


class GateRunsDeferred(unittest.TestCase):
    def gate(self, trailer):
        tree = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, tree)
        realRun(["git", "init", "-q", tree], check=True)
        commit = lambda message: realRun(["git", "-C", tree, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "--allow-empty", "-m", message], check=True)
        commit("base")
        base = run.git(tree, "rev-parse", "HEAD")
        commit("landing\n\nGate-runs: " + trailer)
        gate = run.Gate.__new__(run.Gate)
        gate.arguments = types.SimpleNamespace(tree=tree, base=base, sha=run.git(tree, "rev-parse", "HEAD"), out=tree, full=False)
        gate.result, gate.failure, gate.lock, gate.started = {}, None, threading.Lock(), 0.0
        gate.complete, gate.killSessions = False, lambda: None
        gate.spawn = lambda command, stdout, stderr=None, directory=None, environment=None: realPopen(command, stdout=stdout, stderr=stderr, cwd=directory, text=True)
        native, lint = run.module + "/internal/native", run.module + "/stage1/cohere/lint"
        gate.deferred = {native: {"TestSplitTSGoAgrees", "TestWASI"}, lint: {"TestShardsAgree"}}
        return gate, native, lint

    def test_deferred_runs_every_deferred_test_and_each_must_reach_a_verdict(self):
        gate, native, lint = self.gate("deferred")
        gate.runRequested()
        self.assertEqual(gate.deferred, {})
        self.assertTrue(gate.result["deferred_all_requested"])
        self.assertEqual(sorted(gate.result["deferred_run_by_request"]),
                         [native + " TestSplitTSGoAgrees", native + " TestWASI", lint + " TestShardsAgree"])
        with open(os.path.join(gate.arguments.out, "test.jsonl"), "w") as handle:
            handle.write(json.dumps({"Action": "pass", "Package": native, "Test": "TestSplitTSGoAgrees"}) + "\n")
            handle.write(json.dumps({"Action": "skip", "Package": native, "Test": "TestWASI"}) + "\n")
        with mock.patch("builtins.print"):
            gate.requestedRan()
        self.assertEqual(gate.result["deferred_run_results"], {native + " TestSplitTSGoAgrees": "pass", native + " TestWASI": "skip", lint + " TestShardsAgree": "missing"})
        self.assertEqual(gate.failure["step"], "deferred")
        self.assertIn("TestWASI: skip", gate.failure["detail"])
        self.assertIn("TestShardsAgree: missing", gate.failure["detail"])

    def test_a_requested_name_is_its_split_s_family_never_a_bare_prefix(self):
        # Trio 114a6439, Oct 9: TestNormalizeMatchesNode runs as Points00-16, TestRecordMutants as Unit00-05, and an exact
        # anchor ran none of them. Ruled 12:19Z: the name plus the split's generated suffixes, never a bare prefix.
        gate, native, lint = self.gate("deferred")
        gate.runRequested()
        with open(os.path.join(gate.arguments.out, "test.jsonl"), "w") as handle:
            for package, test, action in ((native, "TestSplitTSGoAgreesUnit00", "pass"), (native, "TestSplitTSGoAgreesUnit01", "pass"),
                                          (native, "TestWASIUnit00", "skip"), (native, "TestWASIUnit01", "skip"),
                                          (native, "TestWASIRefusesUnsupportedOptions", "pass"), (native, "TestWASITargetFlags", "pass"),
                                          (lint, "TestShardsAgree_000", "pass")):
                handle.write(json.dumps({"Action": action, "Package": package, "Test": test}) + "\n")
        with mock.patch("builtins.print"):
            gate.requestedRan()
        results = gate.result["deferred_run_results"]
        self.assertEqual(results[native + " TestSplitTSGoAgrees"], "pass")
        self.assertEqual(results[lint + " TestShardsAgree"], "pass")
        # Over-match: two tests that only share TestWASI's prefix don't make it run while its 36 shards all skipped.
        self.assertEqual(results[native + " TestWASI"], "missing")
        self.assertEqual(gate.failure["step"], "deferred")
        # Under-match: Points00 is TestNormalizeMatchesNode; a failed shard fails the family; Contexts is no shard.
        self.assertTrue(run.inFamily("TestNormalizeMatchesNodePoints00", {"TestNormalizeMatchesNode"}))
        self.assertFalse(run.inFamily("TestNormalizeMatchesNodeContexts", {"TestNormalizeMatchesNode"}))
        self.assertFalse(run.inFamily("TestNormalize", {"TestNormalizeMatchesNode"}))
        self.assertEqual(run.familyOutcome({"p TestAUnit01": "pass", "p TestAUnit02": "fail"}, "p", "TestA"), "fail")
        self.assertEqual(run.familyOutcome({"p TestAUnit01": "pass"}, "p", "TestNoSuchFamily"), "missing")
        self.assertEqual(run.familyOutcome({"p TestA": "skip"}, "p", "TestA"), "skip")

    def test_every_requested_test_reaching_a_verdict_passes_the_check(self):
        gate, native, lint = self.gate("internal/native TestWASI")
        gate.runRequested()
        with open(os.path.join(gate.arguments.out, "test.jsonl"), "w") as handle:
            handle.write(json.dumps({"Action": "fail", "Package": native, "Test": "TestWASI"}) + "\n")
        gate.requestedRan()
        self.assertIsNone(gate.failure)
        self.assertEqual(gate.deferred[native], {"TestSplitTSGoAgrees"})


class DeletedAFiles(unittest.TestCase):
    def test_a_deleted_a_file_is_not_checked_and_does_not_crash_the_gate(self):
        with tempfile.TemporaryDirectory() as tree:
            with open(os.path.join(tree, "kept.a"), "w") as handle:
                handle.write("const x = 1\n")
            gate = run.Gate.__new__(run.Gate)
            gate.arguments = types.SimpleNamespace(tree=tree, out=tree)
            gate.result, gate.steps, gate.exits, gate.failure = {}, {}, {}, None
            ran = []

            def spawn(command, stdout, stderr=None, directory=None, environment=None):
                ran.append(command[-1])
                return types.SimpleNamespace(communicate=lambda: ("", ""), returncode=0)

            with mock.patch.object(gate, "step", lambda *arguments: True), mock.patch.object(gate, "spawn", spawn):
                gate.aCheck(["kept.a", "stage3/interface-downcasts/default-boxed-write.a"])
            self.assertIsNone(gate.failure)
            self.assertEqual(ran, ["kept.a"])
            self.assertEqual(gate.result["a_check"]["stage3/interface-downcasts/default-boxed-write.a"]["outcome"], "deleted")
            self.assertEqual(gate.exits["a-check"], 0)


class CompleteMode(unittest.TestCase):
    """Landing and area gates run on after a failure (run.py --complete): every test and fixture still
    runs and the verdict names them all; an ordinary fast gate still stops everything at the first."""

    def gate(self, complete):
        gate = run.Gate.__new__(run.Gate)
        arguments = types.SimpleNamespace(full=False, out=tempfile.mkdtemp(), sha="a" * 40)
        if complete:
            arguments.complete = True
        gate.arguments = arguments
        gate.complete = vars(arguments).get("complete") is True
        gate.lock = threading.Lock()
        gate.started = 0.0
        gate.failure = None
        gate.killed = 0
        gate.killSessions = lambda: setattr(gate, "killed", gate.killed + 1)
        return gate

    def test_complete_mode_keeps_running_after_the_first_failure(self):
        gate = self.gate(True)
        with mock.patch("builtins.print"):
            gate.fail("tests", "first")
            gate.fail("tests", "second")
        self.assertEqual(gate.killed, 0)
        self.assertEqual(gate.failure["detail"], "first")

    def test_an_ordinary_fast_gate_still_stops_at_the_first_failure(self):
        gate = self.gate(False)
        with mock.patch("builtins.print"):
            gate.fail("tests", "first")
        self.assertEqual(gate.killed, 1)

    def test_a_gate_tool_crash_stops_even_a_complete_run(self):
        # A complete run on broken tools teaches nothing (@system_adamic, Oct 8): a traceback stops it.
        for full in (False, True):
            gate = self.gate(True)
            gate.arguments.full = full
            gate.status = lambda line: None
            with mock.patch("builtins.print"):
                gate.fail("a-check", "Traceback (most recent call last):\n  File run.py\nFileNotFoundError: x.a")
            self.assertEqual(gate.killed, 1, "full=%s" % full)
            self.assertTrue(gate.failure["tool_crash"])
        gate = self.gate(True)
        with mock.patch("builtins.print"):
            gate.fail("tests", "p TestFailed\n")
        self.assertEqual(gate.killed, 0)
        self.assertNotIn("tool_crash", gate.failure)


class OracleSelection(unittest.TestCase):
    """A fixture-only change to internal/oracle runs its fixtures in the lanes, plus every test over a
    table of its own that names a changed fixture; a table filled outside its literal runs it whole."""

    lanes = 'package oracle\n\nvar fixtures = []string{\n\t"internal/oracle/testdata/a.a",\n\t"internal/oracle/testdata/c.a",\n}\n\nfunc TestNativeAgreesWithNode(t *testing.T) {\n\tfor _, fixture := range fixtures {\n\t}\n}\n'
    cast = 'package oracle\n\nvar castFixtures = []string{\n\t"c",\n}\n\nfunc TestCast(t *testing.T) {\n\tfor _, fixture := range castFixtures {\n\t}\n}\n'

    def setUp(self):
        self.tree = tempfile.mkdtemp()
        os.makedirs(os.path.join(self.tree, "internal/oracle/testdata"))
        self.git("init", "-q")
        self.write("internal/oracle/oracle_test.go", self.lanes)
        self.write("internal/oracle/cast_test.go", self.cast)
        for name in ("a", "c"):
            self.write("internal/oracle/testdata/%s.a" % name, "console.log(1)\n")
        self.base = self.commit()

    def git(self, *arguments):
        return realRun(["git", "-C", self.tree] + list(arguments), check=True, capture_output=True, text=True).stdout.strip()

    def write(self, path, text, mode="w"):
        with open(os.path.join(self.tree, path), mode) as handle:
            handle.write(text)

    def commit(self):
        self.git("add", "-A")
        self.git("-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "change")
        return self.git("rev-parse", "HEAD")

    def select(self, base):
        gate = run.Gate.__new__(run.Gate)
        gate.arguments = mock.Mock(tree=self.tree, base=base, sha=self.commit(), full=False)
        gate.buildStoreEnvironment = {}  # This bare fixture only spawns git for selection.
        gate.lock, gate.processes, gate.failure = threading.Lock(), [], None
        return gate.selectOracle()

    def test_a_fixture_no_table_names_skips_that_tables_test(self):
        self.write("internal/oracle/testdata/a.a", "console.log(2)\n")
        selection = self.select(self.base)
        self.assertFalse(selection["whole"], selection)
        self.assertEqual(selection["fixtures"], ["internal/oracle/testdata/a.a"])
        self.assertNotIn("TestCast", selection["tests"])

    def test_a_fixture_a_table_names_runs_that_tables_test(self):
        self.write("internal/oracle/testdata/c.a", "console.log(2)\n")
        selection = self.select(self.base)
        self.assertFalse(selection["whole"], selection)
        self.assertIn("TestCast", selection["tests"])

    def test_a_table_filled_outside_its_literal_runs_whole(self):
        self.write("internal/oracle/cast_test.go", "\nfunc init() { castFixtures = append(castFixtures, \"a\") }\n", "a")
        base = self.commit()
        self.write("internal/oracle/testdata/a.a", "console.log(2)\n")
        selection = self.select(base)
        self.assertTrue(selection["whole"], selection)
        self.assertIn("castFixtures", selection["reason"])


@unittest.skipUnless(sys.platform == "linux", "requires Linux /proc sessions")
class KillDescendants(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.tree = self.directory.name
        for name, content in {
            "go.mod": "module example.com/kill\n\ngo 1.20\n",
            "kill_test.go": r'''package kill
import ("os"; "os/exec"; "strconv"; "syscall"; "testing"; "time")
func TestGuard(t *testing.T) {
    child := exec.Command("sleep", "600")
    child.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
    if err := child.Start(); err != nil { t.Fatal(err) }
    defer child.Process.Kill()
    if err := os.WriteFile(os.Getenv("CHILD_PID"), []byte(strconv.Itoa(child.Process.Pid)), 0600); err != nil { t.Fatal(err) }
    for { time.Sleep(time.Second) }
}
''',
        }.items():
            with open(os.path.join(self.tree, name), "w") as handle:
                handle.write(content)
        realRun(["git", "init", "-q", self.tree], check=True)
        realRun(["git", "-C", self.tree, "add", "."], check=True)
        realRun(["git", "-C", self.tree, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "scratch"], check=True)
        self.binary = os.path.join(self.tree, "guard.test")
        realRun(["go", "test", "-c", "-o", self.binary, "."], cwd=self.tree, check=True, timeout=90)

    def state(self, pid):
        try:
            with open("/proc/%d/stat" % pid) as handle:
                return handle.read().rsplit(")", 1)[1].split()
        except FileNotFoundError:
            return None

    def alive(self, pid):
        state = self.state(pid)
        return state is not None and state[0] not in ("Z", "X")

    def probe(self, action):
        pidFile = os.path.join(self.tree, "child.pid")
        if os.path.exists(pidFile):
            os.unlink(pidFile)
        sha = run.git(self.tree, "rev-parse", "HEAD")
        arguments = mock.Mock(tree=self.tree, tools=self.tree, out=self.tree, sha=sha, base=sha,
                              full=False, branch="", branch_source="", session="", session_source="")
        gate = run.Gate(arguments)
        # The exact testSplit test-binary launch path, including test2json and stream.
        command = ["go", "tool", "test2json", "-t", "-p", "example.com/kill", self.binary,
                   "-test.v=test2json", "-test.run", "^TestGuard$"]
        thread = gate.guarded("tests", gate.stream, "tests", command, None, self.tree, {"CHILD_PID": pidFile})
        thread.start()
        pid = None
        try:
            deadline = time.monotonic() + 15
            while not os.path.exists(pidFile) and time.monotonic() < deadline:
                time.sleep(0.01)
            self.assertTrue(os.path.exists(pidFile), "Go child did not start")
            with open(pidFile) as handle:
                pid = int(handle.read())
            state = self.state(pid)
            self.assertTrue(self.alive(pid), "control cannot see sleep")
            self.assertEqual(int(state[2]), pid, "sleep must have its own process group")
            self.assertIn(int(state[3]), [process.pid for process in gate.processes])
            if action == "fail":
                gate.step("other", ["sh", "-c", "exit 1"])
            elif action == "finish":
                gate.planned = []
                gate.finish()
            deadline = time.monotonic() + 2
            while action != "control" and self.alive(pid) and time.monotonic() < deadline:
                time.sleep(0.01)
            survived = self.alive(pid)
            print("kill probe: action=%s pid=%d pgrp=%s sid=%s survived=%s" %
                  (action, pid, state[2], state[3], survived))
            return survived
        finally:
            # Independent cleanup also handles the intentionally broken mutant and old gate.
            if pid is not None and self.alive(pid):
                os.kill(pid, signal.SIGKILL)
            for process in gate.processes:
                try:
                    os.killpg(process.pid, signal.SIGKILL)
                except ProcessLookupError:
                    pass
            thread.join(10)
            for process in gate.processes:
                if process.stdout is not None:
                    process.stdout.close()
            self.assertFalse(thread.is_alive(), "test stream did not stop")

    def test_control_sees_survivor(self):
        self.assertTrue(self.probe("control"))

    def test_fail_kills_separate_process_group(self):
        self.assertFalse(self.probe("fail"), "Setpgid child survived Gate.fail")

    def test_finish_kills_separate_process_group(self):
        self.assertFalse(self.probe("finish"), "Setpgid child survived Gate.finish")

    def test_mutant_without_session_cleanup_is_caught(self):
        with mock.patch.object(run.Gate, "killSessions", lambda gate: None):
            with self.assertRaisesRegex(AssertionError, "Setpgid child survived"):
                self.test_fail_kills_separate_process_group()


class LongestFirst(unittest.TestCase):
    def test_queue_priorities_and_boost_fallback(self):
        queue = run.TestQueue(4)
        queue.add([(seconds, parallel, "p", name, []) for seconds, name, parallel in [(10, "short", False), (60, "unknown", False), (400, "long", True)]])
        item = queue.take()
        self.assertEqual(item[1:4], ("long", 400, 4))
        queue.release(4)
        self.assertEqual(queue.take()[1], "unknown")
        queue.add([(500, True, "p", "boost", [])])
        self.assertEqual(queue.take()[3], 1)  # three free: retain parallel=2
        queue.release(1)
        queue.release(1)
        self.assertEqual(queue.free, 4)

    def test_record_decay_corrupt_and_concurrent_writers(self):
        with tempfile.TemporaryDirectory() as directory:
            record = run.TestSeconds(os.path.join(directory, "seconds.tsv"))
            self.assertEqual(record.read(), {})
            record.update({("p", "TestA"): (100, True)})
            record.update({("p", "TestA"): (20, False), ("p", "TestB"): (90, True)})
            self.assertEqual(record.read()[("p", "TestA")], (80, False))
            writers = [threading.Thread(target=record.update, args=({("q", "Test%d" % i): (i, False)},)) for i in range(20)]
            for writer in writers:
                writer.start()
            for writer in writers:
                writer.join()
            self.assertEqual(len(record.read()), 22)
            with open(record.path, "w") as handle:
                handle.write("corrupt\n")
            self.assertEqual(record.read(), {})
            record.update({("p", "TestC"): (120, True)})
            self.assertEqual(record.read(), {("p", "TestC"): (120, True)})

    def split(self, parallel=4, broken=None, complete=False):
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        tree = directory.name
        arguments = types.SimpleNamespace(tree=tree, out=tree, parallel=parallel)
        gate = object.__new__(run.Gate)
        gate.arguments = arguments
        gate.lock = threading.Lock()
        gate.failure = None
        gate.complete = complete
        gate.watchers = []
        gate.result = {}
        gate.steps = {}
        gate.exits = {}
        gate.packageDirectories = {"p": tree}
        gate.deferred = {}
        gate.fail = lambda stage, detail: setattr(gate, "failure", detail)
        commands = []
        def stream(stage, command, log, *args):
            if "-c" in command:
                if broken == "build":
                    return 1
                with open(command[command.index("-o") + 1], "w") as handle:
                    handle.write("binary")
                return 0
            with gate.lock:
                gate.recordTestStart(command)
            commands.append(command)
            if broken == "dead" and len(commands) == 1:
                raise OSError("dead process")
            if broken == "fail" and len(commands) == 1:
                return 1
            for watch in gate.watchers:
                watch(json.dumps({"Package": "p", "Test": "TestLong/sub", "Action": "output", "Output": "=== PAUSE"}))
                watch(json.dumps({"Package": "p", "Test": "TestLong", "Action": "pass", "Elapsed": 350}))
            return 0
        gate.stream = stream
        gate.capture = lambda *args: "TestShort\nTestUnknown\nTestLong\n"
        record = run.TestSeconds(os.path.join(tree, "record.tsv"))
        record.update({("p", "TestLong"): (400, True), ("p", "TestShort"): (10, False)})
        with mock.patch.object(run, "TestSeconds", return_value=record):
            thread = threading.Thread(target=gate.testSplit, args=(["p"], io.StringIO()))
            thread.start()
            thread.join(5)
            self.assertFalse(thread.is_alive(), "permits leaked")
        return gate, commands, record

    def test_split_order_boost_and_event_record(self):
        gate, commands, record = self.split()
        self.assertEqual([row["test"] for row in gate.result["test_starts"]], ["TestLong", "TestUnknown", "TestShort"])
        self.assertIn("-test.parallel=8", commands[0])
        self.assertEqual(gate.result["test_starts"][0]["slots"], 4)
        self.assertEqual(record.read()[("p", "TestLong")], (350, True))
        self.assertEqual(gate.result["split_tally"], {"packages": 1, "planned": 3, "passed": 3, "touched": 1})

    def test_a_ready_long_test_starts_before_the_last_package_builds(self):
        # q's build waits until TestLong (p's, the longest known) has started: building every package
        # before any test would deadlock here, and a short test of q can't go first.
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        tree = directory.name
        gate = object.__new__(run.Gate)
        gate.arguments = types.SimpleNamespace(tree=tree, out=tree, parallel=4)
        gate.lock = threading.Lock()
        gate.failure = None
        gate.complete = False
        gate.watchers, gate.result, gate.steps, gate.exits, gate.deferred = [], {}, {}, {}, {}
        gate.packageDirectories = {"p": tree, "q": tree}
        gate.fail = lambda stage, detail: setattr(gate, "failure", detail)
        longStarted = threading.Event()

        def stream(stage, command, log, *args):
            if "-c" in command:
                if command[-1] == "q" and not longStarted.wait(3):
                    return 1
                with open(command[command.index("-o") + 1], "w") as handle:
                    handle.write("binary")
                return 0
            with gate.lock:
                gate.recordTestStart(command)
            if "^TestLong$" in command:
                longStarted.set()
            return 0
        gate.stream = stream
        gate.capture = lambda command, directory: "TestLong\n" if command[0].endswith("p.test") else "TestShortQ\n"
        record = run.TestSeconds(os.path.join(tree, "record.tsv"))
        record.update({("p", "TestLong"): (400, False), ("q", "TestShortQ"): (10, False)})
        with mock.patch.object(run, "TestSeconds", return_value=record):
            thread = threading.Thread(target=gate.testSplit, args=(["q", "p"], io.StringIO()))
            thread.start()
            thread.join(10)
            self.assertFalse(thread.is_alive())
        self.assertEqual([row["test"] for row in gate.result["test_starts"]], ["TestLong", "TestShortQ"])
        self.assertEqual(gate.result["split_tally"], {"packages": 2, "planned": 2, "passed": 2, "touched": 2})

    def test_boost_requires_four_free_slots(self):
        gate, commands, _ = self.split(parallel=3)
        self.assertIn("-test.parallel=2", commands[0])
        self.assertEqual(gate.result["test_starts"][0]["slots"], 1)

    def test_failfast_stops_new_starts_after_dead_process(self):
        gate, commands, _ = self.split(parallel=1, broken="dead")
        self.assertEqual(len(commands), 1)
        self.assertEqual(gate.exits["tests"], 1)

    def test_permits_after_fail_dead_and_build_failure(self):
        for broken in ("fail", "dead", "build"):
            gate, commands, _ = self.split(parallel=1, broken=broken, complete=True)
            self.assertEqual(len(commands), 0 if broken == "build" else 3)
            self.assertEqual(gate.exits["tests"], 1)


class ScopedEnvironment(unittest.TestCase):
    setUp = FailClosed.setUp

    def gate(self):
        arguments = types.SimpleNamespace(tree=self.tree, tools=self.tree, out=self.tree, sha=self.sha, base=self.sha,
                                          full=False, complete=True, branch="cloud/land-x", branch_source="", session="", session_source="")
        return run.Gate(arguments)

    def test_a_scoped_gate_refuses_and_a_clean_one_records_it(self):
        with mock.patch.dict(os.environ, {"ADAMIC_LINT_RULES": "no-debugger"}):
            gate = self.gate()
            gate.run()
        self.assertEqual(gate.result["scoped_env"], ["ADAMIC_LINT_RULES"])
        self.assertEqual(gate.failure["step"], "environment")
        environment = dict(os.environ)
        environment.pop("ADAMIC_LINT_RULES", None)
        with mock.patch.dict(os.environ, environment, clear=True):
            gate = self.gate()
            with mock.patch.object(gate, "git", side_effect=RuntimeError("past the check")):
                with self.assertRaises(RuntimeError):
                    gate.run()
        self.assertEqual(gate.result["scoped_env"], [])

    def test_set_but_empty_still_refuses(self):
        with mock.patch.dict(os.environ, {"ADAMIC_LINT_RULES": ""}):
            gate = self.gate()
            gate.run()
        self.assertEqual(gate.result["scoped_env"], ["ADAMIC_LINT_RULES"])


class LongTests(unittest.TestCase):
    def test_the_ledger_holds_units_over_the_budget_longest_first(self):
        # A test with subtests is split: its subtests are units, and so is its own time outside them.
        events = [
            {"Action": "pass", "Package": "p", "Test": "TestShort", "Elapsed": 29.9},
            {"Action": "pass", "Package": "p", "Test": "TestAtTheLine", "Elapsed": 30},
            {"Action": "pass", "Package": "p", "Test": "TestLong", "Elapsed": 32.5},
            {"Action": "fail", "Package": "q", "Test": "TestLonger", "Elapsed": 400},
            {"Action": "pass", "Package": "q", "Test": "TestLonger/case", "Elapsed": 390},
            {"Action": "pass", "Package": "q", "Test": "TestSplit", "Elapsed": 100},
            {"Action": "pass", "Package": "q", "Test": "TestSplit/a", "Elapsed": 20},
            {"Action": "pass", "Package": "q", "Test": "TestSplit/b", "Elapsed": 20},
            {"Action": "pass", "Package": "q", "Test": "TestParallel", "Elapsed": 40},
            {"Action": "pass", "Package": "q", "Test": "TestParallel/a", "Elapsed": 25},
            {"Action": "pass", "Package": "q", "Test": "TestParallel/b", "Elapsed": 25},
            {"Action": "skip", "Package": "q", "Test": "TestSkipped", "Elapsed": 50},
            {"Action": "pass", "Package": "q", "Elapsed": 600},
            {"Action": "output", "Package": "p", "Test": "TestLong", "Output": "x"},
        ]
        with tempfile.TemporaryDirectory() as directory:
            path = os.path.join(directory, "test.jsonl")
            with open(path, "w") as handle:
                handle.write("not json\n[1]\n")
                handle.writelines(json.dumps(event) + "\n" for event in events)
            units = run.testUnits(path)
            self.assertNotIn(("q", "TestSplit"), units)
            self.assertNotIn(("q", "TestParallel (setup)"), units)
            self.assertEqual(run.longTests(units), [("q", "TestLonger/case", 390, "pass"), ("q", "TestSplit (setup)", 60, "pass"), ("p", "TestLong", 32.5, "pass")])
            self.assertEqual(run.testUnits(os.path.join(directory, "absent.jsonl")), {})


class SetupByTime(unittest.TestCase):
    def test_a_build_before_parallel_shards_is_the_test_s_setup(self):
        # cohere, Oct 8: TestCSSNumbers builds, then runs 357 parallel shards; elapsed less the shards' sum is
        # negative, which hid the build. By the events' times the build is the setup.
        base = "2026-10-08T22:00:%02d.123456789Z"
        events = [{"Action": "run", "Package": "p", "Test": "TestSplit", "Time": base % 0}]
        for index in range(3):
            events.append({"Action": "run", "Package": "p", "Test": "TestSplit/shard-%03d" % index, "Time": base % 40})
        for index in range(3):
            events.append({"Action": "pass", "Package": "p", "Test": "TestSplit/shard-%03d" % index, "Elapsed": 18.0, "Time": base % 58})
        events.append({"Action": "pass", "Package": "p", "Test": "TestSplit", "Elapsed": 59.0, "Time": base % 59})
        with tempfile.TemporaryDirectory() as directory:
            path = os.path.join(directory, "test.jsonl")
            with open(path, "w") as handle:
                handle.writelines(json.dumps(event) + "\n" for event in events)
            units = run.testUnits(path)
        self.assertAlmostEqual(units[("p", "TestSplit (setup)")][0], 41.0, places=3)
        self.assertEqual(units[("p", "TestSplit/shard-000")], (18.0, "pass"))
        self.assertEqual(run.longTests(units), [("p", "TestSplit (setup)", units[("p", "TestSplit (setup)")][0], "pass")])


class Budget(unittest.TestCase):
    """A new unit over the 30 s budget is red at 'budget'; an existing one is drift, and the burn-down holds the rest."""

    def setUp(self):
        self.directory = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, self.directory)
        self.tree = os.path.join(self.directory, "tree")
        os.makedirs(os.path.join(self.tree, "p/sub"))
        os.makedirs(os.path.join(self.tree, "cloud/fast-gate"))
        self.write("p/old_test.go", "package p\n\nfunc TestOld(t *testing.T) {}\nfunc TestListed(t *testing.T) {}\n")
        # A same-named test in a subpackage doesn't make p's new test existing.
        self.write("p/sub/sub_test.go", "package sub\n\nfunc TestNew(t *testing.T) {}\n")
        self.write("cloud/fast-gate/budget-burndown.tsv", "# header\n%s/p\tTestListed\t300.0\t\n%s/p\tTestShrunk\t90.0\t\n" % (run.module, run.module))
        self.base = self.commit("base")
        self.write("p/new_test.go", "package p\n\nfunc TestNew(t *testing.T) {}\n")
        self.head = self.commit("head")

    def write(self, name, text):
        with open(os.path.join(self.tree, name), "w") as handle:
            handle.write(text)

    def commit(self, message):
        for command in (["init", "-q"], ["add", "."], ["-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", message]):
            realRun(["git", "-C", self.tree] + command, check=True, capture_output=True)
        return run.git(self.tree, "rev-parse", "HEAD")

    def budget(self, rows, units=None, census=None):
        gate = run.Gate.__new__(run.Gate)
        gate.arguments = mock.Mock(tree=self.tree, tools=self.tree, base=self.base, sha=self.head, full=False, out=self.directory, census=census)
        gate.result, gate.failure, gate.lock, gate.started, gate.complete, gate.processes = {}, None, threading.Lock(), time.monotonic(), True, []
        ledger = [(run.module + "/p", name, seconds, "pass") for name, seconds in rows]
        allUnits = {(package, name): (seconds, action) for package, name, seconds, action in ledger}
        allUnits.update(units or {})
        with mock.patch("builtins.print"):
            gate.budget(ledger, allUnits)
        return gate

    def test_a_product_over_budget_is_listed_never_red_and_never_burns_down(self):
        # @system_adamic, Oct 9 00:41 MDT: products run to completion; over 60 s is listed beside the verdict.
        self.write("p/product_test.go", "func TestProduct_New(t *testing.T) {}\n")
        self.write("cloud/fast-gate/budget-burndown.tsv", "%s/p\tTestProduct_New\t61.0\n" % run.module)
        self.head = self.commit("new product")
        gate = self.budget([("TestProduct_New", 61.0)])
        self.assertIsNone(gate.failure)
        words = "product %s/p TestProduct_New 61.0 s" % run.module
        self.assertEqual(gate.result["budget_over"], [])
        self.assertEqual(gate.result["products_over_budget"], [words])
        self.assertEqual(gate.result["budget_burndown_units"], [])
        gate.arguments.branch = gate.arguments.session = ""
        gate.kind, gate.planned, gate.exits, gate.steps = "fast", [], {}, {}
        gate.counts, gate.failedTests, gate.census = {"pass": 0, "fail": 0, "skip": 0}, [], {"required_input": [], "unclassified": []}
        gate.result["product_units"] = [{"package": run.module + "/p", "test": "TestProduct_New", "seconds": 61.0, "status": "passed"}]
        with mock.patch("builtins.print"):
            gate.finish()
        with open(os.path.join(self.directory, "status.txt")) as status:
            self.assertIn(words, status.read())

    def test_old_product_over_budget_is_listed_too(self):
        self.write("p/product_test.go", "func TestProduct_Old(t *testing.T) {}\n")
        self.base = self.head = self.commit("old product")
        gate = self.budget([("TestProduct_Old", 61.0)])
        self.assertIsNone(gate.failure)
        self.assertEqual(gate.result["budget_drift"], [])
        self.assertEqual(gate.result["products_over_budget"], ["product %s/p TestProduct_Old 61.0 s" % run.module])

    def test_a_new_unit_over_the_budget_is_red_naming_it_and_the_box(self):
        gate = self.budget([("TestNew/case", 31.5)])
        self.assertEqual(gate.failure["step"], "budget")
        self.assertIn("TestNew/case 31.5 s", gate.failure["detail"])
        self.assertIn(os.uname().nodename, gate.failure["detail"])
        self.assertEqual(gate.result["budget_over"], [run.module + "/p TestNew/case 31.5 s"])
        # Over another run's merged log, the seconds are that log's: the record names it, never this box.
        gate = self.budget([("TestNew/case", 31.5)], census="/pool/merged/test.jsonl")
        self.assertIn("measured by the units of /pool/merged/test.jsonl", gate.failure["detail"])
        self.assertNotIn(os.uname().nodename, gate.failure["detail"])

    def test_an_existing_unit_over_is_drift_and_a_listed_one_is_neither(self):
        gate = self.budget([("TestOld", 45.0), ("TestListed (setup)", 200.0), ("TestListed", 300.0)],
                           units={(run.module + "/p", "TestShrunk"): (12.0, "pass")})
        self.assertIsNone(gate.failure)
        self.assertEqual(gate.result["budget_drift"], [run.module + "/p TestOld 45.0 s", run.module + "/p TestListed (setup) 200.0 s"])
        self.assertEqual(gate.result["budget_burndown_units"], [run.module + "/p TestListed"])
        self.assertEqual(gate.result["budget_can_leave_burndown"], [run.module + "/p TestShrunk"])
        self.assertEqual(gate.result["budget_instrument"]["box"], os.uname().nodename)

    def test_a_test_main_renamed_after_the_fork_isnt_new_in_a_candidate_cut_before(self):
        # Oct 9 05:13Z: a candidate cut from an older main read main's later renames as its own new tests.
        realRun(["git", "-C", self.tree, "checkout", "-q", self.base], check=True, capture_output=True)
        self.write("p/old_test.go", "package p\n\nfunc TestRenamed(t *testing.T) {}\nfunc TestListed(t *testing.T) {}\n")
        realRun(["git", "-C", self.tree, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qam", "main renames TestOld"], check=True, capture_output=True)
        mainTip = run.git(self.tree, "rev-parse", "HEAD")
        realRun(["git", "-C", self.tree, "checkout", "-q", self.head], check=True, capture_output=True)
        self.base = mainTip
        gate = self.budget([("TestOld", 95.0), ("TestNew", 31.0)])
        self.assertEqual(gate.result["budget_drift"], [run.module + "/p TestOld 95.0 s"])
        self.assertEqual(gate.result["budget_over"], [run.module + "/p TestNew 31.0 s"], "the change's own new test is still new")

    def test_a_whole_gate_of_main_never_finds_a_new_unit(self):
        # Base is the candidate itself: every unit exists there.
        self.base = self.head
        self.assertIsNone(self.budget([("TestNew", 500.0)]).failure)


class HeavyBudget(Budget):
    def declare(self, test="TestNew", seconds=120):
        self.write("cloud/fast-gate/heavy-units.tsv", "%s/p\t%s\towner/#task\t%s\twhole main gate\n" % (run.module, test, seconds))

    def test_declared_90_second_unit_is_heavy(self):
        self.declare()
        gate = self.budget([("TestNew", 90)])
        self.assertIsNone(gate.failure)
        self.assertEqual(gate.result["budget_over"], [])
        self.assertEqual(gate.result["budget_drift"], [])
        row = gate.result["heavy_units"][0]
        self.assertEqual((row["class"], row["seconds"], row["budget_seconds"], row["owner"]),
                         ("heavy", 90, 120, "owner/#task"))

    def test_record_and_status_name_heavy_units(self):
        self.declare()
        gate = self.budget([("TestNew", 90)])
        gate.arguments.branch = gate.arguments.session = ""
        gate.kind, gate.planned, gate.exits, gate.steps = "fast", [], {}, {}
        gate.counts, gate.failedTests, gate.census = {"pass": 1, "fail": 0, "skip": 0}, [], {"required_input": [], "unclassified": []}
        with open(os.path.join(self.directory, "test.jsonl"), "w") as handle:
            handle.write(json.dumps({"Package": run.module + "/p", "Test": "TestNew", "Action": "pass", "Elapsed": 90}) + "\n")
        with mock.patch("builtins.print"):
            gate.finish()
        with open(os.path.join(self.directory, "fast.json")) as handle:
            record = json.load(handle)
        self.assertEqual(record["units"][0]["class"], "heavy")
        self.assertEqual(record["units"][0]["budget_seconds"], 120)
        with open(os.path.join(self.directory, "status.txt")) as handle:
            status = handle.read()
        self.assertTrue(status.startswith("green:"), status)
        self.assertIn("heavy:", status)
        self.assertIn("owner/#task", status)

    def test_undeclared_90_second_unit_is_red(self):
        gate = self.budget([("TestNew", 90)])
        self.assertIsNotNone(gate.failure, "undeclared units must not be exempt")
        self.assertEqual(gate.failure["step"], "budget")
        self.assertIn("TestNew 90.0 s", gate.failure["detail"])
        self.assertEqual(gate.result["heavy_units"], [])

    def test_heavy_over_own_budget_names_owner_even_on_main(self):
        self.declare("TestOld")
        gate = self.budget([("TestOld", 121)])
        self.assertEqual(gate.failure["step"], "budget")
        for word in ("heavy", "TestOld", "121.0 s", "120.0 s", "owner/#task"):
            self.assertIn(word, gate.failure["detail"])

    def test_family_descendants_and_setup_but_not_bare_prefix(self):
        self.declare()
        gate = self.budget([("TestNewUnit01/case", 90), ("TestNew (setup)", 90), ("TestNewOther", 90)])
        self.assertEqual([row["test"] for row in gate.result["heavy_units"]], ["TestNew (setup)", "TestNewUnit01/case"])
        self.assertEqual(gate.result["budget_over"], [run.module + "/p TestNewOther 90.0 s"])
        self.assertIsNone(run.heavyUnit(run.heavyUnits(self.tree), run.module + "/other", "TestNew"))

    def test_invalid_declarations_fail_closed(self):
        for value in ("nan", "inf", "0", "-1", "wrong"):
            with self.subTest(value=value):
                self.declare(seconds=value)
                self.assertEqual(self.budget([("TestNew", 90)]).failure["step"], "budget")

    def test_undeclared_exemption_mutant_is_caught(self):
        import inspect
        source = inspect.getsource(run.heavyUnit).replace(
            "return matches[0] if matches else None",
            'return matches[0] if matches else {"package": package, "test": name, "owner": "mutant", "budget_seconds": 1000, "why": "undeclared exemption"}')
        namespace = dict(run.__dict__)
        exec(source, namespace)
        with mock.patch.object(run, "heavyUnit", namespace["heavyUnit"]):
            with self.assertRaises(AssertionError):
                self.test_undeclared_90_second_unit_is_red()


class HeavyCensus(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, self.directory)
        self.path = os.path.join(self.directory, "test.jsonl")
        self.filtered = os.path.join(self.directory, "census-tests.jsonl")
        self.declared = [{"package": "p", "test": "TestHeavy", "owner": "owner/#task",
                          "budget_seconds": 120, "why": "whole main gate"}]

    def events(self, reason, test="TestHeavy"):
        events = [{"Action": "output", "Package": "p", "Test": test, "Output": reason},
                  {"Action": "skip", "Package": "p", "Test": test}]
        with open(self.path, "w") as handle:
            for event in events:
                handle.write(json.dumps(event) + "\n")
        return events

    def test_declared_deferral_reads_heavy_and_preserves_original(self):
        events = self.events("heavy: deferred to main whole gate", "TestHeavy/case")
        heavy, unknown = run.heavyCensus(self.path, self.filtered, self.declared, False)
        self.assertEqual(unknown, [])
        self.assertEqual(heavy[0]["class"], "heavy")
        self.assertEqual(heavy[0]["owner"], "owner/#task")
        with open(self.path) as handle:
            self.assertEqual([json.loads(line) for line in handle], events)
        with open(self.filtered) as handle:
            self.assertNotIn('"Action": "skip"', handle.read())

    def test_removed_declaration_is_unknown_and_whole_gate_cannot_defer(self):
        self.events("heavy: deferred to main whole gate")
        for declarations, full in (([], False), (self.declared, True)):
            heavy, unknown = run.heavyCensus(self.path, self.filtered, declarations, full)
            self.assertEqual(heavy, [])
            self.assertEqual(unknown, ["p TestHeavy"])

    def test_required_input_skip_is_not_exempt(self):
        events = self.events("set REQUIRED_INPUT")
        self.assertEqual(run.heavyCensus(self.path, self.filtered, self.declared, False), ([], []))
        with open(self.filtered) as handle:
            self.assertEqual([json.loads(line) for line in handle], events)

    def test_gate_census_records_heavy_and_reds_unknown(self):
        tools = os.path.join(self.directory, "tools")
        os.makedirs(os.path.join(tools, "cloud/fast-gate"))
        declaration = os.path.join(tools, "cloud/fast-gate/heavy-units.tsv")
        for declared in (True, False):
            with open(declaration, "w") as handle:
                if declared:
                    handle.write("p\tTestHeavy\towner/#task\t120\twhole main gate\n")
            self.events("heavy: deferred to main whole gate")
            gate = run.Gate.__new__(run.Gate)
            gate.arguments = types.SimpleNamespace(out=self.directory, tree=tools, tools=tools, full=False)
            gate.census, gate.result, gate.steps, gate.exits = {"required_input": [], "unclassified": []}, {}, {}, {}
            gate.fail = mock.Mock()
            process = mock.Mock(returncode=0)
            process.communicate.return_value = ("skips=0 required-input=0 unknown=0\n", "")
            gate.spawn = mock.Mock(return_value=process)
            gate.checkCensus()
            if declared:
                gate.fail.assert_not_called()
                self.assertEqual(gate.result["heavy_skips"][0]["class"], "heavy")
            else:
                self.assertEqual(gate.exits["census"], 1)
                self.assertEqual(gate.census["unclassified"], ["p TestHeavy"])
                self.assertIn("unknown\tp\tTestHeavy", gate.fail.call_args.args[1])


@unittest.skipUnless(sys.platform == "linux", "requires Linux /proc sessions")
class StopTests(unittest.TestCase):
    setUp = FailClosed.setUp

    def stopped(self, failed, sibling=True):
        self.addCleanup(shutil.rmtree, self.directory)
        arguments = types.SimpleNamespace(tree=self.tree, tools=self.tree, out=self.tree,
            sha=self.sha, base=self.sha, full=False, complete=True, branch="area/test",
            branch_source="", session="", session_source="")
        gate = run.Gate(arguments)
        gate.packageDirectories = {"p": self.tree}
        gate.result["test_outcomes"] = [{"package": "p", "test": name, "status": "not run"}
                                        for name in ("TestPassed", "TestFailed", "TestKilled", "TestPending")]
        tally = {"passed": 0}
        passed = [sys.executable, "-c", 'import json; print(json.dumps({"Package":"p","Test":"TestPassed","Action":"pass"}))']
        gate.slotted(passed, io.StringIO(), "p", "TestPassed", tally)
        if failed:
            command = [sys.executable, "-c", 'import json; print(json.dumps({"Package":"p","Test":"TestFailed","Action":"fail"})); raise SystemExit(1)']
            gate.slotted(command, io.StringIO(), "p", "TestFailed", tally)
        ready = os.path.join(self.tree, "ready")
        command = [sys.executable, "-c", 'import pathlib,time; pathlib.Path(%r).touch(); time.sleep(60)' % ready]
        thread = gate.guarded("tests", gate.slotted, command, io.StringIO(), "p", "TestKilled", tally)
        thread.start()
        self.addCleanup(thread.join, 5)
        deadline = time.monotonic() + 5
        while not os.path.exists(ready) and time.monotonic() < deadline:
            time.sleep(.01)
        self.assertTrue(os.path.exists(ready))
        reason = "superseded by area/new " + "b" * 40
        if sibling:
            with open(self.tree + ".stop-reason", "w") as handle:
                handle.write(reason)
            self.addCleanup(os.unlink, self.tree + ".stop-reason")
        previous = signal.signal(signal.SIGTERM, gate.stop)
        try:
            with mock.patch.dict(os.environ, {"ADAMIC_FAST_GATE_STOP_REASON": reason}):
                os.kill(os.getpid(), signal.SIGTERM)
                thread.join(5)
        finally:
            signal.signal(signal.SIGTERM, previous)
        self.assertFalse(thread.is_alive())
        before = len(gate.processes)
        with self.assertRaises(SystemExit):
            gate.spawn([sys.executable, "-c", "raise SystemExit(0)"], subprocess.PIPE)
        self.assertEqual(len(gate.processes), before)
        gate.finish()
        with open(os.path.join(self.tree, "fast.json")) as handle:
            result = json.load(handle)
        with open(os.path.join(self.tree, "status.txt")) as handle:
            status = handle.read()
        self.assertEqual(result["stopped"]["reason"], reason)
        self.assertEqual(result["planned_test_counts"], {"passed": 1, "failed": int(failed), "not run": 3-int(failed)})
        self.assertEqual(result["failed_tests"], ["p TestFailed"] if failed else [])
        if failed:
            self.assertTrue(status.startswith("red:"))
            self.assertIn("stopped: " + reason, status)
            self.assertEqual(result["failure"]["detail"], "p TestFailed\n")
        else:
            self.assertEqual(status.strip(), "void: stopped before a verdict: " + reason)

    def test_sigterm_mid_tests_preserves_red_and_marks_killed_not_run(self):
        self.stopped(True)

    def test_sigterm_before_failure_is_void(self):
        self.stopped(False, sibling=False)


class StoppedPublish(unittest.TestCase):
    def test_stopped_remote_run_still_publishes(self):
        with tempfile.TemporaryDirectory() as root:
            cloud = os.path.join(root, "cloud")
            os.mkdir(cloud)
            shutil.copy(os.path.join(os.path.dirname(__file__), "../fast-gate.sh"), cloud)
            with open(os.path.join(cloud, "fast-gate-classify.sh"), "w") as handle:
                handle.write('gateBase() { echo "main ' + 'a'*40 + '"; }\nclassify() { echo B; }\ngateMerge() { echo "same $1"; }\n')
            bindir = os.path.join(root, "bin")
            os.mkdir(bindir)
            fixture = os.path.join(root, "fixture")
            os.mkdir(fixture)
            with open(os.path.join(fixture, "status.txt"), "w") as handle:
                handle.write("red: first failure at tests\nstopped: superseded by newer\n")
            with open(os.path.join(fixture, "fast.json"), "w") as handle:
                json.dump({"stopped": {"reason": "superseded by newer", "at_seconds": 174}}, handle)
            scripts = {
                "ssh": "cat >/dev/null; exit 1\n",
                "scp": 'cp -R "$PUBLISH_FIXTURE" "$4"\n',
                "git": '''case "$*" in
*rev-parse*--absolute-git-dir*) echo "$PUBLISH_ROOT" ;;
*rev-parse*) printf '%040d\\n' 1 ;;
*write-tree*|*commit-tree*) printf '%040d\\n' 2 ;;
*push*) printf '%s\\n' "$*" > "$PUBLISH_ROOT/pushed" ;;
esac
''',
            }
            for name, body in scripts.items():
                path = os.path.join(bindir, name)
                with open(path, "w") as handle:
                    handle.write("#!/bin/bash\n" + body)
                os.chmod(path, 0o755)
            env = dict(os.environ, PATH=bindir + ":" + os.environ["PATH"], PUBLISH_ROOT=root,
                       PUBLISH_FIXTURE=fixture, TMPDIR=root, ADAMIC_FAST_GATE_TOOLS_ON_ORIGIN="1")
            result = realRun(["bash", os.path.join(cloud, "fast-gate.sh"), "b"*40,
                              "--branch", "codex/test", "--session", "test"], env=env,
                             capture_output=True, text=True, timeout=10)
            self.assertEqual(result.returncode, 1, result.stderr)
            self.assertIn("published gate-logs/" + "b"*12, result.stdout)
            self.assertIn("stopped: superseded by newer", result.stdout)
            with open(os.path.join(root, "pushed")) as handle:
                self.assertIn("refs/heads/gate-logs/" + "b"*12, handle.read())


class ReverseDependencies(unittest.TestCase):
    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory()
        self.addCleanup(self.scratch.cleanup)
        self.tree = self.scratch.name
        realRun(["git", "init", "-q", self.tree], check=True)
        self.packages = []
        for name, fields in (("internal/load", {}), ("middle", {"TestImports": [run.module + "/ordinary"]}),
                             ("outer", {"XTestImports": [run.module + "/middle"]}),
                             ("ordinary", {"Imports": [run.module + "/internal/load"]}), ("realouter", {"Imports": [run.module + "/ordinary"]}), ("unrelated", {}), ("internal/oracle", {}), ("stage1/gaps", {})):
            directory = os.path.join(self.tree, name)
            os.makedirs(directory)
            self.packages.append(dict(Dir=directory, ImportPath=run.module + "/" + name, **fields))
        self.gate = run.Gate.__new__(run.Gate)
        self.gate.arguments = types.SimpleNamespace(tree=self.tree, tools=self.tree)
        self.gate.result = {}
        self.gate.deferred = {run.module + "/stage1/gaps": {"TestGap"}}
        self.gate.selectOracle = mock.Mock(return_value={"whole": False})
        self.gate.command = mock.Mock(return_value=types.SimpleNamespace(returncode=0, stdout="".join(json.dumps(p) for p in self.packages)))
        self.gate.git = lambda tree, *args: run.git(tree, *args)
        self.declare({"stage1/gaps": ["internal/load"], "internal/oracle": ["internal/load"]})
        writeUnitNeeds(self.tree)

    def declare(self, packages):
        path = os.path.join(self.tree, "cloud/fast-gate/compiler-dependencies.json")
        os.makedirs(os.path.dirname(path), exist_ok=True)
        with open(path, "w") as handle:
            json.dump({"version": 1, "packages": packages}, handle)

    def test_real_closure_then_one_test_edge(self):
        packages, unowned = self.gate.touched(["internal/load/load.go"])
        self.assertIn(run.module + "/ordinary", packages)
        self.assertIn(run.module + "/realouter", packages)
        self.assertIn(run.module + "/middle", packages)
        self.assertNotIn(run.module + "/outer", packages)
        self.assertNotIn(run.module + "/unrelated", packages)
        self.assertEqual(unowned, [])
        self.assertEqual(self.gate.command.call_args.args[0], run.packageGraphCommand())

    def test_test_variants_normalize_to_real_packages(self):
        self.packages[1]["ImportPath"] += " [" + run.module + "/middle.test]"
        self.packages[1]["TestImports"] = [run.module + "/ordinary [" + run.module + "/middle.test]"]
        self.packages[2]["XTestImports"] = [self.packages[1]["ImportPath"]]
        self.packages.append(dict(Dir=os.path.join(self.tree, "outer"), Name="main", ImportPath=run.module + "/outer.test", Imports=[run.module + "/outer"]))
        self.gate.command.return_value.stdout = "".join(json.dumps(p) for p in self.packages)
        packages, _ = self.gate.touched(["internal/load/load.go"])
        self.assertIn(run.module + "/middle", packages)
        self.assertNotIn(run.module + "/outer", packages)
        self.assertNotIn(run.module + "/outer.test", packages)

    def test_external_test_package_uses_its_real_owner(self):
        self.packages.append(dict(Dir=os.path.join(self.tree, "outer"), ForTest=run.module + "/outer", ImportPath=run.module + "/outer_test [" + run.module + "/outer.test]", Imports=[run.module + "/internal/load"]))
        self.gate.command.return_value.stdout = "".join(json.dumps(p) for p in self.packages)
        packages, unowned = self.gate.touched(["internal/load/load.go"])
        self.assertIn(run.module + "/outer", packages)
        packages, unowned = self.gate.touched(["outer/source.go"])
        self.assertEqual(packages, [run.module + "/outer"])
        self.assertEqual(unowned, [])

    def test_test_source_does_not_seed_closure(self):
        packages, _ = self.gate.touched(["middle/source_test.go"])
        self.assertEqual(packages, [run.module + "/middle"])

    def test_test_inputs_do_not_seed_closure(self):
        # Only once test-reads.json names who reads another package's testdata.
        os.makedirs(os.path.join(self.tree, "cloud/fast-gate"), exist_ok=True)
        with open(os.path.join(self.tree, "cloud/fast-gate/test-reads.json"), "w") as handle:
            handle.write("{}")
        self.packages[1]["TestEmbedFiles"] = ["fixtures/input.txt"]
        self.gate.command.return_value.stdout = "".join(json.dumps(p) for p in self.packages)
        for path in ("middle/testdata/input.txt", "middle/fixtures/input.txt"):
            with self.subTest(path=path):
                packages, _ = self.gate.touched([path])
                self.assertEqual(packages, [run.module + "/middle"])

    def test_without_test_reads_testdata_still_seeds_the_closure(self):
        # A tree that can't name its testdata's readers keeps the wide selection for testdata; _test.go stays narrow.
        wide, _ = self.gate.touched(["middle/source.go"])
        packages, _ = self.gate.touched(["middle/testdata/input.txt"])
        self.assertEqual(packages, wide)
        packages, _ = self.gate.touched(["middle/source_test.go"])
        self.assertEqual(packages, [run.module + "/middle"])

    def test_testdata_the_real_package_embeds_seeds_the_closure(self):
        os.makedirs(os.path.join(self.tree, "cloud/fast-gate"), exist_ok=True)
        with open(os.path.join(self.tree, "cloud/fast-gate/test-reads.json"), "w") as handle:
            handle.write("{}")
        self.packages[1]["EmbedFiles"] = ["testdata/table.txt"]
        self.gate.command.return_value.stdout = "".join(json.dumps(p) for p in self.packages)
        wide, _ = self.gate.touched(["middle/source.go"])
        packages, _ = self.gate.touched(["middle/testdata/table.txt"])
        self.assertEqual(packages, wide)

    def test_explicit_test_reader(self):
        path = os.path.join(self.tree, "cloud/fast-gate/test-reads.json")
        with open(path, "w") as handle:
            json.dump({"outer": ["middle/testdata"]}, handle)
        packages, _ = self.gate.touched(["middle/testdata/input.txt"])
        self.assertEqual(packages, [run.module + "/middle", run.module + "/outer"])

    def test_opaque_gap_and_whole_oracle_are_selected(self):
        packages, _ = self.gate.touched(["internal/load/load.go"])
        self.assertIn(run.module + "/stage1/gaps", packages)
        self.assertTrue(self.gate.oracleSelection["whole"])
        self.assertEqual(self.gate.deferred, {})

    def test_census_rejects_a_new_compiler_runner(self):
        path = os.path.join(self.tree, "stage1/new/probe_test.go")
        os.makedirs(os.path.dirname(path))
        with open(path, "w") as handle:
            handle.write('package new\nvar command = "cmd/adamic"\n')
        realRun(["git", "-C", self.tree, "add", "."], check=True)
        with self.assertRaisesRegex(ValueError, "census.*stage1/new/probe_test.go"):
            self.gate.touched(["internal/load/load.go", "stage1/new/probe_test.go"])

    def test_census_rejects_computed_compiler_outside_stage1(self):
        path = os.path.join(self.tree, "cmd/new/probe_test.go")
        os.makedirs(os.path.dirname(path))
        with open(path, "w") as handle:
            handle.write('package new\nimport "os/exec"\nfunc TestCompiler(t *testing.T) { exec.Command(binary(), "types", "input.a").Run() }\n')
        realRun(["git", "-C", self.tree, "add", "."], check=True)
        with self.assertRaisesRegex(ValueError, "census.*cmd/new/probe_test.go"):
            self.gate.touched(["internal/load/load.go", "cmd/new/probe_test.go"])

    def test_census_rejects_removed_declaration(self):
        path = os.path.join(self.tree, "stage1/gaps/probe_test.go")
        with open(path, "w") as handle:
            handle.write('package gaps\nimport "os/exec"\n')
        realRun(["git", "-C", self.tree, "add", "."], check=True)
        self.declare({"internal/oracle": ["internal/load"]})
        with self.assertRaisesRegex(ValueError, "census.*stage1/gaps/probe_test.go"):
            self.gate.touched(["internal/load/load.go", "stage1/gaps/probe_test.go"])

    def test_an_undeclared_consumer_on_the_base_is_named_not_red(self):
        path = os.path.join(self.tree, "stage1/old/probe_test.go")
        os.makedirs(os.path.dirname(path))
        with open(path, "w") as handle:
            handle.write('package old\nvar command = "cmd/adamic"\n')
        realRun(["git", "-C", self.tree, "add", "."], check=True)
        try:
            self.gate.touched(["internal/load/load.go"])
        except ValueError as error:
            self.fail("a consumer the change didn't touch turned it red: %s" % error)
        self.assertEqual(self.gate.result["compiler_consumers_undeclared_on_base"], ["stage1/old/probe_test.go"])

    def test_a_branch_declares_its_own_consumer_over_the_tools_map(self):
        toolsDirectory = tempfile.TemporaryDirectory()
        self.addCleanup(toolsDirectory.cleanup)
        tools = toolsDirectory.name
        os.makedirs(os.path.join(tools, "cloud/fast-gate"))
        with open(os.path.join(tools, "cloud/fast-gate/compiler-dependencies.json"), "w") as handle:
            json.dump({"version": 1, "packages": {"internal/oracle": ["internal/load"]}}, handle)
        writeUnitNeeds(tools)
        self.gate.arguments = types.SimpleNamespace(tree=self.tree, tools=tools)
        path = os.path.join(self.tree, "stage1/gaps/probe_test.go")
        with open(path, "w") as handle:
            handle.write('package gaps\nimport "os/exec"\n')
        realRun(["git", "-C", self.tree, "add", "."], check=True)
        # The tree's map declares stage1/gaps (setUp); the tools' map doesn't: merged, the consumer is declared.
        try:
            packages, _ = self.gate.touched(["internal/load/load.go", "stage1/gaps/probe_test.go"])
        except ValueError as error:
            self.fail("the branch's own declaration wasn't read: %s" % error)
        self.assertIn(run.module + "/stage1/gaps", packages)

    def test_stage3_driver_declaration_and_selection(self):
        path = os.path.join(self.tree, "cloud/fast-gate/compiler-dependencies.json")
        with open(path) as handle:
            declarations = json.load(handle)
        declarations["drivers"] = {"stage3/drivers/parser": {"dependencies": ["middle"]}}
        with open(path, "w") as handle:
            json.dump(declarations, handle)
        self.gate.touched(["internal/load/load.go"])
        self.assertEqual(self.gate.result["selected_compiler_drivers"], ["stage3/drivers/parser"])
        self.gate.touched(["unrelated/source.go"])
        self.assertEqual(self.gate.result["selected_compiler_drivers"], [])

    def test_invalid_driver_map_fails_closed(self):
        path = os.path.join(self.tree, "cloud/fast-gate/compiler-dependencies.json")
        with open(path) as handle:
            declarations = json.load(handle)
        declarations["drivers"] = {"stage3/drivers/parser": {"dependencies": ["internal/missing"]}}
        with open(path, "w") as handle:
            json.dump(declarations, handle)
        with self.assertRaisesRegex(ValueError, "invalid compiler driver dependencies"):
            self.gate.touched(["internal/load/load.go"])

    def test_census_rejects_undeclared_script_driver(self):
        path = os.path.join(self.tree, "stage3/drivers/new/run.py")
        os.makedirs(os.path.dirname(path))
        with open(path, "w") as handle:
            handle.write('import subprocess\nsubprocess.run([compiler, "build", "input.a"])\n')
        realRun(["git", "-C", self.tree, "add", "."], check=True)
        with self.assertRaisesRegex(ValueError, "census.*stage3/drivers/new/run.py"):
            self.gate.touched(["internal/load/load.go", "stage3/drivers/new/run.py"])

    def test_invalid_map_fails_closed(self):
        self.declare({"stage1/gaps": ["internal/missing"]})
        with self.assertRaisesRegex(ValueError, "invalid compiler dependencies"):
            self.gate.touched(["internal/load/load.go"])

    def test_embedding_and_fixture_ownership_survive(self):
        os.makedirs(os.path.join(self.tree, "cloud/fast-gate"), exist_ok=True)
        with open(os.path.join(self.tree, "cloud/fast-gate/test-reads.json"), "w") as handle:
            handle.write("{}")
        self.packages[0]["EmbedFiles"] = ["runtime/header.h"]
        self.packages.append(dict(self.packages[0], ForTest=run.module + "/internal/load",
                                  ImportPath=run.module + "/internal/load [" + run.module + "/internal/load.test]"))
        self.gate.command.return_value.stdout = "".join(json.dumps(p) for p in self.packages)
        packages, unowned = self.gate.touched(["internal/load/runtime/header.h", "middle/testdata/input.ts", "unknown.txt"])
        self.assertIn(run.module + "/ordinary", packages)
        self.assertNotIn(run.module + "/outer", packages)
        self.assertEqual(unowned, ["unknown.txt"])


class SelectionModule(unittest.TestCase):
    """Exercise go list's real and ForTest records in a small on-disk module."""
    def setUp(self):
        scratch = tempfile.TemporaryDirectory()
        self.addCleanup(scratch.cleanup)
        self.tree = scratch.name
        sources = {
            "go.mod": "module " + run.module + "\ngo 1.24\n",
            "a/a.go": "package a\n",
            "b/b.go": 'package b\nimport _ "' + run.module + '/a"\n',
            "c/c.go": "package c\n",
            "c/c_test.go": 'package c_test\nimport _ "' + run.module + '/b"\n',
            "e/e.go": "package e\n",
            "e/e_test.go": 'package e\nimport _ "' + run.module + '/b"\n',
            "d/d.go": "package d\n",
            "d/d_test.go": 'package d_test\nimport _ "' + run.module + '/c"\n',
            "cloud/fast-gate/compiler-dependencies.json": '{"version":1,"packages":{}}',
            "cloud/fast-gate/unit-needs.json": '{"version":1,"units":[]}',
        }
        for name, source in sources.items():
            path = os.path.join(self.tree, name)
            os.makedirs(os.path.dirname(path), exist_ok=True)
            with open(path, "w") as handle:
                handle.write(source)
        realRun(["git", "init", "-q", self.tree], check=True)
        self.gate = run.Gate.__new__(run.Gate)
        self.gate.arguments = types.SimpleNamespace(tree=self.tree, tools=self.tree)
        self.gate.result = {}
        self.gate.git = run.git
        self.gate.command = lambda command, **kwargs: realRun(command, timeout=30, **kwargs)

    def selected(self, path):
        packages, unowned = self.gate.touched([path])
        self.assertEqual(unowned, [])
        return [package[len(run.module) + 1:] for package in packages]

    def test_real_closure_and_one_test_edge(self):
        self.assertEqual(self.selected("a/a.go"), ["a", "b", "c", "e"])

    def test_test_source_owner_only(self):
        self.assertEqual(self.selected("c/c_test.go"), ["c"])

    def test_explicit_reader_survives(self):
        with open(os.path.join(self.tree, "cloud/fast-gate/test-reads.json"), "w") as handle:
            json.dump({"d": ["c/testdata"]}, handle)
        self.assertEqual(self.selected("c/testdata/input.txt"), ["c", "d"])

    def test_selection_mutants(self):
        with open(run.__file__) as handle:
            original = handle.read()
        mutants = [
            ("transitive test edge", 'reverse.get(pending.pop(), ())',
             'set(reverse.get((node := pending.pop()), ())) | testReverse.get(node, set())',
             "test_real_closure_and_one_test_edge"),
            ("test source closure seed", 'testsOnly.add(owner)', 'packages.add(owner)',
             "test_test_source_owner_only"),
        ]
        for name, before, after, test in mutants:
            with self.subTest(mutant=name):
                self.assertIn(before, original)
                namespace = dict(run.__dict__)
                exec(compile(original.replace(before, after), run.__file__, "exec"), namespace)
                with mock.patch.object(run.Gate, "touched", namespace["Gate"].touched):
                    result = unittest.TestResult()
                    SelectionModule(test).run(result)
                self.assertEqual(result.errors, [])
                self.assertEqual(len(result.failures), 1, (name, result.failures))


class ReverseDependencyMutants(unittest.TestCase):
    def test_each_check_kills_its_mutant(self):
        with open(run.__file__) as handle:
            original = handle.read()
        mutants = [
            ("test edge walked transitively", 'reverse.get(pending.pop(), ())', 'set(reverse.get((node := pending.pop()), ())) | testReverse.get(node, set())', "test_real_closure_then_one_test_edge"),
            ("test source seeds closure", 'testsOnly.add(owner)', 'packages.add(owner)', "test_test_source_does_not_seed_closure"),
            ("external test owner lost", 'package.get("ForTest") or ', '', "test_external_test_package_uses_its_real_owner"),
            ("test variants not normalized", '.split(" [", 1)[0]', '', "test_test_variants_normalize_to_real_packages"),
            ("ordinary imports omitted", 'package.get("Imports", [])', '[]', "test_real_closure_then_one_test_edge"),
            ("test imports omitted", 'package.get("TestImports", [])', '[]', "test_real_closure_then_one_test_edge"),
            ("transitive walk removed", 'pending.append(dependent)', 'pass', "test_real_closure_then_one_test_edge"),
            ("opaque dependencies ignored", 'dependencies = self.compilerDependencies(directories, changed)', 'dependencies = {}', "test_opaque_gap_and_whole_oracle_are_selected"),
            ("compiler oracle narrowed", 'if compilerChanged or oracle not in changedPackages else self.selectOracle()', 'if False else self.selectOracle()', "test_opaque_gap_and_whole_oracle_are_selected"),
            ("computed invocation detection removed", 'return compilerMarker.search(source) or launchMarker.search(source)', 'return compilerMarker.search(source)', "test_census_rejects_computed_compiler_outside_stage1"),
            ("census restricted to stage1", '"ls-files").splitlines()', '"ls-files", "--", "stage1").splitlines()', "test_census_rejects_computed_compiler_outside_stage1"),
            ("removed map entry accepted", 'if missing or dead:', 'if False:', "test_census_rejects_removed_declaration"),
            ("base consumers red the change", 'missing = [named for path, named in missing if path in touchedFiles]', 'missing = [named for path, named in missing]', "test_an_undeclared_consumer_on_the_base_is_named_not_red"),
            ("branch map ignored", 'declarations.setdefault(field, {}).update(layer[field])', 'pass', "test_a_branch_declares_its_own_consumer_over_the_tools_map"),
            ("driver dependencies not validated", 'any(value not in directories for value in inputs)', 'False', "test_invalid_driver_map_fails_closed"),
            ("script census removed", 'if re.search(r"subprocess|child_process|', 'if False and re.search(r"subprocess|child_process|', "test_census_rejects_undeclared_script_driver"),
            ("driver selection removed", 'self.result["selected_compiler_drivers"] = sorted(', 'self.result["selected_compiler_drivers"] = sorted([] if True else ', "test_stage3_driver_declaration_and_selection"),
            ("census disabled", 'if missing or dead:', 'if False:', "test_census_rejects_a_new_compiler_runner"),
            ("invalid dependencies accepted", 'value not in directories', 'False', "test_invalid_map_fails_closed"),
            ("affected tests deferred", 'self.deferred = {}', 'pass', "test_opaque_gap_and_whole_oracle_are_selected"),
            ("fixture ancestry removed", 'while owner is None and ancestor not in', 'while False and ancestor not in', "test_embedding_and_fixture_ownership_survive"),
        ]
        for name, before, after, test in mutants:
            with self.subTest(mutant=name):
                self.assertIn(before, original)
                namespace = dict(run.__dict__)
                exec(compile(original.replace(before, after), run.__file__, "exec"), namespace)
                mutated = namespace["Gate"]
                with mock.patch.object(run.Gate, "touched", mutated.touched), mock.patch.object(run.Gate, "compilerDependencies", mutated.compilerDependencies):
                    result = unittest.TestResult()
                    ReverseDependencies(test).run(result)
                self.assertEqual(len(result.errors), 0, result.errors)
                self.assertEqual(len(result.failures), 1, (name, result.failures))


class PhaseUnitTests(unittest.TestCase):
    def test_inventory_and_patterns(self):
        tree = os.path.abspath(os.path.join(os.path.dirname(__file__), "../.."))
        names = run.wasiFixtures(tree)
        self.assertEqual(len(names), 36)
        self.assertEqual(names[-1], "requests")
        self.assertIn("internal/native/wasm/io.a", names)
        for name in names:
            pattern = run.fixturePattern(["TestWASI"], [name])
            self.assertEqual(len(pattern.split("/")), len(name.split("/")) + 1)
            self.assertTrue(all(part.startswith("^(") and part.endswith(")$") for part in pattern.split("/")))

    def test_inventory_refuses_unknown_registration(self):
        with tempfile.TemporaryDirectory() as tree:
            os.makedirs(os.path.join(tree, "internal/native"))
            with open(os.path.join(tree, "internal/native/wasm_test.go"), "w") as handle:
                handle.write('func TestWASI(t *testing.T) {\nfixtures := []string{\n"a",\n\t}\nt.Run(fixture, f)\nt.Run(dynamic, f)\n}')
            with self.assertRaises(ValueError):
                run.wasiFixtures(tree)

    def bare_gate(self, tree, out):
        gate = object.__new__(run.Gate)
        gate.arguments = types.SimpleNamespace(tree=tree, out=out, parallel=2, sha="abc")
        gate.result, gate.steps, gate.exits = {}, {}, {}
        gate.failures = []
        gate.fail = lambda name, detail: gate.failures.append(name)
        return gate

    def test_the_pool_lists_the_whole_gate_s_units_from_a_plain_checkout(self):
        with tempfile.TemporaryDirectory() as tree:
            os.makedirs(os.path.join(tree, "verify/catalog"))
            open(os.path.join(tree, "verify/catalog/check.sh"), "w").close()
            with open(os.path.join(tree, "verify/catalog/catalog.json"), "w") as handle:
                json.dump([{"number": 1, "name": "one"}, {"number": 12, "name": "twelve"}], handle)
            os.makedirs(os.path.join(tree, "internal/native"))
            with open(os.path.join(tree, "internal/native/wasm_test.go"), "w") as handle:
                handle.write('func TestWASI(t *testing.T) {\n\tfixtures := []string{\n\t\t"a.a",\n\t}\n\tt.Run(fixture, f)\n\tt.Run("requests", f)\n}\n')
            listed = subprocess.run([sys.executable, run.__file__, "--full", "--list-units", "--tree", tree], capture_output=True, text=True, check=True).stdout.split("\n")
            self.assertEqual([line for line in listed if line], ["coverage", "tools", "build", "vet", "wasi a.a", "wasi requests",
                                                                "catalog 1", "catalog 12"])

    def test_the_pool_lists_fast_non_test_units_from_a_plain_checkout(self):
        listed = realRun([sys.executable, run.__file__, "--list-units", "--tree", "/unused"],
                         capture_output=True, text=True, check=True, timeout=90).stdout.splitlines()
        self.assertEqual(listed, run.fastBasePhases)
        # smoke retired (#r8ghgkt): oracle's own units run every fixture it ran, so no plan lists it.
        self.assertNotIn("smoke", listed)
        self.assertIn("census", listed)
        self.assertNotIn("tests", listed)
        self.assertNotIn("wasi", listed)

    def test_one_catalog_entry_runs_alone_by_its_number(self):
        with tempfile.TemporaryDirectory() as tree:
            os.makedirs(os.path.join(tree, "verify/catalog"))
            open(os.path.join(tree, "verify/catalog/check.sh"), "w").close()
            with open(os.path.join(tree, "verify/catalog/catalog.json"), "w") as handle:
                json.dump([{"number": 1, "name": "one"}, {"number": 2, "name": "two"}], handle)
            gate = self.bare_gate(tree, tree)
            gate.arguments.unit = "2"
            entries = []

            def spawn(command, output, environment):
                entry = int(command[command.index("--entry") + 1])
                entries.append(entry)
                output.write("%02d %s: applies-and-fails-as-recorded\n" % (entry, ["one", "two"][entry - 1]))
                return types.SimpleNamespace(wait=lambda: 0)
            gate.spawn = spawn
            gate.catalogFull()
            self.assertEqual(entries, [2])
            self.assertEqual([row["name"] for row in gate.result["catalog_units"]], ["02 two"])
            # One entry alone, rightly skipped, is green: the at-least-one-caught rule is over the whole catalog.
            gate.arguments.unit = "1"

            def skipping(command, output, environment):
                output.write("01 one: skipped (fix not on main)\n")
                return types.SimpleNamespace(wait=lambda: 0)
            gate.spawn = skipping
            gate.failures.clear()
            gate.catalogFull()
            self.assertEqual(gate.failures, [])
            self.assertEqual(gate.exits["catalog"], 0)
            gate.arguments.unit = "9"
            with self.assertRaises(ValueError):
                gate.catalogFull()

    def test_catalog_verdicts_are_per_entry(self):
        cases = [(0, "applies-and-fails-as-recorded", True), (1, "no-longer-applies", True),
                 (0, "skipped (reason)", True), (1, "applies-but-failure-not-as-recorded", False),
                 (1, "error", False), (0, "", False), (1, "applies-and-fails-as-recorded", False)]
        with tempfile.TemporaryDirectory() as tree:
            os.makedirs(os.path.join(tree, "verify/catalog"))
            open(os.path.join(tree, "verify/catalog/check.sh"), "w").close()
            with open(os.path.join(tree, "verify/catalog/catalog.json"), "w") as handle:
                json.dump([{"number": i+1, "name": "case"+str(i)} for i in range(len(cases))], handle)
            gate = self.bare_gate(tree, tree)
            def spawn(command, output, environment):
                i = int(command[command.index("--entry")+1])-1
                self.assertEqual(command[command.index("--jobs")+1], "1")
                self.assertEqual(environment["GOMAXPROCS"], "1")
                self.assertEqual(environment["GOFLAGS"], "-p=1")
                self.assertTrue(os.path.isdir(environment["TMPDIR"]))
                output.write("%02d case%d: %s\n" % (i+1, i, cases[i][1]))
                return types.SimpleNamespace(wait=lambda: cases[i][0])
            gate.spawn = spawn
            gate.catalogFull()
            self.assertEqual([row["ok"] for row in gate.result["catalog_units"]], [case[2] for case in cases])
            self.assertEqual(set(gate.failures), {"catalog/04 case3", "catalog/05 case4", "catalog/06 case5", "catalog/07 case6"})
            cases[0] = (0, "skipped (reason)", True)
            gate.catalogFull()
            self.assertEqual(gate.exits["catalog"], 1)
            self.assertIn("catalog", gate.failures)
            with open(os.path.join(tree, "verify/catalog/catalog.json"), "w") as handle:
                json.dump([{"number": 1,"name":"a"},{"number":1,"name":"b"}],handle)
            with self.assertRaises(ValueError):
                gate.catalogFull()

    def test_wasi_requires_named_pass_and_exact_selection(self):
        with tempfile.TemporaryDirectory() as tree:
            gate = self.bare_gate(tree, tree)
            def stream(name, command, log, environment):
                pattern = command[command.index("-run")+1]
                self.assertEqual(command[command.index("-p")+1], "1")
                self.assertEqual(environment["PATH"], "sdk-first")
                self.assertEqual(environment["GOFLAGS"], "-p=1")
                self.assertEqual(environment["GOMAXPROCS"], "1")
                if "good" in pattern:
                    log.write(json.dumps({"Test":"TestWASI/good","Action":"pass"}))
                if "broad" in pattern:
                    log.write(json.dumps({"Test":"TestWASI/broad","Action":"pass"}))
                    log.write(json.dumps({"Test":"TestWASI/extra","Action":"run"}))
                return 0
            gate.stream = stream
            with mock.patch.object(run, "wasiFixtures", return_value=["good","missing","broad"]):
                gate.wasiSplit(io.StringIO(), {"PATH":"sdk-first"})
            self.assertEqual([row["ok"] for row in gate.result["wasi_units"]], [True,False,False])
            self.assertEqual(set(gate.failures), {"wasi/missing","wasi/broad"})

    def test_real_subprocess_planted_failures_are_named(self):
        tools = os.path.abspath(os.path.join(os.path.dirname(__file__), "../.."))
        with tempfile.TemporaryDirectory() as tree:
            os.makedirs(os.path.join(tree, "verify/catalog"))
            with open(os.path.join(tree, "verify/catalog/catalog.json"), "w") as handle:
                json.dump([{"number":1,"name":"good"},{"number":2,"name":"planted"}], handle)
            with open(os.path.join(tree, "verify/catalog/check.sh"), "w") as handle:
                handle.write('if [ "$2" = 1 ]; then echo "01 good: applies-and-fails-as-recorded"; else echo "02 planted: applies-but-failure-not-as-recorded"; exit 1; fi\n')
            os.makedirs(os.path.join(tree, "internal/native"))
            with open(os.path.join(tree, "go.mod"), "w") as handle:
                handle.write("module example.com/phase-probe\n\ngo 1.25\n")
            with open(os.path.join(tree, "internal/native/wasm_test.go"), "w") as handle:
                handle.write('package native\nimport "testing"\nfunc TestWASI(t *testing.T) {\n\tfixtures := []string{\n\t\t"good", "planted",\n\t}\nfor _, fixture := range fixtures { t.Run(fixture, func(t *testing.T) { if fixture == "planted" { t.Fatal("planted unit failure") } }) }\n}\n')
            for phase in ("catalog", "wasi"):
                out = tempfile.mkdtemp(dir=tree)
                arguments = types.SimpleNamespace(tree=tree, tools=tools, sha="probe", base="probe",
                    out=out, parallel=2, full=True, complete=True, branch="", branch_source="",
                    session="", session_source="")
                gate = run.Gate(arguments)
                gate.planned, gate.kind = [phase], "fast"
                with open(os.path.join(out,"test.jsonl"),"w") as log:
                    if phase == "catalog":
                        gate.catalogFull()
                    else:
                        gate.wasiSplit(log, {"PATH":os.environ["PATH"]})
                gate.finish()
                with open(os.path.join(out,"fast.json")) as handle:
                    result = json.load(handle)
                self.assertEqual(result[phase+"_unit_count"],2)
                self.assertEqual(result["failure"]["step"], "catalog/02 planted" if phase=="catalog" else "wasi/planted")
                # The rows whole on a miss, so a red says what the good unit did, its exit and status among them.
                self.assertEqual([row["name"] for row in result[phase+"_units"] if row["status"]=="failed"],
                                 ["02 planted"] if phase=="catalog" else ["planted"], json.dumps(result[phase+"_units"], indent=1))
                # The good unit either finished before the planted one failed or was cancelled by the whole gate's cancel,
                # never failed (#z89w67t).
                self.assertIn([row["status"] for row in result[phase+"_units"] if row["name"] in ("01 good", "good")],
                              (["passed"], ["cancelled"]), json.dumps(result[phase+"_units"], indent=1))
                if phase == "wasi":
                    self.assertIn("TestWASI/planted", result["failure"]["detail"])
                    self.assertIn("example.com/phase-probe/internal/native TestWASI/planted",result["failed_tests"])

    def test_slot_cpu_quota(self):
        with mock.patch.object(run.os, "sched_getaffinity", return_value=set(range(8)), create=True), \
                mock.patch("builtins.open", mock.mock_open(read_data="400000 100000")):
            self.assertEqual(run.slotCPUs(), 4)

    def test_units_coverage_failure_scratch_and_budget(self):
        with tempfile.TemporaryDirectory() as out:
            gate = object.__new__(run.Gate)
            gate.arguments = types.SimpleNamespace(out=out, parallel=2)
            gate.result, gate.steps, gate.exits = {}, {}, {}
            failures = []
            gate.fail = lambda name, detail: failures.append(name)
            active, maximum, lock = [0], [0], threading.Lock()
            def execute(name, command, scratch):
                with lock:
                    active[0] += 1
                    maximum[0] = max(maximum[0], active[0])
                time.sleep(0.01)
                with lock:
                    active[0] -= 1
                return {"ok": name != "bad", "exit": int(name == "bad")}
            gate.phaseUnits("catalog", ["good", "bad", "last"], [["x"]] * 3, execute)
            rows = gate.result["catalog_units"]
            self.assertEqual([r["name"] for r in rows], ["good", "bad", "last"])
            self.assertEqual(gate.result["catalog_unit_count"], 3)
            self.assertEqual(len(set(r["scratch"] for r in rows)), 3)
            self.assertEqual(maximum[0], 2)
            self.assertEqual(failures, ["catalog/bad"])
            self.assertEqual(gate.exits["catalog"], 1)
            with mock.patch.object(run.time, "monotonic", side_effect=[0, 0, 30.0004, 32]):
                gate.phaseUnits("wasi", ["slow"], [["x"]], lambda *args: {"ok": True})
            # Over the budget is recorded, not red: these phases' units drift over it until their builds are fetched.
            self.assertTrue(gate.result["wasi_units"][0]["over_budget"])
            self.assertEqual(gate.exits["wasi"], 0)
            self.assertNotIn("wasi/slow", failures)

    def test_a_unit_the_whole_gate_cancelled_reads_cancelled(self):
        # Pins #z89w67t without a race: the first unit's failure cancels the whole gate, as fail's full branch does, and
        # the second, which then ends on the cancel's kill (exit -9), reads cancelled and fails nothing. The phase stays
        # red. A unit that failed on its own exit after the cancel still reads failed. Mutant: the cancelled branch taken
        # out, which writes the killed unit failed beside the planted one.
        with tempfile.TemporaryDirectory() as out:
            gate = object.__new__(run.Gate)
            gate.arguments = types.SimpleNamespace(out=out, parallel=1)
            gate.result, gate.steps, gate.exits = {}, {}, {}
            failures = []
            gate.failure = None
            def fail(name, detail):
                failures.append(name)
                gate.failure = gate.failure or {"step": name}
                gate.cancelled = True
            gate.fail = fail
            # The planted unit's failing event fails the gate from inside its stream, as stream does, and the cancel then
            # kills its own process too, so it ends at -9 like its sibling; it caused the cancel, so it stays failed.
            def execute(name, command, scratch):
                if name == "planted":
                    gate.fail("wasi/planted", "TestWASI/planted failed")
                return {"ok": False, "exit": {"planted": -9, "killed": -9, "own": 2}[name]}
            gate.phaseUnits("wasi", ["planted", "killed", "own"], [["x"]] * 3, execute)
            self.assertEqual([row["status"] for row in gate.result["wasi_units"]], ["failed", "cancelled", "failed"])
            self.assertEqual(failures, ["wasi/planted", "wasi/planted", "wasi/own"])
            self.assertEqual(gate.exits["wasi"], 1)

class CacheDrains(unittest.TestCase):
    setUp = FailClosed.setUp
    gate = FailClosed.gate

    def tearDown(self):
        shutil.rmtree(self.directory)

    def publisher(self):
        directory = os.path.join(self.tree, "internal/buildcache/cmd/buildcache-publish")
        os.makedirs(directory)
        with open(os.path.join(directory, "main.go"), "w") as handle:
            handle.write("package main\nfunc main() {}\n")

    def probe(self, full=False, broken=None):
        initialize = FakeProcess.__init__
        calls = []
        key = "a" * 64

        def process(instance, command, stdout):
            initialize(instance, command, stdout)
            name = stage(command)
            calls.append(name)
            if name in ("audit", "upload"):
                instance.wait = lambda: instance.returncode
                if name == broken:
                    instance.returncode = 1
                    text = ("cache poisoning: the stored product for %s differs from a rebuild" % key
                            if name == "audit" else "upload: store unavailable")
                    instance.stdout = io.StringIO(text + "\n")
        with mock.patch.object(FakeProcess, "__init__", process):
            gate, status, result = self.gate(full=full)
        return gate, status, result, calls, key

    def test_poisoned_audit_is_red_naming_key(self):
        self.publisher()
        for full in (False, True):
            with self.subTest(full=full):
                gate, status, result, calls, key = self.probe(full, "audit")
                self.assertIn("first failure at audit", status)
                self.assertIn("cache poisoning", result["failure"]["detail"])
                self.assertIn(key, result["failure"]["detail"])
                self.assertNotEqual(result["stages_exit"]["audit"], 0)
                self.assertEqual(result["cache_drain_units"][0]["status"], "failed")
                self.assertTrue(any(row["test"] == "audit" and row["action"] == "fail" for row in result["units"]))

    def test_failed_upload_is_recorded_and_green(self):
        self.publisher()
        for full in (False, True):
            with self.subTest(full=full):
                gate, status, result, calls, key = self.probe(full, "upload")
                self.assertTrue(status.startswith("green:"), status)
                self.assertIsNone(gate.failure)
                self.assertEqual(result["stages_exit"]["upload"], 0)
                row = result["cache_drain_units"][1]
                self.assertEqual(row["exit"], 1)
                self.assertFalse(row["required"])
                self.assertIn("store unavailable", row["detail"])
                self.assertEqual(result["fail"], 0)
                self.assertLess(calls.index("audit"), calls.index("upload"))
                self.assertTrue(all(name not in ("tests", "wasi") for name in calls[calls.index("audit"):]))
                for name in ("audit", "upload"):
                    self.assertIn(name, result["planned_stages"])
                    self.assertIn(name, result["steps_seconds"])
                    self.assertTrue(any(unit["test"] == name for unit in result["units"]))

    def test_tree_without_publisher_plans_neither_drain(self):
        for full in (False, True):
            gate, status, result, calls, key = self.probe(full)
            self.assertTrue(status.startswith("green:"), status)
            for name in ("audit", "upload"):
                self.assertNotIn(name, result["planned_stages"])
                self.assertNotIn(name, calls)
            self.assertNotIn("cache_drain_units", result)

    def test_upload_spawn_failure_is_optional(self):
        self.publisher()
        gate, status, result = self.gate(broken="upload")
        self.assertTrue(status.startswith("green:"), status)
        self.assertIn("Too many open files", result["cache_drain_units"][1]["detail"])

    def test_every_spawned_go_command_builds_without_the_commit_or_the_checkout_path(self):
        # @system_adamic, Oct 9 16:33Z: floor1's identity check found main's stage 0, fetched at a candidate's commit,
        # 12 ranges away from the candidate's own fresh build (vcs stamp, build id); the path is the same kind of input.
        gate, _, _ = self.gate()
        seen = []
        def popen(command, **options):
            seen.append(options["env"]["GOFLAGS"])
            return realPopen([sys.executable, "-c", "pass"], **options)
        realPopen = subprocess.Popen
        with mock.patch.object(run.subprocess, "Popen", side_effect=popen):
            gate.spawn(["go", "test", "./probe"], subprocess.PIPE).communicate()
            gate.spawn(["go", "test", "./probe"], subprocess.PIPE, environment={"GOFLAGS": "-p=1"}).communicate()
        self.assertEqual(seen[0].split(), ["-buildvcs=false", "-trimpath"])
        self.assertEqual(seen[1].split(), ["-p=1", "-buildvcs=false", "-trimpath"])
        # Loom's pool takes the gate's environment from the selection, so it builds the same way.
        self.assertEqual(run.gateEnvironment["GOFLAGS"].split(), ["-buildvcs=false", "-trimpath"])

    def test_two_runs_keep_test_and_drain_queues_isolated(self):
        self.publisher()
        # Two runs of the very same slot must still get fresh queues. Real child
        # processes consume the environment passed by spawn, just as the publisher does.
        # Each is spawned into after its gate() has finished, which no run does, so their roots are kept until the end
        # and removed then (#tfp4gtm).
        with mock.patch.object(run.Gate, "removeBuildStore", lambda gate: None):
            first, _, _ = self.gate()
            second, _, _ = self.gate(full=True)
        shared = os.path.join(self.directory, "shared-default")
        os.makedirs(shared)
        key = "poisoned-gate-A-key"
        script = """import os, pathlib, sys
spool = pathlib.Path(os.environ.get('ADAMIC_BUILD_STORE_SPOOL', sys.argv[1]))
audits = pathlib.Path(os.environ.get('ADAMIC_BUILD_STORE_AUDITS', sys.argv[1]))
if sys.argv[2] == 'test':
    (audits / 'poison').write_text('poisoned-gate-A-key')
    (spool / 'product').write_text('gate-A-product')
elif sys.argv[2] == 'audit' and (audits / 'poison').exists():
    print('cache poisoning: ' + (audits / 'poison').read_text() + ' differs from a rebuild', file=sys.stderr)
    sys.exit(1)
elif sys.argv[2] == 'upload':
    print('products=' + str(len(list(spool.iterdir()))))
"""
        def popen(command, **options):
            name = stage(command)
            mode = name if name in ("audit", "upload") else "test"
            return realPopen([sys.executable, "-c", script, shared, mode], **options)
        # Explicit inherited defaults simulate the box's shared user cache too.
        with mock.patch.dict(os.environ, ADAMIC_BUILD_STORE_SPOOL=shared, ADAMIC_BUILD_STORE_AUDITS=shared), \
                mock.patch.object(run.subprocess, "Popen", side_effect=popen), mock.patch("builtins.print"):
            process = first.spawn(["go", "test", "./probe"], subprocess.PIPE)
            process.communicate()
            self.assertEqual(process.returncode, 0)
            second.cacheDrains()
            self.assertIsNone(second.failure, second.failure)
            self.assertEqual(second.result["cache_drain_units"][-1]["detail"], "products=0\n")
            first.cacheDrains()
            self.assertEqual(first.failure["step"], "audit")
            self.assertIn(key, first.failure["detail"])
        for gate in (first, second):
            for directory in gate.buildStoreEnvironment.values():
                self.assertEqual(os.path.dirname(os.path.dirname(directory)), os.path.dirname(self.tree))
                self.assertFalse(directory.startswith(gate.arguments.out + os.sep))
            root = gate.buildStoreRoot
            gate.removeBuildStore()
            self.assertFalse(os.path.exists(root))
        self.assertNotEqual(first.buildStoreEnvironment, second.buildStoreEnvironment)

    # The kill walks /proc for the unit's session: without it the sleeper is never killed and the drain waits it out.
    @unittest.skipUnless(sys.platform == "linux", "requires Linux /proc sessions")
    def test_drain_wall_deadlines_kill_processes(self):
        self.publisher()
        for name in ("audit", "upload"):
            gate, _, _ = self.gate()
            gate.failure = None
            gate.arguments.complete = True
            gate.complete = True
            original = gate.spawn
            def spawn(command, *args):
                return original([sys.executable, "-c", "import time; time.sleep(600)"], *args)
            with mock.patch.object(gate, "spawn", spawn), mock.patch.object(run, "unitKillSeconds", 1), mock.patch("builtins.print"):
                gate.cacheDrain(name)
            row = gate.result["cache_drain_units"][-1]
            self.assertNotEqual(row["exit"], 0)
            # The detail names the deadline that applied (18960baa), here the patched one.
            self.assertIn("%s killed at 1 s" % name, row["detail"])
            self.assertTrue(all(process.poll() is not None for process in gate.processes))
            self.assertEqual(gate.failure is not None, name == "audit")
            self.assertEqual(gate.exits[name] == 0, name == "upload")


class CacheDrainMutants(unittest.TestCase):
    def test_shared_default_queues_are_caught(self):
        with open(run.__file__) as handle:
            source = handle.read()
        needle = "            variables.update(self.buildStoreEnvironment)"
        self.assertEqual(source.count(needle), 1)
        namespace = dict(run.__dict__)
        exec(compile(source.replace(needle, "            pass # mutant: shared defaults restored"), run.__file__, "exec"), namespace)
        result = unittest.TestResult()
        with mock.patch.object(run.Gate, "spawn", namespace["Gate"].spawn):
            CacheDrains("test_two_runs_keep_test_and_drain_queues_isolated").run(result)
        self.assertEqual(result.errors, [])
        self.assertEqual(len(result.failures), 1, result.failures)
        self.assertIn("cache poisoning", result.failures[0][1])

    def test_dropped_audit_stage_is_caught(self):
        with open(run.__file__) as handle:
            source = handle.read()
        needle = '        self.cacheDrain("audit")'
        self.assertEqual(source.count(needle), 1)
        namespace = dict(run.__dict__)
        exec(compile(source.replace(needle, "        pass # mutant: audit dropped"), run.__file__, "exec"), namespace)
        result = unittest.TestResult()
        with mock.patch.object(run.Gate, "cacheDrains", namespace["Gate"].cacheDrains):
            CacheDrains("test_poisoned_audit_is_red_naming_key").run(result)
        self.assertEqual(result.errors, [])
        self.assertEqual(len(result.failures), 2, result.failures)


class WASIUnits(unittest.TestCase):
    """Main 7e403e44 split TestWASI's fixtures into top-level units: the whole gate's wasi phase enumerates them by
    name instead of crashing on the missing TestWASI (Oct 9 06:35Z: every pool whole gate's --list-units died)."""

    def tree(self, source):
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        os.makedirs(os.path.join(directory.name, "internal/native"))
        with open(os.path.join(directory.name, "internal/native/wasm_test.go"), "w") as handle:
            handle.write(source)
        return directory.name

    def test_top_level_units_are_the_phase_s_units(self):
        tree = self.tree("package native\n\nfunc TestWASIUnit00(t *testing.T) {\n\trunWASIUnit(t, 0)\n}\n\n"
                         "func TestWASIUnit01(t *testing.T) {\n\trunWASIUnit(t, 1)\n}\n")
        self.assertEqual(run.wasiFixtures(tree), ["TestWASIUnit00", "TestWASIUnit01"])
        self.assertTrue(re.fullmatch(run.wasiSkip, "TestWASIUnit01"))
        self.assertTrue(re.fullmatch(run.wasiSkip, "TestWASI"))
        self.assertFalse(re.fullmatch(run.wasiSkip, "TestWASIRequest"))

    def test_the_old_shape_still_enumerates_its_fixtures(self):
        tree = self.tree('func TestWASI(t *testing.T) {\n\tfixtures := []string{\n\t\t"a.a",\n\t}\n\tt.Run(fixture, f)\n\tt.Run("requests", f)\n}\n')
        self.assertEqual(run.wasiFixtures(tree), ["a.a", "requests"])

    def test_neither_shape_refuses(self):
        with self.assertRaises(ValueError):
            run.wasiFixtures(self.tree("package native\n"))

    def test_a_unit_runs_by_its_own_name(self):
        tree = self.tree("func TestWASIUnit07(t *testing.T) {\n\trunWASIUnit(t, 7)\n}\n")
        seen = {}
        gate = types.SimpleNamespace(arguments=types.SimpleNamespace(tree=tree), selectedUnits=lambda phase, names: names,
                                     phaseUnits=lambda phase, names, commands, execute: seen.update(zip(names, commands)))
        run.Gate.wasiSplit(gate, io.StringIO())
        command = seen["TestWASIUnit07"]
        self.assertEqual(command[command.index("-run") + 1], "^TestWASIUnit07$")

    def test_a_unit_s_own_fixture_subtests_are_expected_and_another_unit_is_not(self):
        # Oct 9 06:47Z: every unit read as red because its own fixture subtests counted as unexpected.
        tree = self.tree("func TestWASIUnit03(t *testing.T) {\n\trunWASIUnit(t, 3)\n}\n")
        executes = {}
        gate = types.SimpleNamespace(arguments=types.SimpleNamespace(tree=tree), selectedUnits=lambda phase, names: names,
                                     phaseUnits=lambda phase, names, commands, execute: executes.update(execute=execute))
        run.Gate.wasiSplit(gate, io.StringIO())
        for extra, ok in (([], True), (["TestWASIUnit04"], False)):
            events = [{"Action": "run", "Test": "TestWASIUnit03"},
                      {"Action": "run", "Test": "TestWASIUnit03/internal/load/testdata/0.1/compile/04_closures.ts"},
                      {"Action": "pass", "Test": "TestWASIUnit03/internal/load/testdata/0.1/compile/04_closures.ts"},
                      {"Action": "pass", "Test": "TestWASIUnit03"}] + [{"Action": "run", "Test": name} for name in extra]
            def stream(name, command, log, environment=None):
                for event in events:
                    log.write(json.dumps(event) + "\n")
                return 0
            gate.stream = stream
            with self.subTest(extra=extra):
                self.assertEqual(executes["execute"]("TestWASIUnit03", [], "/tmp")["ok"], ok)


class DarwinVet(unittest.TestCase):
    """The darwin vet (#gmrwkze): go vet of the darwin build on this Linux box, over the leg's packages and every
    darwin-only file's, so a darwin-only file that doesn't compile is red where it lands (add69cbc59 left
    internal/apple/witness_test.go calling a moved function, and main landed green). These run go for real, in a module
    of their own. Mutants, each red here: the darwin environment left off the vet; the darwin-only files read by name
    alone, or by build line alone, or testdata kept; the extra packages left out of the patterns; the inputs keyed by
    packages, never the darwin listing's directories; the darwin-only files left out of the inputs' names; a listing that
    can't be computed planned as nothing."""

    def setUp(self):
        self.directory = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, self.directory)
        self.tree = os.path.join(self.directory, "tree")
        self.sources = {
            "go.mod": "module " + run.module + "\n\ngo 1.24\n",
            run.compilerDependenciesPath: json.dumps({"version": 1, "packages": {"internal/apple": ["cmd/adamic"]}}),
            "cmd/adamic/main.go": "package main\n\nfunc main() {}\n",
            "internal/apple/apple.go": "package apple\n\n// Resolve is what the darwin-only witness calls.\nfunc Resolve() int { return 1 }\n",
            "internal/apple/witness_test.go": "//go:build darwin\n\npackage apple\n\nimport \"testing\"\n\nfunc TestWitness(t *testing.T) { _ = Resolve() }\n",
            # A darwin-only test outside the leg's closure, and a file every platform but darwin builds, in one package.
            "internal/other/other.go": "package other\n",
            "internal/other/moves_darwin_test.go": "package other\n\nimport \"testing\"\n\nfunc TestMoves(t *testing.T) {}\n",
            "internal/other/lock_other.go": "//go:build !darwin\n\npackage other\n",
            "internal/plain/plain.go": "package plain\n",
            # testdata is never a package, whatever its name says.
            "internal/plain/testdata/x_darwin.go": "package x\n",
        }
        for name, source in self.sources.items():
            self.write(name, source)
        realRun(["git", "-C", self.tree, "init", "-q"], check=True)
        self.commit("the darwin vet's module")

    def write(self, name, source):
        path = os.path.join(self.tree, name)
        os.makedirs(os.path.dirname(path), exist_ok=True)
        with open(path, "w") as handle:
            handle.write(source)

    def commit(self, message):
        realRun(["git", "-C", self.tree, "add", "-A", "."], check=True)
        realRun(["git", "-C", self.tree, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", message], check=True)

    def vet(self):
        """Gate.darwinVet run on a gate of just what it reads, its failure and step kept."""
        out = os.path.join(self.directory, "out")
        os.makedirs(out, exist_ok=True)
        gate = types.SimpleNamespace(arguments=types.SimpleNamespace(tree=self.tree, tools=self.tree, out=out, full=False),
                                     steps={}, exits={}, result={}, failure=None, complete=False, lock=threading.Lock(),
                                     buildStoreEnvironment={}, processes=[])
        gate.fail = lambda step, detail, evidenceMissing=False: setattr(gate, "failure", gate.failure or (step, detail))
        gate.recordTestStart = lambda command: None
        gate.spawn = lambda *arguments, **options: run.Gate.spawn(gate, *arguments, **options)
        run.Gate.darwinVet(gate)
        return gate

    def test_darwin_only_files_are_found_by_name_and_by_build_line_never_in_testdata(self):
        self.assertEqual(run.darwinOnlyFiles(self.tree),
                         ["internal/apple/witness_test.go", "internal/other/lock_other.go", "internal/other/moves_darwin_test.go"])

    def test_the_vet_takes_the_leg_its_compilers_and_each_darwin_only_package(self):
        self.assertEqual(run.darwinVetPatterns(self.tree, self.tree), ["./internal/apple/...", "./cmd/adamic", "./internal/other"])

    def test_a_darwin_only_file_that_doesn_t_compile_is_red_naming_it_where_a_linux_vet_is_green(self):
        gate = self.vet()
        self.assertIsNone(gate.failure)
        self.assertEqual(gate.exits["darwin-vet"], 0)
        # add69cbc59's shape: the darwin-only test calls a name its package no longer holds.
        self.write("internal/apple/witness_test.go", self.sources["internal/apple/witness_test.go"].replace("Resolve()", "native.ResolveTarget()"))
        self.commit("the witness calls a moved function")
        linux = realRun(["go", "vet", "./internal/apple/..."], cwd=self.tree, capture_output=True, text=True, env=dict(os.environ, GOFLAGS="-mod=mod"))
        self.assertEqual(linux.returncode, 0, linux.stderr)
        gate = self.vet()
        self.assertEqual(gate.exits["darwin-vet"], 1)
        self.assertEqual(gate.failure[0], "darwin-vet")
        self.assertIn("internal/apple/witness_test.go:7", gate.failure[1])
        self.assertIn("undefined: native", gate.failure[1])

    def test_a_darwin_only_test_outside_the_leg_that_doesn_t_compile_is_red(self):
        self.write("internal/other/moves_darwin_test.go", self.sources["internal/other/moves_darwin_test.go"].replace("{}", "{ missing() }"))
        self.commit("a darwin-only test outside the leg breaks")
        gate = self.vet()
        self.assertEqual(gate.exits["darwin-vet"], 1)
        self.assertIn("internal/other/moves_darwin_test.go", gate.failure[1])

    def test_the_inputs_cover_each_darwin_only_package_by_path_and_each_darwin_only_file_by_name(self):
        inputs = run.darwinVetInputs(self.tree, "", False, None)
        self.assertEqual(inputs["packages"], [])
        # A change to the darwin-only test outside the leg's closure moves the key, since its directory is a path the
        # planner hashes by git ls-files -s; internal/plain, outside both, never does.
        for path in (":(literal)internal/apple", ":(literal)cmd/adamic", ":(literal)internal/other", ":(literal)go.mod"):
            self.assertIn(path, inputs["paths"])
        self.assertNotIn(":(literal)internal/plain", inputs["paths"])
        self.assertNotIn(".", inputs["paths"])
        self.assertEqual(inputs["names"], run.darwinOnlyFiles(self.tree))
        # A darwin-only file added in a package the vet didn't take moves the key by its name, before its directory does.
        self.write("internal/fresh/fresh_darwin.go", "package fresh\n")
        self.commit("a darwin-only file in a new package")
        self.assertIn("internal/fresh/fresh_darwin.go", run.darwinVetInputs(self.tree, "", False, None)["names"])

    def test_a_listing_that_can_t_be_computed_keys_on_the_whole_tree_and_the_phase_is_red(self):
        os.remove(os.path.join(self.tree, run.compilerDependenciesPath))
        self.commit("no compiler map")
        self.assertEqual(run.darwinVetInputs(self.tree, "", False, None)["paths"], ["."])
        gate = self.vet()
        self.assertEqual(gate.exits["darwin-vet"], 1)
        self.assertIn("can't be listed", gate.failure[1])


class PhaseInputs(unittest.TestCase):
    tree = os.path.abspath(os.path.join(os.path.dirname(__file__), "../.."))

    def listing(self, full=False, inputs=False):
        command = [sys.executable, run.__file__, "--list-units", "--tree", self.tree]
        if full:
            command.append("--full")
        if inputs:
            command.append("--with-inputs")
        return realRun(command, capture_output=True, text=True, check=True, timeout=90).stdout

    def test_plain_listing_is_byte_for_byte_the_base(self):
        base = realRun(["git", "-C", self.tree, "show", "eef6f75d:cloud/fast-gate/run.py"],
                       capture_output=True, text=True, check=True, timeout=10).stdout
        for full in (False, True):
            command = [sys.executable, "-", "--list-units", "--tree", self.tree]
            if full:
                command.append("--full")
            expected = realRun(command, input=base, capture_output=True, text=True, check=True, timeout=90).stdout
            # The determinism phase retired to internal/oracle's TestStage0EmitsDeterministically (#86znq4s), and stage 3's
            # to stage3/lane's tests (#4gp2ytz), and smoke to oracle's own units (#r8ghgkt): the lines the base lists that
            # this doesn't.
            expected = "".join(line for line in expected.splitlines(keepends=True)
                               if line not in ("determinism\n", "smoke\n", "stage3\n") and not line.startswith("stage3 "))
            if not full:
                # The darwin vet (#gmrwkze) is the one fast phase the base doesn't list, after vet.
                expected = expected.replace("vet\n", "vet\ndarwin-vet\n", 1)
            self.assertEqual(self.listing(full), expected)

    def test_every_listed_unit_has_nonempty_inputs(self):
        for full in (False, True):
            rows = [json.loads(line) for line in self.listing(full, True).splitlines()]
            self.assertEqual([row["unit"] for row in rows], self.listing(full).splitlines())
            for row in rows:
                inputs = row["inputs"]
                self.assertTrue(inputs["packages"] or inputs["paths"], row)
                if "." in inputs["paths"]:
                    self.assertTrue(inputs.get("note"), row)
                self.assertTrue(all(not os.path.isabs(path) and ".." not in path.split("/")
                                    for path in inputs["paths"]), row)

    def covered(self, inputs, path):
        return any(root == "." or path == root or path.startswith(root + "/")
                   for root in inputs["paths"])

    def test_wasi_inputs_cover_inventory_and_command_working_directories(self):
        # Independently check the compiler's fixture inventory, not the provider's path list.
        with open(os.path.join(self.tree, "internal/native/wasm_test.go")) as handle:
            source = handle.read()
        fixture = re.search(r'fixtures := \[\]string\{\s*"([^"]+)"', source)[1]
        self.assertIn(fixture, run.wasiFixtures(self.tree))
        commands = []
        gate = types.SimpleNamespace(arguments=types.SimpleNamespace(tree=self.tree),
            selectedUnits=lambda phase, names: [fixture],
            phaseUnits=lambda phase, names, planned, execute: commands.extend(planned))
        run.Gate.wasiSplit(gate, io.StringIO())
        self.assertEqual(commands[0][-1], "./internal/native")
        inputs = run.phaseInputs(self.tree, "wasi " + fixture, True)["inputs"]
        self.assertIn(run.module + "/internal/native", inputs["packages"])
        self.assertIn(run.module + "/cmd/adamic", inputs["packages"])
        # go test's working directory is internal/native, whose test resolves ../..
        # to the repository for compiler/Node commands. Node reads hooks and fixture siblings.
        for path in (fixture, os.path.dirname(fixture) + "/sibling.ts",
                     "oracle/node.mjs", "oracle/adamic.mjs", "internal/native/wasm/run.mjs",
                     "internal/native/wasm/request-abi.c", "internal/oracle/testdata/files/input.txt"):
            self.assertTrue(self.covered(inputs, path), (path, inputs))

    def test_catalog_inputs_cover_inventory_patch_and_root_working_directory(self):
        import shlex
        entry = next(entry for entry in run.catalogEntries(self.tree) if entry.get("command"))
        inputs = run.phaseInputs(self.tree, "catalog %d" % entry["number"], True)["inputs"]
        # check.py runs the inventory's command in a disposable repository-root worktree.
        command = shlex.split(entry["command"])
        package = next(arg for arg in command if arg.startswith("./"))
        self.assertIn(run.module + package[1:], inputs["packages"])
        for path in ("verify/catalog/check.sh", "verify/catalog/check.py", "verify/catalog/catalog.json",
                     "verify/catalog/" + entry["patch"], entry["fixture"], package[2:] + "/testdata/input.a"):
            self.assertTrue(self.covered(inputs, path), (path, inputs))

    def test_only_catalog_apply_declares_it_reads_the_base(self):
        # catalog-apply's patch application reads the base tree beside the candidate's, and its inputs say so in a typed
        # field the planner carries onto the planned unit (#witnessmain, Node's rule 1); no other phase's does, the note's
        # words included. Mutant: readsBase in treeInputs, on every whole-tree phase, red.
        declared = {}
        for phase, provider in run.phaseInputProviders.items():
            try:
                declared[phase] = provider(self.tree, "", False, None)
            except Exception:
                # A provider that needs a unit or a change of its own can't be asked bare; the whole-tree ones can.
                continue
        self.assertIs(declared["catalog-apply"].get("readsBase"), True)
        self.assertNotIn("readsBase", run.treeInputs("any note"))
        self.assertEqual(sorted(phase for phase, inputs in declared.items() if "readsBase" in inputs), ["catalog-apply"])
        self.assertIs(run.phaseInputs(self.tree, "catalog-apply", False)["inputs"].get("readsBase"), True)

    def test_tools_reads_the_go_sources_it_greps_and_no_more(self):
        # The tools phase declares exactly the pathspec its git grep searches (#gw93f6b): one list, so the two can't drift,
        # and never the whole tree. Mutant: the grep searching a wider pathspec than the declaration, red.
        declared = run.phaseInputProviders["tools"](self.tree, "", False, None)
        # What the grep searches, and every .gitattributes, which git grep reads beside each file (the Pupa trace).
        self.assertEqual(declared["paths"], run.toolsSourcePathspec + run.toolsAttributesPathspec)
        attributes = realRun(["git", "-C", self.tree, "ls-files", "--"] + run.toolsAttributesPathspec,
                             capture_output=True, text=True, check=True, timeout=30).stdout.split()
        self.assertIn(".gitattributes", attributes)
        self.assertTrue(any(path.endswith("/.gitattributes") for path in attributes), attributes)
        self.assertNotIn(".", run.phaseInputs(self.tree, "tools", False)["inputs"]["paths"])
        searched = []
        gate = types.SimpleNamespace(arguments=types.SimpleNamespace(tree=self.tree, tools=self.tree), steps={}, result={}, exits={},
                                     command=lambda command, **keywords: searched.append(command) or types.SimpleNamespace(returncode=1, stdout="", stderr=""),
                                     fail=lambda phase, why: None)
        run.Gate.toolsDeclared(gate)
        self.assertEqual(searched[0][searched[0].index("--") + 1:], run.toolsSourcePathspec)

    def test_wasi_fixture_path_mutant_is_caught(self):
        original = run.phaseInputProviders["wasi"]
        fixture = run.wasiFixtures(self.tree)[0]
        def mutant(tree, unit, full, change):
            inputs = original(tree, unit, full, change)
            inputs["paths"].remove(os.path.dirname(fixture))
            return inputs
        with mock.patch.dict(run.phaseInputProviders, wasi=mutant):
            result = unittest.TestResult()
            PhaseInputs("test_wasi_inputs_cover_inventory_and_command_working_directories").run(result)
        self.assertEqual(len(result.failures), 1, result.errors)
        self.assertEqual(result.errors, [])


class CoverageNames(unittest.TestCase):
    """Coverage keys on its own classification of the change (#gw93f6b, Node's ruling): the changed paths neither in a Go
    package nor inert, by run.coverageReduced alone, which cover() decides its red from and --with-inputs names. A README
    alone reduces to nothing, so the planner can reuse an earlier pass. Mutants, each red here: cover() deciding from its
    own rule loop; inert paths kept; owned paths kept; a deleted path dropped."""

    def setUp(self):
        scratch = tempfile.TemporaryDirectory()
        self.addCleanup(scratch.cleanup)
        # One level down: a listing's Gate keeps its build-store directories beside the tree, inside the scratch.
        self.tree = os.path.join(scratch.name, "tree")
        sources = {
            "go.mod": "module " + run.module + "\ngo 1.24\n",
            "a/a.go": "package a\n",
            "README.md": "# probe\n",
            "tools/kept.py": "kept\n",
            "tools/gone.sh": "gone\n",
            "cloud/fast-gate/compiler-dependencies.json": '{"version":1,"packages":{}}',
            "cloud/fast-gate/unit-needs.json": '{"version":1,"units":[]}',
            "cloud/fast-gate/executors.txt": "mode enforce\ninert *.md\ninert cloud/*\nbuild tools/*\n",
        }
        for name, source in sources.items():
            path = os.path.join(self.tree, name)
            os.makedirs(os.path.dirname(path), exist_ok=True)
            with open(path, "w") as handle:
                handle.write(source)
        self.base = self.commit("base")
        # The change: the README, a package's source, a ruled unowned file, a new file no rule matches, a deleted one.
        for name, source in (("README.md", "# probe, edited\n"), ("a/a.go", "package a\n\nvar A = 1\n"),
                             ("tools/kept.py", "kept, edited\n"), ("stray/unruled.txt", "stray\n")):
            os.makedirs(os.path.dirname(os.path.join(self.tree, name)), exist_ok=True)
            with open(os.path.join(self.tree, name), "w") as handle:
                handle.write(source)
        os.remove(os.path.join(self.tree, "tools/gone.sh"))
        self.sha = self.commit("change")

    def commit(self, message):
        for command in (["init", "-q"], ["add", "-A"], ["-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", message]):
            realRun(["git", "-C", self.tree] + command, check=True)
        return run.git(self.tree, "rev-parse", "HEAD")

    def reduced(self, changed):
        return run.coverageReduced(self.tree, self.tree, changed)

    def test_a_readme_alone_reduces_to_nothing(self):
        self.assertEqual(self.reduced(["README.md"]), [])
        # adamic's own map rules a README inert too, its ownership given as cover() gives it.
        repository = os.path.abspath(os.path.join(os.path.dirname(__file__), "../.."))
        self.assertEqual(run.coverageReduced(repository, repository, ["README.md"], ["README.md"]), [])

    def test_a_go_file_in_a_package_reduces_to_nothing(self):
        self.assertEqual(self.reduced(["a/a.go"]), [])

    def test_an_unowned_path_a_rule_runs_stays(self):
        self.assertEqual(self.reduced(["tools/kept.py"]), ["tools/kept.py"])

    def test_an_unowned_path_no_rule_matches_stays(self):
        self.assertEqual(self.reduced(["stray/unruled.txt"]), ["stray/unruled.txt"])

    def test_a_deleted_unowned_path_stays_by_name(self):
        self.assertFalse(os.path.exists(os.path.join(self.tree, "tools/gone.sh")))
        self.assertEqual(self.reduced(["tools/gone.sh"]), ["tools/gone.sh"])

    def test_the_listing_names_what_coverage_judges(self):
        def listed(*arguments):
            output = realRun([sys.executable, run.__file__, "--list-units", "--with-inputs", "--tree", self.tree] + list(arguments),
                             capture_output=True, text=True, check=True, timeout=120).stdout
            return {row["unit"]: row["inputs"] for row in map(json.loads, output.splitlines())}
        rows = listed("--base", self.base, "--sha", self.sha, "--tools", self.tree)
        changed = [path for path in run.git(self.tree, "diff", "--name-only", self.base + "..." + self.sha).split("\n") if path]
        self.assertEqual(rows["coverage"]["names"], run.coverageReduced(self.tree, self.tree, changed))
        self.assertEqual(rows["coverage"]["names"], ["stray/unruled.txt", "tools/gone.sh", "tools/kept.py"])
        # The Go state, and the judged paths by content too (#gamwqs2), never the whole tree.
        self.assertEqual(rows["coverage"]["paths"], sorted(set(run.coverageGoStatePathspec + [":(literal)" + path for path in rows["coverage"]["names"]] + run.phaseSetupPaths)))
        self.assertEqual([unit for unit, inputs in rows.items() if inputs.get("names")], ["coverage"])
        # Without a change there is nothing to classify, and the note says so.
        bare = listed()["coverage"]
        self.assertEqual(bare["names"], [])
        self.assertFalse([path for path in bare["paths"] if path.startswith(":(literal)")])
        # The whole gate's coverage reads the sha against its first parent, and a root commit changes nothing it judges.
        whole = {"base": None, "sha": self.sha, "tools": self.tree}
        self.assertEqual(run.phaseInputs(self.tree, "coverage", True, ([], False), whole)["inputs"]["names"], rows["coverage"]["names"])
        self.assertEqual(run.phaseInputs(self.tree, "coverage", True, ([], False), dict(whole, sha=self.base))["inputs"]["names"], [])

    def test_cover_and_the_listing_both_classify_through_coverage_reduced(self):
        gate = run.Gate.__new__(run.Gate)
        gate.arguments = types.SimpleNamespace(tree=self.tree, tools=self.tree)
        gate.result, gate.steps, gate.exits = {}, {}, {}
        changed = ["README.md", "a/a.go", "stray/unruled.txt", "tools/gone.sh", "tools/kept.py"]
        unowned = ["README.md", "stray/unruled.txt", "tools/gone.sh", "tools/kept.py"]
        with mock.patch.object(run, "coverageReduced", wraps=run.coverageReduced) as reduced:
            gate.cover(unowned, changed, record=False)
            self.assertEqual(reduced.call_count, 1)
            self.assertEqual(gate.result["uncovered_files"], ["stray/unruled.txt"])
            run.phaseInputs(self.tree, "coverage", False, None, {"base": self.base, "sha": self.sha, "tools": self.tree})
            self.assertEqual(reduced.call_count, 2)
        # cover()'s red follows the function: a path it doesn't keep is never uncovered.
        with mock.patch.object(run, "coverageReduced", return_value=[]):
            gate.cover(unowned, changed, record=False)
        self.assertEqual(gate.result["uncovered_files"], [])


class CoverageBounded(unittest.TestCase):
    """Coverage keys on the tree's Go state and, by name, the changed paths it judges, at every change (#gw93f6b and
    #gamwqs2, Node's rulings): two edits of files it doesn't read key alike, anything in the Go state moves its key, and a
    path it judges (unowned and not inert) is in its names. Mutants, each red here: cohere left out of the Go state; the
    names dropped; an inert path named; the whole tree kept at a Go change."""

    def setUp(self):
        scratch = tempfile.TemporaryDirectory()
        self.addCleanup(scratch.cleanup)
        self.tree = os.path.join(scratch.name, "tree")
        sources = {
            "go.mod": "module " + run.module + "\ngo 1.24\n",
            # A workspace beside the root module, as adamic's go.work uses cohere's: its files key coverage too.
            "go.work": "go 1.24\n\nuse (\n\t.\n\t./sub\n)\n",
            "go.work.sum": "",
            "sub/go.mod": "module example.com/sub\ngo 1.24\n",
            "sub/go.sum": "",
            "sub/s.go": "package sub\n",
            "a/a.go": 'package a\n\nimport _ "embed"\n\n//go:embed doc.md\nvar Doc string\n',
            "a/doc.md": "# embedded\n",
            "README.md": "# probe\n",
            "docs/guide.md": "# guide\n",
            "apple/README.md": "# apple\n",
            "stage3/drivers/parser/evidence/probe.py": "print(1)\n",
            # Files nothing reads (#cb8bpe7's wall 2): a log the map rules inert, a C source it rules inert, and a file no
            # rule names.
            "notes/counts.log": "6220\n",
            "notes/probe.c": "int probe;\n",
            "tools/unruled.txt": "x\n",
            "cloud/fast-gate/compiler-dependencies.json": '{"version":1,"packages":{}}',
            "cloud/fast-gate/unit-needs.json": '{"version":1,"units":[]}',
            "cloud/fast-gate/executors.txt": "mode enforce\ndarwin apple/*\ninert *.md\ninert */evidence/*\ninert cloud/*\ninert notes/*\n",
        }
        for name, source in sources.items():
            self.write(name, source)
        for command in (["init", "-q"], ["add", "-A"]):
            realRun(["git", "-C", self.tree] + command, check=True)
        # cohere is a submodule in adamic, keyed by its commit: a gitlink here, at the first commit's own sha.
        self.cohere = self.commit("base")
        realRun(["git", "-C", self.tree, "update-index", "--add", "--cacheinfo", "160000,%s,cohere" % self.cohere], check=True)
        os.makedirs(os.path.join(self.tree, "cohere"))  # a submodule not checked out: its directory, empty
        self.base = self.commit("cohere")

    def write(self, name, source):
        path = os.path.join(self.tree, name)
        os.makedirs(os.path.dirname(path), exist_ok=True)
        with open(path, "w") as handle:
            handle.write(source)

    def commit(self, message):
        realRun(["git", "-C", self.tree, "add", "-A"], check=True)
        realRun(["git", "-C", self.tree, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", message, "--allow-empty"], check=True)
        return run.git(self.tree, "rev-parse", "HEAD")

    def change(self, edit, on=None):
        on = on or self.base
        realRun(["git", "-C", self.tree, "checkout", "-q", "--detach", on], check=True)
        edit()
        return {"base": on, "sha": self.commit("change"), "tools": self.tree}

    def readmeOn(self, edit):
        # A README edit on a Go state the edit made.
        return self.change(lambda: self.write("README.md", "# probe, edited\n"), self.change(edit)["sha"])

    def inputs(self, change):
        return run.phaseInputs(self.tree, "coverage", False, None, change)["inputs"]

    def keyed(self, change):
        # What the planner keys a bounded phase's paths on: git ls-files -s of them, the gitlink by its commit.
        realRun(["git", "-C", self.tree, "checkout", "-q", "--detach", change["sha"]], check=True)
        inputs = self.inputs(change)
        return inputs["paths"], inputs["names"], run.git(self.tree, "ls-files", "-s", "--", *inputs["paths"])

    def test_edits_coverage_neither_reads_nor_judges_key_alike_on_the_go_state(self):
        readme = self.change(lambda: self.write("README.md", "# probe, edited\n"))
        guide = self.change(lambda: self.write("docs/guide.md", "# guide, edited\n"))
        log = self.change(lambda: self.write("notes/counts.log", "6224\n"))
        self.assertNotIn(".", self.inputs(readme)["paths"])
        self.assertIn("cohere", self.inputs(readme)["paths"])
        self.assertEqual(self.keyed(readme), self.keyed(guide))
        self.assertEqual(self.keyed(readme), self.keyed(log))
        self.assertIn("160000 %s 0\tcohere" % self.cohere, self.keyed(readme)[2])

    def test_the_go_state_moves_the_key(self):
        readme = self.change(lambda: self.write("README.md", "# probe, edited\n"))
        states = {
            "cohere moved": lambda: realRun(["git", "-C", self.tree, "update-index", "--cacheinfo", "160000,%s,cohere" % self.base], check=True),
            "the root go.mod": lambda: self.write("go.mod", "module " + run.module + "\ngo 1.24\n\n// edited\n"),
            "a nested module's go.mod": lambda: self.write("sub/go.mod", "module example.com/sub\ngo 1.24\n\n// edited\n"),
            "a nested module's go.sum": lambda: self.write("sub/go.sum", "example.com/x v1.0.0 h1:x=\n"),
            "the workspace's go.work": lambda: self.write("go.work", "go 1.25\n\nuse (\n\t.\n\t./sub\n)\n"),
            "the workspace's go.work.sum": lambda: self.write("go.work.sum", "example.com/x v1.0.0 h1:x=\n"),
            "a nested module's source": lambda: self.write("sub/s.go", "package sub\n\nvar S = 1\n"),
            "a C source": lambda: self.write("c/x.c", "int x;\n"),
            "a driver script": lambda: self.write("stage3/drivers/parser/evidence/probe.py", "print(2)\n"),
        }
        for name, edit in states.items():
            with self.subTest(name):
                self.assertNotEqual(self.keyed(readme)[2], self.keyed(self.readmeOn(edit))[2])

    def test_a_go_change_keys_on_the_go_state_and_names_what_it_judges(self):
        # No whole tree at a Go change (#gamwqs2): the Go state, and by name the unowned paths no rule makes inert.
        for name, edit, names in (
                ("a package's source", lambda: self.write("a/a.go", 'package a\n\nimport _ "embed"\n\n//go:embed doc.md\nvar Doc, Other string\n'), []),
                ("a file no rule names", lambda: self.write("tools/unruled.txt", "y\n"), ["tools/unruled.txt"]),
                ("markdown under a darwin rule", lambda: self.write("apple/README.md", "# apple, edited\n"), ["apple/README.md"]),
                ("an inert log beside a go edit", lambda: (self.write("notes/counts.log", "6224\n"), self.write("a/a.go", "package a\n")), []),
                ("nothing", lambda: None, [])):
            with self.subTest(name):
                inputs = self.inputs(self.change(edit))
                self.assertLessEqual(set(run.coverageGoStatePathspec), set(inputs["paths"]))
                self.assertNotIn(".", inputs["paths"])
                self.assertEqual(inputs["names"], names)
                self.assertEqual([path for path in inputs["paths"] if path.startswith(":(literal)")], [":(literal)" + path for path in names])

    def test_a_path_coverage_judges_moves_its_key_against_its_base(self):
        # A compare keys the change's tree and its base from the same change, so the names alone are alike on both; the
        # judged paths' content moves the key, for an edit, an add and a delete alike (#gamwqs2), so coverage runs and reds.
        for name, edit in (("an edit", lambda: self.write("tools/unruled.txt", "y\n")),
                           ("an add", lambda: self.write("tools/new.txt", "n\n")),
                           ("a delete", lambda: os.remove(os.path.join(self.tree, "tools/unruled.txt")))):
            with self.subTest(name):
                change = self.change(edit)
                tip = self.keyed(change)
                realRun(["git", "-C", self.tree, "checkout", "-q", "--detach", change["base"]], check=True)
                self.assertNotEqual(run.git(self.tree, "ls-files", "-s", "--", *tip[0]), tip[2])

    def test_no_change_names_nothing(self):
        self.assertEqual(run.phaseInputs(self.tree, "coverage", False, None, None)["inputs"]["names"], [])


class PhaseUnreadAgrees(unittest.TestCase):
    """run.py --phase-unread names the tracked files no phase reads when modified (#nz7hm5j, Node's ruling), and the
    listing agrees, both ways. For each file it names, a change editing it alone lists only the base phases, each with the
    same inputs on the change's tree as on its base's, and what those inputs read (their paths' content and their packages'
    own files) equal on both: no phase key can move, so a lookup deciding it empty is no false green. For each regular file
    it doesn't name, some phase reads what the edit moved, so the lookup sends it to the compare (one that moved nothing is
    an over-read). The set is derived from the base phases' declared inputs (#gamwqs2), so every file a phase declares
    reads as read: the Go state, a file in a package's directory (an embedded one among them), cloud/fast-gate, a
    .gitattributes; and so is a file whose change selects a phase (an .a, a catalog patch's target) and an inert
    executable, since the set names plain files alone. An inert-ruled oracle fixture or runtime file reads as unread:
    oracle's own test units read them, and no phase has since the smoke phase retired (#r8ghgkt). Mutants, red here: the
    tools phase's .gitattributes dropped, coverage's Go state dropped, the selection tests dropped (an .a, a catalog
    patch's target), and 100755 listed beside 100644."""

    setUp_, write, commit = CoverageBounded.setUp, CoverageBounded.write, CoverageBounded.commit

    def setUp(self):
        self.setUp_()
        # The shapes a file nothing else reads can still take, each under an inert rule: an .a (a-check), a .gitattributes
        # (tools), a catalog undo patch's target (catalog-apply), and an oracle fixture and runtime file, which no phase
        # reads since smoke retired.
        self.write("notes/gen.a", "a\n")
        self.write("notes/.gitattributes", "*.log text\n")
        self.write("notes/target.txt", "t\n")
        self.write("verify/catalog/01-probe.patch", "--- a/notes/target.txt\n+++ b/notes/target.txt\n@@ -1 +1 @@\n-t\n+u\n")
        self.write("internal/oracle/testdata/note.txt", "fixture\n")
        self.write("oracle/note.txt", "runtime\n")
        with open(os.path.join(self.tree, "cloud/fast-gate/executors.txt"), "a") as handle:
            handle.write("inert internal/oracle/testdata/*\ninert oracle/*\n")
        # An inert executable (Node's must on 426b75d6): the set names plain files alone, as the lookup decides them.
        self.write("notes/tool.log", "x\n")
        os.chmod(os.path.join(self.tree, "notes/tool.log"), 0o755)
        # The darwin leg's package, as adamic always holds it: darwin-vet (#gmrwkze) is a base phase, and a tree with no
        # leg lists no darwin closure, so it keys on the whole tree and nothing reads as unread.
        self.write("internal/apple/apple.go", "package apple\n")
        self.base = self.commit("phase shapes")

    def observed(self, change):
        """The listing of change on its own tree and on its base's, each with what its inputs read there."""
        argv = ["run.py", "--list-units", "--with-inputs", "--tree", self.tree, "--base", change["base"], "--sha", change["sha"], "--tools", self.tree]
        sides = []
        for checkout in (change["sha"], change["base"]):
            realRun(["git", "-C", self.tree, "checkout", "-q", "--detach", checkout], check=True)
            output = io.StringIO()
            try:
                with mock.patch.object(sys, "argv", argv), contextlib.redirect_stdout(output):
                    run.listUnits()
            finally:
                run.listedGoLists = None
            # census is the judge's step, never planned (loom planner.JudgedPhase), so no key of it can move.
            rows = [row for row in map(json.loads, output.getvalue().splitlines()) if row["unit"] != "census"]
            directories = {package: directory for directory, package in run.goPackages(self.tree, run.packageGraph(self.tree).stdout)[0].items()}
            read = []
            for row in rows:
                inputs = row["inputs"]
                owned = [":(glob)%s/*" % directories[package] for package in inputs["packages"] if package in directories]
                read.append((row["unit"], run.git(self.tree, "ls-files", "-s", "--", *inputs["paths"]) if inputs["paths"] else "",
                             run.git(self.tree, "ls-files", "-s", "--", *owned) if owned else ""))
            sides.append((rows, read))
        realRun(["git", "-C", self.tree, "checkout", "-q", "--detach", self.base], check=True)
        return sides

    def edited(self, path):
        realRun(["git", "-C", self.tree, "checkout", "-q", "--detach", self.base], check=True)
        with open(os.path.join(self.tree, path), "a") as handle:
            handle.write("\n")  # new bytes every file here still parses with: Go, a module file, JSON, a patch
        return {"base": self.base, "sha": self.commit("edit " + path), "tools": self.tree}

    def test_the_set_and_the_listing_agree(self):
        unread = run.phaseUnread(self.tree, self.tree)
        for shaping in ("go.mod", "a/a.go", "notes/probe.c", "a/doc.md", "cloud/fast-gate/executors.txt", "notes/.gitattributes",
                        "notes/gen.a", "notes/target.txt", "verify/catalog/01-probe.patch", "cohere", "notes/tool.log"):
            self.assertNotIn(shaping, unread)
        self.assertTrue({"README.md", "docs/guide.md", "notes/counts.log", "internal/oracle/testdata/note.txt", "oracle/note.txt"} <= set(unread), unread)
        tracked = [entry.split("\t", 1)[1] for entry in run.git(self.tree, "ls-files", "-s").splitlines() if entry.split(" ", 1)[0] in ("100644", "100755")]
        overRead = []
        for path in tracked:
            with self.subTest(path):
                (tipRows, tipRead), (baseRows, baseRead) = self.observed(self.edited(path))
                if path in unread:
                    self.assertEqual([row["unit"] for row in tipRows], [phase for phase in run.fastBasePhases if phase != "census"], "%s: a false green, it selected a phase" % path)
                    self.assertEqual(tipRows, baseRows, "%s: a false green, a phase's inputs moved" % path)
                    for tip, base in zip(tipRead, baseRead):
                        self.assertEqual(tip, base, "%s: a false green, phase %s reads it" % (path, tip[0]))
                elif (tipRows, tipRead) == (baseRows, baseRead):
                    overRead.append(path)
        # An over-read sends a change to the compare for nothing, a should, never a false green (Node's ruling): here only
        # the inert executable, which the set leaves out by mode alone, as Node ruled plain files alone.
        self.assertEqual(overRead, ["notes/tool.log"])


class GoListFieldsHoldEveryRead(unittest.TestCase):
    """go list writes only run.goListFields, so a field a reader reads that the list leaves out would read as missing, a
    graph with edges or embeds gone, and fail open. Over a real go list of a tree with every shape the readers meet (a
    package embedding a file, a test and an external test each embedding one and importing, a main package and its
    test binary), every field goPackages and goCommandInputs read is in the list, and each reads the same from the list's
    answer as from go list's whole one. Mutants, red here: a field the readers read left out of the list (Dir, Imports,
    TestEmbedFiles); a reader reading a field the list doesn't name."""

    def setUp(self):
        scratch = tempfile.TemporaryDirectory()
        self.addCleanup(scratch.cleanup)
        self.tree = scratch.name
        sources = {
            "go.mod": "module " + run.module + "\ngo 1.24\n",
            "a/a.go": 'package a\n\nimport _ "embed"\n\n//go:embed doc.md\nvar Doc string\n',
            "a/doc.md": "# embedded\n",
            "a/a_test.go": 'package a\n\nimport (\n\t_ "embed"\n\t"testing"\n\n\t"' + run.module + '/c"\n)\n\n//go:embed test.txt\nvar T string\n\nfunc TestA(t *testing.T) { _ = c.C }\n',
            "a/test.txt": "t\n",
            "a/x_test.go": 'package a_test\n\nimport (\n\t_ "embed"\n\t"strings"\n\t"testing"\n)\n\n//go:embed x.txt\nvar X string\n\nfunc TestX(t *testing.T) { _ = strings.ToUpper(X) }\n',
            "a/x.txt": "x\n",
            "b/main.go": 'package main\n\nimport "' + run.module + '/a"\n\nfunc main() { _ = a.Doc }\n',
            "b/main_test.go": 'package main\n\nimport "testing"\n\nfunc TestMain_(t *testing.T) {}\n',
            "c/c.go": "package c\n\nconst C = 1\n",
        }
        for name, source in sources.items():
            os.makedirs(os.path.dirname(os.path.join(self.tree, name)), exist_ok=True)
            with open(os.path.join(self.tree, name), "w") as handle:
                handle.write(source)

    def read(self, graph, packages):
        """Every field the readers read of graph (go list -deps -test) and packages (go list of ./...), and what they read."""
        read, records = set(), run.goListRecords

        class Recording(dict):
            def get(self, key, *default):
                read.add(key)
                return dict.get(self, key, *default)

            def __getitem__(self, key):
                read.add(key)
                return dict.__getitem__(self, key)

        answer = subprocess.CompletedProcess([], 0, packages, "")
        with mock.patch.object(run, "goListRecords", side_effect=lambda text: (Recording(record) for record in records(text))), \
                mock.patch.object(run, "listedGoList", return_value=answer):
            out = (run.goPackages(self.tree, graph), run.goCommandInputs(self.tree, ["go", "vet", "./..."]),
                   run.goCommandInputs(self.tree, ["go", "build", "./..."]))
        return read, out

    def test_the_list_holds_every_field_read(self):
        listing = lambda *arguments: realRun(["go", "list"] + list(arguments), cwd=self.tree, capture_output=True, text=True, check=True).stdout
        whole = listing("-deps", "-test", "-json", "./...")
        read, wholeOut = self.read(whole, listing("-json", "./..."))
        self.assertTrue({"Dir", "Imports", "TestEmbedFiles", "XTestEmbedFiles", "EmbedFiles", "TestImports", "XTestImports"} <= read, read)
        self.assertLessEqual(read, set(run.goListFields))
        _, listedOut = self.read(realRun(run.packageGraphCommand(), cwd=self.tree, capture_output=True, text=True, check=True).stdout,
                                 realRun(run.goPackageListCommand(["./..."]), cwd=self.tree, capture_output=True, text=True, check=True).stdout)
        self.assertEqual(listedOut, wholeOut)
        self.assertTrue(wholeOut[0][1] and wholeOut[0][5], wholeOut[0])  # embeds and test edges were there to lose


class GoListRecordsReadInPlace(unittest.TestCase):
    """run.goListRecords reads a go list -json answer as the records it holds, in order, whatever whitespace lies before,
    between and after them, as the loop that sliced off each record did. Mutants, red here: the whitespace after a record
    left unskipped; the first record's leading whitespace left unskipped; the last record dropped."""

    records = [{"ImportPath": "example.com/a", "Imports": ["fmt"], "Dir": "/t/a"},
               {"ImportPath": "example.com/b [example.com/b.test]", "ForTest": "example.com/b", "Embed": "x\ty"},
               {"ImportPath": "example.com/c", "Note": "a } inside a string, then { again"}]

    def sliced(self, text):
        decoder, remaining, out = json.JSONDecoder(), text.strip(), []
        while remaining:
            record, end = decoder.raw_decode(remaining)
            remaining = remaining[end:].lstrip()
            out.append(record)
        return out

    def test_records_read_in_place_are_the_sliced_ones(self):
        for name, text in {"go list's own shape": "".join(json.dumps(record, indent="\t") + "\n" for record in self.records),
                           "whitespace around and between": "\n \t" + "\r\n\n".join(json.dumps(record) for record in self.records) + " \n\n",
                           "no whitespace between": "".join(json.dumps(record) for record in self.records),
                           "empty": "", "whitespace alone": " \n\t "}.items():
            with self.subTest(name):
                self.assertEqual(list(run.goListRecords(text)), self.sliced(text))
                self.assertEqual(list(run.goListRecords(text)), self.records if text.strip() else [])


class ListingRunsGoListOnce(unittest.TestCase):
    """One --list-units --with-inputs pass runs each go list once a tree (run.listedGoList, Hotty's measures on #q11k6k6
    and #cb8bpe7): packageGraphCommand, which touched() and coverage's provider share, and go list -json ./...,
    which build's and vet's providers share. It lists what it did without the share. Every process the pass starts is
    counted at Popen, so touched()'s, started through the gate's own spawn, is too. Mutants, red here: the pass's share
    dropped (each asker runs its own); touched() running its own in a listing; build's and vet's each running theirs."""

    setUp, write, commit, change = (CoverageBounded.setUp, CoverageBounded.write, CoverageBounded.commit,
                                    CoverageBounded.change)
    listings = {"graph": run.packageGraphCommand(), "packages": run.goPackageListCommand(["./..."])}

    def listed(self, change, shared):
        realRun(["git", "-C", self.tree, "checkout", "-q", "--detach", change["sha"]], check=True)
        calls = {name: 0 for name in self.listings}
        realPopen = run.subprocess.Popen

        def counted(command, *arguments, **options):
            for name, listing in self.listings.items():
                if list(command) == listing:
                    calls[name] += 1
            return realPopen(command, *arguments, **options)

        def unshared(tree, command, timeout, environment=None):
            return run.subprocess.run(command, cwd=tree, capture_output=True, text=True, timeout=timeout, env=environment)

        argv = ["run.py", "--list-units", "--with-inputs", "--tree", self.tree, "--base", change["base"], "--sha", change["sha"], "--tools", self.tree]
        output = io.StringIO()
        try:
            with mock.patch.object(sys, "argv", argv), mock.patch.object(run.subprocess, "Popen", side_effect=counted), contextlib.redirect_stdout(output):
                if shared:
                    run.listUnits()
                else:
                    with mock.patch.object(run, "listedGoList", side_effect=unshared):
                        run.listUnits()
        finally:
            run.listedGoLists = None
        return calls, output.getvalue()

    def test_each_go_list_runs_once_a_listing(self):
        # Without the share: touched()'s and coverage's names (coverageReduced); build's and vet's.
        for name, edit, graphs in (("a markdown change", lambda: self.write("README.md", "# probe, edited\n"), 2),
                                   ("a go change", lambda: self.write("sub/s.go", "package sub\n\nvar S = 2\n"), 2)):
            with self.subTest(name):
                change = self.change(edit)
                counts, rows = self.listed(change, True)
                self.assertEqual(counts, {"graph": 1, "packages": 1})
                alone, unshared = self.listed(change, False)
                self.assertEqual(alone, {"graph": graphs, "packages": 2})
                self.assertEqual(rows, unshared)


class ListUnitsRefusal(unittest.TestCase):
    """A change the census refuses is refused, not crashed on: --list-units exits RefusalExit with one line naming the
    census and how to declare it, which the planner sends its owner as a red (Queue's unplannable). Any other error is
    still a crash, never the change's red. Mutants: the refusal crashing (a traceback, exit 1); a plain ValueError read
    as a refusal; the hint's wording lost."""

    def listed(self, error):
        stderr = io.StringIO()
        with mock.patch.object(run, "listUnits", side_effect=error), mock.patch.object(sys, "argv", ["run.py", "--list-units", "--tree", "."]), \
                mock.patch.object(sys, "stderr", stderr):
            try:
                run.main()
            except SystemExit as exit:
                return exit.code, stderr.getvalue()
        return None, stderr.getvalue()

    def test_a_census_refusal_exits_with_its_line(self):
        code, said = self.listed(run.ChangeRefusal("compiler dependency census: undeclared compiler consumers: a/b_test.go"))
        self.assertEqual(code, run.RefusalExit)
        self.assertEqual(said.count("\n"), 1)
        self.assertTrue(said.startswith("refused: compiler dependency census: undeclared compiler consumers: a/b_test.go. Declare each in "
                                        "cloud/fast-gate/compiler-dependencies.json"), said)

    def test_any_other_error_is_still_a_crash(self):
        with self.assertRaises(ValueError):
            self.listed(ValueError("a gate-tool bug"))


class CensusLaunchers(unittest.TestCase):
    """A launcher declaration (#9qqvs57, system_adamic's ruling of Oct 14) lets a package that launches processes, and
    never the compiler, through the census honestly: it names what it runs and why. It never covers a file that
    references the compiler, one launching by a literal name it doesn't list, or one launching anything it computes
    (system_adamic's review: the compiler-reference rule is per file, so a computed path could be the compiler), and a
    declaration naming adamic, or giving no why, is a bad map. Mutants: a launcher covering a file that references the
    compiler (the declared launcher that runs adamic passes); the literal names unread; the declaration unchecked; a
    computed executable read as no launch (exec.Command(compilerPath()) passes)."""

    launcher = 'package launch\n\nimport "os/exec"\n\nfunc run() error {\n\treturn %s.Run()\n}\n'
    # The same launcher launching through internal/toolchain alone, as the house's code does: no os/exec import, so a
    # file the census read only when marked was never read (Keys' finding on #ry3pfc6).
    toolchainLauncher = 'package launch\n\nimport "github.com/system-inc/adamic/internal/toolchain"\n\nfunc run() error {\n\treturn %s.Run()\n}\n'
    # Every other road a launch takes, its packages imported as a launcher would.
    roadsLauncher = 'package launch\n\nimport (\n\t"os"\n\t"os/exec"\n\t"syscall"\n\n\t"golang.org/x/sys/unix"\n)\n\nfunc run() error {\n\t%s\n\treturn nil\n}\n'

    def census(self, source, launchers, files=None, executors="", changed=None, packages=None):
        """The census over a tree holding source as tool/launch/launch_test.go and files, {path: source}, with the gate
        tools' map declaring launchers and executors.txt holding executors, changed the change's paths."""
        with tempfile.TemporaryDirectory() as root:
            tree, tools = os.path.join(root, "tree"), os.path.join(root, "tools")
            os.makedirs(os.path.join(tree, "tool/launch"))
            os.makedirs(os.path.join(tools, "cloud/fast-gate"))
            with open(os.path.join(tree, "tool/launch/launch_test.go"), "w") as handle:
                handle.write(source)
            for path, text in (files or {}).items():
                os.makedirs(os.path.dirname(os.path.join(tree, path)), exist_ok=True)
                with open(os.path.join(tree, path), "w") as handle:
                    handle.write(text)
            with open(os.path.join(tools, "cloud/fast-gate/compiler-dependencies.json"), "w") as handle:
                json.dump({"version": 1, "packages": packages or {}, "drivers": {}, "launchers": launchers}, handle)
            with open(os.path.join(tools, "cloud/fast-gate/executors.txt"), "w") as handle:
                handle.write(executors)
            realRun(["git", "init", "-q", tree], check=True)
            realRun(["git", "-C", tree, "add", "."], check=True)
            gate = types.SimpleNamespace(arguments=types.SimpleNamespace(tree=tree, tools=tools), result={},
                                         git=lambda directory, *arguments: realRun(["git", "-C", directory] + list(arguments), capture_output=True, text=True, check=True).stdout.strip())
            directories = {"tool/launch", "internal/load"} | {os.path.dirname(path) for path in (files or {})}
            run.Gate.compilerDependencies(gate, directories, ["tool/launch/launch_test.go"] if changed is None else changed)
            return gate.result

    declared = {"tool/launch": {"runs": ["go", "node"], "why": "launches go test -c and Node, never adamic"}}

    def test_a_declared_launcher_passes(self):
        for call in ('exec.Command("node", "x.js")', 'exec.CommandContext(context.Background(), "go", "test", "-c")',
                     'exec.CommandContext(ctx, `node`)'):
            self.census(self.launcher % call, self.declared)
        for call in ('toolchain.CommandIn(directory, "go", arguments...)', 'toolchain.Command("node", "x.js")',
                     'toolchain.CommandContextIn(ctx, directory, "node", arguments...)'):
            self.census(self.toolchainLauncher % call, self.declared)

    def test_every_road_internal_toolchain_launches_by_is_read(self):
        # The hole: toolchain.Command and CommandContextIn were never read, and a file importing only toolchain wasn't
        # read at all. Mutants: toolchain's launchers dropped from launchRoads (each computed one passes); the marked gate
        # put back for a declared launcher (the unmarked file passes).
        for call in ("toolchain.CommandContextIn(ctx, root, program, arguments...)", "toolchain.Command(name)",
                     "toolchain.CommandIn(directory, tool)"):
            self.refused(self.toolchainLauncher % call, "launches a computed executable, " + call.split("(")[0])
        self.refused(self.toolchainLauncher % 'toolchain.CommandContextIn(ctx, root, "python3", "x.py")',
                     "launches python3, which its declaration's runs don't name")
        # git's two launchers run git whatever their arguments.
        self.refused(self.toolchainLauncher % 'toolchain.GitCommandIn(directory, "status")', "launches git, which its declaration's runs don't name")
        self.census(self.toolchainLauncher % 'toolchain.GitCommand(arguments...)', {"tool/launch": {"runs": ["git"], "why": "x"}})
        # Read by the name its file imports the package as, never its text.
        aliased = self.toolchainLauncher.replace("import ", "import chain ")
        self.refused(aliased % "chain.Command(name)", "launches a computed executable, chain.Command(name)")

    def test_every_other_road_needs_a_literal_executable(self):
        # Adamic Loom's rule on #ry3pfc6: os.StartProcess, syscall's and x/sys/unix's exec calls each name a literal
        # executable, or refuse. Mutant: os's StartProcess dropped from launchRoads (the computed one passes).
        for call in ("os.StartProcess(path, nil, nil)", "syscall.Exec(path, nil, nil)", "syscall.ForkExec(path, nil, nil)",
                     "syscall.StartProcess(path, nil, nil)", "unix.Exec(path, nil, nil)"):
            self.refused(self.roadsLauncher % call, "launches a computed executable, " + call.split("(")[0] + "(path)")
        for call in ('os.StartProcess("node", nil, nil)', 'syscall.Exec("go", nil, nil)'):
            self.census(self.roadsLauncher % call, self.declared)

    def test_what_cant_be_read_refuses(self):
        # Unknown refuses: a launcher named as a value, a call whose executable argument isn't there (half-parsed), a
        # command's Path set by hand, a raw system call, cgo, and a file whose imports can't be read. Mutants: the value
        # check (a func value of exec.Command passes), the argument count (exec.CommandContext("node") read as naming
        # node, half-parsed), the Path, the raw call and cgo each unread.
        for call, reason in (("run := exec.Command", "names exec.Command without calling it"),
                             ("exec.CommandContext(ctx)", "launches by exec.CommandContext with arguments it can't read"),
                             ('exec.CommandContext("node")', "launches by exec.CommandContext with arguments it can't read"),
                             ('command := exec.Command("node"); command.Path = compilerPath()', "writes a Path field by hand"),
                             ("syscall.RawSyscall(59, 0, 0, 0)", "makes a raw system call, syscall.RawSyscall"),
                             ("unix.Syscall(59, 0, 0, 0)", "makes a raw system call, unix.Syscall")):
            self.refused(self.roadsLauncher % call, reason)
        self.refused(self.launcher.replace('import "os/exec"', 'import (\n\t"C"\n\t"os/exec"\n)') % 'exec.Command("node")', "uses cgo")
        # Two specs on one line of a group are Go, and read; a raw string's path isn't read, nor is anything between the
        # package clause and the first declaration that isn't an import.
        self.census(self.launcher.replace('import "os/exec"', 'import (\n\t"os/exec"; "os"\n)') % 'exec.Command("node")', self.declared, self.helper)
        self.refused(self.launcher.replace('import "os/exec"', 'import exec `os/exec`') % 'exec.Command(compilerPath())', "imports that can't be read")
        self.refused(self.launcher.replace('import "os/exec"', 'import "os/exec"\n\nnonsense\n') % 'exec.Command(compilerPath())', "imports that can't be read")
        # A blank import of a package that launches runs its init unread.
        self.refused(self.launcher.replace('import "os/exec"', 'import (\n\t"os/exec"\n\t_ "syscall"\n)') % 'exec.Command("node")', "imports syscall as _")

    # A house package that launches (an os/exec import), one that only imports it, and one that launches nothing.
    helper = {"internal/helper/helper.go": 'package helper\n\nimport "os/exec"\n\ntype Plain struct{ Name string }\n\ntype Runner struct{}\n\nfunc (r *Runner) Run() {}\n\nconst Limit = 1\n\nvar Hook = Build\n\nfunc Build() error {\n\treturn exec.Command("go", "build").Run()\n}\n\nfunc Exports() {}\n',
              "internal/helper/helper_test.go": 'package helper\n',
              "internal/wrapper/wrapper.go": 'package wrapper\n\nimport "github.com/system-inc/adamic/internal/helper"\n\nfunc Wrap() error {\n\treturn helper.Build()\n}\n',
              "internal/quiet/quiet.go": 'package quiet\n\nimport "os"\n\nfunc Read() {\n\tos.ReadFile("x")\n}\n'}
    houseLauncher = 'package launch\n\nimport (\n\t"os/exec"\n\n\t%s"github.com/system-inc/adamic/internal/%s"\n)\n\nfunc run() error {\n\t%s\n\treturn exec.Command("node").Run()\n}\n'

    def allowLine(self, function, launcher="tool/launch", package="internal/helper", why="its runs traced on J-abc1234"):
        return "# allow-launch: %s %s.%s: %s\n" % (launcher, package, function, why)

    def test_a_house_package_that_launches_is_refused_without_its_allow(self):
        # Mutants: the house import read as no launch (the helper passes), and the launch read only directly (the wrapper,
        # which launches through helper, passes).
        self.refused(self.houseLauncher % ("", "helper", "helper.Build()"), "imports internal/helper, which launches (internal/helper/helper.go imports os/exec)")
        self.refused(self.houseLauncher % ("", "wrapper", "wrapper.Wrap()"), "imports internal/wrapper, which launches (internal/wrapper/wrapper.go imports internal/helper, where internal/helper/helper.go imports os/exec)")
        # One that launches nothing is no launch, and neither are its test files.
        self.census(self.houseLauncher % ("", "quiet", "quiet.Read()"), self.declared, self.helper)
        self.refused(self.houseLauncher % ("", "absent", "absent.Read()"), "imports internal/absent, which launches (internal/absent isn't in this tree)")

    def refusedWith(self, source, reason, executors, files=None, changed=None):
        with self.assertRaises(run.ChangeRefusal, msg=source) as raised:
            self.census(source, self.declared, self.helper if files is None else files, executors, changed)
        self.assertIn(reason, str(raised.exception), source)

    def refused(self, source, reason):
        with self.assertRaises(run.ChangeRefusal, msg=source) as raised:
            self.census(source, self.declared, self.helper)
        self.assertIn("tool/launch is a declared launcher, refused at line ", str(raised.exception), source)
        self.assertIn(reason, str(raised.exception), source)
        return str(raised.exception)

    def test_an_allowed_house_package_is_read_as_its_allowed_functions(self):
        # Adamic Loom's narrowing on #ry3pfc6: the allow names each function, so another one, a variable or a type with
        # methods refuses, read by the name the package is imported as. Mutants: the allowed set unread (Exports passes,
        # adaptedtree calling buildcache.GoExports), and a type's methods unread (Runner passes).
        allows = self.allowLine("Build")
        self.census(self.houseLauncher % ("", "helper", "_ = helper.Plain{Name: \"x\"}; _ = helper.Limit; helper.Build()"), self.declared, self.helper, allows)
        self.census(self.houseLauncher % ("built ", "helper", "built.Build()"), self.declared, self.helper, allows)
        for call, named in (("helper.Exports()", "internal/helper.Exports"), ("_ = helper.Hook", "internal/helper.Hook"),
                            ("_ = &helper.Runner{}", "internal/helper.Runner"), ("other := helper.Exports; other()", "internal/helper.Exports")):
            self.refusedWith(self.houseLauncher % ("", "helper", call), "names %s, and no allow-launch line names it for tool/launch" % named, allows)
        self.refusedWith(self.houseLauncher % ("bc ", "helper", "bc.Exports()"), "names internal/helper.Exports", allows)
        self.refusedWith(self.houseLauncher % ("", "helper", "helper.Missing()"), "names helper.Missing, which internal/helper doesn't declare", allows)
        self.refusedWith(self.houseLauncher % (". ", "helper", "Build()"), "imports github.com/system-inc/adamic/internal/helper as .", allows)

    def test_an_allow_is_one_function_with_its_trace_or_the_tools_are_bad(self):
        # A glob, a package alone, or a why naming no trace job raises, never read loosely. Mutant: the job check dropped.
        for line in (self.allowLine("*"), "# allow-launch: tool/launch internal/*.Build: J-abc1234\n",
                     "# allow-launch: tool/launch internal/helper: J-abc1234\n", self.allowLine("Build", why="traced"),
                     self.allowLine("Build") + self.allowLine("Build")):
            with self.assertRaises(ValueError, msg=line) as raised:
                self.census(self.houseLauncher % ("", "helper", "helper.Build()"), self.declared, self.helper, line)
            self.assertNotIsInstance(raised.exception, run.ChangeRefusal)
            self.assertIn("allow-launch line", str(raised.exception), line)

    def test_a_packages_entry_naming_no_compiler_leaves_the_launches_read(self):
        # Lane keeps a packages entry naming only the adapted tree, whose product it holds, so that entry must never
        # answer the census for what lane launches. Mutant: any packages entry exempting its package (the computed launch
        # passes, and so does the undeclared one).
        computed = self.launcher % "exec.Command(compilerPath())"
        self.census(computed, self.declared, self.helper, packages={"tool/launch": ["internal/load"]})
        with self.assertRaises(run.ChangeRefusal) as raised:
            self.census(computed, self.declared, self.helper, packages={"tool/launch": ["internal/quiet"]})
        self.assertIn("refused at line 6: launches a computed executable, exec.Command(compilerPath())", str(raised.exception))
        with self.assertRaises(run.ChangeRefusal) as raised:
            self.census(self.launcher % 'exec.Command("node")', {}, self.helper, packages={"tool/launch": ["internal/quiet"]})
        self.assertIn("undeclared compiler consumers: tool/launch/launch_test.go", str(raised.exception))

    def test_a_launchers_tree_is_exact_places_each_citing_its_trace(self):
        # Adamic Loom's ruling on #ry3pfc6: tree lists what the process tree may exec beyond runs, each an exact place
        # citing its trace, never a bare name, a glob or a compiler. Mutants: the place unchecked (a bare name passes),
        # the citation unchecked, and the compiler names unread (go passes).
        source = self.launcher % 'exec.Command("node")'
        good = [{"place": "/usr/lib/git-core/git-upload-pack", "trace": "J-5bejz3b"},
                {"place": "stage3/typescript/node_modules/@esbuild/linux-x64/bin/esbuild", "trace": "J-crks9vh"}]
        self.census(source, {"tool/launch": dict(self.declared["tool/launch"], tree=good)}, self.helper)
        for tree, reason in ((["sh"], "is {\"place\", \"trace\"}"), ([{"place": "sh", "trace": "J-5bejz3b"}], "is no exact place"),
                             ([{"place": "/usr/bin/*", "trace": "J-5bejz3b"}], "is no exact place"),
                             ([{"place": "keyed:stage3/x/bin/esbuild", "trace": "J-crks9vh"}], "is no exact place"),
                             ([{"place": "../stage3/x/bin/esbuild", "trace": "J-crks9vh"}], "is no exact place"),
                             ([{"place": "stage3/./x/esbuild", "trace": "J-crks9vh"}], "is no exact place"),
                             ([{"place": "/usr/bin/sh", "trace": "traced"}], "cites no trace job"),
                             ([{"place": "/usr/local/go/bin/go", "trace": "J-5bejz3b"}], "is a compiler"),
                             ([{"place": "/usr/bin/clang", "trace": "J-5bejz3b"}], "is a compiler"),
                             ([{"place": "/usr/local/go/pkg/tool/linux_amd64/compile", "trace": "J-5bejz3b"}], "is a compiler"),
                             ([{"place": "/usr/local/go/pkg/tool/linux_amd64/buildid", "trace": "J-5bejz3b"}], "is a compiler"),
                             ([{"place": "/opt/tools/bin/vet", "trace": "J-5bejz3b"}], "is a compiler"),
                             ([{"place": "/usr/bin/sh", "trace": "J-5bejz3b"}] * 2, "is named twice"),
                             ({"place": "/usr/bin/sh"}, "isn't a list")):
            with self.assertRaises(ValueError, msg=tree) as raised:
                self.census(source, {"tool/launch": dict(self.declared["tool/launch"], tree=tree)}, self.helper)
            self.assertNotIsInstance(raised.exception, run.ChangeRefusal)
            self.assertIn(reason, str(raised.exception), tree)

    def test_the_gate_tools_own_allows_read(self):
        # The tools' own lines parse as one launcher, package and function each, so a bad one never reaches a gate.
        tools = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(run.__file__))))
        allows = run.launchAllows(tools)
        self.assertEqual(sorted(allows.get(("stage3/lane/adaptedtree", "internal/buildcache"), {})), ["LeanProductOf", "PinnedGit", "Tool"])

    def test_a_dead_allow_is_the_changes_when_it_touches_the_launcher(self):
        # An allow naming a package the launcher doesn't import, a function it doesn't call, or a launcher the map doesn't
        # declare is dead: the change's red when it touched the launcher's own files, else named on the base, a map edit
        # included. Mutants: references never gathered (every allow reads dead), and every dead line read the change's.
        allows = self.allowLine("Build") + self.allowLine("Exports") + self.allowLine("Build", package="internal/quiet") + self.allowLine("Build", launcher="tool/other")
        source = self.houseLauncher % ("", "helper", "helper.Build()")
        self.refusedWith(source, "dead allow-launch lines: executors.txt:2, allow-launch tool/launch internal/helper.Exports: no file of tool/launch names it", allows)
        result = self.census(source, self.declared, self.helper, allows, changed=["internal/quiet/quiet.go", "cloud/fast-gate/compiler-dependencies.json"])
        self.assertEqual(result.get("launch_allows_dead_on_base"), [
            "executors.txt:2, allow-launch tool/launch internal/helper.Exports: no file of tool/launch names it",
            "executors.txt:3, allow-launch tool/launch internal/quiet.Build: no file of tool/launch names it",
            "executors.txt:4, allow-launch tool/other internal/helper.Build: tool/other isn't a declared launcher"])

    def test_a_comment_or_a_definition_is_no_launch(self):
        source = self.launcher % 'exec.Command("node")' + '\n// exec.Command(name) is what it would be\n/* exec.Command(other) */\nfunc CommandIn(directory, name string) {}\nfunc command() *exec.Cmd {\n\treturn nil\n}\nvar text = "exec.Command(name)"\n'
        self.census(source, self.declared)

    def test_cohere_harness_commands_are_no_reference_to_the_compiler(self):
        # Cohere's finding on #1byp10f: cmd/adamic matched cohere's cmd/adamic_<name> commands by prefix, so a launcher
        # running one was refused as running the compiler. The bound keeps every real reference. Mutant: the bare prefix
        # (the harness command refused).
        self.census(self.launcher % 'exec.Command("go", "run", "./cmd/adamic_path", "x")', self.declared)
        self.census(self.launcher % 'exec.Command("go", "build", "github.com/system-inc/cohere/cmd/adamic_helper_oracle")', self.declared)
        for reference in ('"./cmd/adamic"', '"github.com/system-inc/adamic/cmd/adamic/internal"', '"go run ./cmd/adamic c x.ts"',
                          'path := "cmd/adamic"', 'x := `cmd/adamic`', "// cmd/adamic\n"):
            self.assertIsNotNone(run.compilerMarker.search(reference), reference)
        self.refused(self.launcher % 'exec.Command("go", "run", "./cmd/adamic", "c", "x.ts")', "references the compiler (cmd/adamic)")

    def test_another_command_or_a_possessive_is_no_reference_to_the_compiler(self):
        # Seats' tables on #8v9f77x and #mgvwehq: parallel.go:5 names cmd/adamic-gate/parallelcensus and repowalk.go:6
        # says cmd/adamic's, and the bound at a word character took both for the compiler. Mutant: the old bound
        # (?![A-Za-z0-9_]) back, both refused.
        for text in ("// cmd/adamic-gate/parallelcensus reads this\n", "// cmd/adamic's walk\n", '"cmd/adamic-gate"'):
            self.assertIsNone(run.compilerMarker.search(text), text)
        # Every launch spelling still names it (Adamic Loom's must on #r2a9vd4): a backslash in a Go string holding
        # generated source, a shell's ' and ;, a newline, as the old bound read them. Mutant: the lookahead
        # (?=["`/\s]|$), the backslash, the quote and the semicolon missed.
        for reference in ("// go run ./cmd/adamic\n", "cmd/adamic", '"./cmd/adamic"', "x := `./cmd/adamic`",
                          '"go run ./cmd/adamic\\nx"', '"./cmd/adamic\\"', "sh -c 'go run ./cmd/adamic'", "go run ./cmd/adamic; echo"):
            self.assertIsNotNone(run.compilerMarker.search(reference), reference)

    def test_a_launcher_that_runs_the_compiler_is_refused(self):
        self.refused(self.launcher % 'exec.Command("adamic", "c", "x.a")', 'references the compiler ("adamic")')
        self.refused(self.launcher % 'exec.Command(os.Getenv("ADAMIC_BIN"))', "references the compiler (ADAMIC_BIN)")

    def test_a_computed_executable_is_refused_naming_its_line(self):
        # The fourth mutant: nothing in the file references the compiler, and the path it runs could be it.
        self.assertIn("refused at line 6: launches a computed executable, exec.Command(compilerPath())",
                      self.refused(self.launcher % "exec.Command(compilerPath())", "computed executable"))
        for call in ("exec.CommandContext(ctx, name, arguments...)", 'exec.Command(name, "x")', "exec.Command(names[0])",
                     'exec.Command("no" + "de")', "exec.CommandContext(context.WithTimeout(ctx, limit), path)"):
            self.refused(self.launcher % call, "computed executable")

    def test_a_hand_built_command_is_refused(self):
        self.refused(self.launcher % "(&exec.Cmd{Path: path})", "names exec.Cmd other than as a pointer another call made")

    def test_the_reviews_three_cheap_fixes_and_the_exact_name(self):
        # Adamic Loom's review on #ry3pfc6. Mutants: the package line's imports skipped (the old line anchor), a Path
        # target unread, a Cmd composite read only as exec.Cmd{, and the runs matched by basename; each passes its case.
        # (1) A Path written in any form, in a file importing neither os/exec nor toolchain (a sibling's write reaches a
        # Cmd another file made); a Path read is no write.
        plain = 'package launch\n\nfunc set(c *thing, row record) {\n\t%s\n}\n'
        for statement in ('c.Path, c.Dir = compilerPath(), ""', "c.Path += suffix", "p := &c.Path", "c.Path++"):
            self.refused(plain % statement, "writes a Path field by hand")
        for statement in ('if row.Path[0] != "x" || len(row.Path) == 2 {', 'name := strings.Join(row.Path, ".")', "for _, part := range row.Path {"):
            self.census(plain % statement, self.declared, self.helper)
        # (2) A Cmd value in any spelling; a pointer another call made, as a parameter, result or variable, is no build.
        for statement in ("commands := []*exec.Cmd{{Path: path}}", "type command = exec.Cmd", "c := new(exec.Cmd)", "c := &exec.Cmd {Path: path}",
                          "var commands [2]exec.Cmd"):
            self.refused(self.roadsLauncher % statement, "names exec.Cmd other than as a pointer another call made")
        self.census(self.roadsLauncher % "var c *exec.Cmd" + "\nfunc wrap(c *exec.Cmd) *exec.Cmd {\n\treturn c\n}\n", self.declared, self.helper)
        # (3) Imports read as Go places them, the package line and two on one line included.
        self.refused('package launch; import "os/exec"\n\nfunc run() error {\n\treturn exec.Command(compilerPath()).Run()\n}\n', "launches a computed executable")
        self.refused('package launch\n\nimport "fmt"; import "os/exec"\n\nfunc run() error {\n\treturn exec.Command(compilerPath()).Run()\n}\n', "launches a computed executable")
        # The runs, by exact name: a path ending in a named program is no program the runs name.
        for executable in ("/usr/bin/node", "./node", "/tmp/built/go"):
            self.refused(self.launcher % ('exec.Command("%s")' % executable), "launches %s, which its declaration's runs don't name" % executable)

    def test_an_aliased_os_exec_is_read_and_a_dot_import_refused(self):
        aliased = self.launcher.replace('import "os/exec"', 'import run "os/exec"')
        self.census(aliased % 'run.Command("node")', self.declared)
        self.refused(aliased % "run.Command(compilerPath())", "computed executable")
        self.refused(self.launcher.replace('import "os/exec"', 'import . "os/exec"') % 'Command("node")', "imports os/exec as .")

    def test_a_literal_name_the_declaration_doesnt_list_is_refused(self):
        self.refused(self.launcher % 'exec.Command("python3", "x.py")', "launches python3, which its declaration's runs don't name")

    def test_an_undeclared_launcher_is_refused(self):
        with self.assertRaises(run.ChangeRefusal):
            self.census(self.launcher % 'exec.Command("node", "x.js")', {})

    def test_a_bad_declaration_is_a_bad_map(self):
        for declaration in ({"runs": ["node", "adamic"], "why": "x"}, {"runs": ["node"]}, {"runs": [], "why": "x"}):
            with self.assertRaises(ValueError, msg=declaration) as raised:
                self.census(self.launcher % 'exec.Command("node")', {"tool/launch": declaration})
            self.assertNotIsInstance(raised.exception, run.ChangeRefusal)


class CensusMemoryFloors(unittest.TestCase):
    """A package whose tests size themselves from the unit's memory limit (testguard.MemoryLimit) refuses to start under
    what they need, so the planner's fit (--fit-memory) must never share it under a declared floor (Node's ruling). A file
    the change touches that reads its limit, by testguard, an alias or a dot import, in a package with no memoryMegabytes
    refuses; one on the base is named; testguard itself, comments and strings are no read; a change lowering or removing
    a reader's floor refuses; a table that can't be read is a bad table, never no floors. Mutants: the alias unfollowed;
    comments unstripped; the touched reader named, never refused; the floor-lowering compare skipped; a missing table
    or a floor of 0 read as no floors."""

    reader = ('package json\n\nimport (\n\t"testing"\n\n\t"github.com/system-inc/adamic/internal/testguard"\n)\n\n'
              'func TestBudget(t *testing.T) {\n\tif limit, ok := testguard.MemoryLimit(); ok {\n\t\tt.Log(limit)\n\t}\n}\n')
    needs = "cloud/fast-gate/unit-needs.json"

    def census(self, files, floors=None, changed=None, edits=None):
        """memoryFloors over a tree of files, the gate tools' table giving floors (none: no table), changed defaulting
        to every file. With edits, files are the base's commit and edits the change's over it."""
        with tempfile.TemporaryDirectory() as root:
            tree, tools = os.path.join(root, "tree"), os.path.join(root, "tools")
            os.makedirs(tools)
            if floors is not None:
                writeUnitNeeds(tools, floors)
            realRun(["git", "init", "-q", tree], check=True)
            commits = []
            for layer in [files] + ([edits] if edits is not None else []):
                for path, source in layer.items():
                    os.makedirs(os.path.dirname(os.path.join(tree, path)), exist_ok=True)
                    with open(os.path.join(tree, path), "w") as handle:
                        handle.write(source)
                realRun(["git", "-C", tree, "add", "-A"], check=True)
                realRun(["git", "-C", tree, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "t"], check=True)
                commits.append(run.git(tree, "rev-parse", "HEAD"))
            gate = types.SimpleNamespace(arguments=types.SimpleNamespace(tree=tree, tools=tools, base=commits[0], sha=commits[-1]), result={},
                                         git=lambda directory, *arguments: realRun(["git", "-C", directory] + list(arguments), capture_output=True, text=True, check=True).stdout.strip())
            run.Gate.memoryFloors(gate, list((edits or files).keys()) if changed is None else changed)
            return gate.result

    def refused(self, files, floors=None, **options):
        with self.assertRaises(run.ChangeRefusal, msg=files) as raised:
            self.census(files, {} if floors is None else floors, **options)
        return str(raised.exception)

    def test_a_touched_reader_without_a_floor_refuses_naming_its_file_and_package(self):
        """A file the change adds that reads its memory limit, in a package with no floor, refuses, naming both."""
        with self.assertRaises(run.MemoryFloorRefusal) as raised:
            self.census({"stage1/cohere/fresh/budget_test.go": self.reader}, {"stage1/cohere/json": 8192})
        self.assertEqual(str(raised.exception), "memory floor census: stage1/cohere/fresh/budget_test.go reads its memory limit (testguard.MemoryLimit) "
                                                "and stage1/cohere/fresh declares no memoryMegabytes in cloud/fast-gate/unit-needs.json")
        self.assertIn("cloud/fast-gate/unit-needs.json, its package with the memoryMegabytes", raised.exception.hint())
        self.assertNotIn("compiler-dependencies.json", raised.exception.hint())

    def test_a_reader_with_a_floor_passes(self):
        """The same reader passes once its package has a floor in the gate tools' needs."""
        self.assertEqual(self.census({"stage1/cohere/fresh/budget_test.go": self.reader}, {"stage1/cohere/fresh": 8192}), {})

    def test_an_aliased_import_is_followed(self):
        """A reader that imports testguard under an alias, in a block or on one line, still refuses."""
        block = self.reader.replace('\t"github.com', '\ttg "github.com').replace("testguard.MemoryLimit()", "tg.MemoryLimit()")
        line = block.replace('import (\n\t"testing"\n\n\ttg "github.com/system-inc/adamic/internal/testguard"\n)',
                             'import "testing"\nimport tg "github.com/system-inc/adamic/internal/testguard"')
        self.assertNotEqual(line, block)
        for source in (block, line):
            self.assertIn("stage1/cohere/fresh declares no memoryMegabytes", self.refused({"stage1/cohere/fresh/budget_test.go": source}), source)

    def test_a_dot_import_counts_a_bare_call(self):
        """Under a dot import of testguard, a bare MemoryLimit() call is a read and refuses."""
        source = self.reader.replace('\t"github.com', '\t. "github.com').replace("testguard.MemoryLimit()", "MemoryLimit()")
        self.assertIn("stage1/cohere/fresh declares no memoryMegabytes", self.refused({"stage1/cohere/fresh/budget_test.go": source}))

    def test_a_comment_or_string_is_no_read(self):
        """A file that imports testguard and names testguard.MemoryLimit() only in comments and strings passes."""
        source = self.reader.replace("if limit, ok := testguard.MemoryLimit(); ok {", "if limit, ok := testguard.Share(); ok {") + (
            "\n// The budget follows the unit's limit (testguard.MemoryLimit, 16 GiB on a runner).\n// testguard.MemoryLimit() sizes it.\n"
            "/* testguard.MemoryLimit() */\nvar why = \"testguard.MemoryLimit()\"\nvar raw = `testguard.MemoryLimit()`\n")
        self.assertEqual(self.census({"stage1/cohere/fresh/budget_test.go": source}, {}), {})

    def test_an_undeclared_reader_on_the_base_is_named_not_refused(self):
        """A reader with no floor that the change didn't touch is named in the record and never refuses."""
        result = self.census({"stage1/cohere/fresh/budget_test.go": self.reader, "stage1/other/other.go": "package other\n"}, {},
                             changed=["stage1/other/other.go"])
        self.assertEqual(result, {"memory_readers_undeclared_on_base": ["stage1/cohere/fresh/budget_test.go"]})

    def test_testguard_itself_is_exempt(self):
        """internal/testguard's own files calling MemoryLimit are the reading, never a reader of it."""
        files = {"internal/testguard/memory.go": "package testguard\n\nfunc MemoryLimit() (int64, bool) { return 0, false }\n\nfunc share() { MemoryLimit() }\n",
                 "internal/testguard/memory_test.go": self.reader.replace("package json", "package testguard_test")}
        self.assertEqual(self.census(files, {}), {})

    def test_a_change_lowering_or_removing_a_readers_floor_refuses(self):
        """With the floors in the tree, a change lowering a reader's floor or removing it refuses naming the package;
        raising it, or adding the table where the base had none, passes."""
        def table(floors):
            return json.dumps({"version": 1, "units": [{"package": package, "memoryMegabytes": floor, "cpus": 8} for package, floor in floors.items()]})
        base = {"stage1/cohere/json/memory_budget_test.go": self.reader, self.needs: table({"stage1/cohere/json": 8192})}
        floors = {"stage1/cohere/json": 8192}
        self.assertIn("stage1/cohere/json reads its memory limit (testguard.MemoryLimit) and this change lowers its memoryMegabytes to 4096 in "
                      "cloud/fast-gate/unit-needs.json (8192 on the base)", self.refused(base, floors, edits={self.needs: table({"stage1/cohere/json": 4096})}))
        self.assertIn("stage1/cohere/json reads its memory limit (testguard.MemoryLimit) and this change removes its memoryMegabytes in "
                      "cloud/fast-gate/unit-needs.json (8192 on the base)", self.refused(base, floors, edits={self.needs: table({})}))
        self.assertEqual(self.census(base, floors, edits={self.needs: table({"stage1/cohere/json": 16384})}), {})
        self.assertEqual(self.census({"stage1/cohere/json/memory_budget_test.go": self.reader, "stage1/other/other.go": "package other\n"}, floors,
                                     edits={self.needs: table({"stage1/cohere/json": 4096})}), {})

    def test_a_table_that_cant_be_read_is_a_bad_table_not_a_refusal(self):
        """A missing gate tools table, or one with a floor of 0, raises ValueError, never ChangeRefusal or no floors."""
        for floors in (None, {"stage1/cohere/fresh": 0}):
            with self.assertRaises(ValueError, msg=floors) as raised:
                self.census({"stage1/cohere/fresh/budget_test.go": self.reader}, floors, changed=[])
            self.assertNotIsInstance(raised.exception, run.ChangeRefusal)


class FastCheck(unittest.TestCase):
    """run.py --list-units --fast-check, the census door's submit-time check (#bwnptrb, Adamic Loom's ruling): the units
    as without the flag, then coverage and tools judged on the listing's own Gate and darwin-vet's need from its inputs,
    the verdict file written whatever the exit. A red is exit 3, "refused: " and one line a phase; a check that fails is
    exit 2, "fast check couldn't run: ", the units still printed; the listing's own refusal still wins. These run git and
    go for real, in a module of their own. Mutants, each red here: the darwin leg asked under --fast-check (a leg's
    record can't exist at submit, so every Apple change would be refused); the uncovered check dropped; coverage refusing
    under a map in report mode (M4); the tools check reading the whole tree, so a stacked change answers for its parent's
    literal (M6)."""

    def setUp(self):
        scratch = tempfile.TemporaryDirectory()
        self.addCleanup(scratch.cleanup)
        self.scratch = scratch.name
        self.tree = os.path.join(scratch.name, "tree")
        self.verdictPath = os.path.join(scratch.name, "verdict.json")
        sources = {
            "go.mod": "module " + run.module + "\n\ngo 1.24\n",
            "a/a.go": "package a\n",
            "README.md": "# probe\n",
            "internal/apple/apple.go": "package apple\n",
            # A declared launcher, so the tools literal it gains passes the compiler census and reaches the tools check.
            "probe/probe.go": "package probe\n",
            run.compilerDependenciesPath: json.dumps({"version": 1, "packages": {}, "launchers": {
                "probe": {"runs": ["go", "wasmtime", "wasmer"], "why": "the fast check's tools probe, never adamic"}}}),
            "cloud/fast-gate/unit-needs.json": '{"version":1,"units":[]}',
            "cloud/fast-gate/executors.txt": "mode enforce\ndarwin internal/apple/*\nbuild go.*\ninert *.md\ninert cloud/*\n",
            "cloud/fast-gate/tools.txt": "# declared tools\ngo\tall\tcommand -v go\n",
        }
        for name, source in sources.items():
            self.write(name, source)
        realRun(["git", "-C", self.tree, "init", "-q"], check=True)
        self.base = self.commit("base")

    def write(self, name, source):
        path = os.path.join(self.tree, name)
        os.makedirs(os.path.dirname(path), exist_ok=True)
        with open(path, "w") as handle:
            handle.write(source)

    def commit(self, message):
        realRun(["git", "-C", self.tree, "add", "-A"], check=True)
        realRun(["git", "-C", self.tree, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", message], check=True)
        return run.git(self.tree, "rev-parse", "HEAD")

    def change(self, files, on=None):
        """A commit on the base (or on) writing files, the tree left checked out at it."""
        realRun(["git", "-C", self.tree, "checkout", "-q", "--detach", on or self.base], check=True)
        for name, source in files.items():
            self.write(name, source)
        return self.commit("change")

    def checked(self, sha, base=None, tools=True):
        """main() as the census runs it on base..sha (the base's by default): its exit, its stdout and stderr, and the
        verdict file it wrote. tools False leaves --tools out."""
        if os.path.exists(self.verdictPath):
            os.remove(self.verdictPath)
        argv = ["run.py", "--list-units", "--tree", self.tree, "--base", base or self.base, "--sha", sha] + (["--tools", self.tree] if tools else []) + ["--fast-check", self.verdictPath]
        stdout, stderr, code = io.StringIO(), io.StringIO(), 0
        try:
            with mock.patch.object(sys, "argv", argv), contextlib.redirect_stdout(stdout), contextlib.redirect_stderr(stderr), \
                    mock.patch.object(run.Gate, "npmCli", side_effect=AssertionError("the fast check installs nothing")), \
                    mock.patch.object(run.Gate, "npmPackages", side_effect=AssertionError("the fast check installs nothing")):
                run.main()
        except SystemExit as exit:
            code = exit.code
        finally:
            run.listedGoLists = None
        verdict = None
        if os.path.exists(self.verdictPath):
            with open(self.verdictPath) as handle:
                verdict = json.load(handle)
        # Nothing it ran wrote in the tree.
        self.assertEqual(run.git(self.tree, "status", "--porcelain", "--untracked-files=all"), "")
        return code, stdout.getvalue(), stderr.getvalue(), verdict

    def units(self, *extra):
        return "\n".join(run.fastBasePhases + list(extra)) + "\n"

    def test_an_uncovered_path_is_refused_naming_it_and_the_units_still_list(self):
        strays = {"stray/%02d.txt" % number: "stray\n" for number in range(22)}
        sha = self.change(strays)
        with mock.patch.object(run, "fastCheck", wraps=run.fastCheck) as fastCheck:
            code, stdout, stderr, verdict = self.checked(sha)
        self.assertEqual(code, run.RefusalExit, stderr)
        self.assertEqual(stdout, self.units())
        self.assertEqual(stderr, "refused: coverage: no gate covers 22 changed paths: %s, ...\n" % ", ".join(sorted(strays)[:20]))
        self.assertEqual(verdict["version"], 1)
        self.assertEqual(verdict["darwinVet"], "unneeded")
        self.assertEqual(sorted(verdict["seconds"]), ["coverage", "darwinVetInputs", "tools"])
        self.assertTrue(all(isinstance(value, (int, float)) and value >= 0 for value in verdict["seconds"].values()), verdict)
        # The check ran on the listing's own Gate, whose out directory is gone once it ends.
        gate = fastCheck.call_args[0][0]
        self.assertTrue(gate.arguments.out)
        self.assertFalse(os.path.exists(gate.arguments.out))

    def test_an_undeclared_tool_literal_is_refused_naming_its_line_beside_coverage(self):
        sha = self.change({"probe/probe.go": 'package probe\n\nimport "os/exec"\n\nvar a = exec.Command("go", "version")\nvar b = exec.Command("wasmtime")\n',
                           "stray/unruled.txt": "stray\n"})
        code, stdout, stderr, verdict = self.checked(sha)
        self.assertEqual(code, run.RefusalExit, stderr)
        self.assertEqual(stdout, self.units())
        self.assertEqual(stderr, "refused: coverage: no gate covers 1 changed path: stray/unruled.txt\n"
                                 "tools: probe/probe.go:6 runs wasmtime, which cloud/fast-gate/tools.txt doesn't declare\n")
        self.assertEqual(verdict["darwinVet"], "unneeded")

    def test_an_apple_change_with_no_leg_record_is_never_refused_and_asks_for_the_vet(self):
        # No darwin-legs record exists, and none can at submit: the miss is the line's darwinRecordMissing park (#ycqnr4b).
        sha = self.change({"internal/apple/apple.go": "package apple\n\n// Edited is the change.\nconst Edited = 1\n"})
        with mock.patch.object(run, "darwinLegVerdict", wraps=run.darwinLegVerdict) as verdictRead, \
                mock.patch.object(run, "darwinPathsDigest", wraps=run.darwinPathsDigest) as digest:
            code, stdout, stderr, verdict = self.checked(sha)
        self.assertEqual((code, stderr), (0, ""))
        # The darwin executor is still listed for the line, where its leg is judged.
        self.assertEqual(stdout, self.units("darwin"))
        verdictRead.assert_not_called()
        digest.assert_not_called()
        self.assertEqual(verdict["darwinVet"], "needed")

    def test_a_change_meeting_nothing_passes_and_needs_no_vet(self):
        for name, files in (("a readme", {"README.md": "# probe, edited\n"}),
                            ("a package outside the darwin closure", {"a/a.go": "package a\n\nvar A = 1\n"})):
            with self.subTest(name):
                code, stdout, stderr, verdict = self.checked(self.change(files))
                self.assertEqual((code, stderr), (0, ""))
                self.assertEqual(stdout, self.units())
                self.assertEqual(verdict["darwinVet"], "unneeded")

    def test_a_root_module_file_or_a_listing_that_can_t_be_computed_needs_the_vet(self):
        self.assertTrue(run.darwinVetMeets({"paths": [":(literal)internal/apple"], "names": []}, ["go.sum"]))
        self.assertTrue(run.darwinVetMeets({"paths": ["."], "names": []}, ["README.md"]))
        self.assertTrue(run.darwinVetMeets({"paths": [":(literal)internal/apple"], "names": ["x/y_darwin.go"]}, ["x/y_darwin.go"]))
        self.assertFalse(run.darwinVetMeets({"paths": [":(literal)internal/apple"], "names": []}, ["internal/applesauce/x.go"]))
        code, stdout, stderr, verdict = self.checked(self.change({"go.mod": "module " + run.module + "\n\ngo 1.24\n\n// edited\n"}))
        self.assertEqual((code, stderr), (0, ""))
        self.assertEqual(verdict["darwinVet"], "needed")

    def test_an_exception_in_the_check_can_t_run_and_the_units_still_list(self):
        sha = self.change({"stray/unruled.txt": "stray\n"})
        with mock.patch.object(run, "darwinVetInputs", side_effect=RuntimeError("planted\nsecond line")):
            code, stdout, stderr, verdict = self.checked(sha)
        self.assertEqual(code, run.FastCheckCantRunExit)
        self.assertEqual(stdout, self.units())
        self.assertEqual(stderr, "fast check couldn't run: RuntimeError: planted; second line\n")
        # Written whatever the exit, and a check that never finished asks for the vet.
        self.assertEqual(verdict["darwinVet"], "needed")

    def test_the_listing_s_own_refusal_wins_and_no_check_runs(self):
        # An undeclared compiler consumer the change adds: the compiler census refuses it in the listing, before any check.
        sha = self.change({"a/run_test.go": 'package a\n\nimport (\n\t"os/exec"\n\t"testing"\n)\n\nfunc TestRun(t *testing.T) { _ = exec.Command }\n',
                           "stray/unruled.txt": "stray\n"})
        with mock.patch.object(run, "fastCheck", wraps=run.fastCheck) as fastCheck:
            code, stdout, stderr, verdict = self.checked(sha)
        self.assertEqual(code, run.RefusalExit)
        self.assertTrue(stderr.startswith("refused: compiler dependency census: undeclared compiler consumers: a/run_test.go"), stderr)
        self.assertEqual(stderr.count("\n"), 1)
        self.assertNotIn("coverage:", stderr)
        fastCheck.assert_not_called()
        self.assertEqual(verdict["darwinVet"], "needed")

    def test_an_uncovered_path_under_a_map_in_report_mode_passes(self):
        # M4: report mode records uncovered paths without failing, until the rulings are in (executorMap).
        with open(os.path.join(self.tree, "cloud/fast-gate/executors.txt")) as handle:
            rules = handle.read()
        self.write("cloud/fast-gate/executors.txt", rules.replace("mode enforce\n", "mode report\n"))
        self.base = self.commit("the map reports")
        code, stdout, stderr, verdict = self.checked(self.change({"stray/unruled.txt": "stray\n"}))
        self.assertEqual((code, stderr), (0, ""))
        self.assertEqual(stdout, self.units())
        self.assertEqual(verdict["darwinVet"], "unneeded")

    def test_a_stacked_change_answers_for_its_own_tools_never_its_parent_s(self):
        # M6: the census's base for a stacked change is its unlanded parent, so the parent's literal is in the tree.
        parent = self.change({"probe/probe.go": 'package probe\n\nimport "os/exec"\n\nvar a = exec.Command("wasmtime")\n'})
        quiet = self.change({"README.md": "# probe, edited\n"}, on=parent)
        code, stdout, stderr, verdict = self.checked(quiet, base=parent)
        self.assertEqual((code, stderr), (0, ""))
        own = self.change({"probe/child.go": 'package probe\n\nimport "os/exec"\n\nvar b = exec.Command("wasmer")\n'}, on=parent)
        code, stdout, stderr, verdict = self.checked(own, base=parent)
        self.assertEqual(code, run.RefusalExit, stderr)
        self.assertEqual(stderr, "refused: tools: probe/child.go:5 runs wasmer, which cloud/fast-gate/tools.txt doesn't declare\n")
        # The phase itself still reads the whole tree, as the block's gate does.
        self.assertEqual(len(run.undeclaredTools(self.tree, self.tree, realRun)), 2)

    def test_without_tools_it_can_t_run_and_never_reads_the_tree_s(self):
        code, stdout, stderr, verdict = self.checked(self.change({"stray/unruled.txt": "stray\n"}), tools=False)
        self.assertEqual(code, run.FastCheckCantRunExit)
        self.assertEqual(stdout, self.units())
        self.assertTrue(stderr.startswith("fast check couldn't run: ValueError: --fast-check judges a change by the gate tools"), stderr)
        self.assertEqual(verdict["darwinVet"], "needed")

    def test_the_darwin_listing_runs_offline_and_one_that_can_t_be_computed_needs_the_vet(self):
        listings = []

        def recorded(command, *arguments, **options):
            if command[:3] == ["go", "list", "-deps"] and "{{.Dir}}" in command:  # the darwin listing, never touched()'s graph
                listings.append(options.get("env") or {})
            if "grep" in command and "-l" in command and broken:
                return subprocess.CompletedProcess(command, 128, "", "fatal: planted grep failure\n")
            return realRun(command, *arguments, **options)
        sha = self.change({"README.md": "# probe, edited\n"})
        for broken in (False, True):
            with self.subTest(broken=broken):
                listings.clear()
                with mock.patch.dict(os.environ, {"GOPROXY": "direct"}), mock.patch.object(run.subprocess, "run", side_effect=recorded):
                    code, stdout, stderr, verdict = self.checked(sha)
                self.assertEqual((code, stderr), (0, ""))
                if broken:
                    # A git grep that failed is no listing: the vet is needed, and it says it can't run (never a refusal).
                    self.assertEqual(verdict["darwinVet"], "needed")
                    self.assertEqual(listings, [])
                else:
                    self.assertEqual(verdict["darwinVet"], "unneeded")
                    self.assertEqual([env.get("GOPROXY") for env in listings], ["off"])

    def test_without_the_flag_the_listing_is_as_it_was(self):
        sha = self.change({"stray/unruled.txt": "stray\n"})
        argv = ["run.py", "--list-units", "--tree", self.tree, "--base", self.base, "--sha", sha, "--tools", self.tree]
        stdout = io.StringIO()
        try:
            with mock.patch.object(sys, "argv", argv), contextlib.redirect_stdout(stdout), \
                    mock.patch.object(run, "fastCheck", side_effect=AssertionError("no check without the flag")):
                run.main()
        finally:
            run.listedGoLists = None
        self.assertEqual(stdout.getvalue(), self.units())
        self.assertFalse(os.path.exists(self.verdictPath))


class FastCheckDarwinVet(unittest.TestCase):
    """run.py --fast-check-darwin-vet, the darwin vet the census runs in the background for a change whose fast check
    needs it (#bwnptrb, Adamic Loom's call 2): darwin-vet's own step on the tree, offline and with -json, its seconds one
    JSON line on stdout; exit 0 green, 3 with "refused: darwin-vet: " only for a diagnostic in a package, and 2 with
    "fast check couldn't run: " for anything else, a machine's failure never refusing a change (M5); on SIGTERM it ends
    its go's session before it exits (M2). DarwinVet's module, with go run for real. Mutants, each red here: any nonzero
    vet refusing (M5); SIGTERM left to its default, so go outlives python (M2)."""

    setUp, write, commit = DarwinVet.setUp, DarwinVet.write, DarwinVet.commit

    def vetted(self, sha=None):
        sha = sha or run.git(self.tree, "rev-parse", "HEAD")
        argv = ["run.py", "--fast-check-darwin-vet", "--tree", self.tree, "--sha", sha, "--tools", self.tree]
        stdout, stderr, code = io.StringIO(), io.StringIO(), 0
        with mock.patch.object(sys, "argv", argv), contextlib.redirect_stdout(stdout), contextlib.redirect_stderr(stderr), \
                mock.patch.object(run.Gate, "npmCli", side_effect=AssertionError("the vet installs nothing")), \
                mock.patch.object(run.Gate, "npmPackages", side_effect=AssertionError("the vet installs nothing")):
            try:
                run.main()
            except SystemExit as exit:
                code = exit.code
        lines = stdout.getvalue().splitlines()
        self.assertEqual(len(lines), 1, stdout.getvalue())
        seconds = json.loads(lines[0])
        self.assertEqual(list(seconds), ["seconds"])
        self.assertGreaterEqual(seconds["seconds"], 0)
        return code, stderr.getvalue()

    def standIn(self, script):
        """A directory holding a stand-in go, script its sh body, for PATH's front."""
        directory = os.path.join(self.directory, "stand-in")
        os.makedirs(directory, exist_ok=True)
        with open(os.path.join(directory, "go"), "w") as handle:
            handle.write("#!/bin/sh\n" + script)
        os.chmod(os.path.join(directory, "go"), 0o755)
        return directory

    def test_green_exits_zero(self):
        self.assertEqual(self.vetted(), (0, ""))

    def test_a_darwin_only_file_that_doesn_t_compile_is_refused_with_the_vet_s_last_lines(self):
        self.write("internal/apple/witness_test.go", self.sources["internal/apple/witness_test.go"].replace("Resolve()", "native.ResolveTarget()"))
        self.commit("the witness calls a moved function")
        code, stderr = self.vetted()
        self.assertEqual(code, run.RefusalExit, stderr)
        self.assertTrue(stderr.startswith("refused: darwin-vet: vet: internal/apple/witness_test.go:7:"), stderr)
        self.assertIn("undefined: native", stderr)
        self.assertLessEqual(len(stderr.encode()), len("refused: darwin-vet: \n") + 1500)

    def test_an_analyzer_s_finding_is_refused(self):
        # -json exits 0 with a finding, so the verdict is read from what the vet printed, never its exit alone.
        self.write("internal/apple/say.go", 'package apple\n\nimport "fmt"\n\nfunc Say() { fmt.Printf("%d\\n", "s") }\n')
        self.commit("a printf finding")
        code, stderr = self.vetted()
        self.assertEqual(code, run.RefusalExit, stderr)
        self.assertEqual(stderr, 'refused: darwin-vet: internal/apple/say.go:5:26: fmt.Printf format %d has arg "s" of wrong type string (printf)\n')

    def assertCantRun(self, said):
        code, stderr = self.vetted()
        self.assertEqual(code, run.FastCheckCantRunExit, stderr)
        self.assertTrue(stderr.startswith("fast check couldn't run: "), stderr)
        self.assertIn(said, stderr)

    def test_go_s_own_setup_failure_can_t_run(self):
        # A go.mod past the box's go under GOTOOLCHAIN=local: go's "go: " setup failure.
        self.write("go.mod", self.sources["go.mod"].replace("go 1.24", "go 1.99"))
        self.commit("a go the box doesn't run")
        with mock.patch.dict(os.environ, {"GOTOOLCHAIN": "local"}):
            self.assertCantRun("go: go.mod requires go >= 1.99")

    def test_a_module_the_box_s_cache_doesn_t_hold_can_t_run(self):
        # Required, summed, imported, and in no cache: under GOPROXY=off it fails to load at the import, never a diagnostic.
        self.write("go.mod", self.sources["go.mod"] + "\nrequire example.com/nope v1.0.0\n")
        self.write("go.sum", "example.com/nope v1.0.0 h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=\n"
                             "example.com/nope v1.0.0/go.mod h1:47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=\n")
        self.write("internal/apple/nope.go", 'package apple\n\nimport _ "example.com/nope"\n')
        self.commit("a module no cache holds")
        self.assertCantRun("example.com")

    def test_a_cache_it_can_t_write_can_t_run(self):
        with mock.patch.dict(os.environ, {"GOCACHE": "/proc/fast-check-no-cache"}):
            self.assertCantRun("build cache")

    def test_packages_that_can_t_be_listed_can_t_run(self):
        os.remove(os.path.join(self.tree, run.compilerDependenciesPath))
        self.commit("no compiler map")
        self.assertCantRun("the darwin vet's packages can't be listed: ")

    def test_a_failed_git_grep_can_t_run(self):
        def grep(command, *arguments, **options):
            if "grep" in command and "-l" in command:
                return subprocess.CompletedProcess(command, 128, "", "fatal: planted grep failure\n")
            return realRun(command, *arguments, **options)
        with mock.patch.object(run.subprocess, "run", side_effect=grep):
            self.assertCantRun("git grep for darwin build lines failed")

    def test_a_vet_that_can_t_start_can_t_run(self):
        # A tree at another commit than the sha asked.
        parent = run.git(self.tree, "rev-parse", "HEAD")
        self.write("internal/plain/more.go", "package plain\n")
        self.commit("one more")
        code, stderr = self.vetted(parent)
        self.assertEqual(code, run.FastCheckCantRunExit)
        self.assertTrue(stderr.startswith("fast check couldn't run: ValueError: %s is checked out at " % self.tree), stderr)
        # No go on PATH, git still there.
        bare = os.path.join(self.directory, "bin")
        os.makedirs(bare)
        os.symlink(shutil.which("git"), os.path.join(bare, "git"))
        with mock.patch.dict(os.environ, {"PATH": bare}):
            code, stderr = self.vetted()
        self.assertEqual(code, run.FastCheckCantRunExit)
        self.assertTrue(stderr.startswith("fast check couldn't run: FileNotFoundError: "), stderr)
        self.assertIn("'go'", stderr)

    def test_the_vet_runs_offline_with_json(self):
        said = os.path.join(self.directory, "said")
        standIn = self.standIn('echo "$GOPROXY $*" > "%s"\necho "{}"\n' % said)
        with mock.patch.dict(os.environ, {"PATH": standIn + os.pathsep + os.environ["PATH"], "GOPROXY": "direct"}):
            self.assertEqual(self.vetted(), (0, ""))
        with open(said) as handle:
            self.assertTrue(handle.read().startswith("off vet -json ./internal/apple/... "))

    @unittest.skipUnless(os.path.isdir("/proc/self"), "reads each process's state from /proc, as killSessions does")
    def test_sigterm_ends_the_vet_s_go_and_its_session_before_it_exits(self):
        # M2: Loom's bound signals python's group, which never reaches go's own session (spawn's start_new_session).
        pids = os.path.join(self.directory, "go-pids")
        standIn = self.standIn('sleep 30 &\necho "$$ $!" > "%s.partial" && mv "%s.partial" "%s"\nwait\n' % (pids, pids, pids))
        python = realPopen([sys.executable, run.__file__, "--fast-check-darwin-vet", "--tree", self.tree, "--sha", run.git(self.tree, "rev-parse", "HEAD"),
                            "--tools", self.tree], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True,
                           env=dict(os.environ, PATH=standIn + os.pathsep + os.environ["PATH"]))
        try:
            deadline = time.monotonic() + 60
            while not os.path.exists(pids) and time.monotonic() < deadline and python.poll() is None:
                time.sleep(0.05)
            self.assertTrue(os.path.exists(pids), python.communicate(timeout=30) if python.poll() is not None else "no go started")
            python.send_signal(signal.SIGTERM)
            stdout, stderr = python.communicate(timeout=60)
        finally:
            if python.poll() is None:
                python.kill()
                python.wait()
        self.assertEqual(python.returncode, run.FastCheckCantRunExit, stderr)
        self.assertTrue(stderr.startswith("fast check couldn't run: RuntimeError: SIGTERM at "), stderr)
        self.assertEqual(list(json.loads(stdout)), ["seconds"])

        def state(pid):
            try:
                with open("/proc/%d/stat" % pid) as handle:
                    return handle.read().rsplit(")", 1)[1].split()[0]
            except OSError:
                return None
        with open(pids) as handle:
            started = [int(pid) for pid in handle.read().split()]
        deadline = time.monotonic() + 10
        while any(state(pid) not in (None, "Z", "X") for pid in started) and time.monotonic() < deadline:
            time.sleep(0.05)
        # go and the child it started in its session are both gone (or dead, awaiting a reaper).
        self.assertEqual([pid for pid in started if state(pid) not in (None, "Z", "X")], [])

    def test_the_last_lines_fit_their_bound_whole(self):
        self.assertEqual(run.lastLines("a\nb\n", 1500), "a\nb")
        self.assertEqual(run.lastLines("first line\n" + "x" * 10 + "\nlast", 12), "last")
        self.assertEqual(run.lastLines("y" * 20, 5), "yyyyy")

    def test_a_finding_replayed_from_another_checkout_is_said_tree_relative(self):
        # Go replays a cached vet's output with the paths of the run that made it: on a box whose build cache is shared,
        # another checkout's (J-wxqxnre on Golem). The file is named by its tail in this tree. Mutant: the tail search
        # dropped, so the other checkout's absolute path is refused.
        self.write("internal/apple/say.go", 'package apple\n\nfunc Say() {}\n')
        elsewhere = "/srv/loom-keeper/ahra/lh-other/tmp.x/tree/internal/apple/say.go"
        replayed = '{\n\t"example.com/fixture/internal/apple": {\n\t\t"printf": [\n\t\t\t{\n\t\t\t\t"posn": "%s:5:26",\n' \
                   '\t\t\t\t"message": "fmt.Printf format %%d has arg \\"s\\" of wrong type string"\n\t\t\t}\n\t\t]\n\t}\n}\n' % elsewhere
        diagnostics, errors = run.vetDiagnostics(replayed + "vet: %s:3:1: undefined: x\n" % elsewhere, self.tree)
        self.assertEqual(errors, [])
        self.assertEqual(diagnostics, ['internal/apple/say.go:5:26: fmt.Printf format %d has arg "s" of wrong type string (printf)',
                                       "vet: internal/apple/say.go:3:1: undefined: x"])
        # A file no tail of which is in the tree stays as go said it, and a relative one is kept.
        self.assertEqual(run.treePosition("/nowhere/at/all.go:1:2", self.tree), "/nowhere/at/all.go:1:2")
        self.assertEqual(run.treePosition("internal/apple/say.go:1:2", self.tree), "internal/apple/say.go:1:2")


class VetDiagnostics(unittest.TestCase):
    """vetDiagnostics over go vet's output as Loom's box printed it on go 1.27.1 (J-2vtq883 and J-5v7y2cj, recorded
    here): a package's type or syntax error and an analyzer's -json finding are diagnostics, and none of go's setup or
    loading failures is. Plain vet's finding shares the loader's file:line:col: shape, which is why the census vets with
    -json."""

    root = "/tmp/adamic-gate/tmpn74oig4_/tmpj80mpf3u"
    finding = ('{\n\t"example.com/m/apple": {\n\t\t"printf": [\n\t\t\t{\n\t\t\t\t"posn": "%s/apple/a.go:5:24",\n\t\t\t\t"end": "%s/apple/a.go:5:26",\n'
               '\t\t\t\t"message": "fmt.Printf format %%d has arg \\"s\\" of wrong type string"\n\t\t\t}\n\t\t]\n\t}\n}\n') % (root, root)

    def test_diagnostics_are_read_apart_from_setup_failures(self):
        printed = 'apple/a.go:5:24: fmt.Printf format %d has arg "s" of wrong type string (printf)'
        for name, text, diagnostics in (
                ("clean", "{}\n", []),
                ("a finding", self.finding, [printed]),
                ("a type error", "# example.com/m/apple\n# [example.com/m/apple]\nvet: apple/w_test.go:7:32: undefined: native\n",
                 ["vet: apple/w_test.go:7:32: undefined: native"]),
                ("a syntax error", "# example.com/m/apple\nvet: apple/a.go:3:9: expected ')', found '{'\n", ["vet: apple/a.go:3:9: expected ')', found '{'"]),
                ("a type error and a finding", "# example.com/m/other\nvet: other/o.go:3:13: cannot use \"s\" (untyped string constant) as int value in variable declaration\n" + self.finding,
                 ['vet: other/o.go:3:13: cannot use "s" (untyped string constant) as int value in variable declaration', printed]),
                ("a module the cache can't hold", "go: downloading example.com/nope v1.0.0\napple/a.go:3:8: mkdir /opt/adamic-block/gomodcache/cache/download/example.com: read-only file system\n", []),
                ("a go.sum entry missing", "apple/a.go:3:8: missing go.sum entry for module providing package example.com/nope (imported by example.com/m/apple); to add:\n\tgo get example.com/m/apple\n", []),
                ("a syntax error in the imports", "apple/a.go:4:2: string literal not terminated\n", []),
                ("a go past the box's", "go: go.mod requires go >= 1.99 (running go 1.27.1; GOTOOLCHAIN=local)\n", []),
                ("a cache it can't open", "failed to initialize build cache at /proc/no-cache-here: mkdir /proc/no-cache-here: no such file or directory\n", []),
                ("an import cycle", "package example.com/m/apple\n\timports example.com/m/other from a.go\n\timports example.com/m/apple from o.go: import cycle not allowed\n", []),
                ("plain vet's finding", 'apple/a.go:5:24: fmt.Printf format %d has arg "s" of wrong type string\n', [])):
            with self.subTest(name):
                self.assertEqual(run.vetDiagnostics(text, self.root), (diagnostics, []))

    def test_an_analyzer_s_own_error_is_an_error_never_a_diagnostic(self):
        self.assertEqual(run.vetDiagnostics('{"example.com/m/apple": {"printf": {"error": "analysis failed"}}}\n', self.root),
                         ([], ["example.com/m/apple printf: analysis failed"]))


if __name__ == "__main__":
    unittest.main()
