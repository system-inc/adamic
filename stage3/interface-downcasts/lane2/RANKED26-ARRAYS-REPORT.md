Certified original static-block modifier arrays and JSX type-argument arrays, including the HasTypeArguments union.
Commits: this group adds fixtures, oracle checks, measured counts and reports using existing production adapters.
Commands: fifty-six fixtures and four mutants pass in 67.088s; scoped count refresh passes in 44.961s; oracle vet passes.
Mutants: four reached element-kind replacements are caught by pinned stopping oracles in sanitized native, release native and JavaScript.
Not covered: SourceFile.amdDependencies intersection, remaining ranked pairs and production consumer/intrinsic totals.

The four pairs are ClassStaticBlockDeclaration.modifiers, JsxSelfClosingElement.typeArguments, JsxOpeningElement.typeArguments and HasTypeArguments.typeArguments. Preparation checks all twelve original source spans and full declared types, emits 78 complete original declarations and records complete fields for every receiver arm and reached structural element. The HasTypeArguments union is retained as originally declared; each arm has valid controls. ExportKeyword supplies the original modifier element tag and TypeReferenceNode the original type-node element tag, both read from the original checker.

Fifty-six fixtures cover valid reads, unread invalid arrays, wrong array and element kinds, reached pos kind and readiness failures, later unread invalid elements, map consumers and absent/undefined optional arrays. Thirty-six finishing source controls and four finishing mutants pass native leak checks. Each mutant changes only the reached element descriptor to number. The negative observes typeof, so it remains valid C, runs with Node output and no stderr, and is caught by the unchanged exit-70 stopping oracle in all three modes.

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/original26/generate.cjs /workspace/lane2-original-pin
node stage3/interface-downcasts/lane2/original26/prepare.cjs /workspace/lane2-original-pin /workspace/lane2-original-declarations26
ADAMIC_ARRAY26_ORIGINAL_DECLS=/workspace/lane2-original-declarations26 go test ./internal/oracle -run '^(TestCheckedViewRanked26OriginalArrays|TestCheckedViewRanked26ArrayMutants)$' -count=1 -v
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
ADAMIC_ARRAY26_ORIGINAL_DECLS=/workspace/lane2-original-declarations26 go test ./internal/oracle -run '^TestCheckedViewRanked26ArrayCounts$' -count=1 -v -args -update-counts
go vet ./internal/oracle
```

Logs are retained in original26/evidence. The required global count refresh fails in existing fixtures in 38.270s. The scoped refresh adds fifty-six measured rows; removing them reproduces the previous table byte for byte. No whole package test or full gate ran. No production or cross-lane files changed.

This adds four pairs / twelve original reads. Lane 2 totals are 196 pairs / 2939 reads, remaining 138 / 250 of 334 / 3189. Own fields remain 3 / 26 of 30 / 72. The lane 7 intersection list remains separate while function-like and JSX-child pairs proceed.
