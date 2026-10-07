Built: regexp.modifierGroup in regexp_modifier_group.a; four prerequisite edges across four rules, zero final blockers alone.
Commits: efb00d444 claim pushed before code; implementation SHA is in the delivery.
Checks: focused owned helper test PASS 13.416s, 1,066,656 Go/source-Node/emitted-JavaScript/sanitized-native queries; 6,799,954 identical bytes per runtime; vet and diff check clean.
Mutant: duplicate_guard_disabled compiles, exits successfully with empty stderr on all three runtimes, and is caught only by comparison against private Go.
Limits: bounded flag grammar and source-derived inputs, not four completed rule ports or full repository validation.

The frozen readiness ledger names @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports. All four original Go test families pass: next 0.224s, TypeScript 0.076s, core exports 0.012s and core imports 0.130s. Those suites validate unchanged upstream consumers; the Adamic helper is compared independently rather than substituted into Go rules.

The adapter encodes Go rewriteOptions as an integer mask from 0 through 15: multiline=1, dotAll=2, ignoreCase=4, unicode=8. The result carries options, opener byte width and match status. Rejection returns the incoming options unchanged. Successful openers consist entirely of ASCII, so the UTF-16 colon offset equals the Go UTF-8 byte offset there; any non-ASCII flag rejects with width zero. Payload after the first colon is untouched. Duplicate flags across both sides reject, a second dash rejects, an empty flag set rejects, and a trailing dash after named flags is accepted exactly as Go does. The input mask's adapter domain is the sixteen possible Go boolean vectors.

Go parser/strconv.Unquote extracted 5,755 string literals from all seven test files spanning those four consumers. Each literal and every suffix opening with (? is included. Exhaustive words over ims- through length seven are tested as openers with and without a colon and with ignored payloads, plus invalid flags, NUL, whitespace, BMP/supplementary Unicode and unpaired-surrogate controls. Deduplication yields 66,666 sources, each tested under all sixteen incoming states. These are source-derived helper inputs, not a measurement of private helper calls during consumer execution. Expected bytes come from an added export that calls the unchanged private Go implementation in a scratch overlay.

Commands (each output directly to its corresponding evidence log):

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/helpers/from-wave1-10/testdata/generate_modifier.py
go test -count=1 -v ./stage1/cohere/lint/helpers/from-wave1-10 -run '^TestModifierGroup'
go vet ./stage1/cohere/lint/helpers/from-wave1-10
# In cohere/, each original consumer family:
go test -count=1 -v ./internal/lint/rules/core -run '^TestNoRestrictedExports'
go test -count=1 -v ./internal/lint/rules/core -run '^TestNoRestrictedImports'
go test -count=1 -v ./internal/lint/rules/typescript -run '^TestNoEmptyObjectType'
go test -count=1 -v ./internal/lint/rules/next -run '^TestNoHtmlLinkForPages'
```

The first generation attempt passed a .go.txt template directly to go run and failed before producing inputs; its empty-input test failures are retained. The generator now builds that template as a .go file in scratch, and the complete baseline and mutant rerun passes. No compilation failure or runtime error is credited as a killed semantic mutant. No shared lint harness, compiler or Go rule source was edited. All new Adamic sources are .a.
