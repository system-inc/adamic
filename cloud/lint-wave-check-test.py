"""Disposable-Git CLI tests. Fake Go events test the evidence reader, not parity."""

import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


SCRIPT = Path(__file__).with_name("lint-wave-check.py")
LINT = "stage1/cohere/lint/"
PACKAGE = "github.com/system-inc/adamic/stage1/cohere/lint"
BRANCH = "codex/wave-example"
CLAIM = LINT + "claims/" + BRANCH + ".json"
LABEL = "wave diagnostic changed"


def events():
    rows = []
    def event(action, test=None, output=None):
        row = {"Package": PACKAGE, "Action": action}
        if test:
            row["Test"] = test
        if output is not None:
            row["Output"] = output
        rows.append(row)
    for test in ("TestRulesAgree", "TestOwnedWitnesses", "TestCompilerAndStage1Agree"):
        event("run", test)
        if test == "TestRulesAgree":
            event("output", test, "cohere cases: 218 unique source/rule/options combinations\n")
        event("output", test, "Go, Node, native identical: 100 bytes\n")
        event("pass", test)
    event("run", "TestMutants")
    subject = "TestMutants/" + LABEL.replace(" ", "_")
    event("run", subject)
    for side in ("Node", "native"):
        event("output", subject, LABEL + " caught on " + side + ": different diagnostics\n")
    event("pass", subject)
    event("pass", "TestMutants")
    event("run", "TestRegistrationMutant")
    event("pass", "TestRegistrationMutant")
    event("pass")
    return rows


class WaveCheckTests(unittest.TestCase):
    def git(self, root, *args):
        return subprocess.check_output(["git", *args], cwd=root, stderr=subprocess.STDOUT, text=True).strip()

    def write(self, root, path, content):
        target = root / path
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(content)

    def commit(self, root):
        self.git(root, "add", "-A")
        self.git(root, "commit", "-qm", "Test input")

    def claim(self, branch=BRANCH):
        return {"version": 1, "branch": branch, "rules": ["wave-example"]}

    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix="lint-wave-tests-")
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.origin = self.root / "origin.git"
        self.repo = self.root / "worker"
        self.git(self.root, "init", "--bare", "-q", "--initial-branch=main", str(self.origin))
        self.git(self.root, "clone", "-q", str(self.origin), str(self.repo))
        self.identity(self.repo)
        self.write(self.repo, LINT + "rules/base-rule/rule.json", '{"name":"base-rule","order":1}\n')
        self.write(self.repo, LINT + "lint.ts", "// shared dispatcher\n")
        self.write(self.repo, ".gitignore", "/stage1/cohere/lint/.generated/\n")
        self.commit(self.repo)
        self.git(self.repo, "push", "-q", "origin", "main")
        self.git(self.repo, "checkout", "-qb", BRANCH)
        self.write(self.repo, CLAIM, json.dumps(self.claim()))
        self.write(self.repo, LINT + "rules/wave-example/rule.json", '{"name":"wave-example"}\n')
        self.write(self.repo, LINT + "rules/wave-example/rule.ts", "// independent listener\n")
        self.write(self.repo, LINT + "rules/wave-example/mutant.json", json.dumps({"name": LABEL}))
        self.commit(self.repo)
        self.event_file = self.root / "events.jsonl"
        self.save_events(events())
        self.bin = self.root / "bin"
        self.bin.mkdir()
        fake_go = self.bin / "go"
        fake_go.write_text("#!/usr/bin/env python3\nimport os,sys\nfrom pathlib import Path\n"
                           "if sys.argv[1] == 'run': print('wave-example')\n"
                           "else: print(Path(os.environ['FAKE_EVENTS']).read_text(),end=''); sys.exit(int(os.environ.get('FAKE_EXIT','0')))\n")
        fake_go.chmod(0o755)
        self.environment = dict(os.environ, PATH=str(self.bin) + os.pathsep + os.environ["PATH"],
                                FAKE_EVENTS=str(self.event_file), ADAMIC_TYPESCRIPT_SOURCE=str(self.root))

    def identity(self, root):
        self.git(root, "config", "user.name", "Wave test")
        self.git(root, "config", "user.email", "wave-test@example.invalid")

    def save_events(self, rows):
        self.event_file.write_text("".join(json.dumps(row) + "\n" for row in rows))

    def run_check(self, *args):
        return subprocess.run([sys.executable, "-B", str(SCRIPT), *args], cwd=self.repo,
                              env=self.environment, text=True, capture_output=True)

    def accepted(self, *args):
        result = self.run_check(*args)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        for message in ("PASS claim:", "PASS origin:", "PASS registration:",
                        "PASS parity and mutants: current commit"):
            self.assertIn(message, result.stdout)
        if "--verify" not in args:
            self.assertIn("Running parity and mutants; log:", result.stdout)
        self.assertTrue(self.receipt().is_file(), "validation must write a receipt")
        record = json.loads(self.receipt().read_text())
        self.assertEqual(record["head"], self.git(self.repo, "rev-parse", "HEAD"))
        self.assertEqual(record["claim"], self.claim())
        self.assertEqual(record["mutants"], {"wave-example": LABEL})
        log = self.receipt().parents[1] / record["log"]
        self.assertTrue(log.is_file())
        self.assertEqual(record["sha256"], hashlib.sha256(log.read_bytes()).hexdigest())
        rows = [json.loads(line) for line in log.read_text().splitlines()]
        self.assertEqual(rows, events())
        return result

    def rejected(self, category, message, *args):
        result = self.run_check(*args)
        self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
        self.assertIn("FAIL " + category + ": " + message, result.stderr)
        print("mutant caught:", result.stderr.strip().splitlines()[0])
        return result

    def foreign(self, branch, files):
        root = self.root / ("other-" + branch.replace("/", "-"))
        self.git(self.root, "clone", "-q", str(self.origin), str(root))
        self.identity(root)
        self.git(root, "checkout", "-qb", branch)
        for path, content in files.items():
            self.write(root, path, content)
        self.commit(root)
        self.git(root, "push", "-q", "origin", branch)

    def receipt(self):
        return self.repo / ".git/lint-wave-check" / (BRANCH + ".json")

    def test_positive_run_and_commit_bound_verify(self):
        self.accepted()
        self.accepted("--verify")

    def test_missing_claim(self):
        self.git(self.repo, "rm", CLAIM)
        self.commit(self.repo)
        self.rejected("claim", f"missing committed {CLAIM}")

    def test_claim_owner(self):
        self.write(self.repo, CLAIM, json.dumps(self.claim("codex/somebody-else")))
        self.commit(self.repo)
        self.rejected("claim", f"{CLAIM}: path must mirror branch")

    def test_duplicate_claimed_rule(self):
        claim = self.claim()
        claim["rules"] *= 2
        self.write(self.repo, CLAIM, json.dumps(claim))
        self.commit(self.repo)
        self.rejected("claim", f"{CLAIM}: duplicate claimed rules")

    def markdown_reservation(self):
        path = LINT + "claims/wave-test.md"
        self.git(self.repo, "rm", CLAIM)
        self.write(self.repo, path, "# Wave claim\n\nBranch: `" + BRANCH + "`. Base: origin/main.\n\n"
                   "- 35: wave-example. Claimed here.\n"
                   "- 36: already-done. Skipped, already ported.\n"
                   "\n## Unit report\n- unrelated-name: a finding from a corpus, not an assignment.\n")
        self.commit(self.repo)
        return path

    def test_markdown_claim_discovered_with_skips_and_appended_report(self):
        self.markdown_reservation()
        self.accepted()

    def test_explicit_markdown_claim_and_owned_evidence(self):
        path = self.markdown_reservation()
        self.write(self.repo, LINT + "claims/wave-test-evidence/overlay.json", '{"Replace":{}}')
        self.write(self.repo, LINT + "claims/wave-test-REPORT.md", "Test evidence")
        self.commit(self.repo)
        self.accepted("--claim", path)

    def test_markdown_wrong_owner(self):
        path = self.markdown_reservation()
        self.write(self.repo, path, "Branch: codex/foreign.\n1. wave-example: claimed.\n")
        self.commit(self.repo)
        self.rejected("claim", f"{path} belongs to codex/foreign", "--claim", path)

    def test_markdown_duplicate_assignment(self):
        path = self.markdown_reservation()
        self.write(self.repo, path, "Owner: " + BRANCH + ".\n1. wave-example\n2. wave-example\n")
        self.commit(self.repo)
        self.rejected("claim", f"{path}: duplicate or invalid assignment wave-example", "--claim", path)

    def test_markdown_not_skipped_is_still_reserved(self):
        path = self.markdown_reservation()
        self.write(self.repo, path, "Branch: " + BRANCH + ".\n1. wave-example: claimed, not skipped.\n")
        self.commit(self.repo)
        self.accepted()

    def test_foreign_markdown_unreadable_assignments_fail_closed(self):
        self.foreign("codex/unreadable", {LINT + "claims/unreadable.md": "Branch: codex/unreadable.\nUnstructured assignment: wave-example\n"})
        self.rejected("claim", LINT + "claims/unreadable.md: no readable assignment list before the report section")

    def test_competing_claim_fetched_despite_main_only_config(self):
        self.git(self.repo, "config", "--replace-all", "remote.origin.fetch", "+refs/heads/main:refs/remotes/origin/main")
        path = LINT + "claims/codex/other.json"
        self.foreign("codex/other", {path: json.dumps(self.claim("codex/other"))})
        self.rejected("origin", f"wave-example claimed in origin/codex/other:{path}")

    def test_competing_directory_port_with_different_slug(self):
        self.foreign("codex/ported", {LINT + "rules/different-slug/rule.json": '{"name":"wave-example"}'})
        self.rejected("origin", "wave-example ported in origin/codex/ported:" + LINT + "rules/different-slug/rule.json")

    def test_competing_legacy_port(self):
        self.foreign("codex/legacy", {LINT + "lint.ts": "if(this.enabled('wave-example')) { run(); }\n"})
        self.rejected("origin", "wave-example referenced by legacy port origin/codex/legacy:" + LINT + "lint.ts")

    def test_competing_a_port(self):
        self.foreign("codex/a-port", {LINT + "older-rule.a": "context.report(index, 'wave-example', 'id');\n"})
        self.rejected("origin", "wave-example referenced by legacy port origin/codex/a-port:" + LINT + "older-rule.a")

    def test_competing_markdown_claim_without_implementation(self):
        self.foreign("codex/wave-other", {LINT + "claims/wave-other.md": "Branch: codex/wave-other.\n1. wave-example: claimed here.\n"})
        self.rejected("origin", "wave-example claimed in origin/codex/wave-other:" + LINT + "claims/wave-other.md")

    def test_foreign_claim_evidence_json_is_not_a_claim(self):
        self.foreign("codex/evidence", {LINT + "claims/wave-evidence/overlay.json": '{"Replace":{}}'})
        self.accepted()

    def test_competing_legacy_selection_without_implementation(self):
        self.foreign("codex/selected", {LINT + "batch9-selection.json": '{"selection":[{"name":"wave-example"}]}'})
        self.rejected("origin", "wave-example reserved by legacy selection origin/codex/selected:" + LINT + "batch9-selection.json")

    def test_competing_legacy_batch_reservation(self):
        self.foreign("codex/batch", {LINT + "BATCH9.md": "Selected `wave-example`; implementation pending.\n"})
        self.rejected("origin", "wave-example mentioned by legacy batch reservation origin/codex/batch:" + LINT + "BATCH9.md")

    def test_fixture_mentions_are_not_ports(self):
        self.foreign("codex/corpus", {LINT + "testdata/fixture.ts": "const name = 'wave-example';\n"})
        self.accepted()

    def test_own_published_branch_is_allowed(self):
        self.git(self.repo, "push", "-q", "origin", BRANCH)
        self.accepted()

    def test_pruned_claim_does_not_linger(self):
        path = LINT + "claims/codex/deleted.json"
        self.foreign("codex/deleted", {path: json.dumps(self.claim("codex/deleted"))})
        self.rejected("origin", f"wave-example claimed in origin/codex/deleted:{path}")
        self.git(self.origin, "update-ref", "-d", "refs/heads/codex/deleted")
        self.accepted()

    def test_origin_fetch_failure(self):
        self.git(self.repo, "remote", "set-url", "origin", str(self.root / "absent.git"))
        self.rejected("git", f"fatal: '{self.root / 'absent.git'}' does not appear to be a git repository")

    def test_shared_registration_lines(self):
        for path in (LINT + "lint.ts", LINT + "testdata/oracle.go", LINT + "lint_test.go", "cmd/lint-registry/main.go"):
            with self.subTest(path=path):
                self.write(self.repo, path, "// worker appends to a shared file\n")
                self.commit(self.repo)
                self.rejected("registration", f"shared or unclaimed lint change: {path}")
                self.git(self.repo, "reset", "--hard", "HEAD^")

    def test_unclaimed_rule_directory(self):
        self.write(self.repo, LINT + "rules/another/rule.json", '{"name":"another"}')
        self.commit(self.repo)
        self.rejected("registration", "shared or unclaimed lint change: " + LINT + "rules/another/rule.json")

    def test_shared_file_renamed_into_owned_directory(self):
        self.git(self.repo, "mv", LINT + "lint.ts", LINT + "rules/wave-example/stolen.ts")
        self.commit(self.repo)
        self.rejected("registration", "shared or unclaimed lint change: " + LINT + "lint.ts")

    def test_committed_generated_registry(self):
        self.write(self.repo, LINT + ".generated/registry.ts", "// generated")
        self.git(self.repo, "add", "-f", LINT + ".generated/registry.ts")
        self.commit(self.repo)
        self.rejected("registration", "shared or unclaimed lint change: " + LINT + ".generated/registry.ts")

    def test_shared_ordinal(self):
        self.write(self.repo, LINT + "rules/wave-example/rule.json", '{"name":"wave-example","order":2}')
        self.commit(self.repo)
        self.rejected("registration", "new rule wave-example must omit order")

    def test_missing_descriptor(self):
        self.git(self.repo, "rm", LINT + "rules/wave-example/rule.json")
        self.commit(self.repo)
        self.rejected("registration", "claimed rule wave-example has no directory descriptor")

    def test_missing_owned_mutant(self):
        self.git(self.repo, "rm", LINT + "rules/wave-example/mutant.json")
        self.commit(self.repo)
        self.rejected("mutants", "wave-example has no owned mutant.json")

    def test_missing_parity(self):
        self.save_events([row for row in events() if row.get("Test") != "TestRulesAgree"])
        self.rejected("parity", "TestRulesAgree did not run and pass")

    def test_empty_parity_corpus(self):
        rows = events()
        for row in rows:
            if "cohere cases:" in row.get("Output", ""):
                row["Output"] = "cohere cases: 0 unique\n"
        self.save_events(rows)
        self.rejected("parity", "upstream corpus is empty or unreported")

    def test_absent_three_way_parity(self):
        rows = events()
        for row in rows:
            if row.get("Test") == "TestOwnedWitnesses" and "Output" in row:
                row["Output"] = "Node only identical: 100 bytes\n"
        self.save_events(rows)
        self.rejected("parity", "TestOwnedWitnesses: missing nonempty three-way comparison")

    def test_skipped_compiler_parity(self):
        rows = events()
        for row in rows:
            if row.get("Test") == "TestCompilerAndStage1Agree" and row["Action"] == "pass":
                row["Action"] = "skip"
        self.save_events(rows)
        self.rejected("evidence", "test TestCompilerAndStage1Agree skip")

    def test_missing_mutant_suite(self):
        self.save_events([row for row in events() if not row.get("Test", "").startswith("TestMutants")])
        self.rejected("mutants", "TestMutants did not run and pass")

    def test_owned_mutant_not_run(self):
        self.save_events([row for row in events() if not row.get("Test", "").startswith("TestMutants/")])
        self.rejected("mutants", "wave-example: owned mutant did not run and pass")

    def test_mutant_not_caught_on_native(self):
        self.save_events([row for row in events() if "caught on native" not in row.get("Output", "")])
        self.rejected("mutants", "wave-example: no successful execution and caught comparison on native")

    def test_command_failure_even_with_pass_events(self):
        self.environment["FAKE_EXIT"] = "1"
        self.rejected("evidence", "parity/mutant command failed; read ")

    def test_no_receipt(self):
        self.rejected("evidence", "no receipt; run this script without --verify first", "--verify")

    def test_stale_commit_receipt(self):
        self.accepted()
        self.write(self.repo, "worker-notes.md", "new commit")
        self.commit(self.repo)
        self.rejected("evidence", "receipt is stale for this commit, claim or mutant set", "--verify")

    def test_changed_log(self):
        self.accepted()
        record = json.loads(self.receipt().read_text())
        log = self.receipt().parents[1] / record["log"]
        log.write_text(log.read_text() + "{}\n")
        self.rejected("evidence", "test log missing or changed", "--verify")

    def test_dirty_tree(self):
        self.write(self.repo, LINT + "rules/wave-example/rule.ts", "// changed since tests")
        result = self.rejected("evidence", "commit changes and remove untracked files first:")
        self.assertIn("M " + LINT + "rules/wave-example/rule.ts", result.stderr)

    def test_uncommitted_go_overlay(self):
        self.environment["GOFLAGS"] = "-overlay=/tmp/unapplied-compatibility.json"
        self.rejected("evidence", "remove GOFLAGS -overlay; an unapplied compatibility patch is not the committed candidate")

    def test_origin_changes_after_receipt(self):
        self.accepted()
        self.foreign("codex/late-claim", {LINT + "claims/codex/late-claim.json": json.dumps(self.claim("codex/late-claim"))})
        self.rejected("origin", "wave-example claimed in origin/codex/late-claim:" + LINT + "claims/codex/late-claim.json", "--verify")


if __name__ == "__main__":
    unittest.main(verbosity=2)
