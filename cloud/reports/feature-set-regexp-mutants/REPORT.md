Built: monolithic native builds use sourceFlags, matching split builds and the selected runtime archive.
Commits: repair on compiler/feature-set-link, directly above f5f37525; delivery SHA is the commit containing this report.
Commands and outputs: the requested oracle families, closure test, feature-set mutants, checker controls, build and vet passed; compressed logs are beside this report.
Mutants: regexp mutations reach Node differences or pinned checks; count removal reaches typed arity; feature-set mutants retain their link witnesses.
Not covered: the unrelated cohere TestProfileCompilation failure, the full gate and the optional WASI profile. No new fixtures or counts rows.

The regression was in internal/native/native.go: Build passed Flags(options) to clang, although RuntimeLibraryForSource selected the runtime using featureFlags(source). A prepended regexp_replace.c includes adamic.h before the emitted feature defines, so its gated replacement implementation and closure declarations were unavailable. Build now passes sourceFlags(source, options), the existing shared source of feature flags. These compiler defines apply before all includes. The header and its gates are unchanged; there is no second feature list.

Every affected regexp mutation now first builds and runs the same renamed, prepended runtime without its mutation. Native controls run with address and undefined-behavior sanitizers and leak detection. Valid controls agree with source Node; guard controls stop at the same pinned exit-70 diagnostic as the JavaScript backend. Mutants must compile and complete without sanitizer diagnostics before their Node or checked-output mismatch can count. The closure runtime count test also has a sanitized, Node-held control, and requires both the too-few-arguments diagnostic and expected 3, have 2.

Observed on the unmodified parent: the four oracle families and native closure test failed with the reported incompatible integer-to-pointer clang diagnostics (both commands exit 1). After repair, all commands below exit 0.

Commands:

- go test ./internal/oracle -run '^TestRegExpReplacement(NodeMutants|TypeGuardMutants|DropCount|CountNodeMutant)$|^TestNotYetLibraryRegex(Callback|Offset)Mutants$' -count=1 -v
- go test ./internal/native -run '^TestClosureConventionRuntimeDropCount$|^TestRuntimeFeatureMismatch$' -count=1 -v
- ADAMIC_CLANG_TSGO_ARCHIVE=/tmp/stack-c2-tsgo.a go test ./internal/native -run '^TestSplitTSGoRuntimeFeatures$|^TestSplitTSGoAgrees$' -count=1 -v
- ADAMIC_CLANG_TSGO_ARCHIVE=/tmp/stack-c2-tsgo.a python3 internal/native/testdata/run-feature-set-link-mutants.py
- go build ./cmd/adamic
- go vet ./internal/native ./internal/oracle

All use the configured toolchain and GOTMPDIR=/workspace/scratch/hidden-go-build. The real pinned checker archive is enabled for the split checker tests and mutants. Both cached and uncached checker execution agree with the checker observation. An initial additional checker run without that environment variable skipped those tests; the recorded checker-positive log is their subsequent enabled run. An initial closure control used an incorrect Node harness path; the recorded after-native log is the corrected run.

Mutant witnesses:

| Family | Mutations | Witness |
| --- | --- | --- |
| RegExpReplacementNodeMutants | offset off by one, groups wrong order, missing named groups | Native and JavaScript finish cleanly and disagree with source Node stdout |
| RegExpReplacementTypeGuardMutants | argument_guard, group_guard | Guard removal finishes cleanly instead of the pinned exit-70 declared-type stop, in native and JavaScript |
| NotYetLibraryRegexCallbackMutants | match, literal, unicode, global, offset, reset, collection-order | Native finishes cleanly and disagrees with source Node stdout |
| NotYetLibraryRegexOffsetMutants | offset, input | Native finishes cleanly and disagrees with source Node stdout |
| RegExpReplacementCountNodeMutant | physical packed count replaces logical callback count | Native and JavaScript disagree with source Node stdout |
| RegExpReplacementDropCount and ClosureConventionRuntimeDropCount | remove counted callback argument | Intended typed-arity clang rejection: expected 3, have 2; valid control compiles and agrees with Node |
| Feature-set link | fixed-symbol, drop-reference, drop-retain, header-forces-runtime-features | The mismatched runtime links, which the link-rejection test catches |
| Feature-set link | split-runtime-default-flags | Intended feature symbol link rejection; ordinary matching-build test catches the mutant too |
| Feature-set link | split-original-checker-default-flags | Original checker fixture reaches the intended unsanitized feature symbol link rejection |

Setup succeeded. nproc: 5, cpu.max: 400000 100000. Reported setup timing lines: submodules 0.083s; markdown dependencies 0.097s; clang 0.180s; go build 208.389s; build cache warm 208.553s; done 208.580s. Toolchain: Go 1.27.1, Node 24.19.0, clang 20.1.8. Setup log retained beside this report. This repairs the landing regression and preserves the roadmap's visible checks and feature-set contract.
