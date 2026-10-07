Built parser, exact Unicode tables, bytecode compiler and runtime V8 refusals; lowering is approved and applied.
Pushed parser 70ff637, compiler 3dbc626, V8 refusals 1206448, reference followups 7a7f1b5 dd0f01b and 304ad82.
Checks: Linux ASan/UBSan, WASI, corpus bytecode/property identity, Node backends and statics/layout guards pass.
Mutants: all 18 table, reference, emission, rejection, refusal and ownership mutations are caught; details below.
Approved native.go hook applied; unrelated baseline concurrency failures prevent a full native gate.

The lowering proposal is in the working tree, with its excluded-file hook saved
at /tmp/regex-runtime-native-build-hook.patch and tested through
/tmp/regex-runtime-build-hook-overlay.json. The user explicitly approved both the hook and complete lowering patch.
internal/native/native.go:133 now calls runtimeLibraryForSource(source, options)
in place of RuntimeLibrary("", options), the only change in that file.

Parser/refusal layers are already pushed and independently usable by their C
identity tests. Matching uses the existing VM. The only regexp.c changes select
counted ownership for dynamic programs, clone that ownership, and retain it in
escaped named-group dictionaries. These hooks are compiled only in the dynamic
ownership archive variant. All compiler/table definitions are confined to the
dynamic generated module; ordinary archives contain empty compiler units.

| Linux RegExp directory | Pass | Disagreement | Refused | Crashed | Skipped |
|---|---:|---:|---:|---:|---:|
| Before | 53 | 0 | 424 | 73 | 1329 |
| Lowering proposal | 56 | 0 | 421 | 73 | 1329 |

Every newly passing test agrees with Node:
- built-ins/RegExp/S15.10.2.10_A2.1_T1.js
- built-ins/RegExp/S15.10.2.10_A2.1_T2.js
- built-ins/RegExp/S15.10.4.1_A8_T2.js

cmd/adamic-test262 links runtime objects itself. Its small companion hook selects
the same ownership variant when generated C contains the compiler directive.
Without that hook, the three new cases failed to link; that integration mistake
was corrected and the directory rerun. The runner's tests pass in 26.268s.

The shared corpus contains 5,746 test262 patterns (including the required 2,776),
875 cohere patterns pinned in runtime-cohere.json, 4,000 random patterns seeded
with 0xEC2025, and nine reference probes. Raw native bytecode compares 7,433
programs and 3,197 rejected/refused cases. Compatibility adds 400 shape probes
and a refusal-precedence probe. Both bytecode fields and exact V8 refusal messages
are compared on Linux and wasm32-wasi. Property entry identity compares all
1,722 aliases, 448 unique properties, 23,045 ranges and 7,906 strings.

Seven ordinary dynamic fixtures agree with source Node and the JavaScript
backend: dynamic_gap, id-length exceptions, inline-comment ignore patterns,
warning terms/decorations, flags/SyntaxErrors, ownership and evaluation order.
WASI runs all seven. Native oracle variants include ASan/UBSan, release, slabs,
leak checks and allocation counts. Every fixture's allocations equal frees. Named groups cannot expose the private compiler owner; that fixture and the counted NUL SyntaxError fixture pass against Node.
Four divergence shapes and one counter-width refusal are compared separately to Go's exact refusal
reasons; source Node and the host-RegExp backend still demonstrate V8 acceptance.

Final compiler/statics/layout gate: 47.136s. Dynamic oracle/WASI/refusal/ownership
mutant gate: 64.372s. Lower/reference/fresh packages: 19.612s/10.288s/24.555s.
Final allocation-count regeneration: 87.264s. Repository vet passed with the
documented integration overlay. The later Identifier-property fix passed seven
Node/refusal mutants in 4.465s and the compiler gate was rerun afterward.

WASI uses the 32-bit SDK with ADAMIC_TARGET_WASI=1 and no atomics. Linux is the
sanitizer gate; WASI runs are not sanitizer instrumented because this SDK does
not provide ASan/UBSan runtimes.

| Fixture | Platform | Raw before | Raw after | Brotli before | Brotli after |
|---|---|---:|---:|---:|---:|
| hello | wasm32-wasi | 245068 | 245068 | 71934 | 71934 |
| hello | native | 385632 | 385872 | 128837 | 129086 |
| request | wasm32-wasi | 289230 | 289230 | 87461 | 87461 |
| request | native | 386368 | 386608 | 129699 | 129579 |
| dynamic_gap | native | — | 1826984 | — | 271283 |
| dynamic_gap | wasm32-wasi | — | 783380 | — | 148727 |

The common baseline is a7f6f64 (library plus statics integration), with identical
fixture/output basenames, release defaults and toolchain. Hello/request WASI
artifacts are byte-identical. Native loaded section sizes are unchanged. Each
native file gains 240 bytes solely from five empty compiler translation-unit
filename symbols. Native Brotli differences include the accepted assertion-cache
path noise. There are no compiler/table or ownership function symbols in the
ordinary program. For comparison the user's earlier 14,236-byte hello used a
different build; this report uses one common baseline.

| Mutant | What caught it |
|---|---|
| Change ASCII property endpoint | entry-by-entry C/Go identity |
| Accept lowercase ascii alias | rejected-alias comparison |
| Remove compiler guard | ordinary-program symbol proof |
| Break property lookup only on WASI | WASI entry comparison/trap |
| Change generated provenance | regeneration identity |
| Stop a top-level pattern at NUL | Node reference validity comparison |
| Drop Other_ID_Start | raw/escaped ℘ Node comparison |
| Drop Other_ID_Continue | raw/escaped a· Node comparison |
| Accept Pattern_Syntax U+2E2F name | Node rejection comparison |
| Wrap oversized braced Unicode escape | Node rejection comparison |
| Silently accept clamped reversed bounds | exact V8DivergenceError assertion |
| Return SyntaxError for that divergence | exact V8DivergenceError assertion |
| Emit ASSERT instead of SET | bytecode comparison, case 0 byte 288 |
| Accept ? in C | syntax status comparison, case 4700 byte 0 |
| Bypass runtime V8 check | compatibility status/reason comparison |
| Drop compiled storage owner | ASan heap-use-after-free |
| Always select ownership archive | ordinary-program ownership-symbol assertion |
| Corrupt SyntaxError message after NUL | full counted-byte error comparison |

No working-tree mutant remains. Mutants reach their intended assertions rather
than being killed by compiler diagnostics. The table-only WASI mutant traps in
the entry comparison, as its original proof records.

Full-native-package limitations: its 464.462s run found map TSan races plus
intermittent shared-string-index and signal-mutant survival. The map race also
reproduces on a7f6f64 in 7.040s; map.c is byte-identical between baseline and this
branch. The newly exposed field-layout omission and stale dynamic-RegExp refusal
test were corrected, and their gates rerun. No map/concurrency code was edited.
The merged map_hash_test.go uses slabs instead of Slabs; tests use a temporary
Go overlay for that typo, without editing its owned file. The full repository
oracle has not been rerun; targeted regex/backend/WASI gates are the evidence.
Stress-scale nesting/allocation exhaustion is not measured by this corpus.

All subsequent Go commands use GOPROXY=https://proxy.golang.org|direct per integration.

Setup: Go 1.27.1 ready 0s, clang 20.1.8 ready 0s, Node 24.19.0 ready 0s,
submodules ready 0s, nproc 5. Cache warming initially failed at the unrelated
slabs/Slabs typo. The pinned stage3/api dependency was installed with npm ci
--ignore-scripts (3 packages, 842ms) to resolve the merged lowering prerequisite.

Relay to #adamic_runtime_platforms: branch codex/regex-runtime-compiler; pushed
layers 70ff637, 3dbc626, 1206448, 7a7f1b5, dd0f01b, 304ad82. Construction proposal overlaps
regexp.c/regexp.h for ownership only, native/library.go for archive selection,
native/emit_expressions.go and fields.go, lower/exceptions.go and regexp.go, and
cmd/adamic-test262/run.go. native.go's one-line Build hook is explicitly approved and applied. No matcher algorithm, atomics or mutable runtime statics were added.
Baseline map TSan race is attached in baseline-map-tsan.txt for the concurrency
owner. The size table above is the platform evidence for the reviewed proposal.
