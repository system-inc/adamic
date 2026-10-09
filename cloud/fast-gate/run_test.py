#!/usr/bin/env python3
"""The fast gate fails closed: a stage that dies without reporting is red, never green.

Every process the gate starts goes through subprocess.Popen, so these tests replace it with fake
processes that pass, then make one stage at a time raise what a loaded box raises (too many open
files) or quietly do nothing, and require a red status, a nonzero exit and the stage named. The
all-pass case proves the harness can go green at all. Run: python3 -m unittest cloud/fast-gate/run_test.py
"""

import io
import re
import shutil
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
        return "wasi" if any("^" in part and "TestWASI" in part for part in command) and "-run" in command else ("smoke" if any(run.smokeTest in part for part in command) else "tests")
    return "tests"


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
    planned = ("build", "vet", "tests", "smoke", "census")
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
            lines = [json.dumps({"Action": "pass", "Package": "p", "Test": name}) + "\n"]
        self.lines = lines
        self.stdout = io.StringIO("".join(lines)) if stdout == subprocess.PIPE else None

    def wait(self):
        return 0

    def poll(self):
        return 0

    def communicate(self):
        return "".join(self.lines), ""


class FailClosed(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.mkdtemp()
        self.tree = os.path.join(self.directory, "tree")
        os.makedirs(os.path.join(self.tree, "cloud/fast-gate"))
        with open(os.path.join(self.tree, "cloud/fast-gate/smoke.txt"), "w") as handle:
            handle.write("internal/oracle/testdata/a.a\n")
        with open(os.path.join(self.tree, "cloud/fast-gate/executors.txt"), "w") as handle:
            handle.write("inert *.md\nstage3 stage3/*\n")
        with open(os.path.join(self.tree, "cloud/fast-gate/tools.txt"), "w") as handle:
            handle.write("# declared tools\ngo\tall\tcommand -v go\n")
        # The full gate runs stage 3's lane on every main, so the tree has one.
        os.makedirs(os.path.join(self.tree, "stage3/lane"))
        with open(os.path.join(self.tree, "stage3/lane/run.sh"), "w") as handle:
            handle.write("exit 0\n")
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
                mock.patch.object(run.Gate, silent, lambda *arguments: None) if silent else mock.patch.object(run, "smokeTest", run.smokeTest), \
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
        for full, stages in ((False, ["build", "vet", "tests", "smoke", "census"]), (True, ["build", "vet", "tests", "wasi", "census"])):
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
        self.assertIn("smoke", selection["executors_beyond_go_tests"])
        # Nothing but git ran: no build, vet or test.
        self.assertEqual([command[0] for command in launched if command[0] != "git"], [])

    def test_emission_that_differs_between_two_runs_is_red_at_determinism(self):
        # Oct 9: one tree emitted a 43,250-line main.c on one box and 5,470 lines on another. Stage 0 emits each smoke
        # fixture twice, in two processes; an emitter ordering by a map differs between them.
        os.makedirs(os.path.join(self.tree, "internal/oracle/testdata"), exist_ok=True)
        with open(os.path.join(self.tree, "internal/oracle/testdata/a.a"), "w") as handle:
            handle.write("console.log(1);\n")
        calls = []

        def steady(gate, binary, path):
            calls.append(path)
            return 0, "int main(void) {}\n"

        def mapOrdered(gate, binary, path):
            calls.append(path)
            return 0, "static void f(void);\nint main(void) {}\n" if len(calls) % 2 else "int main(void) {}\nstatic void f(void);\n"

        for full in (False, True):
            with self.subTest(full=full):
                calls.clear()
                with mock.patch.object(run.Gate, "emitC", steady):
                    gate, status, result = self.gate(full=full)
                self.assertTrue(status.startswith("green:"), status)
                self.assertEqual(calls, ["internal/oracle/testdata/a.a"] * 2)
                self.assertEqual(list(result["emission_sha256"]), ["internal/oracle/testdata/a.a"])
                with mock.patch.object(run.Gate, "emitC", mapOrdered):
                    gate, status, result = self.gate(full=full)
                self.assertIn("first failure at determinism", status)
                self.assertIn("internal/oracle/testdata/a.a (exit 0 and 0, first difference at line 1)", gate.failure["detail"])

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
        with mock.patch.object(run.Gate, "emitC", lambda gate, binary, path: (0, "int main(void) {}\n")):
            gate, status, result = self.gate(full=True, extra={"phase": "determinism"})
        self.assertTrue(status.startswith("green:"), status)
        self.assertEqual(result["planned_stages"], ["determinism"])
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
        for phase in ("coverage", "tools", "build", "vet", "smoke", "determinism", "census", "stage3", "catalog"):
            with self.subTest(phase=phase):
                extra = {"phase": phase}
                if phase == "census":
                    extra["census"] = self.poolLog()
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
            self.assertEqual(run.fastUnits(self.tree, base, self.sha), run.fastBasePhases + ["stage3", "workers", "a-check"])
            self.assertEqual(run.fastUnits(self.tree, self.sha, self.sha), run.fastBasePhases)
        gate, status, result = self.gate(base=base, extra={"phase": "coverage"})
        self.assertTrue(status.startswith("green:"), status)
        self.assertIn("workers/probe.mjs", result["changed_files"])
        self.assertEqual(result["planned_stages"], ["coverage"])

    def test_unknown_and_whole_only_fast_phases_are_red(self):
        for phase in ("no-such-phase", "wasi", "tests", "workers", "a-check", "catalog-apply", "darwin"):
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

    def test_several_fast_phases_complete_the_pool_landing_record(self):
        gate, status, result = self.gate(extra={"phases": "build,vet,smoke,census", "census": self.poolLog()})
        self.assertTrue(status.startswith("green:"), status)
        self.assertEqual(result["planned_stages"], ["build", "vet", "smoke", "census"])
        self.assertEqual(result["stages_exit"], {phase: 0 for phase in result["planned_stages"]})
        self.assertEqual(landingProblems(self.sha, [self.poolPart(), (status, result)]), [])
        with open(os.path.join(gate.arguments.out, "test.jsonl")) as handle:
            log = handle.read()
        self.assertIn("TestPool", log)
        self.assertIn("TestOne", log)  # smoke appended, rather than overwriting the pool's log

    def test_a_missing_smoke_or_red_vet_cannot_complete_the_pool_record(self):
        gate, status, result = self.gate(extra={"phases": "build,vet,census", "census": self.poolLog()})
        self.assertTrue(status.startswith("green:"), status)
        self.assertTrue(any("smoke" in problem for problem in landingProblems(self.sha, [self.poolPart(), (status, result)])))
        gate, status, result = self.gate(failing="vet", extra={"phases": "build,vet,smoke,census", "census": self.poolLog()})
        self.assertIn("first failure at vet", status)
        self.assertFalse(result["vet_ok"])
        self.assertTrue(landingProblems(self.sha, [self.poolPart(), (status, result)]))
        gate, status, result = self.gate(extra={"phases": "build,vet,smoke,census", "census": self.poolLog()})
        result["sha"] = "b" * 40
        self.assertTrue(landingProblems(self.sha, [self.poolPart(), (status, result)]))

    def test_a_silent_fast_phase_is_red_even_in_a_combined_record(self):
        for method, phase in (("build", "build"), ("vet", "vet"), ("smoke", "smoke"), ("checkCensus", "census")):
            with self.subTest(phase=phase):
                gate, status, result = self.gate(silent=method, extra={"phases": "build,vet,smoke,census", "census": self.poolLog()})
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
        barrier = threading.Barrier(7)
        stages = {"toolsDeclared": "tools", "build": "build", "vet": "vet", "testSplit": "tests",
                  "smoke": "smoke", "determinism": "determinism", "stage3": "stage3"}
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


class Stage3Red(unittest.TestCase):
    def test_a_lane_red_names_the_lanes_verdict_and_last_lines(self):
        with tempfile.TemporaryDirectory() as root:
            tree, out = os.path.join(root, "tree"), os.path.join(root, "out")
            os.makedirs(os.path.join(tree, "stage3/lane"))
            os.makedirs(out)
            open(os.path.join(tree, "stage3/lane/run.sh"), "w").close()
            for command in (["init", "-q"], ["add", "."], ["-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "t"]):
                realRun(["git", "-C", tree] + command, check=True)
            gate = run.Gate(types.SimpleNamespace(tree=tree, out=out, tools=tree, sha="a" * 40, base="a" * 40, full=False,
                                                  complete=False, branch="", branch_source="", session="", session_source=""))

            def spawn(command, stdout, stderr=None, directory=None, environment=None):
                lane = command[:2] == ["bash", "stage3/lane/run.sh"]
                if lane:
                    os.makedirs(command[2], exist_ok=True)
                    with open(os.path.join(command[2], "verdict.txt"), "w") as handle:
                        handle.write("FAIL stage3 landing lane: test counts: expected 1 failing, observed 4\n")
                stdout.write("".join("noise %d\n" % i for i in range(20)) + ("failed test names: observed ['unusedTypeParameters']\n" if lane else ""))
                return types.SimpleNamespace(wait=lambda: 1 if lane else 0)

            with mock.patch.object(gate, "spawn", spawn):
                gate.stage3()
            detail = gate.failure["detail"]
            self.assertIn("stage3-lane exit 1", detail)
            self.assertIn("FAIL stage3 landing lane: test counts", detail)
            self.assertIn("failed test names: observed ['unusedTypeParameters']", detail)
            self.assertNotIn("noise 3", detail)


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

    def budget(self, rows, units=None):
        gate = run.Gate.__new__(run.Gate)
        gate.arguments = mock.Mock(tree=self.tree, tools=self.tree, base=self.base, sha=self.head, full=False, out=self.directory)
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
                handle.write('gateBase() { echo "main ' + 'a'*40 + '"; }\nclassify() { echo B; }\n')
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
        self.assertEqual(self.gate.command.call_args.args[0], ["go", "list", "-deps", "-test", "-json", "./..."])

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
            ("computed invocation detection removed", '|"os/exec"|os\\.StartProcess|syscall\\.Exec', '', "test_census_rejects_computed_compiler_outside_stage1"),
            ("census restricted to stage1", '"ls-files").splitlines()', '"ls-files", "--", "stage1").splitlines()', "test_census_rejects_computed_compiler_outside_stage1"),
            ("removed map entry accepted", 'if missing:', 'if False:', "test_census_rejects_removed_declaration"),
            ("base consumers red the change", 'missing = [path for path in missing if path in touchedFiles]', 'pass', "test_an_undeclared_consumer_on_the_base_is_named_not_red"),
            ("branch map ignored", 'declarations.setdefault(field, {}).update(layer[field])', 'pass', "test_a_branch_declares_its_own_consumer_over_the_tools_map"),
            ("driver dependencies not validated", 'any(value not in directories for value in inputs)', 'False', "test_invalid_driver_map_fails_closed"),
            ("script census removed", 'if re.search(r"subprocess|child_process|', 'if False and re.search(r"subprocess|child_process|', "test_census_rejects_undeclared_script_driver"),
            ("driver selection removed", 'self.result["selected_compiler_drivers"] = sorted(', 'self.result["selected_compiler_drivers"] = sorted([] if True else ', "test_stage3_driver_declaration_and_selection"),
            ("census disabled", 'if missing:', 'if False:', "test_census_rejects_a_new_compiler_runner"),
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
                                                                "stage3 stage3-apply-tests", "stage3 stage3-lane-tests", "stage3 stage3-lane",
                                                                "catalog 1", "catalog 12", "determinism"])

    def test_the_pool_lists_fast_non_test_units_from_a_plain_checkout(self):
        listed = realRun([sys.executable, run.__file__, "--list-units", "--tree", "/unused"],
                         capture_output=True, text=True, check=True, timeout=90).stdout.splitlines()
        self.assertEqual(listed, run.fastBasePhases)
        self.assertIn("smoke", listed)
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
                self.assertEqual([row["name"] for row in result[phase+"_units"] if row["status"]=="failed"],
                                 ["02 planted"] if phase=="catalog" else ["planted"])
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
                self.assertTrue(all(name not in ("tests", "wasi", "smoke") for name in calls[calls.index("audit"):]))
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

    def test_two_runs_keep_test_and_drain_queues_isolated(self):
        self.publisher()
        # Two runs of the very same slot must still get fresh queues. Real child
        # processes consume the environment passed by spawn, just as the publisher does.
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
        self.assertNotEqual(first.buildStoreEnvironment, second.buildStoreEnvironment)

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
            with mock.patch.object(gate, "spawn", spawn), mock.patch.object(run, "unitKillSeconds", 0.05), mock.patch("builtins.print"):
                gate.cacheDrain(name)
            row = gate.result["cache_drain_units"][-1]
            self.assertNotEqual(row["exit"], 0)
            self.assertIn("killed at 90 s", row["detail"])
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


if __name__ == "__main__":
    unittest.main()


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

    def test_wasi_fixture_path_mutant_is_caught(self):
        original = run.phaseInputProviders["wasi"]
        fixture = run.wasiFixtures(self.tree)[0]
        def mutant(tree, unit, full):
            inputs = original(tree, unit, full)
            inputs["paths"].remove(os.path.dirname(fixture))
            return inputs
        with mock.patch.dict(run.phaseInputProviders, wasi=mutant):
            result = unittest.TestResult()
            PhaseInputs("test_wasi_inputs_cover_inventory_and_command_working_directories").run(result)
        self.assertEqual(len(result.failures), 1, result.errors)
        self.assertEqual(result.errors, [])
