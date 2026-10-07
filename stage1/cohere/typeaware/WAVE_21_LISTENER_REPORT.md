Built: exported numeric syntaxKinds declarations in all thirteen owned rule modules; no shared parser or driver edits.
Commits: the branch remains based on main e8ba3d5d; previous landing evidence is 8338fbb3; this continuation is committed separately.
Checks: TestWave21ListenerDeclarations PASS 22.997s, normal and ASAN; vet and gofmt passed; prior landing oracles remain recorded.
Mutants: enum listener 267 changed to SourceFile 307 compiles, exits 0 with empty stderr, and independent Go contract bytes catch byte 15.
Not covered: numeric parser fields, shared kind-indexed dispatch, handed-node callbacks, removal of existing string-kind reads/refetches, complete native React source lowering or a speed improvement. No new rules claimed.

Each owned module now exports `syntaxKinds: readonly number[]`, with the values below. The private test parses the production Go `rule.Listeners` composite literals using Go's AST, then resolves their keys through the pinned compiler's exported `ast.Kind` constants. It compares a native executable importing all thirteen declarations against that independent contract, and repeats under ASAN. The wrong-kind mutant is a compiled native mutation of the actual exported declaration, not a changed expected value.

| Rule module | Numeric SyntaxKind | Go listener names |
| --- | --- | --- |
| no_object_constructor | 215, 214 | NewExpression, CallExpression |
| no_mixed_enums | 267 | EnumDeclaration |
| correctness_no_uncleared_race_timeout | 214 | CallExpression |
| correctness_no_process_exit_after_output | 307 | SourceFile |
| correctness_no_discarded_outcome | 245 | ExpressionStatement |
| correctness_no_discarded_pure_result | 245 | ExpressionStatement |
| correctness_require_blocking_standard_streams | 307 | SourceFile |
| correctness_no_collection_misuse | 227, 213, 214 | BinaryExpression, ElementAccessExpression, CallExpression |
| no_obj_calls | 307 | SourceFile |
| no_promise_executor_return | 220, 254 | ArrowFunction, ReturnStatement |
| set_state_in_effect | 307 | SourceFile |
| static_components | 307 | SourceFile |
| set_state_in_render | 307 | SourceFile |

The three React rules listen on SourceFile because Go lowers each outermost function with its nested function arena. A function-kind listener would lower nested functions again without the enclosing bindings. These declarations faithfully describe the Go entry point, not a new per-function pipeline.

Observed blocker: `stage1/typescript/parser/nodes.ts` defines only `readonly kind: string`; its constructor also takes a string. There is no numeric SyntaxKind property on the node handed to a rule. The existing shared `Rules`, shape/binding helpers and private suites also use string kinds and index-based node access. Per the unit restrictions, this continuation does not edit that shared parser, registration generator, or harness. Numeric declarations are ready for the arriving shared driver, but existing source ports do not yet satisfy the new no-string-kind/no-refetch requirement. No speed improvement is claimed. Adding a numeric conversion that reads node.kind as a string inside each rule would violate the instruction, so no such adapter was added.

The existing React source-to-HIR lowering, SSA/graph transforms, compilation-unit and full memo annotations remain missing. Its current validator cores accept test-only Go-prepared HIR. Existing JSX refusals and the production Go two-creator phi nondeterminism remain recorded in WAVE_21_REACT_CORE_REPORT.md. Those three source ports remain partial; reservations are retained, and no next batch was taken.

The independent contract caught an implementation error before the final pass: counting enum declarations with a line regex omitted the commented DeferKeyword entry and shifted all chosen values by one. Both failed logs are archived; production Go constants exposed the discrepancy. The corrected declarations passed the compiled normal/ASAN contract and the qualifying wrong-kind mutant.

Command:

```bash
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE21_LISTENER_ARTIFACTS=/workspace/wave21-listeners-final go test ./stage1/cohere/typeaware -run '^TestWave21ListenerDeclarations$' -count=1 -timeout=10m -v > /workspace/wave21-listeners-final.log 2>&1
go vet ./stage1/cohere/typeaware > /workspace/wave21-listeners-vet.log 2>&1
gofmt -l stage1/cohere/typeaware/wave_21_listeners_test.go > /workspace/wave21-listeners-gofmt.log
```

This change adds constant declarations only and does not alter existing findings/fixes/suggestions logic. The previous landing run already exercised all six wave-21 suites, both frozen corpora, each native rule mutant, released handles and sanitizers (880.182s), inherited bridge suites (579.690s), checker (0.558s) and external Node oracle (36.207s), on the same unchanged main. Those complete findings suites were not repeated for the unused metadata declarations. No full repository gate or new throughput benchmark was run. Setup and nproc remain the prior verified 119s and 5 on this same base; WAVE_21_LANDING_REPORT.md records their log and exact commands.
