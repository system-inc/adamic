"""test_matrix.py: matrix.py on the cases the audit met (#a5wmhm0).

	python3 -I test_matrix.py
"""
import json, os, subprocess, sys, tempfile, unittest

here = os.path.dirname(os.path.abspath(__file__))


def run(*arguments):
    return subprocess.run([sys.executable, "-I", os.path.join(here, "matrix.py"), *arguments], check=True, capture_output=True, text=True).stdout


def write(value):
    handle = tempfile.NamedTemporaryFile("w", suffix=".json", delete=False)
    json.dump(value, handle)
    handle.close()
    return handle.name


def classified(**outcomes):
    return write({"tests": [{"test": test, "outcome": outcome} for test, outcome in outcomes.items()]})


def mutant(hash, line=93):
    return write({"hash": hash, "file": "internal/native/runtime/adamic.c", "line": line, "op": "flip-condition", "base": "7b9d4272"})


class Matrix(unittest.TestCase):
    def setUp(self):
        self.path = os.path.join(tempfile.mkdtemp(), "7b9d4272.jsonl")

    def rows(self):
        return {json.loads(line)["hash"]: json.loads(line) for line in open(self.path)}

    def test_two_runs_of_one_diff_are_one_mutant(self):
        # The oracle set's b11-024 and b11-033: one diff, first catcher TestCountsAreRecorded (a pin) in one run and
        # TestNativeAgreesWithNode in the other. Merged, it is one row whose catcher is the real test.
        run("add", self.path, mutant("b11d03"), classified(TestCountsAreRecorded="pin", TestNativeAgreesWithNode="absent"))
        run("add", self.path, mutant("b11d03"), classified(TestCountsAreRecorded="pin", TestNativeAgreesWithNode="fail"))
        rows = self.rows()
        self.assertEqual(len(rows), 1)
        self.assertEqual(rows["b11d03"]["catchers"], ["TestNativeAgreesWithNode"])

    def test_absent_and_load_reds_never_erase_a_catch(self):
        run("add", self.path, mutant("m1"), classified(TestA="fail", TestB="pass"))
        run("add", self.path, mutant("m1"), classified(TestA="absent", TestB="kill"))
        run("add", self.path, mutant("m1"), classified(TestA="deadline"))
        run("add", self.path, mutant("m1"), classified(TestC="pass"))
        run("add", self.path, mutant("m1"), classified(TestC="absent"))  # an interrupted run knows nothing of TestC
        row = self.rows()["m1"]
        self.assertEqual(row["catchers"], ["TestA"])
        self.assertEqual(row["outcomes"]["TestC"], "pass")  # the pass is evidence that TestC doesn't catch m1
        self.assertEqual(row["outcomes"]["TestB"], "kill")  # a kill replaces a pass: it is newer news, and not a catch

    def test_a_rerun_alone_settles_a_load_red(self):
        run("add", self.path, mutant("m6"), classified(TestDecodeASCIIUnit26="kill"))
        run("add", self.path, mutant("m6"), classified(TestDecodeASCIIUnit26="pass"))
        self.assertEqual(self.rows()["m6"]["outcomes"]["TestDecodeASCIIUnit26"], "pass")
        self.assertEqual(run("survivors", self.path).split("\t")[0], "m6")

    def test_unique_kills_and_survivors(self):
        run("add", self.path, mutant("m1", 1), classified(TestA="fail", TestB="fail"))
        run("add", self.path, mutant("m2", 2), classified(TestA="panic-in-code", TestB="pass"))
        run("add", self.path, mutant("m3", 3), classified(TestA="pass", TestB="pin"))
        self.assertEqual(json.loads(run("unique", self.path)), {"TestA": ["m2"]})
        self.assertEqual([line.split("\t")[0] for line in run("survivors", self.path).splitlines()], ["m3"])

    def test_a_test_that_loses_its_last_unique_kill_is_named(self):
        older = os.path.join(tempfile.mkdtemp(), "old.jsonl")
        run("add", older, mutant("m2", 2), classified(TestA="fail", TestB="pass", TestC="fail"))
        run("add", older, mutant("m4", 4), classified(TestC="pass", TestD="fail"))
        run("add", self.path, mutant("m2", 2), classified(TestA="fail", TestB="fail", TestC="pass"))  # TestB now covers it
        run("add", self.path, mutant("m4", 4), classified(TestD="absent"))  # TestD didn't run: no verdict on it
        lost = run("lost", older, self.path)
        self.assertNotIn("TestA", lost)  # it had no unique kill (TestC shared m2), so it had nothing to lose
        self.assertNotIn("TestD", lost)  # an absent test is unknown, never lost
        older2 = os.path.join(tempfile.mkdtemp(), "old2.jsonl")
        run("add", older2, mutant("m2", 2), classified(TestA="fail", TestB="pass"))
        self.assertEqual(run("lost", older2, self.path).split("\t")[0], "TestA")


if __name__ == "__main__":
    unittest.main()
