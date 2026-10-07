# Batch twenty-three report

Built decodeEscape, classAtoms and RegExp.Test in separate .a files. Each serves @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports. Twelve prerequisite occurrences removed, zero final blockers; no rule is declared ported. Exact mappings are in readiness.json and CONSUMERS.md. Cumulative slot 05: 65 helpers, 355 occurrences across 70 consumers and 50 helper-ready rules under the frozen common-adapter assumption.

## Claim and landing

Claim 2a1ebb368df8893c4ea1e6285c2549910aae8d2b was pushed before source. All twenty origin codex/lint-helpers* branches and every claim tree were read before selection and refreshed before publication. Comments leaves remain owned by the shared bundle. These helpers tie the highest unclaimed concrete fan-out at four consumers each. Final ownership.json confirms unique slot-05 ownership.

All earlier 62 helpers were complete and pushed at 2756d2ec4f238ff8151f63a75b364ee2800a054c. Main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06 and area d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898 remain unchanged ancestors. Compiler, pinned cohere, Node oracle, shared options reader, inventory/corpus and retained helper inputs remain byte-identical by Git object identity in evidence/retained-input-identity.json. All 254 preceding semantic witnesses remain applicable; earlier packages were not rerun again this unit. Only codex/lint-helpers-05 is pushed. No main/area push or PR, no unfinished claim or fourth reservation.

## Observed comparisons

Actual cohere 715ba94f3608a6500086b1076ce5cb7e51b836db decides helper behavior. The private overlay changes dependency call names only inside original decodeEscape, classAtoms and RegExp.Test bodies. Dependencies still execute real Go and return actual data. Nested specialized calls are muted where the whole decodeEscape function is classAtoms' external dependency; engine compilation is muted outside RegExp.Test's MatchString boundary. Production cohere, shared harness, rule registry and compiler are untouched. No matcher is added.

The oracle parses all four consumers' test files with Go's parser. Captured string occurrences are 383, 65, 480 and 1266 respectively, yielding 904 distinct consumer strings. It tests every position immediately after a backslash plus EOF for decodeEscape, and each whole string as a candidate class body for classAtoms. Contexts cover Unicode, incoming inClass, named-group presence and group counts 0/9. Class options also cover ignoreCase. Multiline/dotAll are false. These are bounded helper inputs, not complete rule findings or a claim that every selected string is a regex token. Strings include prose and incomplete programs.

Controls include all 256 byte values after a backslash, trailing escapes, every fixed control/set/boundary/named branch, UTF-8 supplementary and malformed bytes, Unicode/hex/control/property/numeric forms, empty bodies, dashes, reversed ranges and word/nonword expansions. Final decoder/class source projections: 1182. **23520 decoder rows; 37824 class-atom rows.** The operation pool has 2009 decoder and 5416 class dependency sets, preserving every row and every ordered expected trace. Ordinary decoded runes are supplied from actual Go per-byte decoding rather than a ported Unicode decoder.

RegExp.Test compiles every captured/control pattern with actual Go and tests seven subjects: empty, a, A, abc, supplementary Unicode, digits and the pattern itself. The 98 rejected pattern attempts are retained as actual nil-wrapper Test calls, not omitted cases. Extra controls separately cover a nil wrapper, a wrapper with nil engine, a controlled dependency returning matched true plus error, and a real Go engine timeout with a one-nanosecond budget on an exponential pattern. Final wrapper rows: **8313**, including error controls whose observed matched/error values are saved in engine-error-controls.json. Total: **69657 helper invocations**. Repeated source/control rows are not claimed unique.

Source Node, emitted JavaScript and ASan/UBSan native match actual Go bytes. The comparator holds every escape field/error, class atom and raw set-byte projection, nil-versus-empty arrays, exactness, wrapper verdict, and dependency arguments/order. Engine handles preserve the actual receiver identity checked by the Go wrapper; subject bytes are reversibly projected as hex and passed unchanged by the port. Native execution holds the Test wrapper against supplied engine results; it does not implement or validate the external matching engine itself. Pattern compilation and timeout execution occur in Go, not native.

Class joining and specialized dependencies return their real observed values. The driver records the arguments constructed by the port before replaying those results. A mutant with an unobserved dependency key gets a neutral result with positive progress and a differing trace, preventing panic/hang credit. The argument trace therefore holds class atom construction independently of the delegated join result. No fallback exists in production helpers.

## Commands and outputs

All output went directly to logs. Setup succeeded: Go 0s, clang 0s, Node 0s, submodules 0s, build cache 40s, total 40s. nproc 5, four-core cgroup quota, 17.6 GB. Versions Go 1.27.1, clang 20.1.8 and Node 24.19.0. Sourced /workspace/adamic-tools/env.sh.

```sh
bash cloud/setup.sh > /tmp/lint05-batch23-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_SLOT05_BATCH23_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch23/evidence" ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch23 -count=1 -v -timeout=20m > /tmp/lint05-batch23-complete.log 2>&1
go vet ./... > /tmp/lint05-batch23-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05/batch23 > /tmp/lint05-batch23-format.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch23-oracle.log 2>&1
```

Final strengthened package **PASS 153.808s**, all fifteen independent compiling semantic variants caught. Vet/format pass with empty logs. Six filtered Node input fixtures PASS 0.987s uncached, zero cache hits and six probe misses. git diff --check passes. No selected check skipped, relaxed or removed.

## Every mutant

Each independent variant is a temporary source copy, compiles and completes under native sanitizers, exits zero with empty stderr, then disagrees with Go. Refusal, panic, sanitizer error or compiler warning cannot earn semantic-mutant credit. Evidence/mutants.json records every exact first differing line and values.

| Helper | Independent variants caught |
|---|---|
| decodeEscape (6) | invert class-sensitive b; invert class-sensitive B; invert named k rejection; exclude zero from numeric dispatch; decode n as carriage return; delegate x with four hex digits instead of two |
| classAtoms (5) | leave inClass false; swap dash/rune classification; keep property sets exact under ignoreCase; omit the slash byte from raw set text; delegate word expansion to nonword expansion |
| RegExp.Test (4) | return true for missing wrappers/engines; return true only for nil-engine wrappers in the guard; accept an error-bearing matched result; invert the matched bit |

The two guard variants hold nil wrapper and nil engine separately. The hex-width and word-delegation variants hold dependency selection/arguments as well as final values. The error acceptance variant is held by the controlled matched-plus-error result, while the real timeout is independently observed as unmatched plus error. Raw corpora and expected bytes are compressed losslessly with fixed gzip timestamps; corpus-manifest.json records original sizes and SHA-256 hashes.

## Findings and limits

The preliminary gate rejected shorthand function fields in the private driver and included engine compilation's class-helper observations in the Test trace. Explicit callback arrows and muted compilation corrected those adapter issues. The next gate passed classAtoms and Test but the newline-dispatch mutant survived: the initial consumer/control corpus had no literal slash-n input. All 256 escape-opener bytes were added, and that same mutant was caught in the final gate. Failed attempts remain in helpers.log and final.log, without whole-gate pass credit. Rejected compiled patterns now exercise nil-wrapper behavior instead of being omitted.

The unpooled class corpus made repeated native loads slow. Identical dependency records are now interned and prepared once, without deleting or merging any input row. New byte controls increased the corpus from 29760 to 37824 class rows while the final full gate fell to 153.808s. This is data sharing, not a correctness-check relaxation. No shared file changed.

Not covered: complete rule diagnostics, positions, fixes, suggestions, options or shared registry integration; external dependency implementations; native regex compilation/matching, flags or timeout cancellation; arbitrary malformed adapters, side-effectful dependencies, negative offsets, object-engine generic instantiations or exhaustive combinations. Numeric engine handles and reversible byte-string projections are private test adapters. UTF-8 byte offsets remain bytes; finding-position conversion is outside these helpers.

The full repository gate, including the seventeen required stage1 external-input checks, was not run. This is the authorized bounded touched-package, repository-vet and filtered external-oracle gate. No skip is counted as green. Only these three helpers were reserved.
