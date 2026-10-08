Certified original function-like type parameters, call/new type arguments and JSX child arrays.
Commits: this group adds fixtures, oracle checks, measured counts and reports using existing adapters.
Commands: fifty-three fixtures and three mutants pass in 60.556s; complete JSX child arm checks pass in 11.416s; scoped counts pass in 35.744s; oracle vet passes.
Mutants: three reached element-kind replacements are caught by pinned stopping oracles in sanitized native, release native and JavaScript.
Not covered: SourceFile.amdDependencies intersection, remaining ranked pairs and production consumer/intrinsic totals.

The three pairs are FunctionLikeDeclaration.typeParameters, CallExpression | NewExpression.typeArguments and JsxElement | JsxFragment.children, accounting for nine original static candidate reads. Preparation validates every source span and original declared member type and emits 78 complete declarations. All original receiver union arms remain intact. The JSX fixtures select a JsxText element, while oracle field-set checks retain and verify every original JsxChild arm, including unread arms. No reduced JSX or function-like declarations replace the originals.

Fifty-three fixtures cover valid receiver arms, unread invalid arrays, wrong array and element kinds, reached pos kind and readiness failures, later unread invalid elements, map consumers and absent/undefined optional arrays. Thirty-eight finishing source controls and three finishing mutants pass native leak checks. Each mutant replaces only a reached element descriptor with number; its typeof probe finishes with Node output and no stderr. The unchanged exit-70 stopping oracle catches each mutant in all three modes.

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/original27/generate.cjs /workspace/lane2-original-pin
node stage3/interface-downcasts/lane2/original27/prepare.cjs /workspace/lane2-original-pin /workspace/lane2-original-declarations27
ADAMIC_ARRAY27_ORIGINAL_DECLS=/workspace/lane2-original-declarations27 go test ./internal/oracle -run '^(TestCheckedViewRanked27OriginalArrays|TestCheckedViewRanked27ArrayMutants)$' -count=1 -v
ADAMIC_ARRAY27_ORIGINAL_DECLS=/workspace/lane2-original-declarations27 go test ./internal/oracle -run '^TestCheckedViewRanked27OriginalArrays/jsx-children' -count=1 -v
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
ADAMIC_ARRAY27_ORIGINAL_DECLS=/workspace/lane2-original-declarations27 go test ./internal/oracle -run '^TestCheckedViewRanked27ArrayCounts$' -count=1 -v -args -update-counts
go vet ./internal/oracle
```

Logs are retained in original27/evidence. The required global count refresh fails in existing fixtures in 42.840s. The scoped refresh records fifty-three rows; removing them reproduces the previous table byte for byte. No whole package test or full gate ran. No production or cross-lane files changed.

This adds three pairs / nine original reads. Lane 2 totals are 199 pairs / 2948 reads, remaining 135 / 241 of 334 / 3189. Own fields remain 3 / 26 of 30 / 72. The lane 7 intersection handoff remains listed; configuration string arrays are next.
