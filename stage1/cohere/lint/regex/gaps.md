# Scout regex gaps

## Runtime constructor: closed

`testdata/dynamic_gap.a` with argument `TODO` prints `true` on source Node, emitted JavaScript and sanitized native after merging area/library. TestDynamicPatternGap now fails on any lowering refusal; the 89 shape fixtures require runtime native acceptance as well.

## Cohere esregexp Unicode script property

At f5d1934a, Compile(`\p{Script=Greek}`, `u`) rejects with SyntaxError while Node new RegExp with the same source and flags accepts and matches `α`. Reproducer: testdata/option_property_gap.json, TestOptionPropertyGap. This is a Go-side esregexp gap; no translation or fallback is used. An inline comment `x; // α` with ignorePattern `\p{Script=Greek}` exposes the rule-level consequence: Go drops the invalid option, while JavaScript accepts it.

## Shared upstream capture loses compiled id-length sources

The cohere docs capture calls encoding/json.Marshal on IdLengthSettings. esregexp.RegExp has unexported state and no JSON marshaler. Therefore ExceptionPatterns becomes `[{}]`, losing the source before the shared lint oracle or Node runner receives it. The owned oracle rejects this explicitly instead of constructing a zero-valued matcher or inventing a source.

Shortest reproducer: testdata/esregexp_capture_gap.go, run through the cohere-module overlay in TestCompiledOptionCaptureGap. Observed output: `source=^_ captured=[{}]`. The trace-only regex fixture capture calls Source() before encoding and retains all actual upstream sources; it compares the three migrations independently. The shared harness is unchanged and its full package gate may fail on this named metadata gap.

## Quiet-hundred manifest unavailable

No named quiet-hundred manifest exists on the fetched base or in this workspace. The recorded alternate 100-file corpus is explicit (77 pinned compiler files, 23 stage1 files, 2,104 strings) and is retained without representing it as the fleet's quiet hundred.

## Native split linker rejects generated attributes

All three production fixture graphs lower successfully with EnableTSGo, but BuildSplitTSGo rejects emitted declarations with `native: split: duplicate definition __attribute__`. This is the current shared lint driver's linking path. The regex-only matcher differential uses the existing single-unit native builder and passes; it does not establish split-linker compatibility.

Reproducer for @system_adamic_library: testdata/split_gap.a and TestNativeSplitAttributeGap, building emitted C with Sanitize and Split enabled. No compiler, emitter, splitter or shared lint harness edit is made, The owned finding fixtures require the canonical BuildTSGo backend used by cmd/adamic; its separate unresolved constructor gap is recorded below. The package gate exercises its unchanged split backend. Neither native finding leg is skipped.

## Canonical checker-linked native runtime has an unresolved regex constructor

BuildTSGo on the production fixture graphs fails to link regexp_compile_runtime.c: `undefined reference to adamic_regex_new_owned`. The standalone cached-runtime builder succeeds, including dynamic_gap and all shape/matcher tests; the checker-linked builder compiles the runtime as separate C units.

Shortest reproducer for @system_adamic_library: unchanged testdata/dynamic_gap.a linked via native.BuildTSGo with the current sanitized checker archive; TestNativeCheckerRegexLinkGap captures the exact linker error. No runtime linkage shim is added. The production finding tests require native, so they fail after proving Go, source Node and emitted JavaScript parity rather than skipping this leg.
