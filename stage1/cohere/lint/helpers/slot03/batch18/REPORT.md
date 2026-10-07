Built binaryExpression, classLike and ifStatement in three separate .a files; fifty-four owned helpers total.
Commits: claim 5ec90184 pushed before implementation (rebased as 1ac93d1b), implementation a54f84ee on lint area b4691483 containing main c7991b90; final publication SHA is named in the final response.
Commands: rebased helper gate PASS owned 896.257s/shared 93.953s, 3,046,326 comparisons; shared harness PASS 315.787s; vet/format clean; six uncached probes PASS 2.266s; setup 151s, nproc 5.
Mutants: 203 owned variants (twenty-one new), four inherited variants and the missing-consumer check caught; clean native execution and stdout differences catch semantic variants, the empty-stack exit check catches its guard variant.
Not covered: full rule findings, independent parser/backend integration, arbitrary malformed AST views and the full repository gate, including the seventeen external correctness comparisons.

# Scope, claim and readiness

All fifty-one earlier helpers were complete, oracle-green and published at 22cd74b7 on unchanged fetched main c7991b90 and lint area b84a9d93 before any new reservation. Rebase onto the current lint area reported already up to date. All twenty origin codex/lint-helpers* branches and nineteen claim files were fetched and read. These three helpers tie the highest eligible remaining fan-out at four consumers each. Comments remain reserved by the shared comment bundle. No regex is ported by this unit. Claim 5ec90184 was pushed before source. A later twenty-branch check finds no duplicate claim for these symbols.

This unit owns slot03/batch18, batch18_test.go and claims/03.md. Shared harness, registry, compiler and runtime files are unchanged. Named parser classifications and fields supply the input views; integer node/block identities are arena handles, not AST kinds. There are no new rules or rule dispatch changes.

Each helper removes a dependency for every rule below:

- array-callback-return
- consistent-return
- no-unreachable-loop
- react-hooks/rules-of-hooks

Twelve dependency edges across four unique rules, zero final helper blockers removed by this batch alone in the frozen inventory. readiness.json subtracts only this batch and lists the residual blockers. These are helper-readiness statements, not claims of complete rule ports or findings parity.

# Actual observations and dependency boundary

The oracle uses cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. A temporary capture overlay records all four consumers' actual upstream fixture strings, including dynamically assembled cases. Upstream core and React packages pass, and 2,119 deduplicated source rows are captured. Every expected consumer is checked against the frozen readiness ledger; drift is an error.

Temporary Go overlays rename the original methods and call their unchanged bodies through observation wrappers. Expressions, allocation, edge storage, entry, pattern reads/writes/binding, destructuring classification, decorators, type parameters, member headers, condition branching and statements run their real Go implementations. Nested dependency work is suppressed from the outer helper's direct-call trace. Records preserve requested operation, three arguments, callback result and before/after cursor. The native/source/emitted driver replays dependency results and cursor changes, generates calls independently, and compares the entire trace and final cursor against Go.

The three helpers port the order, branching and identity preservation of this composition. They do not supply a graph backend, parser, binding implementation or complete rule runner. The dependency implementations are explicit callbacks, and their semantics are not proved by replay.

| Consumer | binaryExpression | classLike | ifStatement |
|---|---:|---:|---:|
| array-callback-return | 16 | 0 | 30 |
| consistent-return | 0 | 18 | 66 |
| no-unreachable-loop | 492 | 2 | 371 |
| react-hooks/rules-of-hooks | 32 | 11 | 88 |

Observed 1,126 live calls plus 27 synthetic controls, 1,153 comparisons total. A zero count means that captured corpus did not reach that helper; it does not establish that the rule cannot reach it. All three helpers have real live observations and controls. Controls exercise logical/coalescing and compound/plain/destructuring assignments, ordinary operand order, nested/absent/abrupt else branches, decorated/type-parameterized classes, heritage/member order and nil fields. Flattened payloads intentionally retain parked arena entries behind absent class/list/type fields, and Go's actual nil fields decide whether those payloads are ignored.

The first synthetic control put a SourceFile node in a heritage list. Go rejected it with an interface type assertion, so that control was corrected to a real typed heritage clause with nil Types. The raw exploratory failure is retained in evidence/oracle-control-rejection.log; it is not a parity success or a mutant kill. Wrong node kinds are outside the typed view contract and must be rejected by an AST adapter, rather than converted into missing metadata. No shared code was edited to accommodate the control.

# Commands and outputs

Commands source /workspace/adamic-tools/env.sh. Test stdout/stderr goes straight to log files. Native builds use ASan/UBSan. The owned command wrapper requires successful execution and empty stderr before comparing stdout; compilation, sanitizer, panic and stderr failures do not count as semantic-mutant catches.

```
bash cloud/setup.sh > stage1/cohere/lint/helpers/slot03/batch18/evidence/setup.log 2>&1
python3 stage1/cohere/lint/helpers/slot03/batch18/testdata/regenerate.py > stage1/cohere/lint/helpers/slot03/batch18/evidence/regeneration.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch18' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch18/evidence/helpers.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch18Mutants$/^if_statement.a#0[456]$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch18/evidence/else-mutants.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/slot03/batch18/evidence/vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot03 > stage1/cohere/lint/helpers/slot03/batch18/evidence/format.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch18/evidence/input-oracle.log 2>&1
go test ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch18/evidence/final-helpers.log 2>&1
```

Selected gate PASS 80.495s: baseline 13.90s and eighteen mutants 66.59s. Three additional optional-else mutants PASS 11.710s. Six uncached input probes PASS 2.417s, zero hits and six misses. Vet and formatting logs are empty. Complete helper gate PASS: all fifty-four owned helpers, 3,046,326 comparisons, 203 owned variants, four inherited variants and the missing-consumer check. Owned package 875.662s, shared package 93.618s. Final fetch found lint area advanced to b4691483 while main remained c7991b90; rebase and current-base verification follow.

Setup: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, cache warm 151s, done 151s; nproc 5, cgroup cpu.max 400000 100000, 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0.

# Every new mutant

Each variant compiles and runs successfully on sanitized native, then differs from actual Go's operation/cursor trace. Source Node, emitted JavaScript and sanitized native baseline comparisons all pass. This table reports observed mutant kills on native; it does not claim three-runtime mutant executions.

| Helper | Mutation | First differing line |
|---|---|---:|
| binaryExpression | disable logical short-circuit branching | 101 |
| binaryExpression | disable logical/coalescing assignment branching | 1130 |
| binaryExpression | skip plain-equals destructuring dispatch | 303 |
| binaryExpression | treat compound assignment as ordinary binary evaluation | 962 |
| binaryExpression | omit lazy destructuring predicate | 303 |
| binaryExpression | bind right instead of left destructuring target | 968 |
| binaryExpression | reverse ordinary operand evaluation | 91 |
| classLike | ignore missing class metadata | 1152 |
| classLike | omit decorators | 4 |
| classLike | omit type parameters | 4 |
| classLike | read parked heritage payload when list is absent | 1151 |
| classLike | read parked type payload when Types is absent | 1153 |
| classLike | read parked members when member list is absent | 1151 |
| classLike | visit class instead of each member header | 4 |
| ifStatement | route absent else to then entry | 1 |
| ifStatement | swap condition continuations | 1 |
| ifStatement | evaluate else instead of then | 1 |
| ifStatement | omit after entry | 1 |
| ifStatement | omit present else allocation | 17 |
| ifStatement | omit present else execution | 17 |
| ifStatement | execute absent else | 1 |

Raw first differences:

```
TestBatch18Mutants/binary_expression.a
batch18_test.go:147: compiled semantic mutant caught at line 101: got "2/expr:1195:-1:-1:2>2:-1;expr:1196:-1:-1:2>2:4;" Go "4/expr:1195:-1:-1:2>2:-1;new:-1:-1:-1:2>2:4;link:2:4:-1:2>2:-1;new:-1:-1:-1:2>2:5;link:2:5:-1:2>2:-1;enter:5:-1:-1:2>5:-1;expr:1196:-1:-1:5>5:-1;link:5:4:-1:5>5:-1;enter:4:-1:-1:5>4:-1;"
TestBatch18Mutants/binary_expression.a#01
batch18_test.go:147: compiled semantic mutant caught at line 1130: got "5/reads:21258:-1:-1:5>5:-1;expr:21259:-1:-1:5>5:7;writes:21258:-1:-1:5>5:-1;" Go "7/reads:21258:-1:-1:5>5:-1;new:-1:-1:-1:5>5:7;link:5:7:-1:5>5:-1;new:-1:-1:-1:5>5:8;link:5:8:-1:5>5:-1;enter:8:-1:-1:5>8:-1;expr:21259:-1:-1:8>8:-1;writes:21258:-1:-1:8>8:-1;link:8:7:-1:8>8:-1;enter:7:-1:-1:8>7:-1;"
TestBatch18Mutants/binary_expression.a#02
batch18_test.go:147: compiled semantic mutant caught at line 303: got "0/reads:3989:-1:-1:0>0:0;expr:3990:-1:-1:0>0:-1;writes:3989:-1:-1:0>0:-1;" Go "0/destructure:3989:-1:-1:0>0:0;reads:3989:-1:-1:0>0:-1;expr:3990:-1:-1:0>0:-1;writes:3989:-1:-1:0>0:-1;"
TestBatch18Mutants/binary_expression.a#03
batch18_test.go:147: compiled semantic mutant caught at line 962: got "0/expr:16435:-1:-1:0>0:-1;expr:16436:-1:-1:0>0:-1;" Go "0/reads:16435:-1:-1:0>0:-1;expr:16436:-1:-1:0>0:-1;writes:16435:-1:-1:0>0:-1;"
TestBatch18Mutants/binary_expression.a#04
batch18_test.go:147: compiled semantic mutant caught at line 303: got "0/reads:3989:-1:-1:0>0:0;expr:3990:-1:-1:0>0:-1;writes:3989:-1:-1:0>0:-1;" Go "0/destructure:3989:-1:-1:0>0:0;reads:3989:-1:-1:0>0:-1;expr:3990:-1:-1:0>0:-1;writes:3989:-1:-1:0>0:-1;"
TestBatch18Mutants/binary_expression.a#05
batch18_test.go:147: compiled semantic mutant caught at line 968: got "1/destructure:16656:-1:-1:0>0:1;expr:16657:-1:-1:0>0:-1;bind:16657:-1:-1:0>1:-1;" Go "1/destructure:16656:-1:-1:0>0:1;expr:16657:-1:-1:0>0:-1;bind:16656:-1:-1:0>1:-1;"
TestBatch18Mutants/binary_expression.a#06
batch18_test.go:147: compiled semantic mutant caught at line 91: got "7/expr:954:-1:-1:7>7:-1;expr:953:-1:-1:7>7:-1;" Go "7/expr:953:-1:-1:7>7:-1;expr:954:-1:-1:7>7:-1;"
TestBatch18Mutants/class_like.a
batch18_test.go:147: compiled semantic mutant caught at line 1152: got "0/decorators:21436:-1:-1:0>0:-1;typeParameters:21436:-1:-1:0>0:-1;" Go "0/"
TestBatch18Mutants/class_like.a#01
batch18_test.go:147: compiled semantic mutant caught at line 4: got "0/typeParameters:67:-1:-1:0>0:-1;memberHeader:68:-1:-1:0>0:-1;" Go "0/decorators:67:-1:-1:0>0:-1;typeParameters:67:-1:-1:0>0:-1;memberHeader:68:-1:-1:0>0:-1;"
TestBatch18Mutants/class_like.a#02
batch18_test.go:147: compiled semantic mutant caught at line 4: got "0/decorators:67:-1:-1:0>0:-1;memberHeader:68:-1:-1:0>0:-1;" Go "0/decorators:67:-1:-1:0>0:-1;typeParameters:67:-1:-1:0>0:-1;memberHeader:68:-1:-1:0>0:-1;"
TestBatch18Mutants/class_like.a#03
batch18_test.go:147: compiled semantic mutant caught at line 1151: got "0/decorators:21423:-1:-1:0>0:-1;typeParameters:21423:-1:-1:0>0:-1;expr:21424:-1:-1:0>0:-1;expr:21425:-1:-1:0>0:-1;expr:21426:-1:-1:0>0:-1;" Go "0/decorators:21423:-1:-1:0>0:-1;typeParameters:21423:-1:-1:0>0:-1;"
TestBatch18Mutants/class_like.a#04
batch18_test.go:147: compiled semantic mutant caught at line 1153: got "0/decorators:21423:-1:-1:0>0:-1;typeParameters:21423:-1:-1:0>0:-1;expr:21424:-1:-1:0>0:-1;expr:21425:-1:-1:0>0:-1;expr:21426:-1:-1:0>0:-1;memberHeader:21427:-1:-1:0>0:-1;memberHeader:21428:-1:-1:0>0:-1;" Go "0/decorators:21423:-1:-1:0>0:-1;typeParameters:21423:-1:-1:0>0:-1;expr:21425:-1:-1:0>0:-1;expr:21426:-1:-1:0>0:-1;memberHeader:21427:-1:-1:0>0:-1;memberHeader:21428:-1:-1:0>0:-1;"
TestBatch18Mutants/class_like.a#05
batch18_test.go:147: compiled semantic mutant caught at line 1151: got "0/decorators:21423:-1:-1:0>0:-1;typeParameters:21423:-1:-1:0>0:-1;memberHeader:21427:-1:-1:0>0:-1;memberHeader:21428:-1:-1:0>0:-1;" Go "0/decorators:21423:-1:-1:0>0:-1;typeParameters:21423:-1:-1:0>0:-1;"
TestBatch18Mutants/class_like.a#06
batch18_test.go:147: compiled semantic mutant caught at line 4: got "0/decorators:67:-1:-1:0>0:-1;typeParameters:67:-1:-1:0>0:-1;memberHeader:67:-1:-1:0>0:-1;" Go "0/decorators:67:-1:-1:0>0:-1;typeParameters:67:-1:-1:0>0:-1;memberHeader:68:-1:-1:0>0:-1;"
TestBatch18Mutants/if_statement.a
batch18_test.go:147: compiled semantic mutant caught at line 1: got "1/new:-1:-1:-1:0>0:1;new:-1:-1:-1:0>0:2;condition:4:2:2:0>0:-1;enter:2:-1:-1:0>2:-1;statement:5:-1:-1:2>3:-1;link:3:1:-1:3>3:-1;enter:1:-1:-1:3>1:-1;" Go "1/new:-1:-1:-1:0>0:1;new:-1:-1:-1:0>0:2;condition:4:2:1:0>0:-1;enter:2:-1:-1:0>2:-1;statement:5:-1:-1:2>3:-1;link:3:1:-1:3>3:-1;enter:1:-1:-1:3>1:-1;"
TestBatch18Mutants/if_statement.a#01
batch18_test.go:147: compiled semantic mutant caught at line 1: got "1/new:-1:-1:-1:0>0:1;new:-1:-1:-1:0>0:2;condition:4:1:2:0>0:-1;enter:2:-1:-1:0>2:-1;statement:5:-1:-1:2>3:-1;link:3:1:-1:3>3:-1;enter:1:-1:-1:3>1:-1;" Go "1/new:-1:-1:-1:0>0:1;new:-1:-1:-1:0>0:2;condition:4:2:1:0>0:-1;enter:2:-1:-1:0>2:-1;statement:5:-1:-1:2>3:-1;link:3:1:-1:3>3:-1;enter:1:-1:-1:3>1:-1;"
TestBatch18Mutants/if_statement.a#02
batch18_test.go:147: compiled semantic mutant caught at line 1: got "1/new:-1:-1:-1:0>0:1;new:-1:-1:-1:0>0:2;condition:4:2:1:0>0:-1;enter:2:-1:-1:0>2:-1;statement:-1:-1:-1:2>3:-1;link:3:1:-1:3>3:-1;enter:1:-1:-1:3>1:-1;" Go "1/new:-1:-1:-1:0>0:1;new:-1:-1:-1:0>0:2;condition:4:2:1:0>0:-1;enter:2:-1:-1:0>2:-1;statement:5:-1:-1:2>3:-1;link:3:1:-1:3>3:-1;enter:1:-1:-1:3>1:-1;"
TestBatch18Mutants/if_statement.a#03
batch18_test.go:147: compiled semantic mutant caught at line 1: got "3/new:-1:-1:-1:0>0:1;new:-1:-1:-1:0>0:2;condition:4:2:1:0>0:-1;enter:2:-1:-1:0>2:-1;statement:5:-1:-1:2>3:-1;link:3:1:-1:3>3:-1;" Go "1/new:-1:-1:-1:0>0:1;new:-1:-1:-1:0>0:2;condition:4:2:1:0>0:-1;enter:2:-1:-1:0>2:-1;statement:5:-1:-1:2>3:-1;link:3:1:-1:3>3:-1;enter:1:-1:-1:3>1:-1;"
TestBatch18Mutants/if_statement.a#04
batch18_test.go:150: compiled semantic mutant caught at line 17: got "5/new:-1:-1:-1:0>0:1;new:-1:-1:-1:0>0:2;condition:113:2:1:0>0:3;enter:2:-1:-1:0>0:-1;statement:114:-1:-1:0>2:-1;link:2:1:-1:2>4:-1;enter:1:-1:-1:4>4:-1;statement:115:-1:-1:4>3:-1;link:3:1:-1:3>5:-1;enter:1:-1:-1:5>5:-1;" Go "1/new:-1:-1:-1:0>0:1;new:-1:-1:-1:0>0:2;new:-1:-1:-1:0>0:3;condition:113:2:3:0>0:-1;enter:2:-1:-1:0>2:-1;statement:114:-1:-1:2>4:-1;link:4:1:-1:4>4:-1;enter:3:-1:-1:4>3:-1;statement:115:-1:-1:3>5:-1;link:5:1:-1:5>5:-1;enter:1:-1:-1:5>1:-1;"
TestBatch18Mutants/if_statement.a#05
batch18_test.go:150: compiled semantic mutant caught at line 17: got "3/new:-1:-1:-1:0>0:1;new:-1:-1:-1:0>0:2;new:-1:-1:-1:0>0:3;condition:113:2:3:0>0:-1;enter:2:-1:-1:0>2:-1;statement:114:-1:-1:2>4:-1;link:4:1:-1:4>4:-1;enter:1:-1:-1:4>3:-1;" Go "1/new:-1:-1:-1:0>0:1;new:-1:-1:-1:0>0:2;new:-1:-1:-1:0>0:3;condition:113:2:3:0>0:-1;enter:2:-1:-1:0>2:-1;statement:114:-1:-1:2>4:-1;link:4:1:-1:4>4:-1;enter:3:-1:-1:4>3:-1;statement:115:-1:-1:3>5:-1;link:5:1:-1:5>5:-1;enter:1:-1:-1:5>1:-1;"
TestBatch18Mutants/if_statement.a#06
batch18_test.go:150: compiled semantic mutant caught at line 1: got "1/new:-1:-1:-1:0>0:1;new:-1:-1:-1:0>0:2;condition:4:2:1:0>0:-1;enter:2:-1:-1:0>2:-1;statement:5:-1:-1:2>3:-1;link:3:1:-1:3>3:-1;enter:1:-1:-1:3>1:-1;statement:-1:-1:-1:1>1:-1;link:1:1:-1:1>1:-1;enter:1:-1:-1:1>1:-1;" Go "1/new:-1:-1:-1:0>0:1;new:-1:-1:-1:0>0:2;condition:4:2:1:0>0:-1;enter:2:-1:-1:0>2:-1;statement:5:-1:-1:2>3:-1;link:3:1:-1:3>3:-1;enter:1:-1:-1:3>1:-1;"
```

# Limits

Not a complete rule findings/fix/suggestion corpus, independent Adamic parsing of these sources, arbitrary graph/AST fuzzing or dependency backend validation. Capture and callback replay are bounded. The full repository gate, including its seventeen external stage-1 correctness comparisons, was not run in this helper unit. No executed correctness test was skipped, and no required-input check was weakened or deleted. The shared harness was previously green on this unchanged base; its full tests were not rerun for these helper-only changes.

All previously delivered variants are rerun in final-helpers.log; evidence/all-mutant-witnesses.log names every rerun mutant and its observed first mismatch or guard result. Their mutations are specified in the earlier batch reports and owned tests.

# Landing on the updated lint area

A final fetch advanced the lint area from b84a9d93 to b4691483. The completed helper unit was committed, then all sixty-three own commits rebased cleanly onto origin/area/stage1-lint. Current main c7991b90 remains an ancestor. The new base migrates legacy lint rules to the registry; those integration changes are accepted unchanged. Rebased implementation a54f84ee and claim 1ac93d1b preserve the original pushed-before-source reservation. A complete second helper gate runs against this new base, together with vet, six uncached input probes and six selected shared-harness checks. The rechecked twenty remote helper branches and nineteen claim files have no competing mentions.

```
git rebase origin/area/stage1-lint
go test ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch18/evidence/rebased-helpers.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/slot03/batch18/evidence/rebased-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run "^TestInputAgreesWithNode$" -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch18/evidence/rebased-input-oracle.log 2>&1
go test ./stage1/cohere/lint -run "^(TestEmittedJavaScriptMismatch|TestDotARename|TestCompleteSuggestionSerialization|TestSuggestionAlongsideAutomaticFix|TestWitnessScriptKind|TestProfileCompilation)$" -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch18/evidence/rebased-harness.log 2>&1
```

Rebased auxiliary checks: shared helpers PASS 93.953s; six selected shared-harness tests PASS 315.787s; vet output empty; six uncached input probes PASS 2.266s with zero hits and six misses. The emitted-JavaScript mismatch test intentionally launches a failing child comparison and passes only when that mismatch is caught; its child FAIL text is an expected mutant witness, not a gate failure. Profile compilation reports 139 allocations and 139 frees.

Final landing result: all fifty-four owned helpers PASS 896.257s on b4691483, containing freshly fetched current main c7991b90. Exactly 3,046,326 comparison lines, 202 owned compiled semantic variants, one owned empty-stack guard variant, four inherited semantic variants, and the missing-consumer check pass. No helper tests skip or fail. Source Node and sanitized native are covered throughout; later batches also compare emitted JavaScript, including all 1,153 new cases. The initial full gate passed before rebase as recorded above. No further helpers are claimed. Publication targets only codex/lint-helpers-03 using the exact lease of previously pushed claim 5ec90184186a693a1882f9d6685ecd079bf25e24.
