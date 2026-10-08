d0fe3a28 is the delivery merge of current main efe9f404 into codex/notyet-binary; completed code groups are separately pushed.
Built assignment values, represented object assignments and checked TypeScript non-null compound targets; remaining binary shapes verified.
Commands: complete owned/imported Node oracle PASS 14.107s; recorded counts PASS 24.798s; affected compiler package tests and vet pass.
Mutants: 29 continuation mutants killed and restored, plus nine prior binary mutants; two draft conditional mutants also killed.
Not covered: missing-binding Identifier sites, array-length store, Uint16 representation, conditional dispatch ownership, delete/yield refusals, null tag choice.

Counts below come from exact unique raw CSV sites, with attempted contexts retained separately. They are operator/target family counts, not a measured net census delta or a claim that whole compiler bodies compile. Existing earlier bindings, checker types and storage boundaries still apply; no combined total is claimed.

| SHA | Group | Raw sites and limit |
| --- | --- | --- |
| 91b598f1 | Original operator/checker breakdown | 1,452 unique sites |
| e7320353 | Logical selections | 1,202 operator-family sites |
| 845e49f9 | Strict reference inequalities | 60 operator-family sites |
| c1b2c9f9 | Assignment values | 190 `=` sites; 189 represented-target candidates, one known array-length store boundary |
| 2ce52c28 | Identifier binding triage | 143 root sites; zero new coverage; examples retain binding/type dependencies |
| a38bcefb | Remaining root breakdown | 204 sites: 198 logical and six additional string assignment values; no arithmetic |
| 67254f76 | Boolean/number | 116 sites: 79 `&&`, 37 `||`; existing generic rule verified |
| 1d4b352d | Value/number | 57 sites: 54 `&&`, three `||`; object and array shapes verified |
| 2febdbca | String/string | 31 sites: 22 `||`, three `&&`, six `=`; existing rules verified |
| 3b93cc07 | Named object assignments | Four represented target shapes; before/after replays reach next named dependencies |
| a32f1086 | Requested c41c0e06 merge | Both IR metadata fields retained; checked .ts assertions admitted and .a assertions remain refused |
| c0bc2ef8 | Checked non-null compound targets | Seven of nine raw target shapes have existing representations; two need Uint16Array |
| f5477a90 | String accessor lifetime | No additional raw coverage; ninth non-null repair/control shape and ownership mutant |
| 36483483 | Null identity equality boundary | Zero additional table coverage; prevent demonstrated native null/undefined miscompile |
| a8ad5bf0 | Conditional proposal | Three root sites; zero production coverage pending function ownership clearance |

The first five binary rows' only remainder after the 1,262 original logical/equality sites is the 190 assignment-value family. No arithmetic operator appears in those rows or the additional 204 measured roots. No unsupported arithmetic or representation is manufactured. Every outside-function file is named in its code group's commit message. No runtime C file changed, no code was copied from cohere, no PR was opened, and every push targets the owned branch.

Detailed commands, exact source types, CSVs, per-rule fixtures and replay JSON are in ASSIGNMENTS.md, root-remainders/BREAKDOWN.md and its group reports, object-assignment/REPORT.md, checked-non-null/REPORT.md, null-equality-boundary/REPORT.md and conditional-proposal/REPORT.md. Original binary validation is in LOWERING.md. Logs are tracked with .log.txt names.

The final oracle selection covers 64 owned registered fixtures and seven imported checked-TypeScript fixtures, plus direct compatibility, missing-target, count and boundary controls. Valid cases agree with source Node, the JavaScript IR backend, release native, ASan/UBSan and LeakSanitizer. Intentional non-null failure controls independently show source Node continuing, then hold both compiled backends to the explicitly imported exit-70 rule before RHS effects. All new registered positive fixture rows were refreshed; the final complete count check changes no row.

Final delivery commands (test output redirected, never piped):

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestMixedEqualityBoxedNullBoundary|TestNonNullAssignment|TestCheckedNonNull|TestNativeAgreesWithNode/internal/oracle/testdata/(assignment_value|assignment_object|assignment_non_null|binary_logical|binary_equality|non_null_checked|non_null_refuse_(null|undefined)[.]ts)' -count=1 -timeout 15m > /tmp/notyet-binary-delivery-oracle.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 15m > /tmp/notyet-binary-delivery-counts.log 2>&1
```

Oracle passes 14.107s; counts pass 24.798s. The latest lower package run after the null guard passes 16.519s. The checked non-null group ran full lower/IR/flow packages (40.862s/13.321s/83.617s), the relevant native ownership/dispatch tests (0.276s), command non-null/explain tests (1.818s), and vet (exit 0). Earlier final IR/JavaScript/flow/fresh/lower runs pass (8.385s/no test files/98.756s/59.924s/39.758s); native relevant tests pass 0.276s. No full repository gate ran. The initial exhaustive native package run was interrupted in its unrelated ASCII corpus and is not claimed to pass.

Every continuation mutant and its catcher:

| Mutant | Catcher |
| --- | --- |
| assignment double-right | Source Node/native stdout |
| assignment double-field-receiver | Source Node/native stdout |
| assignment double-element-receiver | Source Node/native stdout |
| assignment double-element-index | Source Node/native stdout |
| assignment old-local | Source Node/native stdout |
| assignment old-field | Source Node/native stdout |
| assignment old-element | Source Node/native stdout |
| assignment late-local-read | Source Node/native stdout |
| assignment lose-result-alias | TestAssignmentValuePreservesCycleChecks, wrongly admitted cycle |
| assignment remove flow tracking boundary | Independent tracked-read census, graph 1 versus IR 0 |
| object double-source | Node stdout in both backends |
| object skip-store | Node stdout in both backends |
| boolean-number and-as-or | Node output differences |
| boolean-number or-as-and | Node output differences |
| value-number and-as-or | Node output differences |
| value-number or-as-and | Node/compiled exit mismatch at checked union narrowing |
| string and-as-or | Node output differences |
| string or-as-and | Node output differences |
| string assignment double-right | Source Node/native stdout |
| string assignment old-string | Source Node/native stdout |
| checked target double-receiver | Source Node/native stdout |
| checked target double-index | Source Node/native stdout |
| checked target double-current/getter | Source Node/native stdout |
| checked target right-before-current | Source Node/native stdout and RHS-order control |
| checked target skip-presence-check | TestNonNullAssignmentChecksBeforeRight |
| optional-number accessor wrong unpack | Valid C, Source Node/native stdout |
| string getter discard ownership | LeakSanitizer, 220 bytes in three allocations |
| null equality remove boundary | Valid C, native exit 0 prints true:false against Node false:true |
| null equality drop same-representation undefined term | Same valid-C, exit-0 miscompile |

The original nine binary mutants are recorded individually in LOWERING.md: wrong `&&` branch, wrong `||` branch, NaN truthiness, empty-string truthiness, owned-left transfer (LeakSanitizer 88 bytes), JavaScript wrong operator, removed null selection boundary, wrong equality operand normalization and boxed pointer equality. The draft's double-condition and wrong-arm mutants fail Node stdout comparison. All sources were restored; no build-error kill is counted. Earlier compile-failing probe attempts are explicitly excluded in the relevant reports.

Replay progress after the requested non-null merge: getTypeAliasForTypeLiteral (checker.ts:11224:5) and writeLane (debug.ts:1198:13) have no remaining lowering findings in the selected attempted units. Debug 1145:21 and 1148:21 advance from assigning to a NonNullExpression to reading connectors, after its earlier declaration failed. The field target 3972:21 loses its target finding while structural calls remain. Uint16 constructor 10491:22 remains unsupported. Other examples retain their ordered next findings in committed JSON. This is lowering evidence from the checker-rejected census context, not proof that the complete entry program now passes checking or emits native code.

Remaining work is explicit. The 143 Identifier failures are missing registered bindings, not the 190 operator stops; the two measured examples still depend on earlier any/branded-string and declaration failures. Array-length assignment builder.ts:1424:16 needs a store and density/holes contract. Uint16Array is not an existing representation. Delete keeps fixed object shapes; yield/generators need suspended-frame ownership and cancellation. Mixed null selection/equality need null identity; NULL currently also represents undefined, and the guard mutants demonstrate that treating them as one would miscompile.

The three conditional-statement sites have a concrete tested proposal at a8ad5bf0. Its only dispatch edit is two lines in expressionStatement(), a function edited by the void-value worker at c2d177c7. The user's territory rule reserves another worker's lower function, so it is not applied. The proposal evaluates its condition once and executes only the chosen arm, passes Node/both backends/sanitizers, kills both mutants and has separate measured counts. Production fixture registration, normal counts refresh and after replays remain pending that specific ownership authorization. Delete/yield are not proposed for admission.

Toolchain setup: export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh; source /workspace/adamic-tools/env.sh. Original successful timing lines: Go ready 0.064s, Node ready 0.066s, clang ready 0.499s, markdown dependencies ready 1.226s, submodules ready 16.766s, build cache warm 221.332s, done 221.364s. nproc 5; cpu.max 400000 100000. Go 1.27.1, clang 20.1.8, Node v24.19.0. The exact original compiler-base SHA is 8cadb46576e2de70791bc60c1e48694cd70e5d87; raw table pin dc6b1529 and replay pin 9a1f14c5 are retained in BREAKDOWN.md. Current main efe9f404 is merged, alongside explicitly requested c41c0e06. No action remains blocked by automatic sandbox review; the broader checker-type-based undefined proposal was replaced with the approved existing-constant-only rule, as recorded in ASSIGNMENTS.md.
