import hashlib
import json
from pathlib import Path
import re
import shutil
import subprocess
import sys
from tempfile import TemporaryDirectory
import unittest

import report


HERE = Path(__file__).resolve().parent


def fixture(name):
    return (HERE / "testdata" / name).read_bytes()


class FirstLinkReportTests(unittest.TestCase):
    def test_pinned_door_groups_body_and_statement(self):
        data = fixture("body-and-statement.json")
        source = json.loads(data)
        summary = report.summarize(data)
        self.assertEqual(summary["sha256"], hashlib.sha256(data).hexdigest())
        self.assertEqual(summary["total"], 2)
        self.assertEqual({name: phase["total"] for name, phase in summary["phases"].items()},
                         {"Body": 1, "Statement": 1})
        for row in source["traps"]:
            phase = summary["phases"][row["Phase"]]
            self.assertEqual(phase["by_kind"], {row["Kind"]: 1})
            self.assertEqual(phase["by_reason"], [
                {"kind": row["Kind"], "reason": row["Reason"], "family": row["Reason"], "count": 1},
            ])
            self.assertEqual(len(phase["sites"]), 1)
            self.assertEqual(phase["sites"][0]["record"], row)
        text = report.markdown(summary)
        self.assertIn("## Body\n", text)
        self.assertIn("## Statement\n", text)

    def test_pinned_door_explains_missing_pass_records(self):
        summary = report.summarize(fixture("body.json"))
        self.assertFalse(summary["has_whole_program_pass_stops"])
        lines = report.markdown(summary).splitlines()
        self.assertIn("this head records no whole-program pass stops", lines[:8])
        self.assertEqual([line for line in lines if line.startswith("## ")], ["## Body"])

    def test_wave2_passes_flags_and_gaps(self):
        summary = report.summarize(fixture("wave2.json"))
        self.assertTrue(summary["has_whole_program_pass_stops"])
        self.assertEqual(summary["flag_counts"], {"echo": 2, "suspect": 2, "without_flags": 2})
        self.assertEqual(set(summary["phases"]), {"Body", "findCycles", "unlowerable"})
        phase = summary["phases"]["findCycles"]
        self.assertEqual((phase["total"], phase["echo"], phase["suspect"], phase["without_flags"]),
                         (2, 1, 1, 0))
        self.assertEqual(phase["by_kind"], {"Refused": 1, "Gap": 1})
        self.assertEqual([site["record"]["Kind"] for site in phase["sites"]], ["Refused", "Gap"])
        self.assertEqual(summary["phases"]["unlowerable"]["without_flags"], 0)
        text = report.markdown(summary)
        self.assertNotIn(report.NO_PASS_STOPS, text)
        self.assertIn("## findCycles\n", text)
        self.assertIn("already recorded \\| in this unit", text)
        self.assertIn("| Echo |", text)
        self.assertIn("| Suspect |", text)
        self.assertIn("| Echo, Suspect |", text)

    def test_normalization_sites_and_unknown_fields(self):
        data = fixture("wave2.json")
        source = json.loads(data)
        summary = report.summarize(data)
        for key, value in source.items():
            self.assertEqual(summary[key], value)
        body = summary["phases"]["Body"]
        self.assertEqual(body["by_reason"], [
            {"kind": "NotYet", "reason": "a value of type example.ts", "family": "a value of type <T>", "count": 2},
        ])
        first, second = body["sites"]
        self.assertEqual(first["record"], source["traps"][0])
        self.assertEqual(first["record"]["Site"], "/fixture/tree/src/compiler/example.ts:19:3")
        self.assertEqual(second["record"]["Site"], "entry.ts:22:1")
        self.assertEqual(summary["unknown_top_level_keys"], ["FutureReport"])
        self.assertEqual((summary["absolute_sites"], summary["empty_sites"]), (3, 1))
        absolute, empty = summary["phases"]["findCycles"]["sites"]
        self.assertEqual(absolute["record"]["Site"].encode(), b"/elsewhere/shared.ts:4:2")
        self.assertEqual(empty["record"]["Site"].encode(), b"")
        self.assertNotIn("Tree-relative site", report.markdown(summary))

    def test_phase_sets_are_closed(self):
        self.assertEqual(len(report.BODY_PHASES), 5)
        self.assertEqual(len(report.WHOLE_PROGRAM_PHASES), 39)
        self.assertFalse(report.BODY_PHASES & report.WHOLE_PROGRAM_PHASES)

    def test_unknown_phase_fails_without_outputs(self):
        source = json.loads(fixture("wave2.json"))
        source["traps"][4]["Phase"] = "futurePass"
        source["by_phase"]["futurePass"] = source["by_phase"].pop("unlowerable")
        with TemporaryDirectory() as directory:
            output = Path(directory)
            (output / "traps.json").write_text(json.dumps(source), encoding="utf-8")
            run = subprocess.run([sys.executable, "-I", "-B", str(HERE / "report.py"), str(output)],
                                 capture_output=True, text=True)
            self.assertEqual(run.returncode, 1)
            self.assertEqual(run.stderr, 'firstlink report: unknown phase "futurePass" in record 4\n')
            self.assertEqual({file.name for file in output.iterdir()}, {"traps.json"})

    def test_named_step_alone_suppresses_missing_pass_header(self):
        source = json.loads(fixture("wave2.json"))
        row = source["traps"][4]
        source.update(traps=[row], total=1, by_phase={"unlowerable": 1}, by_kind={"Panic": 1},
                      by_reason={"Panic: second pass": 1}, unplaced=0, suspect=1, echo=1)
        summary = report.summarize(json.dumps(source).encode())
        self.assertTrue(summary["has_whole_program_pass_stops"])
        self.assertNotIn(report.NO_PASS_STOPS, report.markdown(summary))
        self.assertEqual(set(summary["phases"]), {"unlowerable"})

    def test_echo_count_and_field_are_checked(self):
        source = json.loads(fixture("wave2.json"))
        source["echo"] -= 1
        with self.assertRaisesRegex(ValueError, "echo disagrees"):
            report.summarize(json.dumps(source).encode())
        for replacement in (None, 1):
            with self.subTest(replacement=replacement):
                source = json.loads(fixture("wave2.json"))
                if replacement is None:
                    del source["traps"][0]["Echo"]
                else:
                    source["traps"][0]["Echo"] = replacement
                with self.assertRaisesRegex(ValueError, "invalid recorded stop at index 0"):
                    report.summarize(json.dumps(source).encode())

    def test_rewrites_are_preserved_in_order(self):
        source = json.loads(fixture("wave2.json"))
        summary = report.summarize(fixture("wave2.json"))
        self.assertEqual(summary["rewrites_skipped"], source["rewrites_skipped"])
        self.assertIn("Rewrites skipped past recorded stops, in door order: "
                      "layFunctionConstructorWindows, certifySourceSlots.", report.markdown(summary))
        source["rewrites_skipped"].append(1)
        with self.assertRaisesRegex(ValueError, "rewrites_skipped"):
            report.summarize(json.dumps(source).encode())

    def test_pinned_door_records_whole_program_stop(self):
        data = fixture("setter.json")
        source = json.loads(data)
        summary = report.summarize(data)
        self.assertEqual(summary["findings"], source["findings"])
        self.assertTrue(summary["has_whole_program_pass_stops"])
        phase = summary["phases"]["checkBoundArguments"]
        self.assertEqual(phase["by_kind"], {"NotYet": 1})
        self.assertEqual([site["record"]["Site"] for site in phase["sites"]], ["setter.ts:9:15"])

    def test_inconsistent_door_counts_are_rejected(self):
        source = json.loads(fixture("wave2.json"))
        source["by_phase"]["findCycles"] -= 1
        with self.assertRaisesRegex(ValueError, "by_phase disagrees"):
            report.summarize(json.dumps(source).encode())

    def test_cli_writes_summary_and_digest(self):
        with TemporaryDirectory() as directory:
            output = Path(directory)
            source = output / "traps.json"
            shutil.copyfile(HERE / "testdata" / "body.json", source)
            run = subprocess.run([sys.executable, "-I", "-B", str(HERE / "report.py"), str(output)],
                                 cwd=directory, check=True, capture_output=True, text=True)
            digest = hashlib.sha256(source.read_bytes()).hexdigest()
            self.assertEqual(run.stdout, f"REPORT-SHA256 {digest}\n")
            self.assertEqual(json.loads((output / "firstlink.json").read_bytes())["sha256"], digest)
            self.assertIn(digest, (output / "firstlink.md").read_text(encoding="utf-8"))

    def test_tree_root_leaves_reasons_so_runs_hash_alike(self):
        # Two runs in different scratch trees, whose door named a file in a reason by its absolute path (J-nsy6d40
        # and J-r2sbc7m, alike but for that).
        digests = set()
        for root in ("/run/loom-scratch/firstlink.HTpu5X/tree", "/run/loom-scratch/firstlink.CMSv7m/tree"):
            with self.subTest(root=root), TemporaryDirectory() as directory:
                source = json.loads(fixture("setter.json"))
                reason = f"a reference to f, whose signature trapped at {root}/src/compiler/utilities.ts:7847:17"
                source["traps"][0]["Reason"] = reason
                source["by_reason"] = {f"NotYet: {reason}": 1}
                output = Path(directory)
                (output / "traps.json").write_text(json.dumps(source, indent=2) + "\n", encoding="utf-8")
                run = subprocess.run([sys.executable, "-I", "-B", str(HERE / "report.py"), "--tree-root", root,
                                      str(output)], capture_output=True, text=True)
                self.assertEqual(run.returncode, 0, run.stderr)
                summary = json.loads((output / "firstlink.json").read_bytes())
                self.assertEqual(summary["traps"][0]["Reason"],
                                 "a reference to f, whose signature trapped at src/compiler/utilities.ts:7847:17")
                self.assertEqual(summary["tree_root_stripped"], 2)
                self.assertNotIn(root, (output / "firstlink.md").read_text(encoding="utf-8"))
                self.assertEqual(run.stdout, f'REPORT-SHA256 {summary["sha256"]}\n')
                digests.add(summary["sha256"])
        self.assertEqual(len(digests), 1)
        with self.assertRaisesRegex(ValueError, "tree root"):
            report.summarize(fixture("setter.json"), "/")

    def test_report_alone_after_door_exit_1_is_a_reading(self):
        # Lower's door-2 (#nxb3xh1): a judging pass gave up, so the door wrote traps.json and traps.md alone and
        # exited 1 before emit. That is a completed report, read as a reading. Any other failure prints the row and
        # exits 3.
        cases = (
            ("report alone", "judgment-abandoned.json", 1, (), 0),
            ("door exit 2", "judgment-abandoned.json", 2, (), 3),
            ("C emitted", "judgment-abandoned.json", 1, ("c/main.c",), 3),
            ("no judgment abandoned", "setter.json", 1, (), 3),
        )
        for name, data, door_exit, emitted, want in cases:
            with self.subTest(name), TemporaryDirectory() as directory:
                output = Path(directory)
                shutil.copyfile(HERE / "testdata" / data, output / "traps.json")
                (output / "traps.md").write_text("# traps\n", encoding="utf-8")
                for path in emitted:
                    (output / path).parent.mkdir(parents=True)
                    (output / path).write_text("int main(void) { return 0; }\n", encoding="utf-8")
                run = subprocess.run([sys.executable, "-I", "-B", str(HERE / "report.py"), "--door-exit",
                                      str(door_exit), str(output)], capture_output=True, text=True)
                self.assertEqual(run.returncode, want, run.stdout + run.stderr)
                digest = hashlib.sha256((output / "traps.json").read_bytes()).hexdigest()
                lines = run.stdout.splitlines()
                self.assertEqual(lines[0], f"REPORT-SHA256 {digest}")
                if want == 0:
                    self.assertEqual(lines[1:], ["DOOR exit 1 before emit, its report alone: a judging pass gave up, "
                                                 "so this is a reading, not a failure"])
                else:
                    self.assertEqual(lines[1:], [f"DOOR exit {door_exit} after its report: row 1 is written, "
                                                 "and the run fails"])
                text = (output / "firstlink.md").read_text(encoding="utf-8")
                if data == "judgment-abandoned.json":
                    self.assertIn("Judgments abandoned past a stop, in door order: "
                                  "checkAccessorSpreads at spread.ts (Gap).", text)
                self.assertTrue((output / "firstlink.json").is_file())


if __name__ == "__main__":
    unittest.main()
