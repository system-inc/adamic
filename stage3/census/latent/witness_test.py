"""Run the phase 1 fixture against the current and pre-witness census binaries."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE.parents[1] / "meter"))
from report import report_pair


UNIT_CLOCK_FIELDS = ("wall_seconds", "cpu_seconds")
METER_CLOCK_FIELDS = ("wall_seconds", "cpu_seconds", "slowest_units")


def read_rows(path):
    return [json.loads(line) for line in path.read_text().splitlines()]


def without_fields(row, fields):
    return {key: value for key, value in row.items() if key not in fields}


def timing_units(run):
    rows = read_rows(run / "latent-times.jsonl")
    assert rows, "fixture did not record timings"
    assert all(all(isinstance(row[key], (int, float)) for key in UNIT_CLOCK_FIELDS) for row in rows)
    return [without_fields(row, UNIT_CLOCK_FIELDS) for row in rows]


def meter_report(tree, run):
    # Keep the checker inputs and entry observations fixed. This test changes
    # only the two real directory-census streams consumed by the paired meter.
    compiler = tree / "src/compiler"
    sources = sorted(str(path) for path in compiler.glob("*.a"))
    entry = tree / "src/tsc/tsc.ts"
    entry.parent.mkdir(parents=True, exist_ok=True)
    entry.write_text("export {};\n")
    labels = ("main", "area")
    for label in labels:
        directory = run / label
        entry_run = directory / "tsc"
        entry_run.mkdir(parents=True, exist_ok=True)
        for name in ("latent.jsonl", "latent-times.jsonl"):
            shutil.copyfile(run / name, directory / name)
        records = [{"roots": [name], "kind": "checker"} for name in sources]
        records.append({"roots": sources, "kind": "checker"})
        (directory / "census.jsonl").write_text("".join(json.dumps(row) + "\n" for row in records))
        observation = {"roots": [str(entry)], "kind": "checker", "diagnostics": ["fixture entry diagnostic"]}
        (entry_run / "census.jsonl").write_text((json.dumps(observation) + "\n") * 2)
        (entry_run / "latent.jsonl").write_text(json.dumps({
            "roots": [str(entry)], "status": "blocked", "checker_rejected": True,
            "diagnostics": observation["diagnostics"],
        }) + "\n")
    commit = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=HERE, text=True).strip()
    (run / "compiler-mode.json").write_text(json.dumps({
        "mode": "single", "commits": {label: commit for label in labels},
    }))
    report_pair(tree, tree, run, "fixture", commit, commit)
    report = json.loads((run / "report.json").read_text())
    # The paired report also mirrors the area result at the top level.
    for result in (report, *(report["trees"][label] for label in labels)):
        time = result["latent_time"]
        assert all(key in time for key in METER_CLOCK_FIELDS)
        # File order in the meter is a wall-clock ranking. Preserve the
        # complete membership and every non-clock field in each file row.
        time = without_fields(time, METER_CLOCK_FIELDS)
        time["files"] = sorted(
            (without_fields(row, UNIT_CLOCK_FIELDS) for row in time["files"]), key=lambda row: row["file"],
        )
        result["latent_time"] = time
    return report


class LatentWitnessTests(unittest.TestCase):
    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory(prefix="latent-witness-")
        self.addCleanup(self.scratch.cleanup)
        self.root = Path(self.scratch.name)
        self.source = self.root / "src/compiler"
        shutil.copytree(HERE / "testdata/witness/census", self.source)

    def run_census(self, name, *, binary=None, enabled=False, mutant=None):
        run = self.root / name
        run.mkdir()
        env = {key: value for key, value in os.environ.items() if not key.startswith("LATENT_")}
        env.update(LATENT_ASSERT_NO_OUTPUT="1", LATENT_TIMES=str(run / "latent-times.jsonl"),
                   LATENT_WITNESS_OUT=str(run / "witness.jsonl"))
        if enabled:
            env["LATENT_WITNESS"] = "1"
        if mutant:
            env[mutant] = "1"
        process = subprocess.run([binary or CENSUS, self.source, run / "latent.jsonl"],
                                 env=env, timeout=45, capture_output=True, text=True)
        self.assertEqual(process.returncode, 0, process.stderr)
        return run

    def test_latent_witness_hidden_body_forecasts(self):
        run = self.run_census("on", enabled=True)
        records = read_rows(run / "latent.jsonl")
        units = {unit["name"]: unit for file in records[1:] for unit in file["units"] if "name" in unit}
        findings = [finding for file in records[1:] for finding in file["findings"] if finding["phase"] == "lowering"]
        for caller in ("first", "second", "third"):
            self.assertEqual([f["reason"] for f in findings if f["unit"] == units[caller]["where"]], ["var"])
        self.assertFalse(any(f["unit"] == units["later"]["where"] for f in findings))
        witnesses = {row["name"]: row for row in read_rows(run / "witness.jsonl")}
        if os.environ.get("WITNESS_TEST_MUTANT_DROP_STOP") == "1":
            del witnesses["blocked"]["stop"]
        self.assertFalse(units["blocked"]["instantiated"])
        self.assertFalse(units["clean"]["instantiated"])
        self.assertEqual(witnesses["blocked"], {
            "file": str(self.source / "z-bodies.a"), "name": "blocked", "where": units["blocked"]["where"],
            "status": "forecast", "call_count": 2, "skipped_non_concrete": 0,
            "caller": units["first"]["where"], "call_site": f"{self.source}/a-caller.a:5:12",
            "type_arguments": ["number"], "stop": {
                "kind": "Refused", "where": f"{self.source}/z-bodies.a:2:5", "reason": "var",
                "text": f"{self.source}/z-bodies.a:2:5: Adamic 0.1 refuses var; use const or let",
            },
        })
        self.assertEqual(witnesses["clean"], {
            "file": str(self.source / "z-bodies.a"), "name": "clean", "where": units["clean"]["where"],
            "status": "forecast", "call_count": 1, "skipped_non_concrete": 0,
            "caller": units["third"]["where"], "call_site": f"{self.source}/a-caller.a:15:12",
            "type_arguments": ["string", "number"],
        })

    def test_latent_witness_rejects_open_type_arguments(self):
        run = self.run_census("on", enabled=True, mutant=os.environ.get("WITNESS_TEST_MUTANT"))
        witness = next(row for row in read_rows(run / "witness.jsonl") if row["name"] == "open")
        self.assertEqual((witness["status"], witness["call_count"], witness["skipped_non_concrete"]),
                         ("no-concrete-witness", 1, 1))
        self.assertTrue({"call_site", "type_arguments", "stop"}.isdisjoint(witness))

    def test_latent_witness_default_matches_previous_census(self):
        baseline = self.run_census("baseline", binary=BASELINE)
        current = self.run_census("off", mutant=os.environ.get("WITNESS_TEST_MUTANT"))
        self.assertEqual((baseline / "latent.jsonl").read_bytes(), (current / "latent.jsonl").read_bytes())
        self.assertFalse((current / "witness.jsonl").exists(), "witness implicitly enabled")

    def test_latent_witness_preserves_census_and_meter(self):
        before = self.run_census("off")
        after = self.run_census("on", enabled=True, mutant=os.environ.get("WITNESS_TEST_MUTANT"))
        self.assertEqual((before / "latent.jsonl").read_bytes(), (after / "latent.jsonl").read_bytes(),
                         "witness leaked into census findings")
        self.assertEqual(timing_units(before), timing_units(after), "witness leaked into timed units")
        self.assertEqual(meter_report(self.root, before), meter_report(self.root, after))


if __name__ == "__main__":
    CENSUS, BASELINE = (str(Path(arg).resolve()) for arg in sys.argv[1:3])
    unittest.main(argv=[sys.argv[0], *sys.argv[3:]])
