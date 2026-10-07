import json
from pathlib import Path
import tempfile
import unittest

from proof import canonical_events


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


if __name__ == "__main__":
    unittest.main()
