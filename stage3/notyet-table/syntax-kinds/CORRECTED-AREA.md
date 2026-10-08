Reverted views integration and its repair; kept checked non-null with .a assertions refused; certified existing namespace syntax.
Revert 10259d0a; non-null merge eb74d12b (c41c0e06); compiler-area refresh 835beee7 (b68b2fe1).
Focused non-null, namespace and substr tests passed; corrected-area counts refresh 19.670s and namespace-fixture counts refresh 19.800s passed.
Namespace initializer and scoped-read mutants fail source Node stdout comparison in both backends, finish cleanly and leak nothing.
Five substr root sites remain covered; nine ModuleDeclaration sites are cancelled as syntax census echoes; the other 113 sites remain unresolved.

The user corrected the prerequisite instruction: do not merge views directly; it arrives through compiler area later. Reverted 73bdb65e and 2cd72c73, then git revert -m 1 5aa43d0a, and pushed 10259d0a without rewriting history. The tree returned to the pre-views worker content. Checked non-null c41c0e06 merged without conflicts. The October 8 00:27 ruling explicitly keeps ! refused in .a; that branch's refusal governs. The previous approval-review blockage is resolved by this instruction and the corrected merge path. Newest area/compiler resolved to b68b2fe1 and merged as 835beee7. These prerequisite corrections are already pushed.

Added internal/oracle/testdata/syntax_module_declarations.a and its own registration and semantic-mutant test in internal/oracle/syntax_module_declarations_test.go. The fixture reduces BuilderState's type-only namespace and JsxNames' runtime string exports, including nested type aliases and same-named exports in separate namespaces. No production lowering change is needed: existing namespace flattening retains checker symbol identity and initialization. The two examples no longer stop at ModuleDeclaration: BuilderState has no findings; JsxNames reaches __String representation at checker.ts:54224:18 and later constants. The namespace syntax lesson is therefore cancelled as a census echo. This does not claim that every statement or every one of the nine enclosing sites compiles.

Both mutants preserve the string representation: the initializer mutant substitutes the other namespace's initialized string, and the scoped-read mutant redirects a read to the same-named export in the other namespace. Both valid mutants produce only stdout disagreement against source Node in native and JavaScript, and native leak checks pass. An earlier initializer mutant used undefined in a non-null string slot and failed UBSan; it was rejected as an invalid semantic witness, replaced, and is not counted as a successful kill. Fixture counts: allocations 4, frees 4, retains 4, releases 12, peak 3, regions 0. counts.md gained only this row.

All 22 ranked kinds were rechecked on the corrected compiler-area lineage. Each replay used unchanged adapted project bytes and the full tsc.ts entry, with the exact table position and reason. The overlay never returns production IR; matching a stop or losing a stop is not proof of whole-program compilation. The full findings and stderr are preserved in evidence/corrected-area-replays.jsonl. The table's first three findings illustrate the next blockers, not necessarily the exact target's finding.

| Root sites | Kind | Status | Replay observation |
| ---: | --- | --- | --- |
| 46 | checked-write | Skipped: another worker owns required lower function | assigning an element of a value; a cast the runtime can't check; a value of type unknown |
| 25 | nested-generic | Skipped: still unresolved | a function returning T |
| 10 | spread | Skipped: another worker owns required lower function | a method call through a structural signature in a program with statics; use typeof the declaring class; a method call through a structural signature in a program with statics; use typeof the declaring class; a SpreadElement |
| 9 | namespace | Cancelled: supported syntax census echo | No lowering findings in selected unit |
| 7 | overload-argument | Skipped: still unresolved | No lowering findings in selected unit |
| 5 | overload-value | Skipped: still unresolved | a BinaryExpression with a number and a number; a BinaryExpression with a number and a number; a BinaryExpression with a number and a number |
| 5 | substr | Lowered: prior fixture and seven mutants remain green | an array of never; an array of any; reading res |
| 4 | nested-reference | Skipped: still unresolved | a value of type Path |
| 3 | dictionary | Skipped: still unresolved | a cast the runtime can't check; reading buildOptions; reading buildOptions |
| 1 | class-expression | Skipped: another worker owns required lower function | a ClassExpression |
| 1 | postfix | Skipped: another worker owns required lower function | a PostfixUnaryExpression |
| 1 | boolean-field | Skipped: still unresolved | a BinaryExpression as a statement |
| 1 | comparator | Skipped: still unresolved | a cast the runtime can't check |
| 1 | uninitialized-field | Skipped: still unresolved | No lowering findings in selected unit |
| 1 | array-length | Skipped: still unresolved | a value of type T; Array as a value outside equality or typeof (overloaded calls and static properties need their own representation); reading grid |
| 1 | predicate | Skipped: still unresolved | a cast the runtime can't check; reading expression; reading args |
| 1 | generic-result | Skipped: still unresolved | overload 1 of skipOuterExpressions result T cannot be served by implementation result Node |
| 1 | unwatch | Skipped: still unresolved | a call to a PropertyAccessExpression; a function returning any; a function inside a function (a closure) |
| 1 | watch | Skipped: still unresolved | a function returning any |
| 1 | watch-file | Skipped: still unresolved | a call to a PropertyAccessExpression; a function returning any; a function inside a function (a closure) |
| 1 | stat-options | Skipped: still unresolved | a function returning any |
| 1 | locale-time | Skipped: still unresolved | a method call through a structural signature in a program with statics; use typeof the declaring class |

The generic nested declaration examples now reach a value of type T or a function returning T rather than the measured syntax-kind stop. The current compiler area still lacks the table's nested_functions.go, view_objects.go, view_array_consumers.go and phantom_overload_results.go families. Source-slot checked writes belong to setProperty in class.go, actively changed by class-set-property. Spread needs arrayLiteral (object-small); ClassExpression/PostfixUnaryExpression need enumNeverValue, edited by this-outside and void-value. No worker's lower function was edited in this group. No extra runtime helpers were added.

The uninitialized-field example is a literal undefined! assigned to a never field. Its .ts replay has no findings after checked non-null, while .a must retain the explicitly ruled assertion refusal. The non-null branch's focused tests cover literal-nullish and object-field assertion refusals. This observation is not counted as another completed kind. The first overload-argument example has no findings, but its full representation lesson is not certified; it remains unresolved pending exact reduction and ownership checks. Other kinds reach earlier, different named stops on the compiler-area lineage; none is silently marked lowered.

Commands (test output redirected to the preserved evidence logs):

- ADAMIC_GATE_UNCACHED=1 go test ./cmd/adamic ./internal/lower ./internal/ir ./internal/flow ./internal/oracle -run '^TestNonNull|^TestExplainChecksDriver$|^TestCheckedNonNull|^TestImpossibleNonNull|^TestPossibleNonNull|^TestSyntaxSubstrMutants$|^TestNativeAgreesWithNode/internal/oracle/testdata/syntax_substr.a$' -count=1 -timeout=5m passed: cmd 3.302s, lower .088s, oracle 1.982s; IR/flow had no matching tests.
- ADAMIC_GATE_UNCACHED=1 go test ./cmd/adamic ./internal/lower ./internal/flow ./internal/oracle -run '^TestNonNull|^TestExplainChecksDriver$|^TestCheckedNonNull|^TestImpossibleNonNull|^TestPossibleNonNull|^TestNamespace|^TestSyntaxSubstrMutants$|^TestNativeAgreesWithNode/internal/oracle/testdata/(syntax_substr|namespaces.*).a$' -count=1 -timeout=5m passed: cmd 2.887s, lower .784s, oracle 3.622s; flow had no matching tests.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestSyntaxModuleDeclarationMutants$|^TestNativeAgreesWithNode/internal/oracle/testdata/syntax_module_declarations.a$' -count=1 -v -timeout=5m passed .412s, including both valid mutants and the ordinary fixture's Node, JS, release native, ASan/UBSan and leaks.
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout=10m -args -update-counts passed before the fixture in 19.670s and after adding it in 19.800s. No existing count changed.
- Guarded worker built with python3 stage3/census/latent/make_overlay.py and go build -buildvcs=false -overlay=/tmp/adamic-syntax-area-overlay/overlay.json -o /tmp/adamic-syntax-area-replay ./stage3/census/latent/replay/worker. All 22 exact requests plus the second namespace example were executed.
- git diff --check passed; no whole-package run or full gate was performed.

The non-null branch's no-check mutant is caught by the mandated panic/output mismatch; its representation-changing narrowing mutant is caught before either backend by the return-representation assertion. All seven existing substr mutants remain caught by source Node stdout comparison in both backends. Existing namespace semantic mutants also pass their required stdout-disagreement assertions. Incoming non-null .ts fixtures were retained as prerequisites; all newly authored fixtures in this group are .a.

Remaining work is blocked by absent compiler-area lowering prerequisites and actively owned functions, not permission to edit IR or backends. Per user correction, views must arrive through compiler area; it was not reintroduced. No new design refusal or safety approval is requested. This report supersedes the reverted views-only validation claims.
