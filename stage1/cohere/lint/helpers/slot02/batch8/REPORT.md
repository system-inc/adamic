Built three retained .a helpers: EscapeClassRune, identityEscape and literalRune; four consumers each, twelve prerequisite entries across four rules, zero final blockers.
Commits: initial claims e6aaf5c2, replacement claim 0a261996, implementation a83050048bc34ef221ab3ce17b608e105747a9ba; branch codex/lint-helpers-02 on current main f8013f0.
Commands: setup 45s/nproc 5; retained isolated oracle PASS 22.984s; complete helper oracle PASS 500.874s; inventory PASS 4.695s; uncached input oracle PASS 2.445s; vet exit 0.
Mutants: thirteen retained new compiling semantic mutants caught against actual Go on source Node, emitted JavaScript and sanitized native; all sixty-four prior mutants caught again, seventy-seven total.
Not covered: whole rules, native regexp compilation/matching, full repository gate, arbitrary malformed adapters or exhaustive Unicode/configuration inputs; control-escape duplicate withdrawn.

# Landing and ownership

All twenty-one earlier helpers were pushed and landing-ready at 3e3e2d37 on current main f8013f0, with the fresh complete helper oracle PASS 567.703s and all sixty-four semantic mutants caught. The branch was clean. An explicit wildcard fetch inspected all seventeen claims files across eighteen origin codex/lint-helpers* branches. Every larger concrete count was reserved; the original comments bundle and rule-local strict-options ledger entry remained excluded. The initial three claims tied the highest available concrete count at four consumers each. e6aaf5c2 was pushed before code.

The post-publication refresh exposed slot 05's earlier decodeControlEscape claim 94e0c374 at 04:37:54 UTC, preceding our e6aaf5c2 at 04:39:31 UTC by ninety-seven seconds. This slot withdraws the duplicate and deletes its .a implementation. The superseded 41.806s local gate and control-specific mutant observations are retained as history, not delivered coverage, readiness or mutant credit. EscapeClassRune and identityEscape remain owned here. Refreshed all eighteen branches and claims again, then claimed unreserved literalRune at the same four-consumer count in 0a261996, pushed before its code.

The final refresh includes twenty helper branches and nineteen complete claim files, including two newly published branches. The retained symbols appear only on slot 02. evidence/retained-claims.log records that audit and main ancestry. Only owned claim, helper batch and standalone slot02_batch8_test.go change. No shared harness/generator, protected compiler, rule dispatcher or Diagnostic model changes. Only codex/lint-helpers-02 is pushed; no main/area branch or pull request.

# Actual behavior and coverage

EscapeClassRune escapes exactly backslash, close bracket, caret, dash and open bracket. Its rune-to-string conversion replaces negative, surrogate and out-of-range scalar values with U+FFFD. IdentityEscape preserves the numeric rune and caller width on success. Unicode mode restricts acceptance to syntax characters, slash and in-class dash; without Unicode mode every identity escape is accepted. Group-count and named-group fields do not affect it. Errors retain Go's exact text, zero decoded fields and ErrUnsupportedSyntax cause. The private Go export independently checks errors.Is and direct errors.Unwrap against the actual sentinel.

LiteralRune passes ASCII letters/digits and supplementary runes through Go rune-to-string conversion. All other signed rune values use lower-case minimum-width-four %04x formatting. Surrogates stay hexadecimal text; negative sign counts within minimum width; large negative values are not truncated. The replacement port follows this behavior rather than applying EscapeClassRune's scalar conversion to every input.

Observed: 641 actual runtime source/options captures cover all four consumers. The Go Core, Next and TypeScript rule packages pass during capture. There are 1,735 captured string option values, 272 rune values, eight contexts and six size controls, producing 13,600 direct helper comparisons. The rune set includes every value through U+00FF, every runtime source/options rune and signed/scalar/surrogate/supplementary boundaries. Zero/negative widths and variation in unused context fields are also checked. Returned kind/set/rune/width/negated fields, string bytes expressed as UTF-16 units, error text and exact cause all agree across actual Go, source Node, emitted JavaScript and ASan/UBSan native.

The oracle directly calls pinned cohere 715ba94f3608a6500086b1076ce5cb7e51b836db through temporary exports. Go helper bodies are unchanged. Capture overlays observe ordinary and typed fixtures in temporary files; no cohere worktree source is edited. Options are marshaled at runtime, with typed temporary-directory prefixes normalized to <fixture> for reproducible source/options input data. Every consumer and a zero Go package exit are required. Expected Want is removed before Adamic reads the corpus.

Sources and retained coverage regenerate byte for byte. Source SHA 7a2441355e3c62eb1d32503fd40e6a1e1db9325a51f57c6dd2ec91121f61f111; coverage SHA ee3c3e3d2d92cc4cdd01800b577ee0fc23b0e2f74033136e3cac416bf2860875. See evidence/retained-reproducibility.log.

Inferred: twelve prerequisite entries are removed across four distinct rules. No rule loses its final listed helper blocker. RULES.md names every consumer; readiness.json subtracts only this slot's twenty-four retained helpers, without presuming other branches have landed. This remains helper readiness conditional on the frozen inventory's common AST adapter.

# Commands and outputs

All test output goes directly to logs, never through a pipe. Source /workspace/adamic-tools/env.sh before Go commands. Go 1.27.1, clang 20.1.8, Node 24.19.0.

- bash cloud/setup.sh > evidence/setup.log 2>&1: PASS. Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 45s, done 45s on five processors; nproc prints 5.
- python3 .../batch8/testdata/regenerate.py > evidence/retained-regeneration.log 2>&1: all four consumers, 641 rows, Go Core/Next/TypeScript packages PASS. Capture package output is in evidence/capture.log.
- Retained capture regeneration and source/coverage hash equality: evidence/retained-reproducibility.log, PASS.
- go test ./stage1/cohere/lint/helpers -run '^TestSlot02Batch8$' -count=1 -v -timeout=20m > evidence/retained-isolated.log 2>&1: PASS 22.984s, all thirteen retained new semantic mutants caught.
- ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > evidence/helpers-final.log 2>&1: PASS 500.874s; all seventy-seven semantic mutants caught.
- go test ./stage1/cohere/lint/inventory -count=1 -v > evidence/inventory.log 2>&1: PASS 4.695s.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > evidence/input-oracle.log 2>&1: PASS 2.445s, all six fixtures, zero cache hits and six probe misses.
- go vet ./stage1/cohere/lint/helpers ./stage1/cohere/lint/inventory > evidence/vet.log 2>&1: exit zero, empty log.

# Every retained new mutant

Each mutant must compile and finish successfully on all three Adamic backends, without stderr or sanitizer findings, and all backends must agree before a difference from actual Go is counted. Temporary copies hold the mutations; production files stay unchanged. Compile failures, crashes and sanitizer reports are not semantic kills.

| Mutation | Actual Go witness |
|---|---|
| Omit class dash escape | Go produces backslash-dash; mutant produces bare dash. |
| Omit class open-bracket escape | Go produces backslash-open-bracket; mutant omits the backslash. |
| Wrong invalid-rune replacement | Go writes U+FFFD; mutant writes question mark. |
| Remove Unicode identity restriction | Go returns a wrapped syntax error; mutant succeeds. |
| Refuse slash identity escape | Go accepts slash with unchanged width; mutant errors. |
| Permit dash outside a Unicode class | Go rejects it; mutant succeeds. |
| Change identity width | Go returns exactly caller size, including -3; mutant increments it. |
| Drop identity error cause | Go retains ErrUnsupportedSyntax as direct cause; mutant removes it. |
| Uppercase hexadecimal digits | Go writes lower-case hexadecimal, including newline's a digit. |
| Shorten minimum hexadecimal width | Go pads to four characters; mutant pads to three. |
| Treat U+FFFF as supplementary | Go writes hexadecimal escape text; mutant writes raw U+FFFF. |
| Exclude z from ASCII passthrough | Go returns z; mutant writes hexadecimal text. |
| Add extra negative padding | Go writes backslash-u-minus-001 for -1; mutant writes an extra zero. |

The sixty-four previous mutants are rerun in the complete package. Every executed test name and first independent Go mismatch is recorded in evidence/mutant-witnesses.log. Their detailed mutation definitions remain in ../landing-f8013f0/REPORT.md and batch2 through batch7 reports, plus the original helper reports. Original option/policy and first slot02 tests cover Go/source Node/native; batch2 onward also compares emitted JavaScript. No extra emitted-JavaScript coverage is claimed for original tests.

# Superseded work and limits

superseded-control-overlap.log and superseded-control-reproducibility.log record the initial overlapping selection; no duplicate control helper is delivered or credited. The owned Go exporter, driver, corpus and mutant list now contain only the three retained helpers. No shared harness or compiler blocker was encountered.

No native regexp engine, whole-rule findings/fixes/suggestions or full repository gate is claimed. The bounded corpus is not an exhaustive Unicode sweep or arbitrary-regexp/configuration fuzzer. Caller sizes must be exactly representable integer Go widths, rune inputs must be signed int32, and malformed numeric/context adapters are outside the Go-input projection. The error/cause representation requires the caller's arena adapter for object identity. Earlier external Tailwind live/corpus gaps remain outside this batch. No new rule is introduced, so no rule.json kinds registration or shared Diagnostic integration change is needed.
