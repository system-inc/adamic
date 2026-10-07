Built: Unicode quoting for the constructed-context message component; the three JSX full-rule ports remain partial and blocked.
Commits: implementation bba112af7 on current origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06; evidence accompanies it on codex/typeaware-wave-30.
Commands and outputs: owned JSX component/prerequisite gates PASS 11.718s; vet and reproducible generation PASS; setup 29s, nproc 5.
Mutants: fragment object, component-name exclusion, function remedy and Unicode astral escape all exit successfully and are caught only by production-Go byte comparison.
Not covered: complete JSX findings/fixes/suggestions, requested full-rule corpora, full-rule timings, binding/stability analysis and shared visitor integration.

The only branch pushed by this unit remains based on current origin/main
c01907a70. The final origin refresh did not advance that base. The announced
ab70f38d4 shared harness exists but is not an ancestor of main; area/stage1-lint
also remains at c01907a70. No main or area branch is a push target, and no new
claims were made. The prior twelve-gate landing validation remains documented
in WAVE_30_REGEX_LANDING_REPORT.md; this change touches only owned component,
fixture, generator and private test files.

Previously the named construction message component refused all non-ASCII
names. It now walks UTF-16 scalar values, preserves printable Unicode, escapes
nonprintable BMP values with Go's four-digit \u form and astral values with
Go's eight-digit \U form, and retains the established ASCII special escapes.
Unpaired surrogates explicitly refuse with NotYet rather than pretending to
represent a valid Go UTF-8 identifier. Construction detection and memo-input
stability are still missing.

The printable predicate is a JavaScript RegExp literal generated once from
strconv.IsPrint in Go 1.27.1, Unicode 17.0.0. It uses explicit scalar ranges
rather than depending on differing host Unicode-property versions. Its owned
generator is excluded from normal Go builds and reproducibly regenerates the
literal byte for byte. The generator is not a copied lint decision, and no Go
regular expression exists in this source rule to translate; the existing shared
regex audit remains applicable. No hand-rolled regex matcher, bridge question,
shared compiler/parser/harness or registration edit was introduced.

The production-Go overlay still calls jsxNoConstructedContextValuesMessage,
jsxFragmentsNameIsFragment and isComponentName directly. Construction fixtures
now cover 22 names across ten construction kinds and named/unnamed forms,
440 cases and 148002 bytes. Added names include Korean, accented Latin, emoji,
embedded surrogate pairs, Unicode controls, nonbreaking space, unassigned BMP,
format characters, line separator and astral noncharacters. This is bounded
case coverage, not an exhaustive scalar-value test. The other component suites
remain 48 qualified-fragment cases (1215 bytes) and 1034 name cases (12037 bytes).
All exact bytes agree in native, ASan/UBSan/LSan native and emitted JavaScript
on Node. No checker handles are acquired in these components; unchanged bridge
released-handle proofs remain in the earlier landing evidence.

All mutants have exit zero and empty stderr, so only the independent Go byte
comparison rejects them:

| Mutation | First differing byte |
| --- | ---: |
| qualified fragment React object changed to Other | 98 |
| dash-containing name exclusion changed to underscore | 22 |
| function declaration remedy changed to class expression | 44006 |
| astral nonprintable escape changed from eight to four digits | 13368 |

An unpaired high-surrogate probe also confirms native refusal with exit 70 and
the exact NotYet message. It is a refusal check, not a comparison mutant.

The independent full-rule Go TSX positive controls still report for all three
rules, while native refuses fragments and missing components with expected
GreaterThanToken/got SlashToken at offsets 30/23, and constructed context with
expected GreaterThanToken/got Identifier at offset 42. These observations locate
the native JSX extraction blocker. They do not establish full-rule parity.
The numeric handed-node driver and registration integration also remain shared
prerequisites. Per instructions, shared files are left to their owners and the
partial work is pushed. The three older React HIR/SSA/capture-dependent claims
remain PARKED.

Commands write exclusively to retained logs:

- bash cloud/setup.sh: Go 0s, clang 1s, Node 1s, submodules 1s, cache warm 29s,
  total 29s; nproc 5, CPU quota 4 cores.
- source /workspace/adamic-tools/env.sh, then
  ADAMIC_WAVE30_JSX_COMPONENTS=/workspace/wave-30-unicode-final
  go test ./stage1/cohere/typeaware -run '^TestWave30Jsx(Components|Prerequisites)$'
  -count=1 -v: components PASS 5.26s; prerequisites PASS 6.46s; total 11.718s.
- go vet ./stage1/cohere/typeaware: PASS with empty output.
- go run the owned generate_printable.go, cmp with printable.a, and
  git diff --check: PASS with empty output.

Exact stdout/stderr streams are gzip-preserved with uncompressed SHA-256 hashes
in validation-wave-30-unicode/streams.json. Setup, gate, vet, fetch and generation
logs and base/literal provenance are retained there. Earlier unchanged gates
were not repeated after this isolated change; no full repository gate or whole
new-rule corpus comparison is claimed. Validation durations are not native/Go
rule-performance benchmarks. Those new whole-rule timings remain unavailable
until native JSX extraction is integrated. New Adamic files are .a.
