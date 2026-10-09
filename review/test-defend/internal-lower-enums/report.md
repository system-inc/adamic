# Defense of internal/lower enum audit rows

Main: 7b9d4272c28f59530ab13daa5c49067e47933b06. Branch: test-defend/internal-lower-enums.

All three rows are defended against the current enabled package: each catches one targeted production mutant alone. The matrix discovers and runs all 276 current top-level tests. Each completed matrix has 273 passed rows, one failed row, and two skipped rows. The skipped rows are TestOriginalCycleLedger (opt-in upstream inventory) and TestOptionalWideningCensus (requires project configuration/output). Their responses remain unknown. MixedUnionContractGraph additionally skips its pending array-interface subcase. No assertion failed before the clean baseline's 90-second timeout; the 19 unfinished rows were rerun and all passed in 3.762 seconds. Mutant runs then completed without panic or timeout. No test, oracle, or harness was changed.

## Code and oracles

Code under test: Adamic Go lowering, especially enumDeclaration in enums.go, sameKeeping's callable ABI representation comparison in expression.go, and instantiateFunction's specialization cache in generic.go. The generic fixture is input, not code mutated.

TestConstEnumErasesRuntimeObject uses actual Node output agreement plus self-written IR invariants forbidding an E local and ObjectLiteral. TestFunctionValueUnionViewsStayNotYet uses self-written expectations of a NotYet error class, without checking a particular reason. TestGenericUnionFixtureHasSeparateInstances uses a self-written count of two function names prefixed describe_; it does not verify both specialization signatures separately.

## Coverage and semantic differences

Six separate clean runs generated profiles with -coverpkg=./internal/lower, -coverprofile, -count=1 and exact single-test selectors. coverage-differences.json lists every covered block absent from the subsumer profile: 80 for const enum, 66 for callable views, 266 for generic instances.

D1 drops the const-enum conditional early return at enums.go:221. The exclusive early return at lines 222-223 is precisely the erasure path. Ordinary enum construction remains unchanged, unlike the prior audit's broad registration flip. Node output still agrees, but the target's IR allocation check catches the runtime E binding.

D2 changes the first callable ABI guard's Union constant to Array at expression.go:305. The target feeds a union-parameter function into a scalar-parameter callable view. The subsumer covers enum slot views without that parameter-representation mismatch. Shared comparison code is relevant; the semantic input difference, and exclusive rejection block at lines 317-318, motivated the attempt. The union-result guard remains intact.

D3 drops the whole specialization cache-hit conditional at generic.go:96. The target repeats calls for two distinct union instantiations. Its cache reuse block at lines 97-98 is absent from the JSON-array refusal subsumer's profile. Repeated specialization creates four instances rather than two, while observable answers can remain the same. This proves the row guards reuse/count, not merely runtime output. It does not prove the two retained signatures differ correctly.

## Replay and validation

D1.diff, D2.diff and D3.diff are standalone unified diffs against the stated main. git apply --check succeeded for each. Each corresponding Go overlay was checked with go vet -overlay ... ./internal/lower/ before its matrix. These overlays replace only the production file with the exact standalone-diff result. There is no selector switch. All runs used their own ADAMIC_BUILD_CACHE_DIR. Source files in the checkout were never modified.

To replay a diff: apply it to this main, source /workspace/adamic-tools/env.sh, set TMPDIR to a writable scratch directory, give it a fresh ADAMIC_BUILD_CACHE_DIR, then run timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . with output redirected to a log. Restore the production file afterwards.

## Costs and limitations

Warm toolchain worked, so setup was skipped. npm ci in stage3/api completed before baseline; its log is included. nproc is 5. Initial disk checks showed 6.4 GB free on /tmp and 5.7 GB on /workspace. Removed only earlier /tmp/deletion-load scratch. /tmp's entire volume is 8.8 GB, so the requested 15 GB threshold cannot be met there; deleting /tmp scratch also cannot free the separate /workspace filesystem. No run encountered ENOSPC. Final free space was 6.3 GB /tmp and 5.5 GB /workspace.

The clean package timed out at 90.065 seconds, making recovery necessary. All unfinished rows passed the recovery selector; baseline and recovery logs are included. Compilation is included in command wall time, not in JSON package elapsed. Separate native products were not built by these three Go-lowering matrices. The three matrix test-binary times are recorded in matrix.json. The audit was a bounded 14-row matrix; this defense used all 276 discovered current top-level tests. The opt-in inventory rows remain unmeasured, so unique means unique among enabled current rows. No repo-wide uniqueness claim is made.

No brief wording prevented an attempt. The disk threshold was unattainable on the /tmp volume. The clean timeout cost one recovery run. The coverage requirement was useful here, especially cache reuse, but profiles do not establish exclusivity by themselves. None of these rows is left undefended. Owner caveat: the generic row counts instances by name and does not assert signature identity; the callable row accepts any NotYet reason. Those limitations do not erase the demonstrated catches.

Matrix test-binary seconds: D1=62.439, D2=64.216, D3=62.853.
