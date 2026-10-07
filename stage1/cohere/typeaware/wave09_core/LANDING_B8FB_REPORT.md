Built: rebased wave-09 onto current main b8fb957a and re-greened all existing rule/component checks.
Commits: old remote tip eae82041f; rebased implementation tip f655f7a68; evidence commit follows on wave-09 only.
Commands/output: original suite, checker/bridge, filtered Node oracle, targeted vet and all seventeen component verifiers PASS.
Mutants: all existing semantic, handle and sanitizer mutants reran, including named listener mutations and the dynamic/static RegExp refusal check.
Not covered: full regex rule parity, legacy visitor migration into shared context or the full gate; no new claims.

Main advanced to b8fb957aa839a9e8cb0b54279dd9864fa317bd30 with the
inherited-static-field read fix. Rebase completed without conflicts, preserving
that change. The compiler was rebuilt before native component tests. A final
origin/main fetch after all validation confirms this same base. Publishing
uses an exact lease on old tip eae82041f5e7fdf6f3c7fc89a7210113edf55cc2,
only to codex/typeaware-wave-09. No shared sources were edited or reverted.

Reproduction uses the commands/manifests in LANDING_C019_REPORT.md with
log/artifact prefix /workspace/wave-09-b8fb. The filtered independent Node
oracle additionally includes inherited_static_field_read.a. All seventeen
Python verifiers listed there were rerun sequentially against the rebuilt
compiler. Named listener declarations use the corrected Go ast.Kind names;
no numeric contract is required. Every test writes a file, never a pipe.
Exact new logs are retained in validation-landing-b8fb.

Package checks:

```
source /workspace/adamic-tools/env.sh
go build -o /workspace/wave-09-core/adamic ./cmd/adamic
go test ./stage1/cohere/typeaware -run '^TestWave09' -count=1 -timeout=30m -v
go test ./bridge/tsgo/checker ./bridge/tsgo -count=1 -timeout=15m -v
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures|inherited_static_field_read)\.a$' -count=1 -timeout=10m -v
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware
```

Original suite passes in 135.466 seconds, checker 0.229 seconds,
bridge 113.756 seconds and filtered Node oracle 4.695 seconds. Vet has empty
output. Original rules match 44 control findings, 14 compiler and 4 repository
findings in normal and sanitized native executions with full fixes/suggestions.
Label controls match 18 findings; both corpora zero. The one parser-refused
label fixture remains explicitly outside supported scope. The full literal
slice and pure helpers match production Go, normal/sanitized native, source
Node and emitted JavaScript, within their previously documented scopes.

Mutants rerun: prefer-const eligibility, radix judgment, type-parameter flag,
assertion fix end, label scope membership and value-mask, flags precedence,
emoji-modifier bounds, surrogate folding, malformed-pattern guard, rune quote,
error-prefix trim, suggestion range, cooked/raw offsets, constructor span,
constant coercion, scoped-write and alias eligibility, named listener kind,
ASCII canonicalization, fold representative, equivalence-pair removal,
class bracket escape and trailing-dash parity. Each builds/runs and is caught
by independent output comparison. JSON and .a named listener mutations are
caught at byte 276. Required panic-70 checks catch released-registry and
checker-retention mutations. Bridge ownership/length/memory mutations are
caught by its sanitizer/refusal checks. The dynamic-RegExp probe remains
refused; its successful sanitized static-input mutant removes that expected
refusal and is caught. Fresh logs retain each catcher and difference.

Original suite whole-process native/Go times: compiler 2.506100/0.596120
seconds, repository 0.339827/0.173532 seconds. Label medians:
compiler 1.552034/0.327141 seconds, repository 0.259189/0.205477 seconds.
These tests overlap other checks and are not isolated throughput results.
The restored workspace reuses its successful toolchain setup: 88 seconds,
nproc 5. No full repository gate was run.

The shared regex branch is now b39979305d880d7d05083704b73d405cd3694d68,
which adds option-dialect comparisons and still reports the dynamic constructor
refusal. No applicable fixed-pattern row exists for these claims. This does
not unblock source-derived RegExp validation. No hand-rolled matcher was
added. Existing custom regex components remain historical isolated evidence,
not completion under the latest constraint. Dynamic new RegExp(pattern,'u')
remains a concrete shared lowering blocker. Legacy checker-backed visitor
migration to the shared handed-node context is unfinished work, not a numeric
schema blocker; that schema mismatch was resolved in KIND_NAMES_REPORT.md.

The named harness ab70f38d4 remains off current main. No encountered allocator
leak-helper change was reverted. React parking does not apply to these regex
claims. Both regex claims remain incomplete and no new batch was claimed.
This landing rebase was the unit of work requested by the work-in-progress cap.
