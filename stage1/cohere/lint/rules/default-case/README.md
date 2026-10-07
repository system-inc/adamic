# Default case: default configuration port

Ported the switch judgment, exact message/span, empty-switch exception and last-comment-only default exemption in .a, registered solely through this directory. Claim ae8e7213 preceded all implementation. This is a bounded port, not a completed full-options rule. A nonempty commentPattern explicitly refuses with exit 70 instead of quietly using different regex behavior.

Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db is the independent reference. With default options, 31 distinct upstream source shapes/control sources produce 9,777 identical bytes on Go, source Node, emitted JavaScript and ASan/UBSan native. This replays the source shapes unconfigured; it does not preserve the custom options from seven upstream rows. Another 353 TypeScript compiler/stage1 source files produce 13,471,512 identical bytes, against compiler pin 050880ce59e30b356b686bd3144efe24f875ebc8. The later added owned gap source was compared separately, so it is not represented in that 353-file snapshot. Five actual Unicode trim controls produce 1,072 identical bytes, including the difference between Go TrimSpace and JavaScript trim on FEFF and U+0085. Raw and compressed outputs are in evidence/.

The listener uses the existing parser-anchored all helper once per file with switches. It compares UTF-8 comment ranges to byte-converted clause and case-block boundaries, then reports through the existing UTF-16 context. The Go file-cache sharing optimization across rules is not exposed by this RuleContext; this listener does not edit that shared API. There are no fixes or suggestions, matching Go.

The semantic mutant default-comment-anchors-removed removes the translated literal's start and absolute-end anchors. It compiles and exits successfully on source Node, emitted JavaScript and sanitized native, then only the Go byte comparison rejects it. It silently excuses near-miss comments that Go reports. No compiler error or sanitizer failure kills this mutant.

Measured findings/s, best of five process launches over 1,000 reported switches, includes startup and uses release native: native 76,358.33; Node 7,708.34; Go 229,411.47. These are default-configuration measurements, not general regex-option throughput. Setup PASS: Go/clang/Node/submodules 0s each, build cache warm 27s, total 27s, nproc 5.

Reproduction, from the repository root with the setup environment sourced:

```
go run ./cmd/lint-registry > /tmp/default-registry.log 2>&1
python3 stage1/cohere/lint/rules/default-case/validate.py --scratch /tmp/default-case-check --compiler /tmp/wave07-typescript > /tmp/default-case-run.log 2>&1
go vet ./stage1/cohere/lint/rules/default-case > /tmp/default-vet.log 2>&1
```

The validator builds its owned profile via TestCompileProfiles, compares the source corpus, builds/runs the mutant and measures throughput. The extra Unicode, custom-refusal and gap observations are retained separately. The exact upstream command was go test ./internal/lint/rules/core -run '^TestDefaultCase' -count=1 -v inside cohere: PASS, 0.008s, including the neighboring default-case-last tests selected by that prefix. Vet and descriptor generation pass.

## Remaining block and coverage limits

Go accepts arbitrary RE2 commentPattern strings and falls back to its default when regexp.Compile rejects them. Stage 0 currently requires a constant RegExp pattern. The owned gaps/dynamic-comment-pattern.a reads its pattern from a file: source Node prints true for ^skip default, while adamic build with -o refuses with "stage 0 can't lower RegExp with a nonconstant pattern yet". A runtime Go-compatible regex compiler/validator and its invalid-pattern fallback remain needed. Changing the backend or substituting JavaScript regex without a Go compatibility proof is outside this rule port.

A custom-pattern control makes Go return clean with exit 0; all three Adamic runtimes return identical stdout/stderr with exit 70 and "NotYet: default-case Go-compatible commentPattern compilation". This is an explicit partial-port refusal, not byte parity with Go for that option. Full declared upstream options coverage and malformed-option decoding are not claimed.

The shared TestOwnedWitnesses run fails before reaching this rule's witness, in the existing next-google-font-display witness: the .ts copy of JSX is rejected by Go with '>' expected and exit 2. Its log is retained. No shared harness, compiler, parser or other rule file is edited to bypass it. The full gate and shared all-options/all-rules certification are not claimed.

The reserved consistent-return still lacks compatible Judge / EndReachable integration; constructor-super remains reserved and unstarted. This partial default-case implementation does not meet the user's all-claimed-rules completion prerequisite, so no helper branch or further claim is taken. Unclaimed helpers still exist; no exhaustion claim is made.

The default exemption now uses the shared codex/lint-regex translation at core/default_case.go:29: /^no default(?![\s\S])/iu. The table uses g for whole-match enumeration; this rule tests a boolean and omits g to keep repeated tests independent. Runtime option patterns remain explicitly refused, as also documented in the shared regex gap report. No hand-written default-pattern matcher remains.
