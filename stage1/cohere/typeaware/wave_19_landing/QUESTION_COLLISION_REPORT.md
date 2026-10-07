Built: namespaced two existing checker questions and their files to avoid incompatible concurrent-worker protocols; six rule algorithms unchanged.
Commits: based on current main e8ba3d5d; previous pushed tip 71120dde; this report accompanies the new owned-branch fix.
Commands and outputs: six-rule byte oracle PASS 580.479s; checker, five raw fact mutants, vet and filtered Node PASS; native/Go medians below.
Mutants: ten byte-only mutants, five fact-contract mutants and three retained-handle mutants caught; the filtered Node one-byte mutant passes its detection test.
Not covered: three shared production registrations, lint JavaScript parity and full repository gate; no new claims, no main or area branch push.

## Concrete collision and fix

Refreshing 473 origin refs found resolved-callee registered on wave 20 and program-modules registered on waves 02, 13 and 21. Their files use the same filenames and Go method names as our prepared questions, but their wire schemas differ. Wave 20 omits our declaration-file boolean and body kind/ranges/full source. Other program-module versions expose different library/source-text fields and specifier shapes. These implementations cannot share one question name without breaking a consumer. Exact branch commits, paths and source hashes are retained in evidence/question-names/conflicts.json.

Our questions are now wave19-resolved-callee and wave19-program-modules. Both sides use wave19_resolved_callee.go/.a and wave19_program_modules.go/.a; the Go methods are wave19ResolvedCallee and wave19ProgramModules. Native consumers, owned test overlays, direct fact contracts, mutant helper paths and the unapplied three-line registration patch use those names. The payload fields and native rule decisions are unchanged. declaration-ancestry remains unchanged because the refresh found no conflicting registered question.

No shared facts.go, registration generator, harness or protected compiler file was edited. No coworker implementation was copied. The actual production archive still refuses declaration-ancestry, wave19-resolved-callee and wave19-program-modules with panic 70; this boundary is tested. The updated wave_19_next_registration.patch passes git apply --check and remains unapplied under Ahra's shared-file restriction. Earlier reports describe the historical unprefixed names; this report records the current names.

## Revalidation

All six rules were rerun against unchanged production Go. Full canonical findings, fixes and suggestions match on 212 original controls (171 findings), 49 timeout controls (24), 58 process-output controls (32), 43 blocking-stream controls (24), original isolated-declaration option controls (133) and index controls (4). Both the 77-root compiler and frozen 287-root repository populations match for every runner, including ASan/UBSan/LeakSanitizer. All three continuation questions refuse released handles in normal and sanitized builds. The known let.if() parser refusal remains explicitly checked. No oracle success is claimed for the pending production registrations.

324 stdout/stderr/manifest/contract-log streams are saved in evidence/question-names/streams.tar.gz with hashes.json. Complete agreement, checker, contract mutant, vet and Node logs are beside them. The Node filter reran compiler comparisons and the one-byte mutation check; it reused the existing matching observation cache, as allowed for worker gates. The full bridge gate was not repeated for this naming change; its prior passing evidence remains in the landing report.

Byte-only and released-registry receipts:

```text
    wave_19_stream_test.go:136: no_process_exit_after_output.a mutant exit 0, empty stderr, only Go bytes catch byte 50
    wave_19_stream_test.go:257: require_blocking_standard_streams.a mutant exit 0, empty stderr, only Go bytes catch byte 51
    wave_19_stream_test.go:258: require_blocking_standard_streams.a mutant exit 0, empty stderr, only Go bytes catch byte 14330
    wave_19_stream_test.go:387: retained-registry mutant exits 0 with empty stderr; all three required-panic checks kill it
    wave_19_test.go:192: generic mutant: exit 0, empty stderr, byte oracle catches byte 51
    wave_19_test.go:192: dot mutant: exit 0, empty stderr, byte oracle catches byte 6460
    wave_19_test.go:192: array mutant: exit 0, empty stderr, byte oracle catches byte 10861
    wave_19_test.go:220: isolated-option checker mutant: exit 0, empty stderr, byte oracle catches byte 51
    wave_19_test.go:220: index-key checker mutant: exit 0, empty stderr, byte oracle catches byte 1332
    wave_19_test.go:220: symbol-origin checker mutant: exit 0, empty stderr, byte oracle catches byte 2290
    wave_19_test.go:278: released-registry mutant: exit 0 caught by required panic 70
    wave_19_timeout_test.go:142: timeout mutant: exit 0, empty stderr, Go bytes catch byte 50
    wave_19_timeout_test.go:212: released-registry mutant: exit 0; required panic check kills it
```

All five raw contract mutants (ancestry alias, ancestry name guard, global augmentation, callee return flags and module target) compile and are killed by their checker contracts. They are distinct from rule verdict byte mutants.

## Native against Go

Three alternating whole-process rounds per corpus and runner were measured after builds/tests finished. Count-only outputs agree in every round; full-field parity is established above. Medians are seconds. All these corpora have zero findings, so no findings-per-second claim is made. Native remains slower.

| Runner | Corpus | Go | Native | Native / Go |
| --- | --- | ---: | ---: | ---: |
| original | compiler | 0.468974 | 2.176297 | 4.641x |
| original | repository | 0.190963 | 0.355326 | 1.861x |
| timeout | compiler | 0.415711 | 2.408378 | 5.793x |
| timeout | repository | 0.195234 | 0.383517 | 1.964x |
| process | compiler | 0.454104 | 2.125009 | 4.680x |
| process | repository | 0.182178 | 0.339034 | 1.861x |
| blocking | compiler | 0.464065 | 2.086133 | 4.495x |
| blocking | repository | 0.189758 | 0.364983 | 1.923x |

Raw benchmarks and phase records are in evidence/question-names/bench. Original setup remains 137s (Go/clang/Node 0s, submodules 1s, cache 137s), nproc 5; retained /workspace/adamic-tools/env.sh was sourced.

## Commands

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE19_ARTIFACTS=/tmp/wave19-namespaced/original ADAMIC_WAVE19_TIMEOUT_ARTIFACTS=/tmp/wave19-namespaced/timeout ADAMIC_WAVE19_STREAM_ARTIFACTS=/tmp/wave19-namespaced/process ADAMIC_WAVE19_BLOCKING_ARTIFACTS=/tmp/wave19-namespaced/blocking ADAMIC_WAVE19_RELEASE_ARTIFACTS=/tmp/wave19-namespaced/released ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript go test ./stage1/cohere/typeaware -run '^TestWave19(AgreementAndMutants|TimeoutPendingRegistration|ProcessPendingRegistration|BlockingPendingRegistration|StreamReleasedHandles)$' -v -count=1 -timeout 30m > /tmp/wave19-namespaced-agreement.log 2>&1
go test ./bridge/tsgo/checker -count=1 -v > /tmp/wave19-namespaced-checker.log 2>&1
python3 stage1/cohere/typeaware/testdata/prove_wave_19_next_facts.py /tmp/wave19-namespaced/fact-mutants > /tmp/wave19-namespaced-fact-mutants.log 2>&1
go vet ./... > /tmp/wave19-namespaced-vet.log 2>&1
go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw|call_targets_.*|devirtualize)\.a$' -count=1 -timeout 15m -v > /tmp/wave19-namespaced-node.log 2>&1
```

No test output was piped. Only this owned branch will be pushed. Shared registration completion remains blocked; no further rule reservation is made.
