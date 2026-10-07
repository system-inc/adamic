Built: rebased wave-09 onto current main c01907a7 and re-greened every owned rule/component check.
Commits: old pushed tip c58b587b0; rebased implementation tip 3fcc87e6a; evidence commit follows on wave-09 only.
Commands/output: original rule suite, bridge/checker, filtered Node oracle, targeted vet and all seventeen owned component verifiers PASS.
Mutants: all previously recorded semantic, handle and bridge sanitizer mutants reran; the final dynamic/static constructor check still proves the shared refusal.
Not covered: complete regex rule parity, dynamic native RegExp, numeric handed-node dispatch or the full repository gate; no new claims.

Origin/main advanced to c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06.
The branch rebased without conflicts and rebuilt its compiler before native
component reruns. A final fetch after validation confirms that same base.
Publishing uses an exact lease on old remote tip
c58b587b056de71ffab3b5dc19fc1c8cab652587 and only this unit's branch name.
There are no changes to shared compiler, parser, harness or generator files.
No encountered leak-helper edit was reverted. No implementation change was
made beyond updating this unit's claim status and fresh evidence.

The commands and manifests are those recorded in LANDING_F801_REPORT.md,
with log/artifact prefix /workspace/wave-09-c019 instead of wave-09-f801.
The original suite uses the pinned 77 compiler and 287 repository manifests.
All tests write files, never a pipe. The compiler was rebuilt using
go build -o /workspace/wave-09-core/adamic ./cmd/adamic. The package checks:

```
go test ./stage1/cohere/typeaware -run '^TestWave09' -count=1 -timeout=30m -v
go test ./bridge/tsgo/checker ./bridge/tsgo -count=1 -timeout=15m -v
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -timeout=10m -v
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware
```

Each Python verifier was run after sourcing /workspace/adamic-tools/env.sh.
The fourteen verifiers listed in LANDING_F801_REPORT.md were repeated, plus
invalid_regexp/verify_equivalence.py, invalid_regexp/verify_class.py and
invalid_regexp/verify_dynamic_gap.py. All seventeen passed. Exact logs are
retained in validation-landing-c019, including all successful mutant outputs
and catchers. Setup remains the previous successful 88-second run, nproc 5.

Original rules match 44 control, 14 compiler and 4 repository findings,
normal and sanitized, including complete fixes/suggestions. Their suite took
113.471 seconds; checker took 0.673 seconds; bridge 121.580 seconds;
filtered Node oracle 1.783 seconds. Vet passed with no output. Label controls
match 18 findings and corpora zero; the one parser-refused control remains
explicitly outside supported scope. Full literal-slice corpus comparisons
pass in native, sanitized native, source Node and emitted JavaScript.
All pure component comparisons retain those same execution modes.

Original suite whole-process native/Go times: compiler 2.627072/0.588320
seconds, repository 0.373187/0.169904 seconds. Label medians:
compiler 1.852198/0.344223 seconds, repository 0.510755/0.271742 seconds.
Tests overlapped; these are not isolated throughput measurements.

Mutants: prefer-const eligibility (byte 52), radix judgment (9897),
type-parameter flag (12746), assertion fix end (12158), label scope membership
and value-mask, flags precedence, emoji-modifier bounds, surrogate folding,
malformed-pattern guard, rune quoting, error-prefix trim, suggestion range,
cooked/raw offsets, constructor span, constant coercion, scoped-write and
alias-chain eligibility, listener kind, ASCII canonicalization and fold
representative, equivalence pair removal, class bracket escaping and trailing
dash parity. They compile/run and are caught by independent comparisons.
Released-registry and checker-retention mutants remove the required panic 70
and are caught by that expectation. Bridge ownership/length/memory mutations
are caught by its sanitizer/refusal oracles. The dynamic constructor probe
remains refused; its successful sanitized static-input mutant is caught by
the required-refusal assertion. Detailed differences are retained in the logs.

## Updated shared regex evidence

origin/codex/lint-regex now exists at
071fb012848ce0408428c61aba0857cca472236f. Its README, gaps.md and table.json
were inspected read-only. The table has 107 compile sites, including 82 fixed
patterns and 25 dynamic sites; none names no_invalid_regexp,
no_misleading_character_class or no_label_var. This unit therefore has no
fixed Go regex row to adopt. Source-derived esregexp compilation and class
syntax inspection are not fixed regexp.MustCompile patterns.

The shared branch itself proves the same "RegExp with a nonconstant pattern"
native refusal and records raw option dialect differences. This supersedes
REGEX_POLICY_REPORT.md's historical absent-branch observation. No hand-rolled
matcher fallback was added and the historical custom regex components do not
constitute completion under the latest user constraint. Current shared
parser still exposes string kinds and no numeric handed-node listener API.
Six numeric rule.json declarations remain ready; legacy execution remains
outside that requested speed contract. No Diagnostic migration SHA was named.

These are not React rules, so the React parking exception does not apply.
Both regex claims remain incomplete; no new batch is claimed. Existing
supported-scope label completion is retained. All work is pushed solely to
codex/typeaware-wave-09, never main or an area branch.
