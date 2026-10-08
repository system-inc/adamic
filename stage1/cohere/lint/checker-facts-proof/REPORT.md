# Current checker facts unit

The current merged unit passed 167/0/1 in 1822.973s wall, against area 151/0/1 in 1329.393s on the same box. See NARROW_PROGRESS.md and PAIRED_BENCHMARK.md for the complete current proof, every top-level timing, named skip and archived logs. The material below is retained historical evidence from d845dccde before this merge; its old pending statuses and timings do not describe the current unit.

# Historical shared checker facts proof (d845dccde)

The complete lint package passed: 118 test/subtest passes, zero failures and one named pending skip. Its wall time is 3228.294s on this box. This exceeds the approximate 2700s seat target; the unchanged area baseline reaches its native 600s deadline on this box, so complete before/after budget comparison is blocked.

## What is implemented

`node-symbol-details` supplies presence, stable run identity, flags, name and complete ordered declaration records, including byte spans, source and default-library flags, immediate parent metadata, JSDoc and parameter names. `binding-declarations` selects the shorthand value binding. `symbol-provenance` supplies complete declaration ancestry and shares the identity map. Same-file selectors go through guarded askFile; one program remains owned by the harness. All answers use its existing recording and replay. The structural WritesToBinding helper runs in Adamic.

The five pilots are no-new-wrappers, no-new-native-nonconstructor, no-class-assign, symbol-description and valid-typeof. The last two adapt wave 24. Its identity map and program fields are retained once; provenance registration uses the shared selector handler, source-less declarations return errors, and both rules declare their program reads and use askFile. The unregistered alias helper is not advertised as an available fact.

## Commands and observations

Setup used `GOPROXY='https://proxy.golang.org|direct'` and the printed environment. Its wall time was 125.021s. All Go commands used `/workspace/adamic-tools/env.sh`, GOMAXPROCS=4 and GOFLAGS=-buildvcs=false. nproc is 5; the CPU quota is four cores.

- `go test -count=1 ./bridge/tsgo/checker`: PASS, 0.200s. `go vet ./bridge/tsgo/checker`: PASS.
- `go test -count=1 -json -timeout 30m -run '<five pilot mutant subtests>|^TestOwnedWitnesses$' ./stage1/cohere/lint`: PASS, 410.388s, seven test/subtest pass events. The exact selected names and command events are in pilots.jsonl.gz.
- `go test -json -count=1 -timeout 60m ./stage1/cohere/lint`: PASS, 3222.262s from Go, 3228.294s wall. ADAMIC_TYPESCRIPT_SOURCE points to a clean 6.0.3 checkout at 050880ce; ADAMIC_LINT_PROFILE_DIR and ADAMIC_LINT_PROFILE_SNAPSHOTS point to the same fresh directory; ADAMIC_LINT_BENCH=1. There are 118 test/subtest passes, zero failures and one named pending skip. The wall time exceeds the approximate seat budget; the same-box area baseline was stopped at its first failure.
- TestRulesAgree passed 4,051 unique source/rule/options combinations in 266.76s. Owned witnesses compare live Go and native with both replay runtimes, including findings, fixes and suggestions.
- TestCompilerAndStage1Agree passed the 881-file corpus in 686.07s. Its historical first run reached the native command's 600s deadline; the complete rerun clears that unchanged deadline after the metadata field-name change.

All native pilot checks use -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all and the inherited leak/release checks. The one named skip is TestCheckerBridgeRefusalPending: `awaits codex/tsgo-errors-as-values: tsgoInspect must return TSGoError from the C error buffer`. Library commits are absent.

## Mutants

Each of the five pilot mutations was caught by Go comparison on Node, emitted JavaScript and sanitized native: ignoring local wrapper shadows, ignoring local nonconstructor shadows, dropping class writes, dropping global Symbol reports, and dropping valid-typeof's options. A merged interface/class witness proves class assignments even when the first declaration is an interface; fixture names are isolated to avoid global-symbol pollution.

Four bridge overlays failed their contract check as intended: assigning every symbol identity 1 loses shadow separation; truncating ancestry loses SourceFile; allowing Node.Text on a computed parent name panics. The last failure is runtime behavior, not a compile warning. A fourth overlay returns a valid frame containing only the first declaration; the contract check rejects its missing second merged declaration. The unchanged contract then passes in 0.126s. Direct-versus-guarded selector equality is also checked.

## Timeout diagnosis and remaining coverage

The original area at c8f6d74f passed its 872-file filtered compiler test in 986.984s, including a cold native build. A later exploratory native-only run over the 881 paths took 623.334s for the area scanner and 631.200s for the facts scanner. Those exploratory runs include brief scratch builds and a temporary metadata field-name experiment; they are not a frozen-input timing pair. Both return normally outside the gate deadline. The final isolated area measurement takes 598.691s and hashes every input before and after, with none changed; it is recorded in results.json.

The isolated metadata-name candidate completed the same 881 input paths in 566.115s versus 598.691s for the area, with no changed input hashes. The metadata fields now have declaration/tag/symbol names; every answer field and transcript input remains intact. This is an observation, not a proven explanation of the speed difference. The separate lean decoder projection was slower on an interleaved sample and is discarded. The full package passed after adopting the names. No hash, transcript field, guard, corpus input or timeout was removed or relaxed.

The final full run completed all 80 registered mutants and the 881-file compiler-corpus comparison, which passed in 686.07s. TestRulesAgree passed all 4051 combinations in 266.76s; OwnedWitnesses passed again. Historical first-red counts below are retained as diagnosis, not the current package result. Wave 05's supplied case counts are worker evidence, not local reruns. Its five questions are mapped and ranked in CHECKER_FACTS.md; they are not claimed implemented by this slice. Its other-file client still parses foreign sources and needs migration behind askFile. The unavailable supplied head and fetched replacement are recorded there. Aliased targets, module graphs, generic signature facets, lexical scopes and wave 19 compiler-option facets remain distinct contracts. React HIR and capture/SSA analysis remain outside this unit.

## Complete rerun and budget measurement

The all-input rerun passed in 3222.262s as reported by Go, 3228.294s including runner startup and exit. It recorded 118 passes, zero failures and one named pending skip. All 80 registered mutants were caught. The median one-minute load was 1.28076171875, maximum 5.2607421875. nproc is 5 and the CPU quota is four cores. Every profile input was set and both profile directories were the same fresh directory. See package-complete.jsonl.gz and results.json.

The isolated native corpus candidate remains 566.115s against the area's 598.691s, with all 881 source hashes unchanged. This is not a complete-package budget proof. The complete area baseline at c8f6d74f was attempted on the same box with the same toolchain and all inputs. It fails TestCompilerAndStage1Agree after 727.78s: its native scanner is killed at the unchanged 600s deadline on 872 files. Seventeen tests passed, one failed and the inherited refusal control skipped before stopping. This is a partial baseline, not a full package timing. See area-package-first-red.jsonl.gz. Mutant binaries deliberately bypass the persistent binary cache, so a warm rerun cannot be assumed to remove their build cost. No check, transcript input, source hash or timeout was removed.

The functional facts unit is green, but the approximate 2700s seat budget remains unproved. Its measured full package wall is 3228.294s here. The isolated candidate native corpus is faster than the area on frozen identical inputs, and the candidate clears the full native deadline the area misses, but those observations do not establish a complete-package budget. The initial baseline attempts using the wrong inherited Go executable and a background process terminated before a result are excluded from all test counts and timing claims.

Wave 07's five unnamed consumers and wave 16's unnamed eighth are kept separate from the 48 distinct named symbol consumers. These workers still need their own rule certification. The seven named wave 16 counts total 812 reported upstream cases; they are not local reruns.
