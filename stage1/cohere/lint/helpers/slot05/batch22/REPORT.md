# Batch twenty-two report

Built decodePropertyEscape, decodeNumericEscape and readClass in separate .a files. Each serves @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports. Twelve prerequisite occurrences removed, zero final blockers; no rule is declared ported. Exact mappings are in readiness.json and CONSUMERS.md. Cumulative slot 05: 62 helpers, 343 occurrences across 70 consumers and 50 helper-ready rules under the frozen adapter assumption.

## Claim and landing

Claim 78a3fa2c7c7256c939bfcd3be15915f6937aae99 was pushed before source. All twenty origin codex/lint-helpers* branches and every claim tree were read before selection and refreshed before publication. Higher-count comments leaves remain owned by the shared bundle. These helpers tie the highest unclaimed concrete fan-out at four consumers each. Final ownership.json confirms unique slot-05 ownership.

All earlier 59 helpers were complete and pushed at 6ea2df893f546dd021012f8293b90dafcc3724f1. Main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06 and area d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898 remain unchanged ancestors. Compiler, pinned cohere, Node oracle, shared options reader, corpus/inventory and retained helper inputs remain byte-identical by Git object identity in evidence/retained-input-identity.json. The 222 earlier all-package landing witnesses plus fifteen batch21 witnesses remain applicable; earlier packages were not rerun again this unit. Only codex/lint-helpers-05 is pushed; no main/area push and no PR. No unfinished claim or fourth reservation.

## Observed comparisons

Actual cohere 715ba94f3608a6500086b1076ce5cb7e51b836db decides the helper behavior. The private overlay renames dependency calls only inside the original three Go bodies. Actual Go identityEscape, strings.IndexByte, isDecimalDigit, decimalEscape, decodeLegacyOctal and utf8.DecodeRuneInString still execute and record argument positions and call order. Production cohere, shared harness, rule registry and compiler are untouched. These are syntax helpers; no regex matcher is introduced.

The oracle parses all four consumer test files with Go's source parser. Their captured string occurrences are 383, 65, 480 and 1266 respectively, yielding 904 distinct consumer strings. It selects every p/P byte and digit byte for escape helper calls, and every opening-bracket tail for readClass. Source projections are deduplicated independently from context rows. These are bounded helper inputs, not complete rule programs or a declaration that every selected marker is a real regex token. Test strings also contain prose and incomplete programs.

Property and numeric rows run with Unicode on/off, inside/outside classes and group counts 0, 1, 9, 100 and 1048576. Named is false, since these helpers do not interpret named groups for these markers. Controls include empty input, omitted property delimiters, empty and malformed property bodies, adjacent braces, standalone NUL, octal limits, 8/9 fallback, exact group-number boundaries, decimal overflow saturation, escaped brackets, trailing slash, negation, supplementary runes, multibyte BMP runes and explicit invalid UTF-8 byte sequences.

Final inputs: **18600 property rows (604 source projections), 5440 numeric rows (88 source projections), 81 class rows (81 projections): 24121 helper invocations**. A string with no relevant marker contributes no call to that helper. Property marker size is its parser-produced one-byte width. All class body bytes are compared as hex, including malformed UTF-8; Go decoder sizes are supplied per byte offset, including zero at EOF. UTF-8 offsets and widths remain bytes. Finding-position conversion is outside these helpers.

Source Node, emitted JavaScript and ASan/UBSan native compare byte-for-byte with actual Go. The comparator holds every escape field, exact error text, class body bytes, negation and byte width, plus dependency call order and positions. Dependency results are exported by real Go; they are not computed from the port. Argument-only mutants retain those row-supplied dependency results and are caught independently by traces.

## Commands and outputs

All test output went directly to logs. Setup succeeded: Go 0s, clang 1s, Node 1s, submodules 1s, build cache 39s, total 39s. nproc 5, four-core cgroup quota, 17.6 GB. Versions Go 1.27.1, clang 20.1.8 and Node 24.19.0. Sourced /workspace/adamic-tools/env.sh.

```sh
bash cloud/setup.sh > /tmp/lint05-batch22-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_SLOT05_BATCH22_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch22/evidence" ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch22 -count=1 -v -timeout=20m > /tmp/lint05-batch22-complete.log 2>&1
go vet ./... > /tmp/lint05-batch22-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05/batch22 > /tmp/lint05-batch22-format.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch22-oracle.log 2>&1
```

Final strengthened package **PASS 45.501s**, all seventeen independent compiling semantic variants caught. The preceding result-only corrected run passed 38.066s; trace-strengthened run passed 38.569s before adding the two argument-only variants. Vet and formatting pass with empty logs. Six filtered Node input fixtures PASS 0.857s uncached, zero cache hits and six probe misses. git diff --check passes. No selected check skipped, relaxed or removed.

## Every mutant

Each independent variant is a temporary source copy, compiles and runs successfully under native sanitizers, exits zero with empty stderr, then disagrees with Go. Refusal, panic, sanitizer failure or compiler warning never earns semantic-mutant credit. Evidence/mutants.json retains every exact first differing line and pair of values.

| Helper | Independent variants caught |
|---|---|
| decodePropertyEscape (5) | invert Annex B identity delegation; invert opening-brace guard; omit unclosed-property rejection; reduce width by one; pass the next byte position to closing-brace lookup |
| decodeNumericEscape (7) | invert standalone-NUL next-digit guard; invert outside-class backreference gate; reject exact group-number equality; omit Unicode rejection; omit legacy-octal return; replace 8/9 literal value with zero; pass the next byte position to decimal decoding |
| readClass (5) | change negation marker to underscore; fail to consume the decoded escaped rune; reduce closing width by one; include the caret in a negated body; retain negation on failure |

The brace and decimal argument variants preserve row-supplied result values and are caught only by the dependency-position traces, proving those checks independently of the result comparator. Raw final corpora and expected bytes are losslessly compressed with fixed gzip timestamps; corpus-manifest.json records original sizes and SHA-256 hashes.

## Findings and limits

The first gate passed property and numeric parity/mutants but failed in the private oracle before calling readClass on empty input: it sliced source[1:] unconditionally to prepare a property-only dependency. That setup now occurs only for property mode. The failed gate remains in evidence/helpers.log without pass credit. Explicit malformed UTF-8 controls were added, all three helpers rerun successfully, then argument/call-order checks were strengthened and independently held. No shared file was edited.

Not covered: complete rule diagnostics, positions, fixes, suggestions, options or shared registry integration; external dependency implementations; arbitrary forged indices or non-parser marker sizes, arbitrary callback side effects, named-group variations, exhaustive byte strings or full regex syntax validation. The byte-array API adapts Go's string representation without changing widths; callers must supply the actual UTF-8 projection and matching slice/decoder callbacks. Numeric and property input offsets must be valid and identify the expected marker.

The full repository gate, including the seventeen required stage1 external-input checks, was not run. This is the authorized bounded touched-package, repository-vet and filtered external-oracle gate. No skip is counted as green. Only these three helpers were reserved.
