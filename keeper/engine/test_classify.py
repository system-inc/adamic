"""test_classify.py: classify.py against today's real replay logs and the exact lines the audit judged by hand (#rdrbxnp).

	python3 -I test_classify.py
"""
import json, os, subprocess, sys, tempfile, unittest

here = os.path.dirname(os.path.abspath(__file__))
data = os.path.join(here, "testdata")


def classify(log, base=None):
    command = [sys.executable, "-I", os.path.join(here, "classify.py"), log] + (["--base", base] if base else [])
    result = json.loads(subprocess.run(command, check=True, capture_output=True, text=True).stdout)
    return {row["test"]: row for row in result["tests"]}, result["catchers"]


def log(events):
    handle = tempfile.NamedTemporaryFile("w", suffix=".jsonl", delete=False)
    for test, action, outputs in events:
        for output in outputs:
            handle.write(json.dumps({"Action": "output", "Package": "github.com/system-inc/adamic/internal/native", "Test": test, "Output": output + "\n"}) + "\n")
        if action:
            handle.write(json.dumps({"Action": action, "Package": "github.com/system-inc/adamic/internal/native", "Test": test}) + "\n")
    handle.close()
    return handle.name


class Classify(unittest.TestCase):
    def test_d5_compiler_panics_are_catches(self):
        # internal/lower D5 at class_static.go:177 (test-defend/deletion-set/internal-lower 428ca4a4), judged by hand:
        # both panics start inside the compiler on valid input, so both are catches.
        rows, catchers = classify(os.path.join(data, "d5-panic-0.jsonl"))
        self.assertEqual(rows["TestClassFeaturesStaticSoundness"]["outcome"], "panic-in-code")
        self.assertEqual(rows["TestClassFeaturesStaticSoundness"]["why"], "internal/ir/call_targets.go")
        self.assertIn("TestClassFeaturesStaticSoundness", catchers)
        rows, catchers = classify(os.path.join(data, "d5-panic-1.jsonl"))
        self.assertEqual(rows["TestPrivateAndPublicStaticsAgreeWithNode"]["outcome"], "panic-in-code")
        self.assertEqual(rows["TestPrivateAndPublicStaticsAgreeWithNode"]["why"], "internal/lower/arguments_length.go")
        self.assertEqual(rows["TestOriginalCycleLedger"]["outcome"], "skip")

    def test_tests_a_panic_cut_off_are_absent_not_catches(self):
        # When D5's panic took the binary down, every paused or running test was left with no action at all.
        rows, catchers = classify(log([
            ("TestEnumNeverDefault", None, ["=== RUN   TestEnumNeverDefault", "=== PAUSE TestEnumNeverDefault"]),
            ("TestAMethodReadAsAValueIsRefused", None, ["=== RUN   TestAMethodReadAsAValueIsRefused", "=== CONT  TestAMethodReadAsAValueIsRefused"]),
            ("TestPanics", "fail", ["panic: ir: virtual call has no target set", "\t/workspace/adamic/internal/ir/call_targets.go:13 +0x1"]),
        ]))
        self.assertEqual(rows["TestEnumNeverDefault"]["outcome"], "absent")
        self.assertEqual(rows["TestAMethodReadAsAValueIsRefused"]["outcome"], "absent")
        self.assertEqual(catchers, ["TestPanics"])

    def test_the_counts_golden_is_a_pin_with_its_pair(self):
        # u045 M08 (test-audit/u045-M08k): only TestCountsAreRecorded failed; it is a golden, never a catch alone.
        rows, catchers = classify(os.path.join(data, "m08-counts.jsonl"))
        self.assertEqual(rows["TestCountsAreRecorded"]["outcome"], "pin")
        self.assertIn("TestStatementRegionsAreUsed", rows["TestCountsAreRecorded"]["why"])
        self.assertEqual(catchers, [])

    def test_load_reds_are_reruns_not_catches(self):
        # The exact lines M06c and M08c recorded; each was green when run alone.
        rows, catchers = classify(log([
            ("TestDecodeASCIIUnit26", "fail", ["    decode_ascii_test.go:267: /tmp/replay-e742af58576a/cache/0350/decode-native [/tmp/cases_053248_055296.bin]: signal: killed"]),
            ("TestPortMatchesGoCohere_123", "fail", ["TestPortMatchesGoCohere_123 exceeded its 90s deadline"]),
            ("TestFileDriver_Setup", "fail", ["panic: test timed out after 1m30s", "\t/workspace/adamic/stage1/cohere/yaml/driver.go:40 +0x1"]),
            ("TestCSSPrinterAgreesWithGo_000", "fail", ["panic: TestCSSPrinterAgreesWithGo_000: own work exceeded 60s"]),
        ]))
        self.assertEqual(rows["TestDecodeASCIIUnit26"]["outcome"], "kill")
        self.assertEqual(rows["TestPortMatchesGoCohere_123"]["outcome"], "deadline")
        self.assertEqual(rows["TestFileDriver_Setup"]["outcome"], "deadline")  # a timeout panic is a deadline, not code
        self.assertEqual(rows["TestCSSPrinterAgreesWithGo_000"]["outcome"], "deadline")
        self.assertTrue(all(row["rerun"] for row in rows.values()))
        self.assertEqual(catchers, [])

    def test_witnesses_harness_panics_and_red_at_base_never_count(self):
        mutant = log([
            ("TestMatcherOct6Mutants", "fail", ["    oct6_test.go:12: planted mutant survived"]),
            ("TestHelperPanics", "fail", ["panic: runtime error: index out of range [3]", "\t/workspace/adamic/internal/lower/agree_test.go:142 +0x170"]),
            ("TestAlreadyRed", "fail", ["    already_test.go:9: wrong"]),
            ("TestRealCatch", "fail", ["    real_test.go:20: got 3, want 4"]),
            ("TestGreen", "pass", []),
        ])
        base = log([("TestAlreadyRed", "fail", ["    already_test.go:9: wrong"]), ("TestRealCatch", "pass", [])])
        rows, catchers = classify(mutant, base)
        self.assertEqual(rows["TestMatcherOct6Mutants"]["outcome"], "witness")
        self.assertEqual(rows["TestHelperPanics"]["outcome"], "panic-harness")
        self.assertEqual(rows["TestAlreadyRed"]["outcome"], "red-at-base")
        self.assertEqual(rows["TestRealCatch"]["outcome"], "fail")
        self.assertEqual(rows["TestRealCatch"]["why"], "real_test.go:20: got 3, want 4")
        self.assertEqual(catchers, ["TestRealCatch"])


if __name__ == "__main__":
    unittest.main()
