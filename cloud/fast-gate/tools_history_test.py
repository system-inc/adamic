#!/usr/bin/env python3
"""Tools served before promotion remain acceptable, including pool jobs."""
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest import mock

ROOT = Path(__file__).resolve().parents[2]
spec = importlib.util.spec_from_file_location("history", ROOT / "cloud/tools-history.py")
history = importlib.util.module_from_spec(spec)
spec.loader.exec_module(history)


class ToolsHistory(unittest.TestCase):
    def test_old_served_record_survives_promotion_but_unserved_tools_do_not(self):
        with tempfile.TemporaryDirectory() as directory:
            state = Path(directory)
            history.snapshot(state, "a" * 40)
            old, served = history.snapshot(state)
            prefix = (state / "tools-good-history.tsv").read_bytes()
            history.snapshot(state, "b" * 40)
            new, later = history.snapshot(state)
            self.assertTrue((state / "tools-good-history.tsv").read_bytes().startswith(prefix))
            log = state / "tools-good-history.tsv"
            self.assertTrue(history.was_live(log, dict(tools_sha=old, tools_served_at=served)))
            self.assertTrue(history.was_live(log, dict(tools_sha=new, tools_served_at=later)))
            self.assertFalse(history.was_live(log, dict(tools_sha=new, tools_served_at=served)))
            self.assertTrue(history.was_live(log, dict(tools_sha=old, tools_served_at=later)))
            self.assertFalse(history.was_live(log, dict(tools_sha="c" * 40, tools_served_at=later)))
            self.assertFalse(history.was_live(log, dict(tools_sha=old)))

    def test_no_fake_past_for_external_promotion(self):
        with tempfile.TemporaryDirectory() as directory:
            state = Path(directory)
            history.snapshot(state, "old")
            _, served = history.snapshot(state)
            (state / "tools-good").write_text("new\n")
            new, later = history.snapshot(state)
            self.assertEqual(new, "new")
            self.assertFalse(history.was_live(state / "tools-good-history.tsv",
                                             dict(tools_sha=new, tools_served_at=served)))
            self.assertTrue(history.was_live(state / "tools-good-history.tsv",
                                            dict(tools_sha=new, tools_served_at=later)))

    def test_pool_job_keeps_serve_metadata_in_its_record_environment(self):
        source = (ROOT / "cloud/pool-job.sh").read_text()
        body = source.split("<<'PY'\n", 1)[1].split("\nPY", 1)[0]
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "job.json"
            tools = "a" * 40
            arguments = ["python", str(path), "codex/fixture", "b" * 40,
                         "c" * 40, "main", tools, "10", ""]
            with mock.patch.object(sys, "argv", arguments), mock.patch.dict(
                    os.environ, ADAMIC_FAST_GATE_SERVED_AT="123"):
                exec(compile(body, "pool-job.sh job writer", "exec"), {})
            job = json.loads(path.read_text())
            self.assertEqual(job["tools"], tools)
            self.assertEqual(job["tools_sha"], tools)
            self.assertEqual(job["tools_served_at"], 123)
            self.assertEqual(job["env"]["ADAMIC_FAST_GATE_SERVED_TOOLS"], tools)
            self.assertEqual(job["env"]["ADAMIC_FAST_GATE_SERVED_AT"], "123")

    def test_direct_pool_job_also_records_serve_time(self):
        source = (ROOT / "cloud/pool-job.sh").read_text()
        body = source.split("<<'PY'\n", 1)[1].split("\nPY", 1)[0]
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "job.json"
            arguments = ["python", str(path), "codex/fixture", "b" * 40,
                         "c" * 40, "main", "a" * 40, "10", ""]
            with mock.patch.object(sys, "argv", arguments), mock.patch.dict(os.environ, {}, clear=True):
                exec(compile(body, "pool-job.sh job writer", "exec"), {})
            job = json.loads(path.read_text())
            self.assertGreater(job["tools_served_at"], 0)
            self.assertEqual(job["env"]["ADAMIC_FAST_GATE_SERVED_AT"], str(job["tools_served_at"]))

    def test_corrupt_history_refuses(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "tools-good-history.tsv"
            path.write_text("old\t2\nnew\t1\n")
            with self.assertRaisesRegex(ValueError, "strictly ordered"):
                history.was_live(path, dict(tools_sha="new", tools_served_at=3))


if __name__ == "__main__":
    unittest.main()
