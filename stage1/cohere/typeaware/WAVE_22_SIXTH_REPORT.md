Built: three native JSX rule listeners with numeric dispatch; three React graph claims parked.
Commits: claim fc25360e9 pushed before code; implementation rebased as 1203fc718; landing ancestry ed149b7cf.
Commands and outputs: all 15 rule oracles match Go, normal and sanitized; twelve-port landing refresh PASS 703.509s.
Mutants: three native rule mutants caught only by byte comparison; raw syntax kind mutant caught by the bridge fact test; common bridge mutants caught.
Not covered: shared JSX/harness integration, emitted JavaScript for these listeners, React HIR/SSA analysis and the complete repository gate.

## What is implemented

The sixth batch ports react/jsx-fragments, react/jsx-no-undef and
react/no-adjacent-inline-elements in separate rule.a files. Each rule.json
lists kind names and pinned numeric SyntaxKind values. The isolated driver
caches parser nodes once and dispatches only declared subscribers. Visitors
receive their node and compare numeric kinds. The adapter converts the
current parser's textual kinds once outside the visitors; kindName is retained
only for the existing bridge ABI. No rule refetches its subject from the parser.

The new symbol-declaration-syntax bridge question lives in dedicated Go and
Adamic files. Its sole registration change is one case line in facts.go.
It exports declaration trees, parent links, numeric kinds and raw initializer
roles. All React source/binding decisions happen in Adamic. No Go rule verdict,
React HIR or capture analysis is imported by the native rules.

Following Ahra's instruction, react-hooks/set-state-in-effect,
react-hooks/set-state-in-render and react-hooks/static-components are parked
with their native HIR/SSA/capture/value-propagation/dominator blocker named in
the claim. Their earlier reproduction remains in WAVE_22_FIFTH_REPORT.md.
react/jsx-no-constructed-context-values was skipped because its stability
analysis follows captured and aliased holders, the same missing prerequisite.
The fetched-ref and selection inventory is validation-wave-22-sixth/selection.json.

## Dependency and exact integration blocker

Main does not yet contain the published JSX parser. The validator copies only
parser/scanner sources from origin/codex/stage1-jsx-lint at
**a8a62d62ca49db7415e14c3887dd305022b17309** into a scratch source tree and builds
the actual native listeners there. This is a native JSX parse, never injected
Go syntax. No shared parser, scanner, generator or harness file is edited.

A direct build against the unmodified repository parser exits 1:
`stage1/typescript/parser/parser.ts:34:20: Adamic 0.1 refuses this escaping a constructor before every field is set`.
That check prevents reaching a JSX runtime test on current main. The earlier
fifth-batch reproduction independently records the missing JSX parse support.
The published dependency builds and passes native/sanitizer comparisons.
Therefore these are implemented and validated ports awaiting the shared parser
and harness landing, not claims of working through the current shared lint CLI.
Shared registration/profile adapters, configuration decoding and emitted
JavaScript comparison are not changed or claimed complete. No new Diagnostic
landing SHA has been supplied by Ahra; this uses the existing proven finding model.

## Landing readiness

During final validation main advanced from f8013f0ba to
**c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06**. The branch's first-parent rule history
was rebased onto that main; the published fc25360e9 ancestry was retained with
an ours merge, permitting a normal fast-forward branch push. Comparing all
previously tracked stage1, bridge, internal, cmd, toolchain and instruction
source hashes across the rebase found zero changed inputs. Main's additional
internal/oracle/stage3_hook_test.go concerns stage 3 and is outside the worker gate.
The twelve previous rule oracles passed on the rebased branch in 703.509s
(fourth 172.04s, continuation 253.46s, original 159.53s, third 118.46s).
Only codex/typeaware-wave-22 is pushed; main and area branches are untouched.

## Byte comparison and positive controls

The independent Go executable runs the unmodified production rule.Run methods,
uses its own program loader and numeric listener walk, and imports no bridge
code. Complete canonical output includes findings, spans, fixes and suggestions.
All three production rules intentionally emit no fixes or suggestions; the
fragment rule keeps the upstream type-argument safeguard rather than adding
an unsupported rewrite.

Go's parser accepted 176 controls and rejected 188 extracted strings or
malformed inputs explicitly. Inputs are string literals extracted from the
three production test files plus Unicode/CRLF, nested members, imported and
destructured aliases, globals, type arguments, React.createElement shapes,
U+0085 and U+FEFF whitespace controls. This extraction is not a claim that the
native driver reproduces every production test's option/fixture construction.
All three rules have positive findings in the default control output: 129 total
findings, versus 106 in element mode and 126 with allowGlobals.

| Comparison | Exact output bytes |
| --- | ---: |
| Default controls | 52,437 |
| Fragment element mode | 47,114 |
| Allow globals | 51,412 |
| Frozen repository, 287 files | 18,485 |
| Frozen TypeScript compiler, 77 files | 5,857 |

Default controls and both frozen corpora match again under ASan, UBSan and
LSan. Option comparisons are normal native runs. Both corpora have zero
findings for these three rules, so positive controls carry the rule coverage.
Cohere is pinned to 715ba94f3608a6500086b1076ce5cb7e51b836db;
typescript-go to 8d550c837c90bd1805b047b7eeccc2baac2d5e7a;
the TypeScript corpus is 050880ce at /workspace/wave-22-typescript-pinned.
No whole-TypeScript repository or React HIR analysis coverage is claimed.

## Mutants and handle checks

| Mutant | Check that catches it |
| --- | --- |
| Fragment preferFragment id changed to preferPragma | Complete Go/native byte comparison |
| Undefined-tag diagnostic id changed | Complete Go/native byte comparison |
| Adjacent-inline predecessor condition inverted | Complete Go/native byte comparison |
| Declaration bridge numeric kind erased | TestSymbolDeclarationSyntax raw AST identity assertion |

The twelve previous per-rule mutants were caught again by complete byte
comparison: numeric-prefix, has-own-fix-span, spread-parens, promise-condition,
spread-await-edit, lost-write-span, nullish-suggestion, qualifier-fix,
private-read, global-provenance, global-declaration-span and timer-string.
The additional scope-export-mutant also exits 0 with empty stderr and differs
at byte 26,547. The old bridge registry-retention mutants survive execution
but fail the required released-handle refusal; each emitted panic 70 positive
control is recorded in landing-oracle.log. The filtered Node one-byte mutant
is caught by its independent oracle assertion.

Every native rule mutant compiles, exits 0 and has empty stderr; only the
comparison fails. The bridge mutant builds and fails the targeted assertion,
not compilation. The direct declaration fact test also checks unresolved
bindings, initializer roles, source identity and refusal of a suffixed question.
The new symbol-declaration-syntax and reused node-symbol-details questions
both refuse a released handle with exit 70 under sanitizers. The complete
bridge suite catches retained released handles, altered type extraction,
unlinked checker calls, input/output length defects, leaked output buffers
and region ownership defects. Logs preserve the exact observations.

## Commands and logs

Toolchain setup from this continuing unit is reused: bash cloud/setup.sh
printed go ready 0s, clang ready 0s, node ready 0s, submodules ready 0s,
build cache warm 118s, done in 118s; nproc is 5 (4-core cgroup quota).
Every toolchain command sources /workspace/adamic-tools/env.sh. Test stdout
and stderr are written directly to files. No test process is piped.

```
go build -o /workspace/wave-22-sixth-adamic ./cmd/adamic
go build -buildmode=c-archive -o /workspace/wave-22-sixth-checker.a ./bridge/tsgo/archive
python3 stage1/cohere/typeaware/rules/wave-22-sixth/validate.py
python3 stage1/cohere/typeaware/rules/wave-22-sixth/checks.py
python3 stage1/cohere/typeaware/rules/wave-22-sixth/benchmark.py
go test ./bridge/tsgo/checker -run '^TestSymbolDeclarationSyntax$' -count=1 -v
go test -overlay /workspace/wave-22-sixth-work/question-mutant-overlay.json ./bridge/tsgo/checker -run '^TestSymbolDeclarationSyntax$' -count=1 -v
ADAMIC_TSGO_CORPUS=/workspace/wave-22-typescript-pinned go test ./bridge/tsgo/... -count=1 -timeout=15m -v
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware
go test ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/(classes|library_array_metadata|library_map_set)\.a$' -count=1 -timeout=10m -v
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -timeout=10m -v
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-22-typescript-pinned go test ./stage1/cohere/typeaware -run '^TestWave22(AgreementAndMutants|NextAgreement|ThirdAgreement|FourthAgreement)$' -count=1 -timeout=30m -v
# In cohere:
go test ./internal/lint/rules/react -run '^(Test(Decode)?JsxFragments|TestJsxNoUndef|TestNoAdjacentInlineElements)' -count=1 -timeout=10m -v
```

Bridge PASS 117.989s; checker PASS 0.884s; focused fact PASS 0.078s;
production Go rule tests PASS 0.129s; touched-package vet exits 0.
Filtered Node/native/JavaScript oracle: three class/array/map fixtures PASS
28.326s, with Node observations fresh and ordinary worker caches allowed.
Its independent one-byte mutant check PASS 0.432s (native and Node cache misses).
The bridge run precedes the final rebase; its complete production inputs are
byte-identical across that rebase. All fifteen native rule oracles are rerun
on the rebased tree. Old test artifact directories use the explicit
ADAMIC_WAVE22{,_NEXT,_THIRD,_FOURTH}_ARTIFACTS environment variables recorded
in the evidence. The filtered worker gate replaces the full repository gate.

The pinned cohere CLI rejects the three .a paths as not TypeScript/JavaScript,
exits 1 and says nothing to check. This pre-existing extension gap is preserved
in cohere.log; no lint/format pass is claimed.

## Native timing

Three alternating complete-output runs compare bytes on each round.
Repository median: native 0.738828s, Go 0.224176s, native 3.296x slower.
Compiler median: native 3.948646s, Go 0.565889s, native 6.978x slower.
These whole-manifest measurements include parsing and checker load, and were
recorded on the rebased tree after the other gates finished.
Findings per second is undefined for these zero-finding corpora. The speed
requirement is not met; numeric visitor dispatch alone does not resolve the
remaining native parsing/adapter cost. That last explanation is an inference,
not a measured breakdown.

Committed evidence under validation-wave-22-sixth preserves canonical gzip
streams, stderr, subprocess statuses/times, controls, manifests, source hashes,
mutant first-difference positions and the final landing logs. Build archives,
executables and generated C remain outside Git under /workspace.
