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
import shutil
import threading
import unittest
from unittest import mock

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import run  # noqa: E402
import shards  # noqa: E402
from pathlib import Path

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

    def gate(self, full=False, broken=None, silent=None, unowned=(), base=None, packages_file=None, once_steps=False):
        out = tempfile.mkdtemp(dir=self.directory)
        arguments = mock.Mock(tree=self.tree, sha=self.sha, base=base or self.sha, tools=self.tree, out=out, parallel=4, full=full,
                              branch="", branch_source="", session="", session_source="", weights=None, packages_file=packages_file, once_steps=once_steps)

        self.commands = []

        def popen(command, **options):
            self.commands.append(command)
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

    def test_package_shard_runs_only_its_packages(self):
        path = Path(self.directory) / "packages.txt"
        path.write_text("example.com/p\n")
        gate, status, result = self.gate(full=True, packages_file=str(path))
        self.assertTrue(status.startswith("green:"), status)
        self.assertEqual(result["planned_stages"], ["tests"])
        tests = [command for command in self.commands if command[:2] == ["go", "test"]]
        self.assertEqual(len(tests), 1)
        self.assertEqual(tests[0][-1], "example.com/p")
        self.assertIn("-count=1", tests[0])
        self.assertIn("-skip", tests[0])

    def test_empty_shard_does_not_test_current_directory(self):
        path = Path(self.directory) / "empty.txt"
        path.write_text("")
        gate, status, result = self.gate(full=True, packages_file=str(path))
        self.assertTrue(status.startswith("green:"), status)
        self.assertFalse(any(command[:2] == ["go", "test"] for command in self.commands))

    def test_once_shard_defers_census(self):
        path = Path(self.directory) / "packages.txt"
        path.write_text("example.com/p\n")
        gate, status, result = self.gate(full=True, packages_file=str(path), once_steps=True)
        self.assertTrue(status.startswith("green:"), status)
        self.assertNotIn("census", result["planned_stages"])
        self.assertIn("wasi", result["planned_stages"])
        self.assertIn("build", result["planned_stages"])

    def test_unknown_package_is_red(self):
        path = Path(self.directory) / "packages.txt"
        path.write_text("unknown\n")
        gate, status, result = self.gate(full=True, packages_file=str(path))
        self.assertTrue(status.startswith("red:"), status)


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


class ShardPlan(unittest.TestCase):
    def test_deterministic_balanced_and_exact(self):
        packages = ["p%03d" % index for index in range(144)]
        weights = {name: 10.0 for name in packages}
        boxes = {"home": 64, "server": 64, "chonchon": 16}
        layout = shards.plan(packages, weights, boxes)
        self.assertEqual(layout, shards.plan(list(reversed(packages)), dict(reversed(list(weights.items()))), dict(reversed(list(boxes.items())))))
        assigned = [name for row in layout["shards"] for name in row["packages"]]
        self.assertEqual(sorted(assigned), sorted(packages))
        self.assertEqual(len(assigned), len(set(assigned)))
        self.assertEqual([row["estimated_wall_seconds"] for row in layout["shards"]], [10.0] * 3)
        self.assertEqual(layout["once_box"], "home")
        self.assertEqual(sum(row["once_steps"] for row in layout["shards"]), 1)

    def test_longest_first_and_unknown_weight(self):
        layout = shards.plan(["a", "b", "c"], {"a": 100, "b": 10}, {"small": 1, "large": 4})
        self.assertEqual(layout["once_box"], "large")
        self.assertEqual(layout["shards"][0]["packages"][0], "a")
        self.assertEqual(sum(row["estimated_seconds"] for row in layout["shards"]), 111)

    def test_invalid_inputs_fail(self):
        for packages, weights, boxes in [(["a", "a"], {}, {"x": 1}), (["a"], {}, {"x": 0}), (["a"], {"a": float("nan")}, {"x": 1})]:
            with self.assertRaises(ValueError):
                shards.plan(packages, weights, boxes)


class ShardMerge(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.layout = shards.plan(["a", "b"], {"a": 4, "b": 1}, {"home": 2, "server": 1})
        self.census = {"sha": "sha", "tools_sha": "tools", "finished": True, "stages_exit": {"census": 0}}
        for row in self.layout["shards"]:
            directory = self.root / "shards" / str(row["index"])
            directory.mkdir(parents=True)
            events = [{"Package": package, "Test": "TestOne", "Action": action} for package in row["packages"] for action in ("pass", "skip")]
            (directory / "test.jsonl").write_text("".join(json.dumps(event) + "\n" for event in events))
            stages = ["coverage", "build", "vet", "tests", "wasi", "stage3", "catalog"] if row["once_steps"] else ["tests"]
            shards.write_json(directory / "full.json", {"sha": "sha", "tools_sha": "tools", "package_list": row["packages"], "finished": True,
                "failure": None, "planned_stages": stages, "stages_exit": {stage: 0 for stage in stages}, "steps_seconds": {"tests": 1},
                "pass": len(events) // 2, "skip": len(events) // 2, "fail": 0})

    def merge(self, **options):
        return shards.merge(self.layout, self.root, "sha", "tools", census=self.census, **options)

    def test_counts_add_up_and_compare(self):
        result = self.merge()
        self.assertIsNone(result["failure"])
        self.assertEqual((result["pass"], result["fail"], result["skip"]), (2, 0, 2))
        self.assertTrue((self.root / "status.txt").read_text().startswith("green:"))
        self.assertEqual(shards.event_counts([self.root / "test.jsonl"]), shards.event_counts(list((self.root / "shards").glob("*/test.jsonl"))))
        self.assertEqual(result["stages_exit"]["census"], 0)

    def test_red_shard_names_box_and_is_sticky(self):
        path = self.root / "shards/1/full.json"
        report = json.loads(path.read_text())
        report["failure"] = {"step": "tests", "detail": "bad test", "after_seconds": 2}
        report["stages_exit"]["tests"] = 1
        shards.write_json(path, report)
        result = self.merge(final=False)
        self.assertEqual(result["failure"]["step"], "shard-1@server/tests")
        self.assertFalse((self.root / "full.json").exists())
        report["failure"] = None
        report["stages_exit"]["tests"] = 0
        shards.write_json(path, report)
        self.assertEqual(self.merge()["failure"]["step"], "shard-1@server/tests")

    def test_missing_shard_is_red(self):
        (self.root / "shards/1/full.json").unlink()
        self.assertIsNone(self.merge(final=False)["failure"])
        result = self.merge()
        self.assertEqual(result["failure"]["step"], "shard-1@server")
        self.assertIn("missing shard", result["failure"]["detail"])

    def test_missing_census_is_red(self):
        self.census = None
        self.assertEqual(self.merge()["failure"]["step"], "shard-census@home")

    def test_counts_mismatch_is_red(self):
        path = self.root / "shards/1/test.jsonl"
        path.write_text("")
        self.assertIn("counts disagree", self.merge()["failure"]["detail"])

    def test_partial_snapshot_does_not_create_false_red(self):
        path = self.root / "shards/1/full.json"
        original = path.read_text()
        path.write_text('{"sha":')
        self.assertIsNone(self.merge(final=False)["failure"])
        path.write_text(original)
        self.assertIsNone(self.merge()["failure"])

    def test_wrong_source_is_red(self):
        path = self.root / "shards/1/full.json"
        report = json.loads(path.read_text())
        report["sha"] = "other"
        shards.write_json(path, report)
        self.assertIn("wrong source", self.merge()["failure"]["detail"])

    def test_comparer_counts_repeated_events(self):
        path = self.root / "whole.jsonl"
        path.write_text('{"Package":"a","Test":"T","Action":"pass"}\n' * 2)
        self.assertEqual(shards.event_counts([path])[("a", "pass")], 2)


class ShardCoordinator(unittest.TestCase):
    def test_launcher_collects_and_runs_census_on_owner(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            sha, tools = "a" * 40, "b" * 40
            args = mock.Mock(out=str(root / "run"), sha=sha, tools=tools, boxes="home server", timeout=100)
            ready = {box: threading.Event() for box in ("home", "server")}
            calls = []

            def remote(box, script, arguments, timeout=300):
                calls.append((box, script))
                realRun(["bash", "-n"], input=script, text=True, check=True)
                directory = root / box
                directory.mkdir(exist_ok=True)
                if "nproc --all" in script:
                    return ("2" if box == "home" else "1") + "\na\nb\n\nWEIGHTS\n4 a\n1 b\n"
                if "--packages-file" in script:
                    packages = script.split("'PACKAGES'\n", 1)[1].split("PACKAGES\n", 1)[0].split()
                    stages = ["coverage", "build", "vet", "tests", "wasi", "stage3", "catalog"] if "--once-steps" in script else ["tests"]
                    (directory / "test.jsonl").write_text("".join(json.dumps({"Package": name, "Test": "T", "Action": "pass"}) + "\n" for name in packages))
                    shards.write_json(directory / "full.json", {"sha": sha, "tools_sha": tools, "finished": True, "package_list": packages,
                        "failure": None, "pass": len(packages), "fail": 0, "skip": 0,
                        "steps_seconds": {"tests": 1}, "planned_stages": stages, "stages_exit": {stage: 0 for stage in stages}})
                    ready[box].set()
                if "--census-only" in script:
                    self.assertEqual(box, "home")
                    self.assertEqual(sum(shards.event_counts([root / "uploaded.jsonl"]).values()), 2)
                    directory = root / "remote-census"
                    directory.mkdir()
                    shards.write_json(directory / "full.json", {"sha": sha, "tools_sha": tools, "finished": True, "stages_exit": {"census": 0}, "failure": None})
                return ""

            def copy(command, **options):
                self.assertEqual(command[0], "scp")
                source, destination = command[-2:]
                if ":" not in source:
                    shutil.copyfile(source, root / "uploaded.jsonl")
                elif source.endswith("-census"):
                    shutil.copytree(root / "remote-census", destination)
                else:
                    box = source.split(":", 1)[0]
                    self.assertTrue(ready[box].wait(2))
                    shutil.copytree(root / box, destination, dirs_exist_ok=True)
                return subprocess.CompletedProcess(command, 0, stdout=b"", stderr=b"")

            with mock.patch.object(shards, "remote", side_effect=remote), mock.patch.object(shards.subprocess, "run", side_effect=copy), mock.patch.object(shards.time, "sleep"), mock.patch("builtins.print"):
                shards.coordinate(args)
            result = json.loads((root / "run/full.json").read_text())
            self.assertIsNone(result["failure"])
            self.assertEqual(result["pass"], 2)
            self.assertEqual(sum("--once-steps" in script for _, script in calls), 1)
            self.assertEqual(sum("--census-only" in script for _, script in calls), 1)

    def test_down_box_finishes_red_with_name(self):
        with tempfile.TemporaryDirectory() as temporary:
            args = mock.Mock(out=str(Path(temporary) / "run"), sha="a" * 40, tools="b" * 40, boxes="home server", timeout=100)
            def remote(box, *arguments):
                if box == "server":
                    raise OSError("box down")
                return "2\na\n\nWEIGHTS\n1 a\n"
            with mock.patch.object(shards, "remote", side_effect=remote):
                shards.coordinate(args)
            result = json.loads((Path(args.out) / "full.json").read_text())
            self.assertTrue(result["finished"])
            self.assertEqual(result["failure"]["step"], "setup@server")
            self.assertTrue((Path(args.out) / "status.txt").read_text().startswith("red:"))


if __name__ == "__main__":
    unittest.main()
