Built statements and makeContinue in separate .a files, completing this turn's trio with makeBreak; twelve prerequisites removed across four consumers.
SHAs: current-main landing push 0143e2ae on b8fb957a; replacement claim 98a779ba pushed before source; retained makeBreak implementation f1781baf; final implementation/report SHAs named in final response.
Commands and outputs: replacement pair PASS 148.534s, 28,454 queries; order check PASS 2.274s; vet/types/format PASS; uncached six-fixture oracle PASS 1.150s; setup 35s, nproc 5.
Mutants: thirteen replacement variants caught, including count-preserving reversal; seven retained break variants caught; all 110 earlier variants re-green on current main before claims.
Not covered: complete graph dependencies, rule findings/fixes/integration or full repository test gate; zero final rule blockers removed.

# Delivered helpers and ownership

Each of statements, makeContinue and makeBreak serves array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Four occurrences per helper, twelve across the trio; no final blocker removed. Cumulative slot 05: 41 retained helpers, 259 occurrences, 70 unique consumers, frozen helper readiness 50. This is prerequisite readiness, not implemented rules or diagnostic parity.

Before these replacement claims, all 39 earlier retained helpers were verified on current main b8fb957a and pushed at 0143e2ae. Earlier pushJump/popJump claims were withdrawn because slot 03 had claimed them first, and no duplicate implementation was delivered. The subsequent fresh scan caught an already-claimed reachedStatement before publishing any claim or source. The replacement reservation 98a779ba was pushed before either source file. The final scan of all twenty origin helper branches found all three delivered helpers only on slot 05. Main remained b8fb957aa839a9e8cb0b54279dd9864fa317bd30, an ancestor of HEAD.

# External oracle coverage

Actual Go cohere is pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db. The temporary overlay changes dependency call names in the original methods, preserving their algorithms. Wrappers execute real Go statement, loop, link and unreachable processing. The statement wrapper records only calls at the outer list helper's seam; nested dependency processing still executes the real Go code. The .a callback adapters consume the actual Go observations for graph effects, keeping those externally owned implementations outside these files.

All four consumer Go test files are inspected. Nonempty literal occurrences: array-callback-return 166, consistent-return 272, no-unreachable-loop 71, react-hooks/rules-of-hooks 555. The union, plus explicit controls, is 803 strings, parsed with the actual Go parser into 864 source/loop/switch controls and nine nil/text label controls. Source, configuration and expected text literals are helper inputs, not full-rule diagnostic replay.

statements: 806 nil, empty, nil-item and consumer source-file lists. Callback arena IDs retain exact node identity, repeats and nil (-1). Real Go statements run before expected list-dispatch output is recorded. An order mutant preserves count and nodes while reversing the visit order; its first differing output is line 27, Go 2|stmt:1;stmt:2;, mutant 2|stmt:2;stmt:1;.

makeContinue: 27,648 state/label/target combinations. Depths zero through five; all 256 masks; both reachability states; nil/empty/ASCII/Unicode/unmatched label text; all-nil and alternating nil continue destinations; distinct break versus continue targets; absent/present loop nodes; breakability and initial broken flags. Comparison includes callback ordering, exact destination identity, graph transition observations and the unchanged broken flags through both current and saved jump-list views.

Together with makeBreak's 27,648 comparisons, this trio covers 56,102 distinct invocations. The statement-order check reruns its 806 baseline lists; those repetitions are not counted as new unique inputs. All baseline bytes agree between Go, source Node, emitted JavaScript and ASan/UBSan native. Successful runs and semantic variants must exit zero with empty stderr, including the native leak check.

# Every replacement mutant

| Helper | Mutation | First output line | Check |
| --- | --- | ---: | --- |
| statements | skip first item | 3 | two nil-item calls become one |
| statements | skip final item | 3 | two nil-item calls become one |
| statements | duplicate callback | 3 | two calls become four |
| statements | reverse list | 27 | ordered arena IDs change with count preserved |
| makeContinue | ignore unreachable guard | 1 | unexpected unreachable callback |
| makeContinue | ignore nil continue destination | 4898 | loop/link invoked for nil target |
| makeContinue | ignore exact label membership | 4612 | unmatched label links target |
| makeContinue | omit loop | 4610 | required loop method call absent |
| makeContinue | omit makeUnreachable | 2 | current remains reachable block 0 instead of block 5 false |
| makeContinue | search outermost first | 9218 | destination 1 instead of inner 2 |
| makeContinue | link before loop | 4610 | trace order differs |
| makeContinue | mark break flag | 4610 | broken flag changes despite continue |
| makeContinue | use break destination | 4610 | destination 4 instead of continue destination 1 |

Thirteen exact witnesses are in evidence/mutants.json, helpers.log and order.log. Every variant compiled and ran normally before wrong output was credited. A refusal, panic, warning or sanitizer failure is not a semantic-mutant success. makeBreak's seven witnesses are in ../batch14/REPORT.md. All 110 earlier witnesses are in ../landing-evidence-4/mutants.json.

# Commands and observed outputs

All output went directly to logs, with /workspace/adamic-tools/env.sh sourced:

```
ADAMIC_SLOT05_BATCH15_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch15/evidence" go test ./stage1/cohere/lint/helpers/slot05/batch15 -count=1 -v -timeout=20m > /tmp/lint05-batch15-complete.log 2>&1
ADAMIC_SLOT05_BATCH15_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch15/evidence" go test ./stage1/cohere/lint/helpers/slot05/batch15 -run '^TestSlot05StatementOrder$' -count=1 -v -timeout=20m > /tmp/lint05-batch15-order.log 2>&1
go run ./cmd/adamic types stage1/cohere/lint/helpers/slot05/batch15/main.a > /tmp/lint05-batch15-types.log 2>&1
go vet ./... > /tmp/lint05-batch15-vet-final.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05/batch14 stage1/cohere/lint/helpers/slot05/batch15 > /tmp/lint05-batch15-format-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch15-oracle.log 2>&1
```

The pair run passed in 148.534s (statements 9.11s, makeContinue 139.41s), with twelve mutants. The independently added reversal check passed 2.274s, one more mutant. These elapsed times overlap a superseded earlier run and are not performance claims. Vet and final format logs are empty. Type loading succeeds. The filtered oracle passes all six input fixtures in 1.150s, zero cache hits and six misses. Setup was recorded in ../batch14/evidence/setup.log: Go/clang/Node/submodules ready 0s, cache warm 35s, done 35s on five processors (cpu.max 400000 100000), 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0.

The first private generator attempt had a missing newline; it was fixed. The next run exposed trace contamination from nested makeContinue instrumentation during actual Go statement execution. Dependency traces now respect the outer seam while still executing all nested real Go processing. Those superseded combined runs failed and are not pass evidence. A final formatting check found a trailing blank in the owned batch14 test file; gofmt removed it and the final format log is empty. No executable helper behavior changed in that formatting correction.

# Limits

No shared harness, registration generator, rule, regex matcher, finding model or protected compiler file was changed. Immutable dense parser lists, valid current blocks and contiguous distinct private jump records are assumed. Arbitrary callback mutation of parser lists, aliasing one mutable jump record into several slots, invalid UTF-8/lone surrogate labels and invalid AST/state shapes are outside coverage. No whole-rule findings/fixes/suggestions, full CFG/event parity, dependency implementation or full repository gate claim is made. The initial landing command's extra no-Go-files directory argument and the corrected successful discovery/refusal command are reported in ../LANDING_REPORT_4.md. Upstream MIT attribution stays in the owned ../batch14/NOTICE.md.
