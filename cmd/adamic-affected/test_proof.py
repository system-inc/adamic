import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

from proof import canonical_events, run_set
from finish_proof import calibration_verdicts


class EventComparisonTest(unittest.TestCase):
    def compare(self, left, right):
        with tempfile.TemporaryDirectory() as directory:
            paths = [Path(directory) / name for name in ("left", "right")]
            for path, events in zip(paths, (left, right)):
                path.write_text("".join(json.dumps(event) + "\n" for event in events))
            return canonical_events(paths[0]) == canonical_events(paths[1])

    def test_only_harness_timing_is_ignored(self):
        left = [{"Action": "output", "Test": "Test/Sub", "OutputType": "frame",
                 "Output": "    --- PASS: Test/Sub (1.20s)\n", "Time": "old"},
                {"Action": "pass", "Test": "Test/Sub", "Elapsed": 1.20}]
        right = [{"Action": "output", "Test": "Test/Sub", "OutputType": "frame",
                  "Output": "    --- PASS: Test/Sub (2.40s)\n", "Time": "new"},
                 {"Action": "pass", "Test": "Test/Sub", "Elapsed": 2.40}]
        self.assertTrue(self.compare(left, right))
        right[1]["Action"] = "fail"
        self.assertFalse(self.compare(left, right))

    def test_counts_paths_and_program_frames_survive(self):
        for before, after in (("1040 writes", "1041 writes"),
                              ("/tmp/first/input.a", "/tmp/second/input.a"),
                              ("--- PASS: program (1.00s)\n", "--- PASS: program (2.00s)\n")):
            self.assertFalse(self.compare(
                [{"Action": "output", "Test": "Test", "Output": before}],
                [{"Action": "output", "Test": "Test", "Output": after}]))

    def test_parallel_order_is_per_test(self):
        a = {"Action": "run", "Test": "A"}
        b = {"Action": "run", "Test": "B"}
        c = {"Action": "pass", "Test": "A"}
        self.assertTrue(self.compare([a, b, c], [b, a, c]))
        self.assertFalse(self.compare([a, b, c], [c, b, a]))

    def test_calibration_preserves_names_actions_and_verdicts(self):
        observed = {"Test": [("run", ""), ("output", "1 MB/s"), ("pass", "")]}
        plain = {"Test": [("run", ""), ("output", "2 MB/s"), ("pass", "")]}
        self.assertEqual(calibration_verdicts(observed), calibration_verdicts(plain))
        plain["Test"][-1] = ("fail", "")
        self.assertNotEqual(calibration_verdicts(observed), calibration_verdicts(plain))
        renamed = {"OtherTest": observed["Test"]}
        self.assertNotEqual(calibration_verdicts(observed), calibration_verdicts(renamed))

    def test_skipped_list_and_durable_run_have_distinct_keys(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            state = {"skipped": ["probe"]}
            with patch("proof.run_package", return_value={"exit": 0, "events": ""}), \
                 patch("proof.timing_metadata", return_value={}):
                run_set(["probe"], "skipped_run", state, root / "state.json",
                        {"probe": {"Dir": str(root)}}, root, root / "events", {},
                        {"Packages": {"probe": {"Events": ""}}})
            saved = json.loads((root / "state.json").read_text())
            self.assertEqual(saved["skipped"], ["probe"])
            self.assertTrue(saved["skipped_run"]["packages"]["probe"]["identical"])


if __name__ == "__main__":
    unittest.main()
