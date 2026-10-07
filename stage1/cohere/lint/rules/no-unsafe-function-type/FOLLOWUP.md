Built all twelve already-claimed rule implementations in .a: nine independent-parser ports and three JSX ports blocked on shared parser integration.
Commits: 481f9f0e foundation merge, twelve rule commits through 29d6dcee; all pushed on codex/lint-wave1-06.
Checks: 577 independent-parser cases, 79 Go-AST-projected JSX fixtures, per-rule throughput, registry tests, go vet and filtered Node oracle passed; compiler replay passed (3,213 rows, 119,493,276 identical bytes).
Mutants: one per rule, all twelve ran successfully and were caught solely by byte comparison on Node, emitted JavaScript and sanitized native.
Uncovered: independent JSX integration, thirteen additional JSX fixtures, one malformed-comment recovery fixture, full repository gate and raw Windows paths; no new claims taken.

This report supersedes the earlier reports for this unit. All source edits are in the twelve claimed rule directories. The dependency merge brings the shared .a loader and suggestion serialization; no shared harness, registration generator or compiler implementation was authored here. The original claims were already pushed before these implementations.

## Commits

```
29d6dcee Port no-before-interactive-script-outside-document with explicit JSX adapter refusal
5039f7cb Port no-css-tags with explicit JSX adapter refusal
d954d06a Port google-font-display with explicit JSX adapter refusal
04191a66 Restore no-useless-empty-export Adamic modules
5e98bc02 Restore no-unsafe-function-type Adamic modules
63f556c9 Port no-unnecessary-type-constraint in its own directory
637ec430 Port no-document-import-in-page in its own directory
72b13f2e Port require-description in its own directory
22e63745 Port tailwind-no-physical-direction in its own directory
7332a19a Port no-self-assign in its own directory
ab1fa1e7 Port no-restricted-properties in its own directory
4a97b3c9 Port no-inner-declarations in its own directory
481f9f0e Merge remote-tracking branch 'origin/codex/lint-harness-dot-a' into codex/lint-wave1-06
```

## Verification and reproduction

Source `/workspace/adamic-tools/env.sh`. Run `go run ./cmd/lint-registry` for generated registration. The owned supplementary Go test source is `testdata/comparison.go.txt`. It is added as a new virtual test file using a Go overlay, without replacing any existing shared file:

```json
{"Replace":{"/workspace/adamic/stage1/cohere/lint/wave06_candidate_test.go":"/workspace/adamic/stage1/cohere/lint/rules/no-unsafe-function-type/testdata/comparison.go.txt"}}
```

```bash
go test -overlay /tmp/wave1-06-finish-overlay.json ./stage1/cohere/lint -run '^TestWave06Candidates$' -count=1 -v > /tmp/wave1-06-candidates-corrected.log 2>&1
go test -overlay /tmp/wave1-06-finish-overlay.json ./stage1/cohere/lint -run '^TestWave06(Mutants|Throughput|CompilerThroughput)$' -count=1 -v > /tmp/wave1-06-mutants-throughput.log 2>&1
go test -overlay /tmp/wave1-06-finish-overlay.json ./stage1/cohere/lint -run '^TestWave06Throughput$' -count=1 -v > /tmp/wave1-06-throughput-positive-final.log 2>&1
go test -overlay /tmp/wave1-06-finish-overlay.json ./stage1/cohere/lint -run '^TestWave06Projected$' -count=1 -v > /tmp/wave1-06-jsx-projection-final.log 2>&1
WAVE06_PROJECTED_TIMING=1 go test -overlay /tmp/wave1-06-finish-overlay.json ./stage1/cohere/lint -run '^TestWave06Projected$' -count=1 -v > /tmp/wave1-06-jsx-throughput.log 2>&1
go test -overlay /tmp/wave1-06-finish-overlay.json ./stage1/cohere/lint -run '^TestWave06Refusals$' -count=1 -v > /tmp/wave1-06-refusals.log 2>&1
go test -overlay /tmp/wave1-06-finish-overlay.json ./stage1/cohere/lint -run '^TestWave06Compiler$' -count=1 -v > /tmp/wave1-06-compiler-run-serializer.log 2>&1
go test ./stage1/cohere/lint/registry -count=1 > /tmp/wave1-06-finish-registry.log 2>&1
go vet ./stage1/cohere/lint ./stage1/cohere/lint/registry > /tmp/wave1-06-finish-vet.log 2>&1
go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/wave1-06-finish-oracle.log 2>&1
```

Observed: 568 filename-preserving upstream fixtures plus nine owned witnesses yielded 179,492 identical bytes on Go, Node, emitted JavaScript and sanitized native (80.73s). All nine independent-parser mutants passed execution and failed comparison (367.61s). Registry tests passed in 0.081s, vet exited zero, the filtered external Node oracle passed in 12.593s. Full repository gate was not run.

Observed: 79 complete JSX upstream fixtures yielded identical results using an owned Go AST projection on Node, emitted JavaScript and sanitized native, with all three mutants caught (113.05s). The projection supplies only AST structure, source positions and decoded tokens; findings are produced independently by the Adamic rules. This proves rule logic against Go, but cannot prove Adamic's JSX parser. The owned copied Go oracle selects ScriptKind from the filename because the shared adapter hardcodes TS. Private JSX test drivers live inside no-css-tags/testdata.

Observed: the three production JSX rule guards reject `<div>{/* c */}</div>` with exit 70 and `NotYet: stage-1 JSX parser adapter` on source Node, emitted JavaScript and sanitized native (42.19s), rather than treating the shared parser's non-JSX nodes as a valid no-findings result.

Observed: full compiler/stage1 replay passed in 420.26s, yielding 119,493,276 identical bytes across Go, original Node, owned driver Node, emitted JavaScript and sanitized native.

Compiler corpus: TypeScript 6.0.3 commit 050880ce59e30b356b686bd3144efe24f875ebc8, 77 src/compiler files plus 280 stage1 .ts/.a files, 3,213 file/rule combinations. The owned compiler.a driver changes only the output escaping implementation to batch contiguous ASCII runs; parsing, settings, rules, traversal, fixes and formatting are otherwise the shared driver. It is checked against both the original Node driver and Go. Earlier full-corpus attempts using character-by-character output accumulation were stopped for native serialization cost and are not counted as passing.

## Mutants

Every mutation below compiled and executed successfully on source Node, emitted JavaScript and ASan/UBSan native before comparison found the mismatch. For the three JSX rules the AST projection was used, with the integration limitation above.

### google-font-display

```json
{
  "Name": "wrong-diagnostic-id",
  "File": "rule.a",
  "From": "'googleFontDisplayMissing', googleFontDisplayMissing",
  "To": "'googleFontDisplayNotRecommended', googleFontDisplayMissing"
}
```

### no-before-interactive-script-outside-document

```json
{
  "Name": "wrong-diagnostic-id",
  "File": "rule.a",
  "From": "'noBeforeInteractiveScriptOutsideDocument', noBeforeInteractiveScriptOutsideDocument",
  "To": "'wrongStrategy', noBeforeInteractiveScriptOutsideDocument"
}
```

### no-css-tags

```json
{
  "Name": "wrong-diagnostic-id",
  "File": "rule.a",
  "From": "'noCssTags', noCssTags",
  "To": "'wrongCssTags', noCssTags"
}
```

### no-document-import-in-page

```json
{
  "Name": "wrong-module-specifier",
  "File": "rule.a",
  "From": "node.text === 'next/document'",
  "To": "node.text === 'next/documents'"
}
```

### no-inner-declarations

```json
{
  "Name": "wrong-enclosing-scope",
  "File": "rule.a",
  "From": "let body = 'program';",
  "To": "let body = 'function body';"
}
```

### no-restricted-properties

```json
{
  "Name": "invent-unconfigured-restriction",
  "File": "rule.a",
  "From": "if(root < 0) { panic('invalid no-restricted-properties options'); }",
  "To": "if(root < 0) { panic('invalid no-restricted-properties options'); }\n        this.properties.set('bar', root);"
}
```

### no-self-assign

```json
{
  "Name": "report-left-reference",
  "File": "rule.a",
  "From": "if(left.text === right.text) { this.finding(r); }",
  "To": "if(left.text === right.text) { this.finding(l); }"
}
```

### no-unnecessary-type-constraint

```json
{
  "Name": "delete-parameter-name",
  "File": "rule.a",
  "From": "new SuggestionEdit(name.end, constraint.end, replacement)",
  "To": "new SuggestionEdit(this.context.start(nameIndex), constraint.end, replacement)"
}
```

### no-unsafe-function-type

```json
{
  "name": "wrong-global-type-name",
  "file": "rule.ts",
  "from": "name.kind === 'Identifier' && name.text === 'Function'",
  "to": "name.kind === 'Identifier' && name.text === 'NotFunction'",
  "File": "rule.a"
}
```

### no-useless-empty-export

```json
{
  "name": "omit-empty-export-fix",
  "file": "rule.ts",
  "from": "'uselessEmptyExport', message, 'fix'",
  "to": "'uselessEmptyExport', message, ''",
  "File": "rule.a"
}
```

### require-description

```json
{
  "Name": "accept-empty-description",
  "File": "rule.a",
  "From": "if(description !== '') { continue; }",
  "To": "if(description === '') { continue; }"
}
```

### tailwind-no-physical-direction

```json
{
  "Name": "wrong-logical-margin",
  "File": "rule.a",
  "From": "['ms-', 'me-'",
  "To": "['me-', 'me-'"
}
```

## Throughput

Findings per second, wall time including process startup, 500 repeated positive witness rows per rule. These are bounded witness benchmarks, not compiler-workload speed claims. Native here is optimized unsanitized; parity above is sanitized. no-restricted-properties uses an explicit decoded restriction of foo.bar. Settings in upstream replays use Go-decoded canonical options.

| Rule | Native | Node | Go |
|---|---:|---:|---:|
| @eslint-community/eslint-comments/require-description | 12904.24 | 1360.86 | 7066.47 |
| @next/next/no-document-import-in-page | 16301.52 | 1390.84 | 9918.27 |
| @typescript-eslint/no-unnecessary-type-constraint | 23140.75 | 1451.50 | 13174.00 |
| @typescript-eslint/no-unsafe-function-type | 14631.06 | 1163.28 | 13833.10 |
| @typescript-eslint/no-useless-empty-export | 13991.27 | 1483.73 | 13288.27 |
| no-inner-declarations | 16435.58 | 1505.53 | 12537.97 |
| no-restricted-properties | 9353.56 | 1643.59 | 13545.67 |
| no-self-assign | 21667.61 | 1760.06 | 14473.72 |
| structure/tailwind-no-physical-direction | 22211.68 | 2052.56 | 16896.18 |

All nine rules produced zero findings on the 77 compiler files, so compiler findings/second is zero for native, Node and Go. Their measured durations are preserved in followup-mutants-throughput.log (compiler throughput test passed in 75.37s).

Projected JSX upstream fixtures, sanitized native, startup and AST materialization included. Go parses the original source; the other backends receive Go AST geometry, so these rates compare different front-end work and are not independent-parser performance.

| Rule | Cases / findings | Sanitized native | Node | Emitted JS | Go |
|---|---:|---:|---:|---:|---:|
| google-font-display | 16 / 7 | 91.47 | 32.42 | 105.60 | 432.84 |
| no-css-tags | 16 / 6 | 84.04 | 36.47 | 76.56 | 412.18 |
| no-before-interactive-script-outside-document | 47 / 20 | 71.08 | 100.19 | 221.11 | 912.78 |

## Toolchain and limits

`bash cloud/setup.sh` first failed during a concurrent foundation merge: lint_test.go could not import internal/javascript, volumeGenerated and checkRecoveryRefusal were undefined, Descriptor.Module was absent, and the parser test main referenced missing TestClassMethodInterfaceAgreesWithNode. This was a mixed-revision build, an inference supported by the clean retry after the merge settled. The retry passed: Go ready 0s, clang ready 0s, Node ready 0s, submodules 1s, build cache warm 271s, done 271s. `nproc`: 5. Go 1.27.1, clang 20.1.8, Node 24.19.0. Both setup logs are retained.

Remaining shared gaps: JSX opening/attribute AST nodes are absent from the stage1 parser. google-font-display, no-css-tags and no-before-interactive-script-outside-document have complete rule logic but remain integration-blocked. Their conservative guard also refuses ordinary TypeScript containing `<` without JSX nodes; this explicit partial limitation needs replacement once parser support lands. The other ports' replay excludes exactly thirteen JSX fixtures (require-description 1, no-document-import-in-page 11, tailwind 1), recorded individually in the candidate log. One unterminated block-comment fixture is blocked because the shared Go oracle panics before comparison. Raw Windows filename spelling is normalized for Go's absolute-path requirement, so replay proves normalized path behavior rather than raw backslash spelling. Arbitrary raw configuration validation is not covered: the manifest protocol supplies canonical Go-decoded options.

All existing implementations and evidence were pushed through a098ea5b before the next-rule audit. Fetching every origin head yielded 350 remote refs and 413 unique claim blobs. Every one of the ordered 46 helper-ready rules and the inventory's 54 syntax-ready rules appears in an origin claim. No unclaimed candidate remains; no new claim was created. The selected claim reference for each rule is recorded in testdata/followup-claim-audit.log. Shared files were left untouched. Test output was redirected directly to logs, never piped.
