# Batch 31 handoff

Built borderSideDescription, maskStopDescription and resolveArmColor in separate Adamic .a files. Every helper removes a listed prerequisite for better-tailwindcss/enforce-consistent-class-order, better-tailwindcss/enforce-shorthand-classes and better-tailwindcss/no-unknown-classes. Nine dependency occurrences, three distinct consumers, none newly helper-ready. Cumulative slot05: 89 helpers, 428 occurrences, 76 consumers, 51 helper-ready candidates including the common 46. These are frozen inventory calculations, not completed rules.

Claim bfa869ddbe804c1f7a963499595ff36dbe67e3c6 was pushed successfully before source writes. Selection read all claim files across 20 origin codex/lint-helpers* branches and shared HELPERS.md; the three helpers tied the highest available concrete fan-out, three each. Current origin/main 71d7e491b3c9724f7a0e2ee754592149e7f9790b and origin/area/stage1-lint db2ecc00447f9ebe8adecb190f71ac222e5db860 were unchanged and already ancestors. All prior 86 helpers, compiler/runtime and shared harness inputs are unchanged; their retained proof remains applicable. No shared files were authored, no main or area branch pushed.

## Observed comparisons

The oracle asserts pinned Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db and overlays only thin exports of the original three functions. All consumer test files contribute 406 distinct strings; controls produce 422 values. Each no-argument description is tested 422 times with independently varied returned-array mutation probes. Color resolution tests 3376 cases, eight actual Go theme configurations per value. Total 4220 baseline cases match byte-for-byte on actual Go, source Node, emitted JavaScript on Node and sanitized native, with exit zero and empty stderr.

Descriptions preserve all Go fields, nil/present distinctions, namespace and inference order, defaults and independently allocated arrays. Color resolution preserves exact keywords and currentcolor spelling, and the actual Theme.Resolve text/found pair for every other value, including empty-but-found. Tests include actual Go inline, empty-inline, reference, prefixed and ordinary theme results; empty/nil namespaces, precedence and duplicates; keyword entries whose theme values disagree with keyword answers; near spellings and Unicode/NUL/newline controls. The driver uses actual Go dependency observations and checks the exact value, present=true, unchanged key-array identity and options=0 on delegation. No dependency call counts are claimed observed from Go.

## Commands and output

All output was written directly to logs, with no test pipeline.

- ADAMIC_SLOT05_BATCH31_EVIDENCE=<owned evidence directory> ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch31 -count=1 -v -timeout=20m: final PASS, 37.499s, all 4220 cases and 39 compiling native semantic mutants. Initial 35-mutant suite PASS, 35.846s; four additional order/freshness mutants justified the complete rerun. Both logs retained.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v: PASS, 1.105s, seven actual input probes, zero cache hits and seven misses.
- Final go vet ./...: exit zero, empty output. Final gofmt -l cmd internal stage1/cohere/lint/helpers/slot05: exit zero, empty output.
- export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh: PASS, 24.109s; nproc=5, quota 400000 100000, 17.6 GB. Node ready 0.018s, Go 0.031s, markdown 0.071s, submodules 0.089s, clang 0.204s, Go build 23.960s, test binaries deferred 24.080s, cache warm 24.081s. Go 1.27.1, clang 20.1.8, Node 24.19.0. The printed /workspace/adamic-tools/env.sh was sourced before builds.

Evidence contains lossless logs, every generated case and actual Go verdict, consumer file/count coverage, exact mutant replacements and first witnesses, base identities and dependency hashes. Disk space remained available; no prior published evidence was removed.

## Every mutant

All 39 compiled and ran on native with exit zero and no stderr before an independent Go output mismatch caught them. No compiler error, panic or sanitizer finding counts as a semantic kill. Source and emitted JavaScript baseline comparisons passed; mutants ran on native only. Full records are in evidence/mutants.json and helpers.log.gz.

Both descriptions: eight changed presence/default fields (theme presence, negative, fractions, bare handler presence, negative handler presence, static names presence, modifier-on-arbitrary and arms presence); change default text and presence; reverse color namespace order; reverse the two arms; reuse a single description across calls. Border additionally erases its width suffix and modifier refusal together. These 14 border mutants are all caught.

Mask additionally removes --spacing; substitutes PositiveInteger for SpacingMultiplier; drops modifier refusal; reverses number/percentage inference order; erases inference presence; drops percentage passthrough. These 19 mask mutants are all caught. Reused-description mutants specifically fail the second Go description's independence after mutations to the first.

Color resolver: omit inherit; omit transparent; map current to current instead of currentcolor; always refuse a theme value; always claim a theme value is found; corrupt the returned text. All six are caught by actual Go output.

## Limits

The full repository gate and its 17 required external correctness checks, full shared lint suite, whole-rule diagnostics/fixes/suggestions, CSS emission, arbitrary nil-pointer inputs and performance were not run. No skipped check is credited green. Theme.Resolve is an explicit external dependency, not authored here. False presence flags represent absent values; callers must not use placeholders as supplied dependencies. These three helpers require no regex.
