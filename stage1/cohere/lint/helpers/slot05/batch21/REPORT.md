# Batch twenty-one report

Built joinRanges, caseExtras and buildCaseTables in separate .a files. Each serves @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports. Twelve prerequisite occurrences removed, zero final blockers; no rule is declared ported. Mapping is in readiness.json and CONSUMERS.md. Cumulative slot 05: 59 helpers, 331 occurrences across 70 consumers and 50 helper-ready rules under the frozen common-adapter assumption.

## Claim and landing

Claim 0ef2441ea13406a7a2f92747a57feff351d5ed69 was pushed before any source. All twenty origin codex/lint-helpers* branches and every claim tree were read before selection and refreshed before publication. The higher-count comments leaves remain owned by the shared bundle in HELPERS.md. These three concrete symbols tie the highest unclaimed fan-out at four consumers each. Final ownership.json confirms only slot 05 claims them. A preliminary assertion searched the bare dependency name caseExtras, found its mention as an external dependency in our earlier writeClass claim and stopped before modifying any file. Checking fully qualified reservation names resolved that false collision; no competing claim was found.

All earlier 56 helpers were ported, tested and pushed at d39aed21. The preceding landing unit reran all twenty actual packages and all 222 compiling semantic mutants after taking the changed compiler/runtime. Current main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06 and area d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898 remain unchanged ancestors. Compiler, pinned cohere, Node oracle, shared options reader, corpus/inventory and retained helper inputs remain byte-identical by Git object identity in evidence/retained-input-identity.json. Those earlier packages were not rerun again in this unit. No main or area push and no PR. No claimed helper remains unfinished.

## Observed comparisons

Actual cohere at 715ba94f3608a6500086b1076ce5cb7e51b836db is the oracle. joinRanges executes unchanged. The private overlay renames dependency calls only inside original caseExtras and buildCaseTables bodies. All renamed dependencies still execute actual Go: group selection, literalRune, Canonicalize and unicode.SimpleFold. Production cohere, shared harness, registry and compiler are untouched. These helpers transform syntax and construct case tables; no hand-written regex matcher is added.

All four consumers are scanned using Go's source parser. Captured string occurrences respectively: 383, 65, 480 and 1266; 904 distinct strings. Raw code-point atom sequences and actual classAtoms candidate-body decoding contribute samples; 1723 candidate bodies decode successfully across both Unicode modes. Targeted controls cover empty atoms, every atom kind, reversed/equal/supplementary/surrogate endpoints, set/dash boundaries, chained ranges, Unicode folding outliers and complete equivalence groups. **7379 distinct atom/mode samples per helper** for joinRanges and caseExtras. Test-file strings are helper inputs, not complete lint programs; they can include prose and incomplete programs, and a successful candidate body decode does not identify a real regex literal in source.

buildCaseTables uses all 336 actual Go Unicode CaseRanges and the three explicit out-of-range seed pairs. Four fresh builds run with flags false, true, true, false. Real Go dependency observations supply 5991 canonicalization pairs and 2990 fold mappings; they are not expected helper results. Comparisons include the complete resulting groups, member map and shared group identity, plus ordered canonical/fold call traces. In total: **14762 distinct helper invocations** across the three baseline corpora; the strengthened caseExtras rerun repeats its 7379 rows.

Source Node, emitted JavaScript and sanitized native match actual Go bytes. joinRanges compares exact ordered atom fields and complete error text. caseExtras compares exact expanded text and ordered group-selection/literal formatting calls. Go map traversal leaves table group enumeration unspecified: the test observation sorts groups by their first member, preserving the original member order and map-to-array identity. No canonicalization is performed on members or caseExtras text. caseExtras receives the exact Go group order captured by its own oracle run. The helper itself leaves group traversal in map order. Distinct oracle processes may capture distinct valid group orders; each raw corpus includes its own order and expected bytes.

## Commands and outputs

All output went directly to logs. Setup succeeded: Go 0s, clang 0s, Node 0s, submodules 0s, build cache 99s, total 99s. nproc 5, four-core cgroup quota, 17.6 GB. Versions Go 1.27.1, clang 20.1.8 and Node 24.19.0. Environment sourced from /workspace/adamic-tools/env.sh.

```sh
bash cloud/setup.sh > /tmp/lint05-batch21-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_SLOT05_BATCH21_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch21/evidence" ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch21 -count=1 -v -timeout=20m > /tmp/lint05-batch21-helpers.log 2>&1
ADAMIC_SLOT05_BATCH21_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch21/evidence" ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch21 -run '^TestCaseExtras$' -count=1 -v -timeout=20m > /tmp/lint05-batch21-order.log 2>&1
go vet ./... > /tmp/lint05-batch21-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05/batch21 > /tmp/lint05-batch21-format.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch21-oracle.log 2>&1
```

Complete package PASS 180.971s (join 25.32s, extras 136.83s, build 18.79s), including the first fourteen independent semantic variants. Added group-order reversal: caseExtras rerun PASS 146.120s, all five caseExtras variants caught. Fifteen distinct variants total. Vet and formatting pass with empty logs. Six filtered Node input fixtures PASS 0.938s uncached, zero cache hits and six probe misses. git diff --check passes. No selected check skipped, relaxed or removed.

## Every mutant

Each independent variant is a temporary source copy, compiles and runs successfully with native sanitizers, exits zero and writes no stderr before its output is compared to Go. A refusal, panic, sanitizer failure or warning never receives mutant credit. Exact first differing output line, byte and values for all fifteen variants are in evidence/mutants.json.

| Helper | Independent variants caught |
|---|---|
| joinRanges (4) | invert reversed-endpoint guard; consume one rather than two following atoms; omit Unicode set-endpoint rejection; copy the lower endpoint into the range upper bound |
| caseExtras (5) | widen groups with no covered member; append already covered members; invert Unicode group selection; overwrite rather than append literals; reverse group order while preserving group count |
| buildCaseTables (6) | omit duplicate suppression; omit Unicode fold traversal; discard pairs by requiring three members; replace the second seed members with first members; reverse numeric member sorting; give the member map copies instead of the shared group arrays |

The reversal and shared-array-copy variants separately hold order and identity rather than only presence/count. The no-orbit variant is caught by the actual dependency trace even where canonical seeding already produces the same group. Raw corpus and expected output are compressed losslessly with fixed gzip timestamps; corpus-manifest.json records original sizes and SHA-256 hashes. helpers.log contains the initial fourteen witnesses; order.log and mutants.json carry the final five caseExtras witnesses, including reversal.

## Findings and limits

Preliminary runs exposed an unused private Go import, MapIterator/Array.from narrowing, unsupported nested function declarations, an inferred never[] allocation and closure-cycle refusal when local callbacks capture supplied function values. The owned code uses explicit narrowing, typed fresh arrays, direct loops and top-level helper-local functions. No compiler or shared harness changes were needed; preliminary runs receive no pass credit.

Not covered: complete rule findings, positions, fixes, suggestions or options; shared registration integration; the external dependency implementations; arbitrary malformed adapters, nonclosing fold orbits, forged Unicode ranges, mutation of shared readonly arrays, concurrent cache initialization or exhaustive atom combinations. The port has no process-global table cache; its caller owns caching. Go's unspecified group-list order is not claimed identical across separate processes. The native driver uses value projections for class atoms rather than compiler AST nodes.

The full repository gate, including all seventeen required stage1 external-input checks, was not run. This is the bounded touched-package, repository-vet and filtered external-oracle gate authorized for the unit. No skip is credited as green. Only these three helpers are reserved; there is no fourth reservation.
