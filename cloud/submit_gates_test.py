"""submit-gates.py's sealed tests: a disposable git repository holding a small tree and the real cloud/gates, job rows
written here, no network, no Loom and no live repository. Git runs with no global or system configuration, so the
fixtures name every identity they commit with.

Run from cloud/: python3 -B -m unittest -v submit_gates_test (cloud/submit_gates_test.go runs it in the gate).
"""

import argparse
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile
import unittest
import unittest.mock

sys.dont_write_bytecode = True
Cloud = Path(__file__).resolve().parent
Script = Cloud / "submit-gates.py"
specification = importlib.util.spec_from_file_location("submit_gates", Script)
gates = importlib.util.module_from_spec(specification)
specification.loader.exec_module(gates)
compareSpecification = importlib.util.spec_from_file_location("units_compare", Cloud / "gates" / "units-compare.py")
unitsCompare = importlib.util.module_from_spec(compareSpecification)
compareSpecification.loader.exec_module(unitsCompare)

# Git sealed from the machine: no global or system configuration (a signing key, a hook path), a home of its own.
sealedGit = {"GIT_CONFIG_GLOBAL": os.devnull, "GIT_CONFIG_NOSYSTEM": "1", "GIT_TERMINAL_PROMPT": "0",
             "GIT_AUTHOR_NAME": "Submit gates test", "GIT_AUTHOR_EMAIL": "submit-gates-test@example.invalid",
             "GIT_COMMITTER_NAME": "Submit gates test", "GIT_COMMITTER_EMAIL": "submit-gates-test@example.invalid"}

lowerSource = """package lower

type lowering struct {
	debug   bool // a comment
	modules map[string]int
	*shared
}

func lowerTarget(options Options) (*Program, error) {
	built, err := lowerChecked(options)
	return built.result, err
}

func lowerChecked(options Options) (*lowering, error) {
	return lowerWithElementKey(options)
}

func lowerWithElementKey(options Options) (*lowering, error) {
	lowering := &lowering{}
	lowering.finishClassCalls()
	if err := lowering.certifySourceSlots(); err != nil {
		return nil, err
	}
	return lowering, nil
}

func (l *lowering) finishClassCalls() {}
"""

irSource = """package ir

type (
	Program struct {
		Main  []Node
		Units int
	}
	Function struct {
		Name string
		Body []Node
		kept bool
	}
)

type Call struct {
	Target string
}
"""

baseTree = {
    "go.mod": "module example\n",
    "internal/lower/lower.go": lowerSource,
    "internal/lower/lower_test.go": "package lower\n",
    "internal/lower/calls.go": "package lower\n\nfunc calls() {}\n",
    "internal/ir/ir.go": irSource,
    "internal/native/emit.go": "package native\n",
    "internal/native_runtime/runtime.go": "package native_runtime\n",
    "internal/buildcache/cache.go": "package buildcache\n",
    "internal/oracle/oracle_test.go": "package oracle\n",
    "stage3/census/latent/pass_ledger.go.txt": 'var latentPostLoweringSteps = []latent.PassStep{\n\t{Name: "finishClassCalls"},\n\t{Name: "certifySourceSlots"},\n}\n',
    "stage3/census/latent/statecopy/ir_fields.go": 'package main\n\nvar irSpecFields = map[string]irFieldPolicy{\n\t"Program.Main": {},\n}\n',
    "stage3/upstream/LEDGER.md": "# Ledger\n",
    "stage3/lane/check.py": "print('lane')\n",
    "README.md": "adamic\n",
}


class SubmitGatesTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.temporary = tempfile.TemporaryDirectory(prefix="submit-gates-tests-")
        cls.root = Path(cls.temporary.name) / "adamic"
        cls.environment = dict(os.environ, HOME=cls.temporary.name, **sealedGit)
        cls.patch = unittest.mock.patch.dict(os.environ, {"HOME": cls.temporary.name, **sealedGit})
        cls.patch.start()
        cls.root.mkdir(parents=True)
        cls.git("init", "-q", "--initial-branch=main")
        cls.write_tree(baseTree)
        shutil.copytree(Cloud / "gates", cls.root / "cloud" / "gates")
        cls.git("add", "-A")
        cls.git("commit", "-qm", "main")
        cls.main = cls.git("rev-parse", "HEAD")

    @classmethod
    def tearDownClass(cls):
        cls.patch.stop()
        cls.temporary.cleanup()

    @classmethod
    def git(cls, *arguments):
        return subprocess.run(["git", "-C", str(cls.root), *arguments], check=True, capture_output=True,
                              env=cls.environment, text=True).stdout.strip()

    @classmethod
    def write_tree(cls, files):
        for path, content in files.items():
            target = cls.root / path
            if content is None:
                target.unlink()
                continue
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_text(content)

    def commit(self, files, message="a head", parent=None):
        """A head of one commit on parent (main's first commit unless named), its sha."""
        self.git("checkout", "-q", "--detach", parent or self.main)
        self.write_tree(files)
        self.git("add", "-A")
        self.git("commit", "-q", "--allow-empty", "-m", message)
        return self.git("rev-parse", "HEAD")

    def recipe_hash(self, gate, at):
        text = self.git("show", f"{at}:cloud/gates/{gate}.sh") + "\n"
        return hashlib.sha256(gates.recipe_command(text).encode()).hexdigest()

    def job(self, gate, head, base=None, exit=0, at=None, job_id=None):
        tail = f"PASS {gate}\nVERDICT {'PASS' if exit == 0 else 'FAIL'} {gate} head {head} base {base or 'none'}\n"
        return {"id": job_id or "J-" + hashlib.sha256((gate + head + str(base) + str(exit)).encode()).hexdigest()[:7],
                "atRepository": "adamic", "atSha": at or head, "commandSha256": self.recipe_hash(gate, self.main),
                "outcome": "Exit", "exit": exit, "tailText": tail, "title": f"{gate} at {head[:8]}", "task": "v3q9ayd",
                "endedAt": "2026-11-23T10:00:00Z"}

    def judge(self, head, rows=(), main=None, live_hole=False):
        jobs = Path(self.temporary.name) / "jobs.json"
        jobs.write_text(json.dumps(list(rows)))
        arguments = argparse.Namespace(head=head, main=main or self.main, repository=str(self.root), jobs=str(jobs),
                                       task="v3q9ayd", live_hole=live_hole, gates=None, json=True)
        return gates.judge(arguments)

    def owed(self, report):
        return {entry["name"] for entry in report["gates"]}

    def rows(self, report):
        return {entry["name"]: entry for entry in report["rows"]}

    def every_green(self, head, report, base=None):
        return [self.job(entry["name"], head, base or self.main) for entry in report["gates"] if not entry.get("byHand")]

    # Each rule's trigger, firing and not.

    def assertFires(self, files, rule, gate=None, message="a head"):
        report = self.judge(self.commit(files, message))
        self.assertIn(rule, report["fired"], report["fired"])
        if gate:
            self.assertIn(gate, self.owed(report))
        return report

    def assertQuiet(self, files, rule, message="a head"):
        report = self.judge(self.commit(files, message))
        self.assertNotIn(rule, report["fired"], report["fired"])
        return report

    def test_every_head_owes_the_guards(self):
        report = self.assertFires({"README.md": "adamic, read\n"}, "every-head")
        self.assertEqual(self.owed(report), {"format", "touched-packages-whole", "compilercensus", "parallelcensus"})

    def test_compiler_change(self):
        self.assertFires({"internal/native/emit.go": "package native\n\nvar emitted = 1\n"}, "compiler-change", "native-whole-wasi")
        self.assertQuiet({"internal/native_runtime/runtime.go": "package native_runtime\n\nvar x = 1\n"}, "compiler-change")
        self.assertQuiet({"internal/lower/lower_test.go": "package lower\n\nvar x = 1\n"}, "compiler-change")

    def test_refusal_added(self):
        self.assertFires({"internal/lower/calls.go": "package lower\n\nfunc calls() error { return &Refused{} }\n"},
                         "refusal-added", "oracle-whole")
        self.assertQuiet({"internal/lower/lower_test.go": "package lower\n\nvar x = &Refused{}\n"}, "refusal-added")

    def test_lower_tests_fire(self):
        self.assertFires({"internal/lower/lower_test.go": "package lower\n\nvar x = 1\n"}, "lower-tests-or-pass", "ir-whole")
        self.assertQuiet({"internal/lower/calls.go": "package lower\n\nfunc calls() { _ = 1 }\n"}, "lower-tests-or-pass")

    def test_pass_list_body_changed_by_text(self):
        added = lowerSource.replace("\tlowering.finishClassCalls()\n", "\tlowering.finishClassCalls()\n\tlowering.checkWorkers()\n")
        report = self.assertFires({"internal/lower/lower.go": added}, "pass-added", "latent-whole")
        self.assertIn("internal/lower/lower.go changed", report["fired"]["pass-added"])
        row = self.rows(report)["pass-ledger-row"]
        self.assertFalse(row["satisfied"])
        self.assertIn("checkWorkers (not named)", row["why"])
        # A comment is a change: the file is read byte for byte
        commented = lowerSource.replace("\treturn lowerWithElementKey(options)\n", "\t// forwards\n\treturn lowerWithElementKey(options)\n")
        self.assertFires({"internal/lower/lower.go": commented}, "pass-added")
        # Any change to the file fires, even to a function the pass lists never name: no span is selected
        elsewhere = lowerSource.replace("func (l *lowering) finishClassCalls() {}", "func (l *lowering) finishClassCalls() { _ = l }")
        self.assertFires({"internal/lower/lower.go": elsewhere}, "pass-added")

    def test_a_declaration_shaped_comment_hides_no_change(self):
        # review-r round 3: a block comment before the real lowerTarget holds a column-zero line shaped like its
        # declaration, so a span selected by a declaration's line is the comment's on both sides and the real body's
        # change is unseen. Both sides compile; only the real body changes.
        base = ("package lower\n\n/*\nfunc lowerTarget() is where lowering starts, read before the passes\n*/\n"
                "func lowerChecked() {}\n\nfunc lowerTarget() (result int) {\n\tresult = 1\n\treturn\n}\n")
        first = self.commit({"internal/lower/lower.go": base})
        head = self.commit({"internal/lower/lower.go": base.replace("\tresult = 1\n", "\tresult = 2\n")}, parent=first)
        report = self.judge(head, main=first)
        for rule in ("pass-added", "lower-tests-or-pass"):
            self.assertIn(rule, report["fired"], report["fired"])
        self.assertIn("internal/lower/lower.go changed", report["fired"]["pass-added"])
        self.assertTrue({"latent-whole", "ir-whole"} <= self.owed(report), self.owed(report))
        # A pass list moved to another file of the package is still read there, by its name in that file's bytes
        moved = self.commit({"internal/lower/entry.go": "package lower\n\n// lowerTarget's passes live here now\nvar entry = 1\n"})
        self.assertIn("pass-added", self.judge(moved)["fired"])
        self.assertIn("internal/lower/entry.go changed and holds the word lowerTarget", self.judge(moved)["fired"]["pass-added"])

    def test_a_unicode_statement_split_fires(self):
        # review-r round 2's pair, both forms gofmt-clean and compiling: one assignment of a nested call's result, against
        # assigning the function and then a separate call. They differ by whitespace alone, a newline and a tab, which
        # Go reads as a semicolon. Both named functions are there, unchanged but for that, so only the compare can fire.
        base = ("package lower\n\nfunc α(callback func()) func() int { return func() int { return 1 } }\n\n"
                "func lowerChecked() {}\n\n"
                "func lowerTarget() (result any) {\n\tf := func() {}\n\tresult = α(f)()\n\treturn\n}\n")
        first = self.commit({"internal/lower/lower.go": base})
        split = base.replace("\tresult = α(f)()\n", "\tresult = α\n\t(f)()\n")
        self.assertEqual(re.sub(r"\s+", "", split), re.sub(r"\s+", "", base))
        head = self.commit({"internal/lower/lower.go": split}, parent=first)
        report = self.judge(head, main=first)
        self.assertIn("pass-added", report["fired"], report["fired"])
        self.assertIn("internal/lower/lower.go changed", report["fired"]["pass-added"])
        self.assertTrue({"latent-whole", "ir-whole"} <= self.owed(report))

    def test_inserted_semicolons_split_a_statement(self):
        # review-r's pair: one return of a call against a bare return and then a call, the same tokens but for the
        # semicolon Go inserts at the line end, one statement against two
        split = lowerSource.replace("\treturn lowerWithElementKey(options)\n", "\treturn\n\tlowerWithElementKey(options)\n")
        report = self.assertFires({"internal/lower/lower.go": split}, "pass-added")
        self.assertTrue({"latent-whole", "ir-whole"} <= self.owed(report), self.owed(report))
        self.assertEqual([token[1] for token in gates.go_tokens("return\nf()\n")], ["return", ";", "f", "(", ")", ";"])
        self.assertEqual([token[1] for token in gates.go_tokens("return f()\n")], ["return", "f", "(", ")", ";"])

    def test_the_pass_ledger_row_is_owed_by_hand(self):
        # Whether a call was added is go/ast's question: the ledger naming every callee-shaped word is no proof (an
        # added line can open or close a comment around unchanged calls), so the row is the accepter's read
        added = lowerSource.replace("\tlowering.finishClassCalls()\n", "\tlowering.finishClassCalls()\n\tcheckWorkers(lowering)\n")
        ledger = baseTree["stage3/census/latent/pass_ledger.go.txt"][:-2] + '\t{Name: "checkWorkers"},\n}\n'
        head = self.commit({"internal/lower/lower.go": added, "stage3/census/latent/pass_ledger.go.txt": ledger})
        row = self.rows(self.judge(head))["pass-ledger-row"]
        self.assertFalse(row["satisfied"], row)
        self.assertIn("owed by hand, the accepter checks", row["why"])
        self.assertIn("checkWorkers (named)", row["why"])
        # A change that adds no call still owes the row: the script can't say no call was added
        changed = lowerSource.replace("return nil, err", "return lowering, err")
        report = self.judge(self.commit({"internal/lower/lower.go": changed}))
        self.assertIn("pass-added", report["fired"])
        self.assertFalse(self.rows(report)["pass-ledger-row"]["satisfied"])

    def test_new_call_site(self):
        self.assertFires({"internal/lower/calls.go": "package lower\n\nvar call = ir.Call{}\n"}, "new-call-site", "ir-whole")
        self.assertQuiet({"internal/native/emit.go": "package native\n\nvar call = ir.Call{}\n"}, "new-call-site")

    def test_lowering_struct_field(self):
        added = lowerSource.replace("\tdebug   bool // a comment\n", "\tdebug   bool // a comment\n\tworkers []string\n")
        self.assertFires({"internal/lower/lower.go": added}, "lowering-struct-field", "latent-whole")
        # whether a change adds a field is go/ast's question, so any change to the directory's files fires
        recommented = lowerSource.replace("// a comment", "// a different comment")
        self.assertFires({"internal/lower/lower.go": recommented}, "lowering-struct-field", "latent-whole")
        self.assertQuiet({"internal/lower/lower_test.go": "package lower\n\nvar y = 1\n"}, "lowering-struct-field")

    def test_ir_field_and_its_statecopy_row(self):
        added = irSource.replace("\t\tUnits int\n", "\t\tUnits int\n\t\tRegions []int\n")
        report = self.assertFires({"internal/ir/ir.go": added}, "ir-field", "latent-whole")
        self.assertFalse(self.rows(report)["statecopy-row"]["satisfied"])
        self.assertIn("Program.Regions (not named)", self.rows(report)["statecopy-row"]["why"])
        # irSpecFields naming every field a token read finds is no proof: the row is the accepter's read
        named = baseTree["stage3/census/latent/statecopy/ir_fields.go"].replace("}\n", '\t"Program.Regions": {},\n}\n')
        head = self.commit({"internal/ir/ir.go": added, "stage3/census/latent/statecopy/ir_fields.go": named})
        row = self.rows(self.judge(head))["statecopy-row"]
        self.assertFalse(row["satisfied"], row)
        self.assertIn("Program.Regions (named)", row["why"])
        self.assertIn("owed by hand, the accepter checks", row["why"])
        # any change to internal/ir fires, an unexported field's included
        unexported = irSource.replace("\t\tkept bool\n", "\t\tkept bool\n\t\tseen bool\n")
        self.assertFires({"internal/ir/ir.go": unexported}, "ir-field", "ir-whole")

    def test_test_skip_owes_its_row_by_hand(self):
        report = self.assertFires({"internal/oracle/oracle_test.go": 'package oracle\n\nfunc TestX(t *testing.T) { t.Skip("why") }\n'},
                                  "test-skip")
        row = self.rows(report)["skip-census-row"]
        self.assertFalse(row["satisfied"])
        self.assertIn("oracle_test.go", row["why"])
        self.assertQuiet({"internal/oracle/oracle_test.go": "package oracle\n\nvar skipped = 1\n"}, "test-skip")

    def test_a_skip_row_is_owed_by_hand(self):
        head = self.commit({"internal/oracle/oracle_test.go": 'package oracle\n\nfunc TestX(t *testing.T) {\n\tt.Skip("why")\n}\n'})
        row = self.rows(self.judge(head))["skip-census-row"]
        self.assertFalse(row["satisfied"], row)
        self.assertIn("owed by hand, the accepter checks", row["why"])
        self.assertIn("internal/oracle/oracle_test.go:TestX", row["why"])

    def test_stopping_pin(self):
        report = self.assertFires({"internal/oracle/testdata/x_new.a": "console.log(1);\n",
                                   "internal/oracle/testdata/x_new.stop": "NotYet\n"}, "stopping-pin")
        self.assertFalse(self.rows(report)["pending-sidecar"]["satisfied"])
        report = self.judge(self.commit({"internal/oracle/testdata/x_new.a": "console.log(1);\n",
                                         "internal/oracle/testdata/x_new.stop": "NotYet\n",
                                         "internal/oracle/testdata/x_new.pending": "awaits #v3q9ayd: the lift\n"}))
        self.assertTrue(self.rows(report)["pending-sidecar"]["satisfied"])
        report = self.judge(self.commit({"internal/oracle/testdata/y_new.ts": "console.log(1);\n"}))
        self.assertFalse(self.rows(report)["pending-sidecar"]["satisfied"])
        # the commit message is no file: saying the program agrees holds nothing
        report = self.judge(self.commit({"internal/oracle/testdata/y_new.ts": "console.log(1);\n"},
                                        message="oracle: y_new.ts agrees with Node"))
        self.assertFalse(self.rows(report)["pending-sidecar"]["satisfied"])
        self.assertIn("y_new.stop reading `lowers`", self.rows(report)["pending-sidecar"]["why"])
        report = self.judge(self.commit({"internal/oracle/testdata/y_new.ts": "console.log(1);\n",
                                         "internal/oracle/testdata/y_new.stop": "lowers\n"}))
        self.assertTrue(self.rows(report)["pending-sidecar"]["satisfied"])
        self.assertQuiet({"internal/oracle/testdata/notes.md": "notes\n"}, "stopping-pin")

    def test_a_stop_by_the_programs_whole_name(self):
        # the dead-zone pins' convention, internal/oracle/dead_zone_stops_test.go: x.a.stop beside x.a
        files = {"internal/oracle/testdata/x_whole.a": "console.log(1);\n",
                 "internal/oracle/testdata/x_whole.a.stop": "NotYet: a reason\n"}
        report = self.judge(self.commit(files))
        self.assertFalse(self.rows(report)["pending-sidecar"]["satisfied"])
        self.assertIn("x_whole.a.stop expects a stop", self.rows(report)["pending-sidecar"]["why"])
        report = self.judge(self.commit(dict(files, **{"internal/oracle/testdata/x_whole.pending": "awaits #v3q9ayd: the lift\n"})))
        self.assertTrue(self.rows(report)["pending-sidecar"]["satisfied"], self.rows(report)["pending-sidecar"])
        report = self.judge(self.commit({"internal/oracle/testdata/x_lowers.a": "console.log(1);\n",
                                         "internal/oracle/testdata/x_lowers.a.stop": "lowers\n"}))
        self.assertTrue(self.rows(report)["pending-sidecar"]["satisfied"])

    def test_sidecars_that_disagree_refuse(self):
        report = self.judge(self.commit({"internal/oracle/testdata/x_both.a": "console.log(1);\n",
                                         "internal/oracle/testdata/x_both.stop": "lowers\n",
                                         "internal/oracle/testdata/x_both.a.stop": "NotYet: a reason\n"}))
        self.assertFalse(self.rows(report)["pending-sidecar"]["satisfied"])
        self.assertIn("disagree", self.rows(report)["pending-sidecar"]["why"])

    def test_a_stop_is_read_from_its_sidecar_alone(self):
        # a stop the commit message declares, with its .pending, is still owed its sidecar
        program = {"internal/oracle/testdata/x_declared.a": "console.log(1);\n",
                   "internal/oracle/testdata/x_declared.pending": "awaits #v3q9ayd: the lift\n"}
        report = self.judge(self.commit(program, message="oracle: x_declared.a stops on purpose"))
        self.assertFalse(self.rows(report)["pending-sidecar"]["satisfied"])
        self.assertIn("x_declared.a has no expected-outcome sidecar", self.rows(report)["pending-sidecar"]["why"])
        report = self.judge(self.commit(dict(program, **{"internal/oracle/testdata/x_declared.stop": "NotYet\n"})))
        self.assertTrue(self.rows(report)["pending-sidecar"]["satisfied"], self.rows(report)["pending-sidecar"])
        # a .pending whose words aren't `awaits #<task>: <reason>` whole names no task
        loose = dict(program, **{"internal/oracle/testdata/x_declared.stop": "NotYet\n",
                                 "internal/oracle/testdata/x_declared.pending": "awaits #v3q9ayd:\n"})
        self.assertFalse(self.rows(self.judge(self.commit(loose)))["pending-sidecar"]["satisfied"])

    def test_harness_and_its_mutant_evidence(self):
        report = self.assertFires({"stage3/lane/check.py": "print('lane, checked')\n"}, "harness", "lane-whole-sealed")
        self.assertFalse(self.rows(report)["mutant-evidence-path"]["satisfied"])
        # a test's name in prose is no declaration: the row is held by the machine form's Red cell
        report = self.judge(self.commit({"stage3/lane/check.py": "print('lane, checked')\n",
                                         "stage3/lane/check.mutants.md": "Undo check.py's line and test_check_rows goes red.\n"}))
        self.assertFalse(self.rows(report)["mutant-evidence-path"]["satisfied"])
        self.assertIn("mutants", report["fired"])  # the .mutants.md fires its own rule
        declared = ("| Mutant | Patch | Red | Run |\n|---|---|---|---|\n"
                    "| The line undone | `stage3/lane/check.patch` | `test_check_rows` | `python3 -m unittest stage3.lane.test_check` |\n")
        report = self.judge(self.commit({"stage3/lane/check.py": "print('lane, checked')\n", "stage3/lane/check.mutants.md": declared}))
        row = self.rows(report)["mutant-evidence-path"]
        self.assertTrue(row["satisfied"], row)
        self.assertIn("stage3/lane/check.mutants.md:1 names test_check_rows", row["why"])
        self.assertQuiet({"stage3/upstream/LEDGER.md": "# Ledger\n\nA line.\n"}, "harness")

    def test_adaptation_owes_its_gates_its_ledger_and_its_value_count(self):
        report = self.assertFires({"stage3/adapt/99-x/adapt.cjs": "module.exports = 1;\n"}, "adaptation", "unit-census-pair")
        self.assertIn("mutants", report["fired"])  # by its rule trigger
        self.assertTrue({"stock-js-identical", "declarations-identical", "mutants-red", "value-count-reached"} <= self.owed(report))
        self.assertFalse(self.rows(report)["ledger-line"]["satisfied"])
        report = self.judge(self.commit({"stage3/adapt/99-x/adapt.cjs": "module.exports = 1;\n",
                                         "stage3/upstream/LEDGER.md": "# Ledger\n\n99-x\n"}))
        self.assertTrue(self.rows(report)["ledger-line"]["satisfied"])

    def test_buildcache_routes_alone(self):
        report = self.assertFires({"internal/buildcache/cache.go": "package buildcache\n\nvar x = 1\n"}, "buildcache")
        self.assertTrue(any("alone" in note for note in report["notes"]))

    # Freshness, hashes and the refusal.

    def test_a_head_with_every_job_is_ready(self):
        head = self.commit({"internal/native/emit.go": "package native\n\nvar ready = 1\n"})
        report = self.judge(head)
        report = self.judge(head, self.every_green(head, report))
        self.assertEqual(report["missing"], [], report)

    def test_fresh_base(self):
        # main moves twice past the head's base: once outside lowering, once inside it
        head = self.commit({"internal/native/emit.go": "package native\n\nvar fresh = 1\n"})
        quiet = self.commit({"README.md": "adamic, moved\n"})
        lowered = self.commit({"internal/lower/calls.go": "package lower\n\nfunc calls() { _ = 2 }\n"}, parent=quiet)
        rows = self.every_green(head, self.judge(head))
        self.assertEqual(self.judge(head, rows, main=quiet)["missing"], [], "a base with no lowering change since is fresh")
        report = self.judge(head, rows, main=lowered)
        self.assertEqual(sorted(report["missing"]), ["admission-delta", "unit-census-pair"])
        self.assertIn("stale", next(e for e in report["gates"] if e["name"] == "unit-census-pair")["why"])
        # a base off main is never fresh
        side = self.commit({"README.md": "a side branch\n"})
        rows = [row if row["title"].split()[0] not in gates.PairedGates else self.job(row["title"].split()[0], head, side)
                for row in rows]
        self.assertEqual(sorted(self.judge(head, rows)["missing"]), ["admission-delta", "unit-census-pair"])

    def test_a_recipe_whose_hash_differs_is_missing(self):
        head = self.commit({"README.md": "adamic, hashed\n"})
        rows = self.every_green(head, self.judge(head))
        rows[0] = dict(rows[0], commandSha256=hashlib.sha256(b"go test ./internal/compilercensus/").hexdigest())
        report = self.judge(head, rows)
        self.assertEqual(report["missing"], [rows[0]["title"].split()[0]])
        entry = next(e for e in report["gates"] if e["name"] == report["missing"][0])
        self.assertIn("command isn't the recipe", entry["why"])

    def test_a_job_at_another_sha_or_red_is_missing(self):
        head = self.commit({"README.md": "adamic, elsewhere\n"})
        rows = self.every_green(head, self.judge(head))
        rows[0] = dict(rows[0], atSha=self.main)
        rows[1] = dict(rows[1], exit=1)
        report = self.judge(head, rows)
        self.assertEqual(sorted(report["missing"]), sorted(row["title"].split()[0] for row in rows[:2]))

    def test_the_newest_verdict_at_the_head_decides(self):
        head = self.commit({"README.md": "adamic, flaked\n"})
        rows = self.every_green(head, self.judge(head))
        green = next(row for row in rows if row["title"].startswith("compilercensus "))
        red = dict(self.job("compilercensus", head, exit=1, job_id="J-redlast"), endedAt="2026-11-23T11:00:00Z")
        report = self.judge(head, rows + [red])
        self.assertEqual(report["missing"], ["compilercensus"])
        self.assertIn("J-redlast", next(e for e in report["gates"] if e["name"] == "compilercensus")["why"])
        # red first, then green: the green is the newest verdict
        earlier = dict(red, endedAt="2026-11-23T09:00:00Z")
        report = self.judge(head, rows + [earlier])
        self.assertEqual(report["missing"], [])
        self.assertEqual(next(e for e in report["gates"] if e["name"] == "compilercensus")["job"], green["id"])
        # a run that ended with no verdict of its own is passed over
        for outcome in ("Infrastructure", "Cancelled", "Superseded", None):
            with self.subTest(outcome=outcome):
                passed = dict(red, id="J-noverdc", outcome=outcome, exit=None)
                self.assertEqual(self.judge(head, rows + [passed])["missing"], [])
        # a paired gate's newest verdict, red, refuses it though an older green is fresh
        native = self.commit({"internal/native/emit.go": "package native\n\nvar paired = 1\n"})
        rows = self.every_green(native, self.judge(native))
        paired = dict(self.job("unit-census-pair", native, self.main, exit=1, job_id="J-pairred"), endedAt="2026-11-23T11:00:00Z")
        self.assertEqual(self.judge(native, rows + [paired])["missing"], ["unit-census-pair"])

    def test_ends_read_as_instants_and_a_tie_refuses(self):
        head = self.commit({"README.md": "adamic, at the same time\n"})
        rows = self.every_green(head, self.judge(head))
        green = next(row for row in rows if row["title"].startswith("compilercensus "))
        # a red a tenth of a second later, its end spelled with more precision than the green's
        red = dict(self.job("compilercensus", head, exit=1, job_id="J-tenthrd"), endedAt="2026-11-23T10:00:00.100Z")
        self.assertEqual(self.judge(head, rows + [red])["missing"], ["compilercensus"])
        tied = dict(red, endedAt=green["endedAt"])
        self.assertEqual(self.judge(head, rows + [tied])["missing"], ["compilercensus"])

    def test_an_end_with_no_zone_is_a_missing_proof(self):
        head = self.commit({"README.md": "adamic, zoneless\n"})
        rows = [dict(row, endedAt="2026-11-23T10:00:00") if row["title"].startswith("compilercensus ") else row
                for row in self.every_green(head, self.judge(head))]
        report = self.judge(head, rows)
        self.assertEqual(report["missing"], ["compilercensus"])
        self.assertIn("no end that reads", next(e for e in report["gates"] if e["name"] == "compilercensus")["why"])

    def paired(self, tail, head=None, main=None):
        """The report for a compiler head whose unit-census-pair job ends on this tail."""
        head = head or self.commit({"internal/native/emit.go": "package native\n\nvar verdict = 1\n"})
        rows = [dict(row, tailText=tail(head)) if row["title"].startswith("unit-census-pair ") else row
                for row in self.every_green(head, self.judge(head, main=main), base=self.git("rev-parse", head + "^"))]
        return self.judge(head, rows, main=main)

    def test_a_paired_verdict_names_the_jobs_head(self):
        self.assertEqual(self.paired(lambda head: f"VERDICT PASS unit-census-pair head {head} base {self.main}\n")["missing"], [])
        report = self.paired(lambda head: f"VERDICT PASS unit-census-pair head {'e' * 40} base {self.main}\n")
        self.assertEqual(report["missing"], ["unit-census-pair"])
        self.assertIn("names the head eeeeeeee", next(e for e in report["gates"] if e["name"] == "unit-census-pair")["why"])

    def test_a_paired_verdict_measured_the_heads_parent(self):
        # main moves outside lowering and the head is cut on it: the old main is fresh, but it isn't what HEAD^ measured
        moved = self.commit({"README.md": "adamic, before the head\n"})
        head = self.commit({"internal/native/emit.go": "package native\n\nvar parent = 1\n"}, parent=moved)
        report = self.paired(lambda head: f"VERDICT PASS unit-census-pair head {head} base {self.main}\n", head=head, main=moved)
        self.assertEqual(report["missing"], ["unit-census-pair"])
        self.assertIn("not the head's parent", next(e for e in report["gates"] if e["name"] == "unit-census-pair")["why"])

    def test_a_paired_verdict_is_the_tails_final_line(self):
        report = self.paired(lambda head: f"VERDICT PASS unit-census-pair head {head} base {self.main}\n"
                                          "Traceback (most recent call last):\n  a later failure\n")
        self.assertEqual(report["missing"], ["unit-census-pair"])
        self.assertIn("doesn't end on a passing verdict line", next(e for e in report["gates"] if e["name"] == "unit-census-pair")["why"])

    def test_a_head_is_one_change(self):
        # Adamic's rule (14:33Z): a head is one change's diff against main's tip
        first = self.commit({"README.md": "one\n"})
        head = self.commit({"README.md": "two\n"}, parent=first)
        self.assertTrue(self.rows(self.judge(head))["one-head"]["satisfied"], "two commits are one change")
        # a merge of main into the head is still one head
        moved = self.commit({"cloud/notes.md": "main moved\n"})
        self.git("checkout", "-q", "--detach", head)
        self.git("merge", "-q", "--no-ff", "-m", "main into the head", moved)
        merged = self.git("rev-parse", "HEAD")
        self.assertTrue(self.rows(self.judge(merged, main=moved))["one-head"]["satisfied"])
        # two heads merged together are two, and the report says so
        other = self.commit({"cloud/other.md": "another head\n"})
        self.git("checkout", "-q", "--detach", head)
        self.git("merge", "-q", "--no-ff", "-m", "two heads", other)
        joined = self.git("rev-parse", "HEAD")
        row = self.rows(self.judge(joined))["one-head"]
        self.assertFalse(row["satisfied"], row)
        self.assertIn("merges 2 heads main doesn't hold", row["why"])
        self.assertIn("more than one head", row["why"])

    # A mutant in the machine form cloud/gates/mutants-red.sh runs: an exact patch, its one test, and its run.
    mutantsFile = "stage3/adapt/99-y/adapt.mutants.md"
    swapRow = "| The pair swapped | `stage3/adapt/99-y/swap.patch` | `TestAdapted` | `go test -count=1 -run '^TestAdapted$' ./stage3/adapt/` |\n"
    dropRow = "| The entry dropped | `stage3/adapt/99-y/drop.patch` | `TestAdapted` | `go test -count=1 -run '^TestAdapted$' ./stage3/adapt/` |\n"
    oneMutant = "# Mutants\n\n| Mutant | Patch | Red | Run |\n|---|---|---|---|\n" + swapRow
    twoOnOneTest = oneMutant + dropRow

    def by_hand_head(self, mutants=None, adapter="module.exports = 2;\n", extra=None):
        """An adaptation head, its *.mutants.md as given (one machine-form mutant unless named), and its green recipe jobs."""
        files = {"stage3/adapt/99-y/adapt.cjs": adapter, "stage3/upstream/LEDGER.md": "# Ledger\n\n99-y\n",
                 self.mutantsFile: self.oneMutant if mutants is None else mutants, **(extra or {})}
        head = self.commit(files)
        return head, self.every_green(head, self.judge(head))

    def hand_job(self, head, job_id, title, exit=0, tail=""):
        return {"id": job_id, "atRepository": "adamic", "atSha": head, "commandSha256": "0" * 64, "outcome": "Exit",
                "exit": exit, "tailText": tail, "title": title, "task": "v3q9ayd", "endedAt": "2026-11-23T10:00:00Z"}

    def red_line(self, head, row=1, patch="stage3/adapt/99-y/swap.patch", test="TestAdapted",
                 run="go test -count=1 -run '^TestAdapted$' ./stage3/adapt/"):
        return (f"VERDICT RED mutants-red {self.mutantsFile}:{row} patch {patch} test {test} "
                f"run {hashlib.sha256(run.encode()).hexdigest()} head {head}")

    def mutant_job(self, head, job_id, row=1, exit=1, tail=None):
        """A job of the mutants-red recipe for the declared mutant at row, its command the recipe with the mutant's name
        written in, ending as the recipe ends when the mutant's test is red, unless the tail is given."""
        command = self.git("show", f"{self.main}:cloud/gates/mutants-red.sh").replace("@MUTANT@", f"{self.mutantsFile}:{row}")
        tail = tail if tail is not None else f"--- FAIL: TestAdapted (0.01s)\nFAIL\n{self.red_line(head, row)}\n"
        return dict(self.hand_job(head, job_id, f"mutants-red {self.mutantsFile}:{row}", exit, tail),
                    commandSha256=hashlib.sha256(command.encode()).hexdigest())

    def gate(self, report, name):
        return next(entry for entry in report["gates"] if entry["name"] == name)

    def test_by_hand_gates_say_what_the_accepter_checks(self):
        # Advisory (Adamic, 14:33Z): no attestation counts; a gate no recipe proves is printed with what to check
        head, rows = self.by_hand_head()
        report = self.judge(head, rows + [self.mutant_job(head, "J-mutant1")])
        for gate in ("stock-js-identical", "value-count-reached"):
            with self.subTest(gate=gate):
                entry = self.gate(report, gate)
                self.assertFalse(entry["satisfied"])
                self.assertTrue(entry["why"].startswith("owed by hand, the accepter checks: "), entry["why"])
        self.assertTrue(self.gate(report, "mutants-red")["satisfied"], self.gate(report, "mutants-red"))
        # identity or value, the map's thenIf: neither side found, so the identity and the whole branch are listed
        self.assertEqual([choice["by"] for choice in report["choices"]], [None])
        self.assertEqual(sorted(report["missing"]), ["adapter-output-test", "stock-js-identical", "value-count-reached"])

    def test_a_labelled_green_is_no_evidence(self):
        head, rows = self.by_hand_head()
        labelled = self.hand_job(head, "J-identic", f"stock-js-identical at {head[:8]}: emit byte for byte",
                                 tail="VERDICT PASS stock-js-identical\n")
        entry = self.gate(self.judge(head, rows + [labelled]), "stock-js-identical")
        self.assertFalse(entry["satisfied"])
        self.assertIn("owed by hand", entry["why"])

    def test_a_mutant_is_red_only_by_its_own_line(self):
        # The map's red: exit non-zero with the named test's FAIL line. The recipe prints its RED line only then, so a
        # job is evidence by that line, ending its tail, naming this mutant, its patch, its test, its run and the head
        head, rows = self.by_hand_head()
        judged = lambda job: self.gate(self.judge(head, rows + [job]), "mutants-red")
        self.assertTrue(judged(self.mutant_job(head, "J-mutant1"))["satisfied"])
        # a setup that failed at the right head, exit 1, no test run
        setup = self.mutant_job(head, "J-mutant1", tail="FAIL patch: stage3/adapt/99-y/swap.patch doesn't apply to the head\n"
                                                        f"VERDICT FAIL mutants-red head {head} base none failed: patch\n")
        self.assertFalse(judged(setup)["satisfied"])
        self.assertIn("doesn't end on the line that names TestAdapted red", judged(setup)["why"])
        # another test failed, the named one not
        other = self.mutant_job(head, "J-mutant1", tail="--- FAIL: TestOther (0.01s)\nFAIL\n"
                                                        f"VERDICT FAIL mutants-red head {head} base none failed: red\n")
        self.assertFalse(judged(other)["satisfied"])
        renamed = self.mutant_job(head, "J-mutant1", tail=self.red_line(head, test="TestOther") + "\n")
        self.assertFalse(judged(renamed)["satisfied"])
        # the right line and exit, but another command, or a green, or another sha
        self.assertIn("no job ran the command", judged(dict(self.mutant_job(head, "J-mutant1"), commandSha256="0" * 64))["why"])
        self.assertIn("not red", judged(self.mutant_job(head, "J-mutant1", exit=0))["why"])
        self.assertFalse(judged(dict(self.mutant_job(head, "J-mutant1"), atSha=self.main))["satisfied"])

    def test_a_mutant_not_in_the_machine_form_is_owed(self):
        head, rows = self.by_hand_head(mutants="Undo the table and TestAdapted goes red.\n")
        entry = self.gate(self.judge(head, rows + [self.mutant_job(head, "J-mutant1")]), "mutants-red")
        self.assertFalse(entry["satisfied"])
        self.assertIn("no recipe can run its mutant", entry["why"])

    def test_two_mutants_on_one_test_need_two_jobs(self):
        head, rows = self.by_hand_head(mutants=self.twoOnOneTest)
        first = self.mutant_job(head, "J-mutant1")
        second = self.mutant_job(head, "J-mutant2", row=2,
                                 tail="--- FAIL: TestAdapted (0.02s)\n" + self.red_line(head, 2, "stage3/adapt/99-y/drop.patch") + "\n")
        entry = self.gate(self.judge(head, rows + [first]), "mutants-red")
        self.assertFalse(entry["satisfied"])
        self.assertIn(f"{self.mutantsFile}:2: no job ran the command", entry["why"])
        report = self.judge(head, rows + [first, second])
        self.assertTrue(self.gate(report, "mutants-red")["satisfied"], self.gate(report, "mutants-red"))
        # each mutant's ask writes its own name into the recipe
        asks = self.gate(self.judge(head, rows), "mutants-red")["ask"]
        self.assertEqual(len(asks), 2)
        self.assertIn(f"show {self.main}:cloud/gates/mutants-red.sh | sed 's|@MUTANT@|{self.mutantsFile}:2|'", asks[1])

    def test_an_adapter_pins_nodes_never_a_whole_file(self):
        hashing = "const crypto = require('crypto');\nfunction hash(text) { return crypto.createHash('sha256').update(text).digest('hex'); }\n"
        whole = ("const crypto = require('crypto');\nconst fs = require('fs');\n"
                 "const sha = value => crypto.createHash('sha256').update(value).digest('hex');\n"
                 "assert.equal(sha(fs.readFileSync(path, 'utf8')), pin.fileSha256);\n")
        head, rows = self.by_hand_head(adapter=whole)
        row = self.rows(self.judge(head))["adapter-structural-pins"]
        self.assertFalse(row["satisfied"])
        self.assertIn("a whole file", row["why"])
        # a SourceFile's getText is a whole file's: review-r round 3's receiver
        source_file = (hashing + "const sf = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);\n"
                       "assert.equal(hash(sf.getText()), pin.sha256);\n")
        head, rows = self.by_hand_head(adapter=source_file)
        row = self.rows(self.judge(head))["adapter-structural-pins"]
        self.assertFalse(row["satisfied"], row)
        self.assertIn("hashes 'sf.getText()', a whole file", row["why"])
        # a node's getText is never evidence: whether it is a node is its receiver's type, the accepter's read
        for adapter in (hashing + "assert.equal(hash(owner.getText(file)), pin.ownerBeforeSha256);\n",
                        hashing + "assert.equal(hash(pins[state]), pin.ownerBeforeSha256);\n",
                        "const host = require('../47-host-errors/adapt.cjs');\nhost.pin(file);\n"):
            with self.subTest(adapter=adapter.splitlines()[-1]):
                head, rows = self.by_hand_head(adapter=adapter)
                row = self.rows(self.judge(head))["adapter-structural-pins"]
                self.assertFalse(row["satisfied"], row)
                self.assertIn("owed by hand, the accepter checks", row["why"])
        # an adapter that makes no digest, reads no getText( and requires no code of its own shows no pin
        head, rows = self.by_hand_head(adapter="const owners = require('./owners.json');\nmodule.exports = owners;\n")
        self.assertTrue(self.rows(self.judge(head))["adapter-structural-pins"]["satisfied"])

    def test_a_javascript_changing_adapter_carries_an_output_test(self):
        head, rows = self.by_hand_head()
        row = self.rows(self.judge(head))["adapter-output-test"]
        self.assertFalse(row["satisfied"])
        self.assertIn("no test.cjs", row["why"])
        head, rows = self.by_hand_head(extra={"stage3/adapt/99-y/test.cjs": "require('assert');\n"})
        row = self.rows(self.judge(head))["adapter-output-test"]
        self.assertFalse(row["satisfied"])
        self.assertIn("owed by hand, the accepter checks", row["why"])

    def test_live_hole_submits_on_its_gates(self):
        head = self.commit({"internal/lower/calls.go": "package lower\n\nfunc calls() error { return &Refused{} }\n"})
        rows = [self.job(gate, head) for gate in gates.LiveHoleGates]
        report = self.judge(head, rows, live_hole=True)
        self.assertEqual(report["missing"], [], report)
        self.assertIn("oracle-whole", report["owedToday"])
        self.assertTrue(self.judge(head, rows[1:], live_hole=True)["missing"])

    def test_a_missing_gate_is_reported_missing(self):
        # Adamic's one mutant (14:33Z): a gate whose job is deleted is reported as missing, by name, with its ask
        head = self.commit({"internal/native/emit.go": "package native\n\nvar mutant = 1\n"})
        rows = self.every_green(head, self.judge(head))
        jobs = Path(self.temporary.name) / "mutant-jobs.json"

        def run(kept):
            jobs.write_text(json.dumps(kept))
            return subprocess.run([sys.executable, "-B", str(Script), "--head", head, "--main", self.main, "--repository",
                                   str(self.root), "--jobs", str(jobs), "--task", "v3q9ayd"],
                                  capture_output=True, text=True, env=self.environment)

        whole = run(rows)
        self.assertEqual(whole.returncode, 0, whole.stdout + whole.stderr)
        self.assertIn("verdict: evidence found for every owed gate and row", whole.stdout)
        self.assertIn(gates.Advisory, whole.stdout)
        deleted = next(row for row in rows if row["title"].startswith("native-whole-wasi "))
        missing = run([row for row in rows if row is not deleted])
        self.assertEqual(missing.returncode, 1, missing.stdout + missing.stderr)
        self.assertIn("verdict: 1 owed with no evidence found, each for the accepter to check: native-whole-wasi",
                      missing.stdout)
        self.assertRegex(missing.stdout, r"\n  unproven +native-whole-wasi ")
        self.assertIn(gates.Advisory, missing.stdout)
        self.assertIn(f'ahra loom job "$(git -C {self.root} show {self.main}:cloud/gates/native-whole-wasi.sh)" --kind proof '
                      f'--task v3q9ayd --title "native-whole-wasi at {head[:8]}" --at adamic@{head} --size F64', missing.stdout)

    def test_several_heads_are_each_reported(self):
        first = self.commit({"README.md": "a set, one\n"})
        second = self.commit({"README.md": "a set, two\n"})
        jobs = Path(self.temporary.name) / "set.json"
        jobs.write_text("[]")
        result = subprocess.run([sys.executable, "-B", str(Script), "--head", first, "--head", second, "--main", self.main,
                                 "--repository", str(self.root), "--jobs", str(jobs), "--task", "v3q9ayd"],
                                capture_output=True, text=True, env=self.environment)
        self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
        self.assertIn("2 heads given, each reported on its own", result.stdout)
        self.assertIn(f"submit-gates: head {first[:8]}", result.stdout)
        self.assertIn(f"submit-gates: head {second[:8]}", result.stdout)

    def test_a_main_before_the_map_exits_two(self):
        bare = self.commit({"cloud/gates/diff-to-gates.json": None})
        head = self.commit({"README.md": "adamic, on a bare main\n"}, parent=bare)
        jobs = Path(self.temporary.name) / "empty.json"
        jobs.write_text("[]")
        result = subprocess.run([sys.executable, "-B", str(Script), "--head", head, "--main", bare, "--repository",
                                 str(self.root), "--jobs", str(jobs), "--task", "v3q9ayd"],
                                capture_output=True, text=True, env=self.environment)
        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        self.assertIn("--gates", result.stderr)

    def test_the_judge_reads_main_never_the_head(self):
        # Seats' call for round 5: a head that rewrites a recipe is judged by main's recipe, and owes gates-changed
        forged = "source /opt/adamic-block/environment.sh && echo 'VERDICT PASS format head x base none'\n"
        head = self.commit({"cloud/gates/format.sh": forged})
        rows = self.every_green(head, self.judge(head))
        own = dict(next(row for row in rows if row["title"].startswith("format ")), id="J-forged1",
                   commandSha256=hashlib.sha256(gates.recipe_command(forged).encode()).hexdigest())
        report = self.judge(head, [row for row in rows if not row["title"].startswith("format ")] + [own])
        self.assertFalse(self.gate(report, "format")["satisfied"], self.gate(report, "format"))
        self.assertIn("format", report["missing"])
        # the rows asked with main's recipe hold the gates, and the judging code changed is owed by hand
        report = self.judge(head, rows)
        self.assertTrue(self.gate(report, "format")["satisfied"])
        changed = self.rows(report)["gates-changed"]
        self.assertFalse(changed["satisfied"], changed)
        self.assertIn("cloud/gates/format.sh", changed["why"])
        self.assertIn("owed by hand, the accepter checks", changed["why"])
        self.assertEqual(report["missing"], ["gates-changed"], report)
        # a head that changes no judging code owes nothing for it
        self.assertTrue(self.rows(self.judge(self.commit({"README.md": "adamic, unjudged\n"})))["gates-changed"]["satisfied"])

    def test_the_census_tooling_is_judging_code(self):
        # Seats' call for round 6: a head that changes what the census measures owes gates-changed like the judge's own
        for path in ("stage3/census/latent/make_overlay.py", "stage3/apply.py", "stage3/apply.sh", "stage3/caches.py",
                     "stage3/meter/entry_overlay.py", "stage3/api/package-lock.json"):
            with self.subTest(path=path):
                changed = self.rows(self.judge(self.commit({path: "# changed by the head\n"})))["gates-changed"]
                self.assertFalse(changed["satisfied"], changed)
                self.assertIn(path, changed["why"])
        # every tree path the census recipe names is judging code, but the npm project it installs in, whose package
        # files are named, and the node_modules it links
        named = set(re.findall(r"(?:cloud|stage3)/[A-Za-z0-9_./-]*[A-Za-z0-9_]", (Cloud / "gates" / "unit-census-pair.sh").read_text()))
        self.assertEqual(sorted(path for path in named - {"stage3/api", "stage3/api/node_modules"}
                                if not path.startswith(gates.JudgePaths)), [])

    def test_the_script_runs_as_main_holds_it(self):
        other = self.commit({"cloud/submit-gates.py": "print('another judge')\n"})
        head = self.commit({"README.md": "adamic, judged\n"}, parent=other)
        with self.assertRaises(gates.Unreadable) as refused:
            self.judge(head, main=other)
        self.assertIn("isn't main's", str(refused.exception))
        same = self.commit({"cloud/submit-gates.py": Script.read_text()})
        head = self.commit({"README.md": "adamic, judged by main's own\n"}, parent=same)
        report = self.judge(head, main=same)
        self.assertFalse(any(note.startswith("bootstrap:") for note in report["notes"]), report["notes"])
        # a main with no script is the bootstrap, and says so
        self.assertTrue(any(note.startswith("bootstrap:") for note in self.judge(self.commit({"README.md": "x\n"}))["notes"]))


class GateTreeTests(unittest.TestCase):
    """The map and recipes as committed in cloud/gates."""

    def setUp(self):
        self.map = gates.read_map((Cloud / "gates" / "diff-to-gates.json").read_text())

    def test_every_gate_has_a_recipe_or_is_owed_by_hand(self):
        recipes = {path.stem for path in (Cloud / "gates").glob("*.sh")} - {"steps"}
        self.assertEqual(recipes | set(gates.ByHandGates), set(self.map["gates"]))
        self.assertFalse(recipes & set(gates.ByHandGates))
        self.assertEqual((Cloud / "gates" / "mutants-red.sh").read_text().count(gates.MutantPlaceholder), 1)
        self.assertTrue(set(gates.PairedGates) <= recipes and set(gates.LiveHoleGates) <= recipes)
        # gates-changed is the judge's own row, owed by hand beside the map's
        self.assertEqual(set(gates.RowReaders) | set(gates.ByHandRows) - {"gates-changed"}, set(self.map["sameDiff"]))
        self.assertNotIn("gates-changed", self.map["sameDiff"])
        self.assertTrue((Cloud / "gates" / "README.md").exists())

    def test_every_recipe_opens_sealed_parses_and_ends_on_its_verdict(self):
        for path in sorted((Cloud / "gates").glob("*.sh")):
            if path.stem == "steps":
                continue
            with self.subTest(recipe=path.name):
                text = path.read_text()
                self.assertTrue(text.startswith("source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh "
                                                f"&& source cloud/gates/steps.sh && gate_begin {path.stem} || exit 1\n"))
                self.assertTrue(text.rstrip("\n").endswith("gate_end"))
                gates.ask_flags(text)
                parsed = subprocess.run(["bash", "-n", str(path)], capture_output=True, text=True)
                self.assertEqual(parsed.returncode, 0, parsed.stderr)
                self.assertNotIn("\u2014", text)
                self.assertNotIn("\u2013", text)

    def test_the_verdict_line_reads_as_the_script_reads_it(self):
        line = "VERDICT PASS unit-census-pair head " + "a" * 40 + " base " + "b" * 40
        self.assertEqual(gates.final_verdict("PASS a step\n" + line + "\n"), ("PASS", "unit-census-pair", "a" * 40, "b" * 40))
        self.assertIsNone(gates.final_verdict(line + "\nmore output\n"))
        self.assertIsNone(gates.final_verdict(line + " and more"))
        self.assertIn("VERDICT PASS %s head %s base %s", (Cloud / "gates" / "steps.sh").read_text())

    def test_a_changed_definition_is_refused(self):
        changed = json.loads((Cloud / "gates" / "diff-to-gates.json").read_text())
        changed["definitions"]["compilerProduction"] += ", internal/flow2/"
        with self.assertRaises(gates.Unreadable):
            gates.read_map(json.dumps(changed))
        changed = json.loads((Cloud / "gates" / "diff-to-gates.json").read_text())
        changed["rules"][0]["triggers"].append({"goInterfaceChanged": "internal/ir"})
        with self.assertRaises(gates.Unreadable):
            gates.read_map(json.dumps(changed))


class UnitsCompareTests(unittest.TestCase):
    def census(self, tree, stops):
        units = [{"where": f"{tree}/src/a.ts:{line}:1", "name": name, "kind": "function"}
                 for line, name in enumerate(("first", "second", "third"), 1)]
        findings = [{"unit": f"{tree}/src/a.ts:{line}:1", "phase": phase, "kind": "NotYet",
                     "where": f"{tree}/src/a.ts:{line}:5", "reason": "a reason"} for line, phase in stops]
        return [{"file": f"{tree}/src/a.ts", "units": units, "findings": findings}]

    def units(self, tree, names, stops):
        """A census over src/a.ts holding the named units, each name's stop phase in stops when it stops."""
        units = [{"where": f"{tree}/src/a.ts:{line}:1", "name": name, "kind": "function"} for line, name in enumerate(names, 1)]
        findings = [{"unit": f"{tree}/src/a.ts:{line}:1", "phase": stops[name], "kind": "NotYet",
                     "where": f"{tree}/src/a.ts:{line}:5", "reason": "a reason"}
                    for line, name in enumerate(names, 1) if name in stops]
        return [{"file": f"{tree}/src/a.ts", "units": units, "findings": findings}]

    def test_a_head_only_unit_that_stops_is_newly_stopping(self):
        result = unitsCompare.compare(self.units("/base", ["kept"], {}), self.units("/head", ["kept", "added"], {"added": "lowering"}),
                                      "/base", "/head")
        self.assertEqual([unit[1] for unit in result["newly"]], ["added"])
        self.assertEqual(result["head_only_quiet"], [])

    def test_a_head_only_unit_with_no_stop_fails_nothing(self):
        result = unitsCompare.compare(self.units("/base", ["kept"], {}), self.units("/head", ["kept", "added"], {}),
                                      "/base", "/head")
        self.assertEqual(result["newly"], [])
        self.assertEqual([unit[1] for unit in result["head_only_quiet"]], ["added"])

    def test_a_base_only_unit_fails_nothing(self):
        result = unitsCompare.compare(self.units("/base", ["kept", "gone"], {"gone": "lowering"}), self.units("/head", ["kept"], {}),
                                      "/base", "/head")
        self.assertEqual(result["newly"], [])
        self.assertEqual([unit[1] for unit in result["base_only"]], ["gone"])

    def test_newly_stopping_is_read_by_unit_and_phase(self):
        base = self.census("/base", [(1, "lowering"), (2, "refusal_scan")])
        head = self.census("/head", [(2, "deferred"), (3, "prepass")])
        result = unitsCompare.compare(base, head, "/base", "/head")
        self.assertEqual([unit[1] for unit in result["newly"]], ["second", "third"])
        self.assertEqual([unit[1] for unit in result["cleared"]], ["first"])

    def test_a_newly_stopping_unit_is_allowed_by_its_line(self):
        unit, unique = ("src/a.ts", "third", "function", 1), {("src/a.ts", "third")}
        self.assertIsNone(unitsCompare.allowed(unit, "adapt: a change\n\nnothing named", unique))
        self.assertEqual(unitsCompare.allowed(unit, "x\n\n  allowed: src/a.ts third #abc1234\n", unique),
                         "allowed: src/a.ts third #abc1234")
        self.assertIsNone(unitsCompare.allowed(unit, "allowed: src/a.ts thirdly #abc1234", unique))
        self.assertIsNone(unitsCompare.allowed(unit, "allowed: src/a.ts third #abc1234, though not really", unique))

    def test_a_bare_mention_allows_nothing(self):
        unit, unique = ("src/a.ts", "third", "function", 1), {("src/a.ts", "third")}
        self.assertIsNone(unitsCompare.allowed(unit, "adapt: src/a.ts third now stops\n", unique))
        self.assertIsNone(unitsCompare.allowed(unit, "src/a.ts third stops, ruled on #abc1234\n", unique))
        self.assertIsNone(unitsCompare.allowed(unit, "allowed: src/a.ts third stops\n", unique))

    def test_a_denial_allows_nothing(self):
        unit = ("src/a.ts", "added", "function", 1)
        self.assertIsNone(unitsCompare.allowed(unit, "src/a.ts added stops and is not allowed by #abc1234\n",
                                               {("src/a.ts", "added")}))

    def test_a_ruling_covers_a_name_only_when_it_is_one_unit_on_each_side(self):
        # Seats' default for census instance identity (#v3q9ayd round 4): a line names a file and a name, never an
        # instance, so it covers a unit only when that name is one unit in its file at the base and at the head
        ruling = "allowed: src/a.ts f #abc1234\n"
        result = unitsCompare.compare(self.units("/base", ["f", "g"], {}), self.units("/head", ["f", "f", "g"], {"f": "lowering"}),
                                      "/base", "/head")
        self.assertEqual([unit[1] for unit in result["newly"]], ["f", "f"])
        for unit in result["newly"]:
            self.assertIsNone(unitsCompare.allowed(unit, ruling, result["unique"]), unit)
        result = unitsCompare.compare(self.units("/base", ["g"], {}), self.units("/head", ["g", "f"], {"f": "lowering"}),
                                      "/base", "/head")
        self.assertIsNone(unitsCompare.allowed(result["newly"][0], ruling, result["unique"]), "a name only the head holds")
        result = unitsCompare.compare(self.units("/base", ["f", "g"], {}), self.units("/head", ["f", "g"], {"f": "lowering"}),
                                      "/base", "/head")
        self.assertEqual(unitsCompare.allowed(result["newly"][0], ruling, result["unique"]), ruling.strip())

    def test_a_name_held_a_different_number_of_times_fails_closed(self):
        # review-r round 2: two base functions named f, the first stopping, against one head f, stopping
        base = self.units("/base", ["f", "f"], {})
        base[0]["findings"] = [{"unit": "/base/src/a.ts:1:1", "phase": "lowering", "kind": "NotYet",
                                "where": "/base/src/a.ts:1:5", "reason": "a reason"}]
        result = unitsCompare.compare(base, self.units("/head", ["f"], {"f": "lowering"}), "/base", "/head")
        self.assertEqual(result["newly"], [("src/a.ts", "f", "function", 1)])
        self.assertEqual(result["ambiguous"], [("src/a.ts", "f", "function")])


if __name__ == "__main__":
    unittest.main()
