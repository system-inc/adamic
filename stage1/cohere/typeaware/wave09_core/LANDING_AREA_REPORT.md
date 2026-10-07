Built: rebased wave-09 onto integrated origin/area/stage1-lint and re-greened all owned rule and component checks.
Commits: old remote tip 7cf88c2fc; rebased implementation tip bdf754412; area base 7481e0324 includes current main 39638d9e2; evidence commit follows.
Commands/output: original suite, eighteen component verifiers, registry, bridge/checker, filtered Node oracle and targeted vet PASS.
Mutants: every existing semantic, handle and sanitizer mutant reran, including handed-node refetch and dynamic/static RegExp checks.
Not covered: full regex rule parity, shared installation of this unit's checker-backed rules or the full gate; no new claims.

The user explicitly requested rebase onto origin/area/stage1-lint after the
shared harness landing at 50a5f105. Fetch found that area had advanced to
7481e0324e34a2537aafa9db7eeacda50405611b, including current main
39638d9e278d38bb5aeae887f46d55a70e47aaad. The branch rebased without
conflicts. Shared harness, parser, finding model and allocator-check changes
are retained. No shared files were edited or reverted by this unit.
The compiler was rebuilt before verification. A final main fetch confirms
the same current main and its ancestry. The push destination is exclusively
codex/typeaware-wave-09, with an exact lease on old tip
7cf88c2fc96045a73ebcb1b70dd6c77a19bfa280. No main or area push occurs.

Reproduction uses the manifests and commands recorded in LANDING_B8FB_REPORT.md,
with /workspace/wave-09-area as the log/artifact prefix. All seventeen component
checks there were repeated, plus label_var/verify_handed.py. Registry tests
were added for the integrated shared contract. Tests write files, never pipes.
Exact fresh logs are retained in validation-landing-area.

```
source /workspace/adamic-tools/env.sh
go build -o /workspace/wave-09-core/adamic ./cmd/adamic
go test ./stage1/cohere/typeaware -run '^TestWave09' -count=1 -timeout=30m -v
go test ./bridge/tsgo/checker ./bridge/tsgo -count=1 -timeout=15m -v
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures|inherited_static_field_read)\.a$' -count=1 -timeout=10m -v
go test ./stage1/cohere/lint/registry -count=1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware
```

Original suite PASS 113.538 seconds, bridge PASS 74.350 seconds,
filtered Node oracle PASS 0.329 seconds and registry PASS 0.058 seconds.
Checker passes and vet output is empty. Original rules retain full findings,
fixes and suggestions on 44 control, 14 compiler and 4 repository findings,
normal/sanitized. Label controls retain 18 findings / 7700 bytes; compiler
77 sources zero / 7859 bytes; repository 287 sources zero / 18485 bytes.
The existing one parser-refused control remains recorded outside supported
scope. Full literal-slice corpus comparisons pass normal/sanitized native,
source Node and emitted JavaScript. Pure components retain all those modes.

Mutants rerun: prefer-const eligibility, radix judgment, type-parameter flag,
assertion fix end, label scope membership and value-mask, flags precedence,
emoji-modifier bounds, surrogate folding, malformed-pattern guard, rune quote,
error-prefix trim, suggestion range, cooked/raw offsets, constructor span,
constant coercion, scoped-write and alias eligibility, named listener kind,
ASCII canonicalization, fold representative, equivalence-pair removal,
class bracket escape and trailing-dash parity. All compile/run and fail their
independent output comparisons. Handed-node visits using index zero retain
Go bytes; the clean-running target-refetch mutant differs at byte 43.
Named-kind JSON/.a mutations differ at byte 276. Released-registry and
checker-retention mutants remove required panic 70 and are caught by that
expectation. Bridge length/memory/ownership mutations are caught by sanitizer
or refusal checks. Dynamic RegExp remains refused; the successful sanitized
static-input mutant removes that required refusal. Logs retain each catcher.

Original whole-process native/Go times: compiler 3.799976/0.678732 seconds,
repository 0.415899/0.162877 seconds. Label medians:
compiler 2.002183/0.382157 seconds, repository 0.576896/0.199912 seconds.
Runs overlap other validation, so no isolated throughput claim is made.
The restored workspace reused its successful 88-second toolchain setup,
nproc 5. No full repository gate was run.

The shared harness and handed-node API are now present on this branch after
the area rebase. Historical absent-harness observations no longer describe
the branch. Named descriptor kinds are compatible. no-label-var retains its
handed-node implementation with no per-rule relevance comparison or target
refetch. Its checker question is still certified with an isolated registration
overlay, pending integrator installation. Other legacy visitor migration
remains unfinished work rather than an absent API blocker.

The separate native dynamic RegExp refusal was reproduced on this area base.
No hand-rolled matcher, pattern parser or rewrite was added. Earlier custom
regex components remain historical isolated evidence, not full completion
under the latest constraint. Both regex claims remain incomplete. No React
analysis exception applies, and no new batch is claimed. This rebase and
fresh verification are the landing-first unit explicitly requested.
