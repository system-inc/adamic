#!/usr/bin/env python3
"""The fast gate fails closed: a stage that dies without reporting is red, never green.

Every process the gate starts goes through subprocess.Popen, so these tests replace it with fake
processes that pass, then make one stage at a time raise what a loaded box raises (too many open
files) or quietly do nothing, and require a red status, a nonzero exit and the stage named. The
all-pass case proves the harness can go green at all. Run: python3 -m unittest cloud/fast-gate/run_test.py
"""

import io
import json
import os
import subprocess
import sys
import tempfile
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

    def gate(self, full=False, broken=None, silent=None, unowned=(), base=None):
        out = tempfile.mkdtemp(dir=self.directory)
        arguments = mock.Mock(tree=self.tree, sha=self.sha, base=base or self.sha, tools=self.tree, out=out, parallel=4, full=full,
                              branch="", branch_source="", session="", session_source="", weights=None)

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

    def test_every_stage_raising_is_red(self):
        for full, stages in ((False, ["build", "vet", "tests", "smoke", "census"]), (True, ["build", "vet", "tests", "wasi", "census"])):
            for broken in stages:
                with self.subTest(full=full, stage=broken):
                    gate, status, result = self.gate(full=full, broken=broken)
                    self.assertTrue(status.startswith("red:"), status)
                    self.assertIsNotNone(gate.failure)
                    self.assertNotEqual(result["stages_exit"].get(broken), 0)

    def test_a_stage_that_reports_nothing_is_red(self):
        for silent, stageName in (("build", "build"), ("vet", "vet"), ("testSplit", "tests"), ("checkCensus", "census")):
            with self.subTest(stage=stageName):
                gate, status, result = self.gate(silent=silent)
                self.assertTrue(status.startswith("red:"), status)
                self.assertIn(stageName, status)


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


if __name__ == "__main__":
    unittest.main()
