#!/usr/bin/env python3
"""The fast gate fails closed: a stage that dies without reporting is red, never green.

Every process the gate starts goes through subprocess.Popen, so these tests replace it with fake
processes that pass, then make one stage at a time raise what a loaded box raises (too many open
files) or quietly do nothing, and require a red status, a nonzero exit and the stage named. The
all-pass case proves the harness can go green at all. Run: python3 -m unittest cloud/fast-gate/run_test.py
"""

import io
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
    if command[:2] == ["go", "run"]:
        return "census"
    if command[:2] == ["go", "test"] and "-c" not in command:
        return "wasi" if "^TestWASI$" in command and "-run" in command else ("smoke" if any(run.smokeTest in part for part in command) else "tests")
    return "tests"


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
            lines = [json.dumps({"Action": "pass", "Package": "p", "Test": "TestOne"}) + "\n"]
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

    def gate(self, full=False, broken=None, silent=None, unowned=(), base=None, failing=None, runToEnd=False):
        out = tempfile.mkdtemp(dir=self.directory)
        arguments = mock.Mock(tree=self.tree, sha=self.sha, base=base or self.sha, tools=self.tree, out=out, parallel=4, full=full,
                              branch="", branch_source="", session="", session_source="", weights=None, run_to_end=runToEnd)

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
                mock.patch.object(run.Gate, "command", side_effect=lambda command, **options: listing if command[:2] == ["go", "list"] else realRun(command, **options)), \
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
            self.assertEqual(set(result["stages_exit"]), set(result["planned_stages"]))
            # Every verdict carries its long-test ledger, empty or not.
            self.assertEqual((result["long_tests"], result["long_test_threshold_seconds"]), (0, run.longTestSeconds))
            self.assertIn("0 units over 30 s", status)
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

    def test_a_stage_that_reports_nothing_is_red(self):
        for silent, stageName in (("build", "build"), ("vet", "vet"), ("testSplit", "tests"), ("checkCensus", "census")):
            with self.subTest(stage=stageName):
                gate, status, result = self.gate(silent=silent)
                self.assertTrue(status.startswith("red:"), status)
                self.assertIn(stageName, status)


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
        realRun(["go", "test", "-c", "-o", self.binary, "."], cwd=self.tree, check=True, timeout=120)

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
        for name, fields in (("internal/load", {}), ("middle", {"TestImports": [run.module + "/internal/load"]}),
                             ("outer", {"XTestImports": [run.module + "/middle"]}),
                             ("ordinary", {"Imports": [run.module + "/internal/load"]}), ("unrelated", {}), ("internal/oracle", {}), ("stage1/gaps", {})):
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

    def test_transitive_test_imports_and_unrelated_package(self):
        packages, unowned = self.gate.touched(["internal/load/load.go"])
        self.assertIn(run.module + "/ordinary", packages)
        self.assertIn(run.module + "/middle", packages)
        self.assertIn(run.module + "/outer", packages)
        self.assertNotIn(run.module + "/unrelated", packages)
        self.assertEqual(unowned, [])
        self.assertEqual(self.gate.command.call_args.args[0], ["go", "list", "-deps", "-test", "-json", "./..."])

    def test_test_variants_normalize_to_real_packages(self):
        self.packages[1]["ImportPath"] += " [" + run.module + "/middle.test]"
        self.packages[2]["XTestImports"] = [self.packages[1]["ImportPath"]]
        self.packages.append(dict(Dir=os.path.join(self.tree, "outer"), Name="main", ImportPath=run.module + "/outer.test", Imports=[run.module + "/outer"]))
        self.gate.command.return_value.stdout = "".join(json.dumps(p) for p in self.packages)
        packages, _ = self.gate.touched(["internal/load/load.go"])
        self.assertIn(run.module + "/middle", packages)
        self.assertIn(run.module + "/outer", packages)
        self.assertNotIn(run.module + "/outer.test", packages)

    def test_external_test_package_uses_its_real_owner(self):
        self.packages.append(dict(Dir=os.path.join(self.tree, "outer"), ForTest=run.module + "/outer", ImportPath=run.module + "/outer_test [" + run.module + "/outer.test]", Imports=[run.module + "/internal/load"]))
        self.gate.command.return_value.stdout = "".join(json.dumps(p) for p in self.packages)
        packages, unowned = self.gate.touched(["outer/source.go"])
        self.assertEqual(packages, [run.module + "/outer"])
        self.assertEqual(unowned, [])

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
        self.packages[0]["EmbedFiles"] = ["runtime/header.h"]
        self.gate.command.return_value.stdout = "".join(json.dumps(p) for p in self.packages)
        packages, unowned = self.gate.touched(["internal/load/runtime/header.h", "middle/testdata/input.ts", "unknown.txt"])
        self.assertIn(run.module + "/outer", packages)
        self.assertEqual(unowned, ["unknown.txt"])


class ReverseDependencyMutants(unittest.TestCase):
    def test_each_check_kills_its_mutant(self):
        with open(run.__file__) as handle:
            original = handle.read()
        mutants = [
            ("external test owner lost", 'package.get("ForTest") or ', '', "test_external_test_package_uses_its_real_owner"),
            ("test variants not normalized", '.split(" [", 1)[0]', '', "test_test_variants_normalize_to_real_packages"),
            ("synthetic test binary selected", 'if package.get("Name") == "main"', 'if False and package.get("Name") == "main"', "test_test_variants_normalize_to_real_packages"),
            ("ordinary imports omitted", 'package.get("Imports", [])', '[]', "test_transitive_test_imports_and_unrelated_package"),
            ("test imports omitted", 'package.get("TestImports", [])', '[]', "test_transitive_test_imports_and_unrelated_package"),
            ("external test imports omitted", 'package.get("XTestImports", [])', '[]', "test_transitive_test_imports_and_unrelated_package"),
            ("transitive walk removed", 'pending.append(dependent)', 'pass', "test_transitive_test_imports_and_unrelated_package"),
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


if __name__ == "__main__":
    unittest.main()
