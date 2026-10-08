# Scout regex gaps

## Runtime constructor: closed

`testdata/dynamic_gap.a` with argument `TODO` prints `true` on source Node, emitted JavaScript and sanitized native after merging area/library. TestDynamicPatternGap now fails on any lowering refusal; the 89 shape fixtures require runtime native acceptance as well.

## Cohere esregexp Unicode script property: parked on #7zm085y

At f5d1934a, Compile(`\p{Script=Greek}`, `u`) rejects with SyntaxError while Node new RegExp with the same source and flags accepts and matches `α`. Reproducer: testdata/option_property_gap.json, TestOptionPropertyGap. This is a Go-side esregexp gap, parked on #7zm085y. No cohere pointer to the proposed property branch is used; no translation or fallback is used. An inline comment `x; // α` with ignorePattern `\p{Script=Greek}` exposes the rule-level consequence: Go drops the invalid option, while JavaScript accepts it.

## Shared upstream capture retains compiled sources: closed

The lint port's capture overlay in `../shared_test.go` now replaces compiled matcher JSON leaves with their Source() strings. The helper in testdata/capture_options.go starts with upstream JSON, retaining custom marshalers, JSON tags, omitted fields and exact numbers. Cohere itself and its pin are unchanged.

TestRegexCompiledOptionCapture replays all 350 upstream cases through the shared capture and Go oracle, preserving 14 compiled id-length exception-pattern sources. TestCompiledOptionCaptureSources exercises actual esregexp objects, nested options, nil matchers and custom enum serialization. ADAMIC_CAPTURE_SOURCE_MUTANT=1 replaces a source with an empty object; go test fails with the missing-source comparison. Raw encoding/json still serializes the matcher as {}, which is why the capture overlay is required.

## Quiet-hundred manifest unavailable

No named quiet-hundred manifest exists on the fetched base or in this workspace. The recorded alternate 100-file corpus is explicit (77 pinned compiler files, 23 stage1 files, 2,104 strings) and is retained without representing it as the fleet's quiet hundred.

## Native split linker rejects generated attributes

All three production fixture graphs lower successfully with EnableTSGo, but BuildSplitTSGo rejects emitted declarations with `native: split: duplicate definition __attribute__`. This is the current shared lint driver's linking path. The regex-only matcher differential uses the existing single-unit native builder and passes; it does not establish split-linker compatibility.

Reproducer for @system_adamic_library: testdata/split_gap.a and TestNativeSplitAttributeGap, building emitted C with Sanitize and Split enabled. The shared capture edit changes only compiled option serialization. No compiler, emitter or splitter edit is made. The owned finding fixtures require the canonical BuildTSGo backend used by cmd/adamic; its separate unresolved constructor gap is recorded below. The package gate exercises its unchanged split backend. Neither native finding leg is skipped.

## Canonical checker-linked native constructor: closed

The library-only 6c5b7f03 fix is merged. TestNativeCheckerRegexLinkGap now requires the unchanged dynamic_gap.a to build with the sanitized checker archive and print true. All three production finding graphs pass the canonical BuildTSGo backend: 204 id-length, 58 no-inline-comments and 88 no-warning-comments cases; 216 total findings agree with Go, source Node and emitted JavaScript. Evidence and exact commands are in CAPTURE_LINK_RERUN.md. This establishes the canonical native backend; the shared split backend still waits on the conflicted compiler merge described below.

## Shared legacy mutant anchors and interrupted package run

The unchanged shared package's legacy mutations look for removed/retired matcher text: `this.anchors[0] = true;`, `!space(character)`, and `for(const entry of decoration) {`. These anchors now return a count of zero; owned runtime-RegExp mutants replace their relevant comparison coverage. The shared mutant definitions are unchanged.

The package's pinned stage1 corpus check rejected the initially dirty tree; the implementation is now committed. The single all-input package run was interrupted with SIGKILL after the environment reconnect and has no final summary. Partial counts, exact command, timestamps and every completed test name are in evidence/scout/package-summary.json. No second full-package run or stage1 corpus parity is claimed.

## Split fix integration blocker

The library-only branch codex/regex-lint-link-gaps-library at 6c5b7f03 merged without changing the f5d1934a cohere pin. Both compiler/area-next-fixtures at 547551cb and the exact compiler/split-attribute branch at 3bea093d produce 53 conflicts across compiler, library, oracle and fixture files on this scout base. Both attempts were aborted; no unrelated conflict resolution or compiler workaround is introduced. TestNativeSplitAttributeGap still observes duplicate definition __attribute__. Integration of the compiler branch is required before the shared split gate can pass.
