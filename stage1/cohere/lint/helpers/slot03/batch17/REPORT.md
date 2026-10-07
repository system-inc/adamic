Built three CFG composition helpers in separate .a files: conditionalExpression, bindWithDefault and memberHeader.
Commits: replacement claim 397caa0b pushed before source; prior landing bfeb2be3; final publication SHA is named in the final response.
Commands: selected helper gate PASS 41.568s and 948 comparisons; vet/format clean; six uncached input probes PASS 1.986s; setup 53s, nproc 5.
Mutants: all twelve new variants compile and run successfully, then differ from real Go call traces; compiler refusals are not counted.
Not covered: complete rule findings, parser/backend integration, arbitrary graph dependencies and the full repository gate; main advanced and requires landing verification below.

# Scope and ownership

The original JSX/hook-search claim dd8d7cfa was ten seconds later than slot 04 df5d5097. All three were withdrawn. Their duplicate sources are absent. The abandoned run's raw evidence is compressed in evidence/withdrawn-search.log.gz; seven initial variants passed, while the redundant-search variant was stopped after its recursive work expanded. None is delivered or counted. Replacement claim 397caa0b at 07:39:38 UTC was pushed before source. Final wildcard fetch checked twenty origin helper branches; no other claims mention any replacement. The comment collection bundle remains reserved in shared HELPERS.md; regex-engine internals are excluded by the JS RegExp instruction.

The replacement helpers tie the highest eligible unclaimed fan-out, four rules each. No shared generator, test harness, rule directory, compiler or runtime file was changed. The helpers take parser views and opaque node/block arena handles, rather than dispatching rules or inventing numeric AST kinds. No regex is introduced. Prior forty-eight helpers were green and published on the unchanged fetched bases before these claims.

# Rules and readiness

Each of the three helpers removes a dependency for all four rules:

- array-callback-return
- consistent-return
- no-unreachable-loop
- react-hooks/rules-of-hooks

This is twelve removed dependency edges, four unique consumers and zero final blockers removed by this batch alone in the frozen inventory. readiness.json subtracts only these three symbols and lists every remaining dependency. This is helper readiness, not completed rules or native findings parity.

# What the oracle observes

The pinned upstream cohere revision is 715ba94f3608a6500086b1076ce5cb7e51b836db. The capture overlay records actual fixture strings from the upstream core and React packages without changing their worktrees. Both packages pass, capturing 2,119 sources across every consumer. The comparator parses those sources with the actual Go parser and builds their real CFG roots.

Temporary overlays rename the helper bodies and dependency methods, then invoke the original bodies through owned observation wrappers. Nested dependency work executes the actual Go implementation; only the outer helper's direct calls are recorded. Records hold operation names, node/block identities, callback results and before/after cursor state. The Adamic driver replays callback results/cursor changes and independently generates the requested calls and arguments. It compares the complete trace and final cursor against Go on source Node, emitted JavaScript and sanitized native. Expression evaluation, pattern binding, decorators, graph allocation, edge storage and entry remain explicit dependencies; their semantics are not implemented or proved by replay.

Observed live calls by consumer:

| Consumer | conditionalExpression | bindWithDefault | memberHeader |
|---|---:|---:|---:|
| array-callback-return | 5 | 75 | 262 |
| consistent-return | 0 | 2 | 38 |
| no-unreachable-loop | 0 | 114 | 171 |
| react-hooks/rules-of-hooks | 51 | 72 | 139 |

There are 929 live helper calls plus 19 synthetic-control calls, 948 comparisons total. Zero conditional-expression calls in two captured corpora is an observation, not a claim those rules cannot use the helper. Controls cover nested conditionals, defaults/destructuring, computed member names, decorators, property annotations and an index signature's parameter/return types. All helpers have real consumer observations and controls. Every parsed consumer input is represented in the capture, but only helper calls actually reached by Go produce comparison lines.

# Validation

Toolchain commands source /workspace/adamic-tools/env.sh. Test stdout and stderr go directly to log files, never pipes. Native builds use ASan/UBSan; successful driver runs require exit zero with empty stderr. The owned command wrapper enforces successful compilation/execution before stdout mismatch counts as a mutant witness.

```
bash cloud/setup.sh > stage1/cohere/lint/helpers/slot03/batch17/evidence/setup.log 2>&1
python3 stage1/cohere/lint/helpers/slot03/batch17/testdata/regenerate.py > stage1/cohere/lint/helpers/slot03/batch17/evidence/regeneration.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch17' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch17/evidence/helpers.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/slot03/batch17/evidence/vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot03 > stage1/cohere/lint/helpers/slot03/batch17/evidence/format.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot03/batch17/evidence/input-oracle.log 2>&1
```

Before landing: helper gate PASS 41.568s, source/emitted/native comparison PASS 9.96s, twelve mutants PASS 31.60s. Vet and format logs are empty. Six uncached input probes PASS 1.986s, zero cache hits/six misses. Capture Go package results: core PASS 8.882s, React PASS 5.085s.

Setup timing: Go 0s, clang 0s, Node 0s, submodules 1s, cache warm 53s, done 53s; nproc 5, cpu.max 400000 100000, 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0.

The replay driver initially used an unsupported local function, then a closure refused as cycle-capable; a readonly-list assignment also exposed a typing mismatch. Moving the driver callback to module scope and preserving the reader's readonly list resolved these locally. The retained closure-refusal log is failed exploratory evidence, not a mutant kill. No compiler or harness change was made.

# Every new mutant

| Helper | Mutation | Caught at line |
|---|---|---:|
| conditionalExpression | omit condition evaluation | 335 |
| conditionalExpression | link false branch from current true-branch end instead of original test end | 335 |
| conditionalExpression | evaluate false node in true branch | 335 |
| conditionalExpression | omit final join entry | 335 |
| bindWithDefault | omit fallback fork | 462 |
| bindWithDefault | bind fallback instead of target | 15 |
| bindWithDefault | omit join entry before binding | 462 |
| memberHeader | omit member decorators | 1 |
| memberHeader | decorate member instead of parameter | 14 |
| memberHeader | omit computed key expression | see raw witness below |
| memberHeader | omit property annotation, including nil annotation call | 11 |
| memberHeader | evaluate return type instead of index parameter type | 938 |

All twelve variants compile and run successfully with empty stderr; each is caught by stdout disagreement with actual Go. The following raw first differences identify exact node/block handles and cursor transitions:

```
batch17_test.go:133: compiled semantic mutant caught at line 335: got "3/new:-1:-1:0>0:-1;new:-1:-1:0>0:1;link:0:1:0>0:2;enter:1:-1:0>0:-1;expr:7256:-1:0>2:-1;link:2:-1:2>2:-1;new:-1:-1:2>2:-1;link:0:-1:2>2:3;enter:-1:-1:2>2:-1;expr:7257:-1:2>3:-1;link:3:-1:3>3:-1;enter:-1:-1:3>3:-1;" Go "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 335: got "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:2:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;" Go "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 335: got "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7257:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;" Go "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 335: got "3/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;" Go "1/expr:7255:-1:0>0:-1;new:-1:-1:0>0:1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7256:-1:2>2:-1;link:2:1:2>2:-1;new:-1:-1:2>2:3;link:0:3:2>2:-1;enter:3:-1:2>3:-1;expr:7257:-1:3>3:-1;link:3:1:3>3:-1;enter:1:-1:3>1:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 462: got "0/bind:7852:-1:0>0:1;" Go "1/new:-1:-1:0>0:1;link:0:1:0>0:-1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7853:-1:2>2:-1;link:2:1:2>2:-1;enter:1:-1:2>1:-1;bind:7852:-1:1>1:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 15: got "0/bind:-1:-1:0>0:-1;" Go "0/bind:76:-1:0>0:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 462: got "1/new:-1:-1:0>0:1;link:0:1:0>0:-1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7853:-1:2>2:-1;link:2:1:2>2:-1;bind:7852:-1:2>1:-1;" Go "1/new:-1:-1:0>0:1;link:0:1:0>0:-1;new:-1:-1:0>0:2;link:0:2:0>0:-1;enter:2:-1:0>2:-1;expr:7853:-1:2>2:-1;link:2:1:2>2:-1;enter:1:-1:2>1:-1;bind:7852:-1:1>1:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 1: got "0/" Go "0/decorators:0:-1:0>0:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 14: got "0/decorators:74:-1:0>0:-1;decorators:74:-1:0>0:-1;" Go "0/decorators:74:-1:0>0:-1;decorators:75:-1:0>0:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 5: got "0/decorators:39:-1:0>0:-1;" Go "0/decorators:39:-1:0>0:-1;expr:40:-1:0>0:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 11: got "0/decorators:65:-1:0>0:-1;" Go "0/decorators:65:-1:0>0:-1;expr:-1:-1:0>0:-1;"
batch17_test.go:133: compiled semantic mutant caught at line 938: got "0/decorators:10880:-1:0>0:-1;expr:10882:-1:0>0:-1;expr:10882:-1:0>0:-1;" Go "0/decorators:10880:-1:0>0:-1;expr:10881:-1:0>0:-1;expr:10882:-1:0>0:-1;"
```

# Limits and landing

These are bounded helper-composition observations, not full findings/fix/suggestion corpora, arbitrary AST/graph fuzzing, or a port of dependency implementations. The full repository gate and its seventeen external stage-1 correctness comparators were not run in this bounded helper gate. No required input check was skipped, weakened, deleted or reported green; the executed helper and input-oracle gates contain no skips.

Final pre-publication fetch observed new main c7991b90 and lint area b84a9d93. The batch is committed before rebasing, and will be oracle-checked again on that base before publication. Later landing evidence below supersedes prelanding timing and bases.
