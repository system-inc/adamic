Merged views integration and repaired its inherited-constructor parameter-property panic.
Views merge: 5aa43d0a (incoming ffe428ab26eefb73154adbf570ad99e9c1b4f872); c41c0e06 merge remains blocked.
Focused lowering, representation, native and oracle regressions passed; expanded Node host tests and counts refresh failed.
Record/Uint8Array collision mutant is caught by the tag invariant; removing the constructor guard panics in TestParameterPropertyCheckerContracts.
All 22 kinds replayed against views; no new kind is certified, and the non-null policy ruling remains outstanding.

The requested non-null branch checks .ts assertions and deliberately refuses .a assertions. The requested views branch admits .a assertions, including literal assertions used to reserve uninitialized slots. Automatic approval review rejected a combined resolution across lowering, IR and CLI reporting because that semantic choice was not clearly authorized and risked silent miscompilation. The merge was aborted after saving its conflict diff at /tmp/adamic-syntax-non-null-conflicts.patch. User choice was requested asynchronously; no answer has arrived. No rejected patch was applied.

The constructor repair evaluates assertionInitializer only for PropertyDeclaration nodes. Parameter properties are ParameterDeclaration nodes and were panicking in inherited constructor field creation. Ownership history shows the other class-construction worker changes classViewRefusal, not this guard. Files outside this function: internal/ir/syntax_kind_tags_test.go and this report/evidence. The tag invariant documents the prerequisite merge's separate Record=14 and typed-array=15/16/17 representations. No additional runtime helper was added by the repair. The prerequisite merge includes shared runtime files and requires runtime-owner review, particularly graph-region invalid frees.

Observations below precede the pending non-null merge. Exit 0 means the exact NotYet signature reproduced, not successful compilation. Exit 1 with findings means a different stop; exit 1 without findings is only a census-echo candidate. Existing namespace fixtures and mutants pass and both namespace examples have no findings. None is claimed to cover all roots.

| Root sites | Kind | Replay | Observed findings (first three) |
| ---: | --- | --- | --- |
| 46 | checked-write | Exact stop reproduced | reading Debug; a cast the runtime can't check; a value of type unknown |
| 25 | nested-generic | Signature absent | a function returning T |
| 10 | spread | Exact stop reproduced | a method call through a structural signature in a program with statics; use typeof the declaring class; a method call through a structural signature in a program with statics; use typeof the declaring class; a SpreadElement |
| 9 | namespace | Signature absent | No lowering findings |
| 7 | overload-argument | Signature absent | No lowering findings |
| 5 | overload-value | Exact stop reproduced | a BinaryExpression with a number and a number; a BinaryExpression with a number and a number; a BinaryExpression with a number and a number |
| 5 | substr | Signature absent | an array of any |
| 4 | nested-reference | Exact stop reproduced | a BinaryExpression as a statement; a first-class nested function reference from another nested function; reading result |
| 3 | dictionary | Signature absent | a rest array of DiagnosticArguments |
| 1 | class-expression | Exact stop reproduced | a ClassExpression |
| 1 | postfix | Exact stop reproduced | a PostfixUnaryExpression |
| 1 | boolean-field | Signature absent | a BinaryExpression as a statement |
| 1 | comparator | Signature absent | an array of T |
| 1 | uninitialized-field | Exact stop reproduced | an uninitialized object field without a supported declared slot type |
| 1 | array-length | Exact stop reproduced | assigning to a NonNullExpression; a value of type T; an untyped length Array escaping without a proven element representation |
| 1 | predicate | Signature absent | No lowering findings |
| 1 | generic-result | Signature absent | overload 1 of skipOuterExpressions result T cannot be served by implementation result Node |
| 1 | unwatch | Exact stop reproduced | numeric coercion requiring dynamic ToPrimitive; numeric coercion requiring dynamic ToPrimitive; node:fs.watchFile |
| 1 | watch | Exact stop reproduced | node:fs.watch |
| 1 | watch-file | Exact stop reproduced | numeric coercion requiring dynamic ToPrimitive; numeric coercion requiring dynamic ToPrimitive; node:fs.watchFile |
| 1 | stat-options | Exact stop reproduced | statSync: fs options other than a fixed literal or its plain const binding |
| 1 | locale-time | Exact stop reproduced | toLocaleTimeString: Date methods other than getTime/valueOf in the fs file host |

Full findings and replay stderr are retained in evidence/views-replays.jsonl. Additional binder generic replay reaches a value of type T at 1477:71; checker generic replay reaches a function returning T at 6517:18. Second namespace example checker.ts:54223:1 also has no findings. All replays used the unchanged /tmp/adamic-syntax-adapted source and full src/tsc/tsc.ts entry. The reusable worker was built using make_overlay.py and go build -buildvcs=false -overlay=/tmp/adamic-syntax-views-overlay/overlay.json ./stage3/census/latent/replay/worker. Node declarations initially missing; npm ci --prefix stage3/api installed locked @types/node 25.3.3 and TypeScript 6.0.3.

Commands and outcomes (all output redirected, no whole package or full gate):

- go build ./internal/lower ./internal/native ./internal/javascript ./internal/flow ./internal/fresh passed after resolving views conflicts.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/ir ./internal/native ./internal/javascript ./internal/flow ./internal/fresh ./internal/oracle -run '^TestTypedArray|^TestNamespace|^TestParameterPropert|^TestSyntaxKindRepresentationTags$|^TestViewObjectContractsAreAvailableToEraser$|^TestViewObjectWritesNeedSourceCertificate$|^TestSyntaxSubstrMutants$|^TestNativeAgreesWithNode/internal/oracle/testdata/(syntax_substr|typed_arrays_uint8array|params_namespaces_values).a$' -count=1 -timeout=10m passed. Lower 4.082s, IR .039s, native 1.661s, oracle 4.110s. JS/flow/fresh had no matching tests in this selection.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/flow ./internal/fresh ./internal/javascript ./internal/oracle -run '^TestArrayHolesExceptionEdges$|^TestNodeBufferHashUpdateKeepsAlias$|^TestNodeFSFile|^TestViewArraysNode$|^TestViewArrayFieldPresence$|^TestViewCallableProducerCertificateNode$|^TestNativeAgreesWithNode/internal/oracle/testdata/params_namespaces_.*.a$' -count=1 -timeout=5m: flow .013s, fresh .013s, JS .787s passed; oracle failed 7.420s. Node host system deleteFile emits a value return from a C void function; mkdtemp options are refused; errno.code checked-view representation fails. These are not valid mutant kills.
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout=10m -args -update-counts failed 46.004s. Multiple inherited fixture lowering/native errors, including graph-region free(): invalid pointer. The writer leaves counts.md unchanged after failure; no new fixture was added by this group.
- go test ./internal/ir -run '^TestSyntaxKindRepresentationTags$' -v -count=1 passed .007s and logged the valid-input collision mutant.
- Pre-fix worktree go test ./internal/lower -run '^TestParameterPropertyCheckerContracts$' -count=1 failed .125s with the exact ParameterDeclaration to PropertyDeclaration panic. This restores the missing guard; the fixed corresponding test passes.
- git diff --check passed.

The passing focused oracle selection includes the existing namespace semantic mutants, parameter-property semantic and ownership mutants, and all seven substr mutants. Their suite assertions require the intended Node disagreement/ownership check; expanded Node host mutants that fail earlier are explicitly reported as failures above. This group does not add new fixture semantics or claim new root-site coverage. Previously certified substr remains five covered table sites with replay reaching the next stop. The other 122 sites remain unresolved or census-echo candidates pending both prerequisites and proper certification.

Runtime files brought by the views prerequisite merge, including adamic.h, graph_regions.c and environment/checked-view helpers, need review by their owner. No shared-runtime repair was attempted. Additional ranked lowering work remains blocked by the requested non-null prerequisite and these failing inherited validations. We stopped for a concrete ruling instead of choosing incompatible .a semantics.
