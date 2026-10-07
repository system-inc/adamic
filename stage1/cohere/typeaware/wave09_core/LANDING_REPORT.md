Built: rebased the entire wave-09 branch onto current main and repaired the flag-driver callback annotation.
Commits: main e011f8f6; old tip aa38a538 rebased to ebb4bf95; the subsequent landing commit contains this repair and new evidence.
Commands/output: reran the original wave suite, checker/bridge and Node oracles, label suite and every owned regex component verifier.
Mutants: compiling semantic mutants and released-handle checks were rerun; individual results are retained in validation-landing.
Not covered: full constructor tracking, the exact native pattern engine, all options/upstream fixtures and the repository-wide gate.

## Landing and scope

The only branch pushed by this unit is codex/typeaware-wave-09. Fetch completed
and its twelve commits rebased without conflicts onto origin/main
e011f8f60899586d6373a5ccb07335ad82cfbf3c. No further rules were claimed.
The old remote tip was aa38a538d7af5d1d933ab9a43e066b1e4c75c127.
The rebase preserves prior ownership and withdrawals, rule sources and evidence.
The exact remote old tip is the lease for publishing the requested rewritten
history. This is the user's requested landing rebase, not an unrelated rewrite.

The rebased compiler rejected the anonymous panic-only PatternCompileResult
callback in the isolated flag driver under its nominal ancestry check. A named
function with an explicit PatternCompileResult return annotation fixes that
refusal. It still panics immediately if pattern compilation is reached. It does
not supply an empty successful compiler result at runtime. Shared compiler,
generator, harness and registration sources were not edited for this repair.

## Former compiler boundary is now clear

The combined native literal source parser/listener now builds in 6.331 seconds
and completes its oracle. The previous 120-second freshness/lowering timeout
is historical evidence, no longer a blocker for this driver on this main.
This observation does not identify which main commit resolved the slowdown.

Actual source controls pass in normal native, ASan/UBSan/LeakSanitizer, source
Node and emitted JavaScript: default 22 findings / 10971 bytes, allowEscape
17 findings / 8806 bytes. Full findings include fixes and all suggestions.
All four executions also match production Go on the pinned 77 compiler files
(0 findings / 7859 bytes) and 287 repository files (0 / 18485 bytes).
The clean-running native suggestion-range mutant differs at byte 1426.
This certifies the literal slice, not constructor handling or regex flag
overrides through constructor calls.

Single whole-process literal-slice native/Go timings: compiler 1.588440 /
0.595793 seconds; repository 0.218745 / 0.364848 seconds; default positive
controls 0.007812 / 0.118360 seconds. These are single runs during concurrent
validation, not isolated throughput medians. The label suite's three-round
medians were compiler 3.452170 / 0.654381 seconds and repository 0.498045 /
0.329245 seconds, also with other validation running.

## Reproduction

Source /workspace/adamic-tools/env.sh. Setup is reused from this workspace
(previous cloud/setup.sh 88 seconds); nproc remains 5. Every run writes logs.

```
ADAMIC_WAVE09_ARTIFACTS=/workspace/wave-09-landing-original ADAMIC_WAVE09_REPOSITORY_MANIFEST=/workspace/wave-09-validation/repository.manifest ADAMIC_WAVE09_COMPILER_MANIFEST=/workspace/wave-09-validation/compiler.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/TypeScript-050880ce59e30b356b686bd3144efe24f875ebc8 go test ./stage1/cohere/typeaware -run '^TestWave09' -count=1 -timeout=30m -v > /workspace/wave-09-landing-original.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/label_var/verify.py > /workspace/wave-09-landing-label.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/verify_helpers.py > /workspace/wave-09-landing-helpers.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/misleading_character_class/verify_patterns.py > /workspace/wave-09-landing-patterns.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/misleading_character_class/verify_literal_tokens.py > /workspace/wave-09-landing-tokens.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/misleading_character_class/verify_offsets.py > /workspace/wave-09-landing-offsets.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/invalid_regexp/verify_quotes.py > /workspace/wave-09-landing-quotes.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/invalid_regexp/verify_format.py > /workspace/wave-09-landing-format.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/invalid_regexp/verify_frontend.py > /workspace/wave-09-landing-frontend.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/misleading_character_class/verify_literals.py > /workspace/wave-09-landing-literal-boundary.log 2>&1
go test ./bridge/tsgo/checker -count=1 -v > /workspace/wave-09-landing-checker.log 2>&1
go test ./bridge/tsgo -count=1 -timeout=15m -v > /workspace/wave-09-landing-bridge.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -timeout=10m -v > /workspace/wave-09-landing-node.log 2>&1
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware > /workspace/wave-09-landing-vet.log 2>&1
```

The original wave suite passes in 181.375 seconds, checker in 0.400 seconds,
bridge in 126.030 seconds and filtered Node oracle in 14.059 seconds. Vet passes
with an empty log. Every component verifier above passes. The original three rules still match
production Go on all controls and both corpora, normal and sanitized. The
label oracle also reruns scope meaning, released handle and retained-handle
mutants, with its parser-rejected undefined-label fixture explicitly recorded.
The complete regex rules remain partial. The native regexp2-compatible engine
and constructor/reference/constant-expression adapter are still work remaining.
The independent cooked/raw mapper is held, but not yet wired to constructors.
No additional batch is landing-ready merely because these slices are green.

The component token verifier's old success text says the combined native parser
is still blocked. Its scope remains token-only, but that boundary statement is
superseded by the successful full verify_literals.py run documented here.

Rerun mutant coverage: prefer-const write eligibility, radix base-36 judgment,
assertion fix end, type-parameter flag, declaration-file fact, label scope-name
membership and value/variable mask, flags precedence, emoji-modifier lower
bound, class surrogate folding, malformed-pattern guard, rune bell escape,
compiler-error prefix trim, suggestion edit end and cooked/raw byte offset.
All semantic mutants build and run; the independent Go comparison catches
them. Native/source Node/emitted JavaScript comparisons pass for pure helpers
and the full literal slice. Released-handle retention mutants fail the required
panic expectation. Bridge input/output length mutations trigger ASan; leaked
output and heap/region ownership mutations trigger LSan; wrong positions fail
Go bytes and removed link opt-in fails its refusal expectation. Exact outputs,
first differing bytes and run timings are saved in validation-landing.
