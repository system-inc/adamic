Updated all twelve owned native listener exports and rule.json declarations to AST kind names.
Branch codex/typeaware-wave-24 remains based on current main b8fb957a; no new rule is claimed.
Original Go test PASS 133.444s; next, third, fifth and named-listener validators PASS; vet and diff checks clean.
Caught 12 rule and 11 checker-question byte mutants, released-registry mutation, native listener and two JSON listener mutants.
Full repository gate, nondefault options and older node-only migration remain uncovered; React analysis claims remain parked.

## Corrected contract

Public kinds are strings such as CallExpression, BinaryExpression and SourceFile, matching the shared registry's strings.TrimPrefix(kind.String(), "Kind") convention. All twelve native exports now have string[] type. JSON declarations contain no numeric kinds. The owned Go listener oracles and compiled native probes compare the actual names byte for byte. The newest three still receive cached handed nodes and the private numeric dispatch implementation is unchanged; no new string-kind relevance checks or parser node refetches were introduced. The nine older engines retain their documented legacy traversal limits. These JSON files remain declaration metadata rather than fabricated shared factory adapters.

No rule predicate, diagnostic, fix, suggestion, bridge implementation, shared registry or shared harness changed. Main does not yet contain ab70f38d4; no extra harness commit was cherry-picked. No developer-tools leak-helper diff is present in the current main advance, and none was reverted. Earlier numeric declaration evidence is historical and superseded by this named contract.

The first older-listener comparison rejected KindClassDeclaration versus ClassDeclaration. That exposed the prefix normalization needed by the registry, and the Go oracle was corrected. This failed attempt is preserved in prefix-refusal.log. Both a runnable native Identifier-to-CallExpression declaration mutant and the matching JSON mutant are caught by the comparison; the newest three also catch a CallExpression-to-BinaryExpression JSON mutant.

## Validation

Every complete suite was rebuilt after the declaration changes. Original controls yield 43 findings, next 65, third 295, and fifth 125 across 217 valid controls. All suites compare the full canonical findings, fixes and ordered suggestions on both frozen populations (287 repository roots, 77 compiler roots), normal and ASan/UBSan. The frozen corpora yield zero findings. All released-handle checks pass with panic 70. The original released-registry mutant is also caught. Sources with Go parse errors are filtered only by the independent Go oracle, as documented in the existing report.

Reproduction commands, with direct output files:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE24_ARTIFACTS=/workspace/wave24-kindnames-original ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-24-corpus ADAMIC_WAVE24_REPOSITORY_MANIFEST=/workspace/wave-24-repository.manifest ADAMIC_WAVE24_COMPILER_MANIFEST=/workspace/wave-24-compiler.manifest go test ./stage1/cohere/typeaware -run TestWave24AgreementAndMutants -count=1 -v > /tmp/wave24-kindnames-original.log 2>&1
python3 stage1/cohere/typeaware/wave-24-next/validate.py /workspace/wave24-kindnames-next --stage0 /tmp/wave24-kindnames-main-adamic --compiler /workspace/wave-24-corpus > /tmp/wave24-kindnames-next.log 2>&1
python3 stage1/cohere/typeaware/wave-24-third/validate.py /workspace/wave24-kindnames-third --stage0 /tmp/wave24-kindnames-main-adamic --compiler /workspace/wave-24-corpus > /tmp/wave24-kindnames-third.log 2>&1
python3 stage1/cohere/typeaware/wave-24-fifth/validate.py /workspace/wave24-kindnames-fifth --stage0 /tmp/wave24-kindnames-main-adamic --compiler /workspace/wave-24-corpus > /tmp/wave24-kindnames-fifth.log 2>&1
python3 stage1/cohere/typeaware/wave-24-fourth/validate-listeners.py /workspace/wave24-kindnames-listeners --stage0 /tmp/wave24-kindnames-main-adamic --checker /workspace/wave24-kindnames-fifth/checker.a > /tmp/wave24-kindnames-listeners.log 2>&1
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware > /tmp/wave24-kindnames-vet.log 2>&1
python3 stage1/cohere/typeaware/wave-24-fifth/timing.py /workspace/wave24-kindnames-fifth --repository /workspace/adamic --compiler /workspace/wave-24-corpus > /tmp/wave24-kindnames-timing.log 2>&1
```

No bridge source changed, so its previously passing package tests were not broadened or repeated. Setup succeeds: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 56s, done 56s. nproc=5; cgroup quota four CPUs; memory 17.6 GB. Obsolete named scratch binaries were removed to free disk; repository sources and prior evidence were preserved.

## Mutants and timing

Original: receiver, return-await wording, unknown callback; class, then and handler facts. Next: exit, timeout and blocking rules; provenance, callee origin and module facts. Third: promise reject, rest and regex rules; direct-symbol and symbol-count facts. Fifth: require-await wording, Symbol provenance and typeof undefined wording; generic signature, index and heritage facts. Each compiled byte mutant exits successfully with empty stderr and only the independent oracle comparison catches it. Released-registry mutation fails the required panic, and listener mutants fail name comparison. Complete logs and canonical outputs are in evidence/named-kinds/; source-hashes.json hashes the updated native declaration files and JSON metadata.

Three alternating whole-process count rounds after builds stop: repository native median 270.736 ms versus Go 86.198 ms, 3.141x slower; compiler native 1520.198 ms versus Go 296.434 ms, 5.128x slower. These measure the newest three-rule suite with load/parse/walk overhead on zero-finding populations and show no native speed advantage. The sample arrays and Go phase stderr are retained. Shared-registry factory integration, arbitrary-input exhaustiveness, nondefault options and the complete repository gate remain outside this verification.

## Final current-main landing refresh

After the first named-declaration push (f8cae5767), main advanced from c01907a70 to b8fb957aa839a9e8cb0b54279dd9864fa317bd30 with the inherited-static-field emission fix. Rebased cleanly, preserving that change without editing protected files. Rebuilt the compiler with go build -o /tmp/wave24-kindnames-main-adamic ./cmd/adamic, with output in main-build.log. All four complete rule suites and both named-listener checks pass again with that compiler. Original Go test passes in 133.444s; next, third, fifth and older-listener validators print PASS. All control counts, complete corpus byte comparisons, sanitizers, 23 compiled byte mutants, released-registry mutation, three named metadata/native mutations and released-handle checks pass. Timing figures above are this final compiler's alternating samples, after builds stopped.

Also ran go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/inherited_static_field_read' -count=1 -v with direct output in main-oracle-filtered.log: the exact fixture passes in 9.393s against Node and the JavaScript/native backends; native cache hits=1, misses=2. Worker oracle caching was enabled. An initial abbreviated filter ran only the parent and is retained separately; it is not counted as the fixture check. No new compiler change was authored by this unit.

After the completed named-declaration push, refreshed all origin heads and inspected 584 refs and 33 distinct claim blobs. The conservative mention-based audit counts 172 claimed rules; no unclaimed rule remains in the frozen ranking after the known main/bridge port exclusions. No additional claim is made. The audit is claim-audit.log. Main was independently verified unchanged at b8fb957a after the final tests and timings. Full repository gate, nondefault options, older node-only migration and the three parked React analysis ports remain uncovered. Current-main artifacts are in evidence/named-kinds/rebased/.
