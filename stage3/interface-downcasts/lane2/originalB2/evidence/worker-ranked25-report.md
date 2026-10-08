Certified original accessor, call-signature and construct-signature parameter arrays and signature type-parameter arrays.
Commits: this group adds fixtures, oracle checks, measured counts and reports; existing production adapters suffice.
Commands: fifty fixtures and five mutants pass in 59.961s; scoped count refresh passes in 37.905s; oracle vet passes.
Mutants: one reached element-kind replacement per original pair is caught by its pinned stopping oracle in sanitized native, release native and JavaScript.
Not covered: SourceFile.amdDependencies intersection, later ranked pairs and production consumer/intrinsic totals.

The five pairs are AccessorDeclaration.parameters, CallSignatureDeclaration.parameters/typeParameters and ConstructSignatureDeclaration.parameters/typeParameters. They account for fifteen original static candidate reads. Preparation checks every original source span and declared type and emits 78 complete original declarations. AccessorDeclaration retains both complete GetAccessorDeclaration and SetAccessorDeclaration arms. ParameterDeclaration and TypeParameterDeclaration retain every original member, including unread members.

Fifty fixtures cover valid reads, lazy unread arrays, wrong array and element kinds, reached pos kind and readiness failures, a later unread invalid element, map consumers and absent or undefined optional type parameters. Twenty-five finishing controls pass native leak checks. Each pair has a number element negative with an exact message. Its mutant replaces only that reached element descriptor with number; typeof lets the mutated program remain valid C and finish with Node output. All five finishing mutants pass leak checks and are caught by the unchanged exit-70 stopping oracle in all three modes.

```sh
source /workspace/adamic-tools/env.sh
node stage3/interface-downcasts/lane2/original25/generate.cjs /workspace/lane2-original-pin
node stage3/interface-downcasts/lane2/original25/prepare.cjs /workspace/lane2-original-pin /workspace/lane2-original-declarations25
ADAMIC_ARRAY25_ORIGINAL_DECLS=/workspace/lane2-original-declarations25 go test ./internal/oracle -run '^(TestCheckedViewRanked25OriginalArrays|TestCheckedViewRanked25ArrayMutants)$' -count=1 -v
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
ADAMIC_ARRAY25_ORIGINAL_DECLS=/workspace/lane2-original-declarations25 go test ./internal/oracle -run '^TestCheckedViewRanked25ArrayCounts$' -count=1 -v -args -update-counts
go vet ./internal/oracle
```

Logs are retained in original25/evidence. The required global count refresh fails in existing fixtures in 38.305s. The scoped refresh records fifty measured rows; removing those rows reproduces the previous table byte for byte. No whole package test or full gate ran. No production or cross-lane files changed.

This adds five pairs / fifteen original reads. Lane 2 totals are 192 pairs / 2927 reads, remaining 142 / 262 of 334 / 3189. Own fields remain 3 / 26 of 30 / 72. The lane 7 handoff remains listed while independent ranked work continues with modifier and JSX arrays.
