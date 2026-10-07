Rebased: the existing seventeen-helper branch onto origin/main e8ba3d5; helper source unchanged and no new claims.
Commits: original pushed b2ed825, final rebased source head b846322, with landing evidence finalized in this report commit.
Checks: all six helper packages PASS on the latest fetched main; vet/format clean; six-fixture uncached input oracle PASS 1.358s.
Mutants: all forty-one compiling semantic mutants caught by actual Go output comparisons (37 owned, 4 inherited).
Not covered: full repository gate, production rule integration and new helper work; this turn completes the landing unit first.

## Landing cap and branch

This unit has pushed only codex/lint-helpers-05. The local work branch is environment bootstrap state and was not pushed by this unit. The published helper head b2ed825 was not on origin/main, so the user's landing-first instruction made this turn's unit a rebase, oracle rerun and push. No new helpers were reserved.

All 38 commits rebased cleanly onto e011f8f initially. During validation main advanced to e8ba3d5d81de4d3773c723914fccd4c76248b965, bringing compiler call-target and devirtualization changes. The first complete gate passed, but did not cover that newer compiler. The branch was rebased cleanly again, and every retained helper oracle and mutant was rerun on e8ba3d5. Final source head before this documentation commit: b846322e1d3bafda04e5ab09a25a04f00a850a40.

The helper tree before adding this landing evidence is byte-identical to the previously pushed tree. [rebase.json](landing-evidence/rebase.json) and [range-diff.txt](landing-evidence/range-diff.txt) retain the base/head and patch mapping. No helper implementation, shared registration generator, shared rule harness or compiler implementation was manually edited. All seventeen retained helpers remain delivered in .a files. Inherited shared prerequisites retain their existing file extensions.

The original published reservation hashes/timestamps in claims/05.md and earlier ownership evidence remain the basis for claim precedence; rebasing changes commit hashes and committer timestamps, not when those reservations were first published. Historical reports intentionally keep their original validation commits. This report maps the landing history rather than rewriting prior evidence.

## Validation on e8ba3d5

All test output was redirected directly to logs, with no test-output pipelines. Source /workspace/adamic-tools/env.sh first.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/... -p 2 -count=1 -v -timeout=30m > /tmp/lint05-landing-helpers.log 2>&1
go vet ./stage1/cohere/lint/helpers/... > /tmp/lint05-landing-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > /tmp/lint05-landing-oracle.log 2>&1
gofmt -l stage1/cohere/lint/helpers > /tmp/lint05-landing-format.log
```

| Package | Time | Result |
|---|---:|---|
| helpers | 140.285s | PASS |
| batch2 | 106.179s | PASS |
| batch3 | 17.263s | PASS |
| batch4 | 45.477s | PASS |
| batch5 | 70.470s | PASS |
| batch6 | 150.519s | PASS |

Every owned helper's actual pinned Go oracle, source Node comparison, sanitized native run and compiling mutants were executed again. Existing emitted-JavaScript comparisons in batches 4, 5 and 6 also passed. Earlier batches preserve their existing source-Node/native scope; this landing unit adds no new backend coverage claim. The root package also checks the four inherited option/message prerequisite helpers and their mutants. Unsupported input/refusal and metadata-drift checks pass as part of these packages.

Vet exits 0 with empty output; formatting output is empty. The filtered uncached oracle passes in 1.358s with all six fixtures, zero hits and six probe misses. Full repository tests were not run; this is the complete helper subtree plus filtered input oracle gate. The earlier e011f8f gate is retained separately as first-main-helpers.log and first-main-oracle.log and is not substituted for the final gate.

## Every mutant

[mutants.json](landing-evidence/mutants.json) enumerates all 41 executed mutants, their test, log line and actual Go mismatch. [helpers.log](landing-evidence/helpers.log) contains the complete final gate. All credited variants compile, run successfully without stderr/sanitizer failure, and differ from actual Go. Compile errors, crashes and sanitizer findings do not qualify.

| Helper group | Mutations caught | Count |
|---|---|---:|
| Inherited prerequisites | raw JSON control allowed; overlapping oneOf accepted; unknown target field skipped; message value interpolation omitted | 4 |
| First batch | parenthesis unwrapping skipped; namespace binding replaced by its identifier; non-React namespace accepted | 3 |
| Second batch | identifier callee guard inverted; later call arguments dropped; variable pattern-match guard inverted | 3 |
| Third batch | JSX equality case folded; only first filename backslash normalized; non-ASCII hook uppercase rejected | 3 |
| Fourth batch | visitor stop inverted; entry option bits erased; nested namespace variables admitted | 3 |
| Fifth batch | clearAll retains dead count/map/order; compaction at equality, blank/stale slots admitted, dead count retained; deletion fails to blank, ignores absent guard, removes wrong map key, scans forward, omits counter increment, blanks every duplicate | 13 |
| Sixth batch | wrong initial dead count/prefix, shared map/order; candidate presence ignored, fallback disabled, ignored namespace admitted, only first dot replaced, namespace order reversed, missing fallback accepted; key returned as value, empty stored value treated missing | 12 |
| Total | Actual Go comparison kills every credited variant | 41 |

The earlier withdrawn helper experiments are not rebuilt or counted as retained work. Their historical evidence remains labeled withdrawn in the batch reports.

## Setup and push

bash cloud/setup.sh succeeded. Toolchain: Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc printed 5. Timing lines:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (1s)
setup: build cache warm (134s)
setup: done in 134s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

The latest user instruction explicitly requests rebasing and pushing the existing branch. The rebase push uses an exact force-with-lease on the observed old head b2ed8252714f780719eda8c010d1d63a8a1ee1fb, preventing replacement of an intervening remote update. No main push or pull request is made. The final response gives the report commit SHA after remote verification.

Committed log copies trim trailing whitespace; original raw logs remain under /tmp. Only owned landing documentation/evidence and the slot README are changed after the final execution gate, so no extra execution rerun is needed for those documentation changes.

## Scope left open

No new helper or individual rule is claimed this turn. Full rule findings/fixes/suggestions, production AST/theme/linter integration, and the full repository gate remain outside this bounded landing check. Earlier batch reports retain their precise corpus and representation limits. This branch is based on the tested fetched main commit e8ba3d5; future main changes require the next landing check.
