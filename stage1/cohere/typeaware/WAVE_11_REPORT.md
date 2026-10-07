Built: prefer-regexp-exec, correctness-no-identical-branches and no-for-in-array as native Adamic rules.  
Commits: claim `4811ac7a` pushed before implementation; implementation `81bd2b16`, based on `0d540f413625f016f20fea39761c7b184f335de6`.  
Checks: byte oracle, ASan, UBSan, LeakSanitizer, bridge packages, checker questions, filtered Node oracle and vet passed.  
Mutants: three rule edits, three raw-question answers and released registry each caught by their intended comparison.  
Not covered: full repository test gate, complete upstream rule fixtures, JSX and cross-file static initializer folding; pinned cohere cannot gate `.a` sources.

## Selection and implementation

The claim documents the combined descending volume ranking, excluding all 26
base ports. Positions 31, 32 and 33 are the three rules above. All 263 distinct
fetched origin trees were audited for ports and claims; none needed skipping.
No pull request was opened.

Each rule lives in its own `.a` file. New checker questions are isolated in
`number_index_type.go`, `resolved_signature_equal.go` and
`regular_expression_syntax.go`, with matching Adamic helper files. Only three
switch registrations (six gofmt lines) change shared `facts.go`. The bridge
returns numeric-index presence, resolved-signature pointer equality, and raw
regular-expression syntax validity. Adamic makes the lint decisions, ranges,
fixes and suggestions. No protected compiler implementation file changed.
The regex syntax substrate preserves the pinned cohere engine; its internal Go
package cannot be imported from this module. See WAVE_11_NOTICE.md and the MIT
license beside the engine. One dependency registration was added to go.mod.

## Independent evidence

The Go oracle imports the three unchanged production cohere rules at
`715ba94f3608a6500086b1076ce5cb7e51b836db`. Both implementations independently
load the same roots and serialize complete findings, fixes and suggestions.
The compiler corpus is TypeScript v6.0.3,
`050880ce59e30b356b686bd3144efe24f875ebc8`, with all 77 roots from the frozen
compiler manifest. The repository corpus uses the original 287-root manifest,
without adding this wave's sources. Source hashes, manifests, logs, complete
gzipped Go/native/sanitized outputs and SHA-256 summaries are committed in
`validation-wave-11/`. Absolute output paths identify the execution workspace;
source manifests are relative.

| Population | Regexp | Array | Identical branches | Equal bytes |
| --- | ---: | ---: | ---: | ---: |
| 25 control roots | 31 | 9 | 13 | 22350 |
| Repository, 287 roots | 0 | 0 | 3 | 19747 |
| Compiler, 77 roots | 3 | 2 | 0 | 7268 |

Normal and sanitized native outputs matched Go byte for byte for all three
populations. Controls include flags, regex syntax and escapes, static bindings,
reassignments including destructuring, shadowing, constraints and intersections,
array-like objects, chained branches, comments, narrowed overloads, arithmetic,
spread calls, Unicode and CRLF. An initial shorthand-assignment discrepancy was
corrected by using raw symbol identity rather than the binding helper's
shorthand substitution.

Each mutant compiled, exited zero, emitted no stderr and changed genuine control
input diagnostics. Only the independent complete-byte comparison rejected it:

| Mutation | First unequal byte |
| --- | ---: |
| Regexp fix end +1 | 458 |
| Array finding end +1 | 13759 |
| Identical branches finding end +1 | 16454 |
| Signature identity always equal | 20441 |
| Numeric index always absent | 13756 |
| Regex syntax always valid | 1317 |

A query on a released program exits 70 with exactly
`adamic: panic: invalid or released checker handle`. An overlay mutant retaining
the released registry entry exits zero, so the required-panic check rejects it.
The bridge package's existing lifetime, buffer and sanitizer checks also passed.

## Commands and observed output

Run from the repository after sourcing `/workspace/adamic-tools/env.sh`.
All test output was redirected to log files.

```sh
bash cloud/setup.sh > /workspace/wave-11-logs/setup.log 2>&1
nproc
ADAMIC_WAVE_11_ARTIFACTS=/workspace/wave-11-logs/validation ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-11-typescript go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave11AgreementAndMutants$' > /workspace/wave-11-logs/agreement.log 2>&1
go test -count=1 -v ./bridge/tsgo/checker > /workspace/wave-11-logs/checker-questions.log 2>&1
go test -v -count=1 -timeout 15m ./bridge/tsgo/... > /workspace/wave-11-logs/bridge.log 2>&1
go test -v -count=1 -timeout 10m ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(sorting|string_index|functions|closures)\.a$' > /workspace/wave-11-logs/node-oracle.log 2>&1
go vet ./... > /workspace/wave-11-logs/vet.log 2>&1
gofmt -l bridge/tsgo/checker bridge/tsgo/regular_expression_syntax stage1/cohere/typeaware/wave_11_test.go stage1/cohere/typeaware/testdata/oracle_wave_11.go > /workspace/wave-11-logs/gofmt.log 2>&1
git diff --check > /workspace/wave-11-logs/diff-check.log 2>&1
```

Setup printed ready in 0s, cache warm in 86s, done in 86s. `nproc` printed 5;
the container CPU quota is four cores. Agreement passed in 99.095s. Checker
questions passed in 0.162s; a final post-import-format checker run also passed.
Filtered Node oracle passed in 12.533s, including generic_functions and
method_closures matched by Go's unanchored subtest regex and its one-byte mutant.
Vet, gofmt and whitespace checks produced no findings.

## Timing and limits

Three alternating Go/native process runs per population emitted the same full
diagnostics, not counts alone. Medians from `validation-wave-11/timing.json`:

| Population | Go | Native | Native / Go |
| --- | ---: | ---: | ---: |
| Repository | 132.984 ms | 263.424 ms | 1.98 |
| Compiler | 319.789 ms | 1889.704 ms | 5.91 |

These are observed wall times including program loading, traversal and output.
Native is slower here. Initial instrumented runs recorded 74 compiler bridge
queries and zero repository queries; phase timings are in agreement.log.
No performance improvement is claimed.

The pinned cohere CLI rejects `.a` selection. Scratch Go overlays allowing
nonstandard root extensions and specifying TS parsing got further, but the
pinned resolver still reports six TS2307 errors for `.a` imports and proposes
consistent-type-imports/prefer-template edits in prefer_regexp_exec.a. Therefore
cohere formatting/lint is not a passing gate for this wave. Overlays never
changed the submodule. Native compilation, independent oracle and sanitizers
are passing gates. The full repository gate and exhaustive upstream fixture
matrices were not run. Cross-file initializer folding and JSX have not been
verified; corpus agreement does not prove universal equivalence.
