Built: regexp.simpleFold in regexp_simple_fold.a; four prerequisite edges across four rules, zero final blockers alone.
Commits: 005c910cb claim pushed before code; implementation SHA is in the delivery.
Checks: focused owned helper test PASS 10.656s, 1,195,698 private-Go/source-Node/emitted-JavaScript/sanitized-native queries; 8,047,532 identical bytes per runtime; focused vet clean.
Mutant: maximum_fold_selected compiles, finishes successfully with empty stderr on all three runtimes, and is caught only by the Go byte comparison.
Limits: pinned Go 1.27.1 Unicode 17 data, rune inputs are signed 32-bit integers; not four completed rules or a full repository gate.

The frozen readiness ledger names @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports. It removes four prerequisite entries; each rule retains other prerequisites. The unchanged original Go consumer families all passed earlier in this session on the same cohere/toolchain/base: see evidence/modifier-consumer-*.log.txt. No Go source changed between those runs and this helper comparison. The Adamic helper is compared independently, not substituted into Go rules.

The helper selects the least member of the Unicode.SimpleFold orbit exactly as private Go simpleFold does. The same file contains a sorted transition table and an internal binary lookup for the standard-library operation. The generator queries Go Unicode.SimpleFold on every Unicode value and stores all 2,994 nonidentity transitions. These are generated standard-library data, not copied answers for the private helper's minimum selection. Golden outputs independently call the unchanged private Go helper through an added export. Future Go/Unicode changes that affect any output fail the exhaustive comparison until the table is explicitly regenerated.

Inputs cover each integer from zero through 0x110000 inclusive, including all surrogate code points and one out-of-range value, plus signed-rune boundary controls (-2147483648, -1048576, -1, 1114113, 2097151, 2147483647). The seven real consuming test files contribute 5,755 string literals and all 81,579 runes in those literals, retaining duplicates and their order. Those are source-derived helper queries rather than an observation of runtime private helper calls. Invalid UTF-8 Go test literal bytes are normalized by Go's JSON string encoder; the adapter corpus is Unicode text. Negative and above-MaxRune controls agree with Go identity behavior, but the entire 2^32 signed-rune domain is not exhaustively enumerated.

Commands (each output directly to its corresponding evidence log):

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/helpers/from-wave1-10/testdata/generate_fold.py
go test -count=1 -v ./stage1/cohere/lint/helpers/from-wave1-10 -run '^TestSimpleFold'
go vet ./stage1/cohere/lint/helpers/from-wave1-10
```

The first driver compile failed because console.log takes a string. The driver now explicitly converts numeric results with String. That failure is retained in evidence/fold-tests-first.log.txt and is not a killed mutant. The corrected complete baseline and compiling mutant pass. No shared harness, compiler or original Go implementation was changed; all new Adamic files use .a.
