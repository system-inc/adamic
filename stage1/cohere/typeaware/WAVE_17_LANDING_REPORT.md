Built: landing unit only, rebasing the existing wave-17 work onto main; no new rules claimed.
Commits: original tip f28732f8, first rebase 53aaf76c on e011f8f6, final rebase 7138b515 on e8ba3d5d; evidence accompanies this report.
Commands and outputs: all eight selected branch oracles PASS 893.495s; checker PASS 0.148s; uncached Node fixtures PASS 3.015s; byte mutant PASS 0.295s; vet clean; setup 19s, nproc 5.
Mutants: all rule, fact, tracker, Unicode, retained-registry and Node byte mutants caught again; every catcher is recorded in results.json and final-oracles.log.
Not covered: full shared gate, production JSX, configurable BooleanPropNaming matching and complete props/component integration; no further claims.

The branch was the only branch pushed by this session. It was not on main and
was 20 commits ahead and 245 behind the first fetched origin/main. The requested
rebase completed without conflicts. All eight selected tests then passed in
943.444s, but a final remote check observed main advancing to e8ba3d5d. That
update changes method call targets, devirtualization and element borrowing, so
its effects could not be dismissed as documentation. The branch was rebased
again, without conflicts, and the complete selected oracle set reran against
that compiler. No rule-source repair was needed. The old-to-rebased commit map
is validation-wave-17-landing/rebase.json. Prior reports retain their historical
commit identities and measurements.

The final base is e8ba3d5d81de4d3773c723914fccd4c76248b965. The final validated
source tip is 7138b515663aa684215ad963e5243f62d730b97d. This landing evidence
commit changes only reports and validation artifacts, not the validated source.
The user's explicit rebase-and-push instruction takes precedence over CLAUDE's
usual no-history-rewrite rule. Push uses a lease against the original published
tip f28732f86eba1c14c1b62f61e0ecabc4d0aa4d56 to avoid overwriting remote updates.
No other branch is pushed, deleted or changed.

| Final oracle | Result | Seconds |
| --- | --- | ---: |
| inherited ten inventory rules | PASS | 162.84 |
| fifth batch, including partial PropTypes scope | PASS | 150.41 |
| fourth batch | PASS | 145.79 |
| Head native decisions and parser boundary | PASS | 90.45 |
| Nexus rules | PASS | 118.52 |
| original async rule and JSX boundary | PASS | 66.40 |
| global rules | PASS | 156.64 |
| Unicode uppercase, all 1,114,112 code points | PASS | 2.44 |

All normal and sanitizer comparisons pass. Each full rule batch compares the
frozen compiler and repository roots byte for byte, including fixes and
suggestions. The inherited inventory test's optional corpus environment was
not supplied in the combined command, so both corpora were separately compared
using the final inventory binaries and frozen manifests. Go, normal native and
sanitized native streams are identical: repository 180 findings, 85151 bytes;
compiler 16589 findings, 7120921 bytes. Complete hashes and output streams are
preserved in inventory-corpora.json and raw/coverage. No new fixture manifests
were substituted for the frozen 77 compiler and 287 repository roots.

The final comparison-only rule and helper mutants all exit zero with empty
sanitizer stderr and differ solely from Go's expected bytes. Their first
mismatching offsets are:

| Mutant | Byte |
| --- | ---: |
| inventory before / cast / methods / coercion / caller | 56 / 1556 / 2203 / 1162 / 10248 |
| inventory parameter / invariant / optional / alias / unused | 368 / 12759 / 15522 / 15998 / 16815 |
| unsupported-syntax / use-memo / partial boolean acceptance | 55 / 9283 / 1203 |
| throw / backreference / arrow fix / RegExp tracker | 17497 / 1235 / 334 / 15013 |
| duplicate Head / Script in Head | 54 / 472 |
| collection / discarded outcome / discarded pure result | 3687 / 16784 / 8422 |
| awaited ancestry fact | 15716 |
| async finding end | 59 |
| global assign / implicit globals / implied eval | 13247 / 52 / 52 |
| source JS flag / shorthand symbol fact | 55157 / 4716 |
| Unicode uppercase boundary | 1 |

Released-handle questions reject with panic 70. Every retained-registry
counterpart finishes cleanly and is caught by the required-panic assertion.
The external Node byte mutant is caught independently. The initial Node run
used main's new cache; the final run explicitly sets ADAMIC_GATE_UNCACHED=1,
reports zero hits and covers functions, closures, generic functions, method
closures, devirtualize and all five call-target fixtures. No cache hit is
represented as a fresh final comparison.

The previous implementation gaps remain explicit. Production JSX is still
refused by the shared parser and is validated through independent raw AST
fixtures for native decisions. Main now supports constant RegExp construction,
but the dynamic-pattern probe is still refused with the updated exact text:

```
adamic: /workspace/wave-17-fifth-regexp-gap.a:3:24: stage 0 can't lower RegExp with a nonconstant pattern yet
```

Thus BooleanPropNaming remains partial: its tested default-pattern PropTypes
checks and message interpolation are not complete configured-pattern parity.
Its full component detector and TypeScript props integration remain unwired.
The inherited shared unknown-question mutation test still expects the former
facts.go refusal line; no shared harness or registration-generator edit was
made to repair that test. The full repository gate was not run or claimed green.
The green result here is the exact eight-oracle selection and listed bridge,
Node and vet checks, rather than every test in the repository.

Quiet three-round alternating whole-process timings for the final fifth-batch
ordinary-parser binaries, after all concurrent checks finished:

| Corpus | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 2.564267s | 0.367275s | 6.98 |
| repository | 0.339624s | 0.127977s | 2.65 |

Every timed round has identical diagnostic bytes. Native is slower. Individual
rounds and phases are in timing/measurements.json. The original async oracle
also records its in-test process timings; those were taken during the larger
validation run and are not advertised as quiet measurements.

Final commands, with output saved to the named logs:

```
git fetch --no-recurse-submodules origin '+refs/heads/*:refs/remotes/origin/*'
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
git rebase origin/main
ADAMIC_WAVE17_ARTIFACTS=/workspace/wave-17-landing/original ADAMIC_WAVE17_HEAD_ARTIFACTS=/workspace/wave-17-landing/heads ADAMIC_WAVE17_NEXT_ARTIFACTS=/workspace/wave-17-landing/next ADAMIC_WAVE17_THIRD_ARTIFACTS=/workspace/wave-17-landing/third ADAMIC_WAVE17_FOURTH_ARTIFACTS=/workspace/wave-17-landing/fourth ADAMIC_WAVE17_FIFTH_ARTIFACTS=/workspace/wave-17-landing/fifth ADAMIC_COVERAGE_ARTIFACTS=/workspace/wave-17-landing/coverage ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-17-typescript TMPDIR=/workspace/wave-17-artifacts go test ./stage1/cohere/typeaware -run '^(TestWave17|TestCoverageAgreementAndMutants)' -count=1 -v -timeout=30m
go vet ./...
go test ./bridge/tsgo/checker -count=1 -v
ADAMIC_GATE_UNCACHED=1 TMPDIR=/workspace/wave-17-artifacts go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(functions|closures|generic_functions|method_closures|devirtualize|call_targets_closure|call_targets_element|call_targets_region|call_targets_reuse|call_targets_sort)[.]a$' -count=1 -v
ADAMIC_GATE_UNCACHED=1 TMPDIR=/workspace/wave-17-artifacts go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v
python3 stage1/cohere/typeaware/validation-wave-17-landing/measure.py /workspace/wave-17-landing/fifth stage1/cohere/typeaware/validation-wave-17-landing/timing /workspace/wave-17-typescript
```

Setup reports Go, clang, Node and submodules ready in 0s each; cache warm and
total 19s on five visible processors with four CPU cgroup quota. Both rebases,
both full selected runs, final uncached Node checks, the current regex refusal,
full comparison streams, exact mutant results and timing evidence are preserved.
