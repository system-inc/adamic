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
        # The full gate runs stage 3's lane on every main, so the tree has one.
        os.makedirs(os.path.join(self.tree, "stage3/lane"))
        with open(os.path.join(self.tree, "stage3/lane/run.sh"), "w") as handle:
            handle.write("exit 0\n")
        for command in (["init", "-q"], ["add", "."], ["-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "t"]):
            realRun(["git", "-C", self.tree] + command, check=True)
        self.sha = run.git(self.tree, "rev-parse", "HEAD")

    def gate(self, full=False, broken=None, silent=None, unowned=()):
        out = tempfile.mkdtemp(dir=self.directory)
        arguments = mock.Mock(tree=self.tree, sha=self.sha, base=self.sha, tools=self.tree, out=out, parallel=4, full=full,
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
                mock.patch.object(run.Gate, "touched", lambda gate, changed: (gate.packageDirectories.update({"p": self.tree}) or ["p"], list(unowned))), \
                mock.patch.object(run.subprocess, "run", side_effect=lambda command, **options: listing if command[:2] == ["go", "list"] else realRun(command, **options)), \
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


if __name__ == "__main__":
    unittest.main()
