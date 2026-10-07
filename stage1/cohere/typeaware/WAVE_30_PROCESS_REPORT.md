Built: completed `.a` process-exit-after-output and require-blocking-standard-streams ports; all third-batch rules complete.
Commits: claim 952a900c; timer 80e56233; process implementations 182d776e; evidence committed separately and pushed on codex/typeaware-wave-30.
Commands and outputs: process gates PASS 135.651s; graph PASS 37.95s; timer recheck PASS 53.924s; full bridge PASS 77.041s; vet PASS; setup 20s, nproc 5.
Mutants: exit empty-state byte 57, blocking empty-state byte 86, graph throwable fork byte 931, timer range byte 56; all exited 0 and only independent comparisons caught them.
Not covered: full repository gate, full oracle fixture matrix, or shared production registration integration; private native suites execute the ports.

Both rules compare every finding, fix and suggestion field against the pinned Go production registry. These rules produce no fixes or suggestions, whose zero counts remain compared. Process-exit: 60 controls, 39 findings, 23989 identical bytes. Blocking-streams: 64 separate program fixtures, 30 positive programs. Graph foundation: 18 controls and 10969 identical bytes against Go's production graph. Controls include foreign helpers, Node ambient overloads, console identity, callbacks, generators, nested catch/finally, destructuring, class initializers, type-only imports/exports, computed require, imported entrypoints, top-level await and await using. Production tests supplied 44 blocking programs; another 20 controls expand boundary coverage.

Each rule agrees on the frozen 77-file TypeScript 6.0.3 compiler corpus (5318 bytes, zero findings) and 287-file repository corpus (18485 bytes, zero findings). ASan, UBSan and LeakSanitizer runs agree on controls and both corpora. Five new questions reject released handles with exact panic text and exit 70. Five strict-suffix bridge mutants are individually caught by TestProcessFactsLocationsAndGuards. The observed compiler BindingPattern Text() panic was fixed in the new declaration metadata file; the subsequent complete corpus gate passes. Initial failing output is preserved alongside the passing rerun.

The native code builds the graph, catch-sensitive output state, module reachability and blocking judgments. New Go files expose raw AST relationships, symbol declaration ancestry, checker-selected call signatures and source bodies, program-resolved module targets, and external-module parse context. Native callers set the existing parser awaitContext from that fact. The sole shared edit adds five question switch registrations in facts.go; no shared harness, generator, parser, or protected compiler implementation was edited.

Commands source /workspace/adamic-tools/env.sh first; output redirected to saved logs:

- go test ./stage1/cohere/typeaware -run '^TestWave30(ProcessGraphAgreement|ProcessExitAgreement|BlockingStreamsAgreement)$' -count=1 -v (initial corpus accessor failure; graph passes).
- go test ./stage1/cohere/typeaware -run '^TestWave30(ProcessExitAgreement|BlockingStreamsAgreement)$' -count=1 -v (passing complete rerun), with ADAMIC_WAVE_30_PROCESS_REPOSITORY_MANIFEST, ADAMIC_WAVE_30_PROCESS_COMPILER_MANIFEST and ADAMIC_TYPESCRIPT_SOURCE set to the preserved corpora.
- go test ./bridge/tsgo/... -count=1; go vet ./bridge/tsgo/... ./stage1/cohere/typeaware.
- go test ./stage1/cohere/typeaware -run '^TestWave30ThirdAgreementAndMutants$' -count=1 -v, with the same manifests through THIRD variables.
- python3 bridge/tsgo/profile/volume_bench.py <native> <oracle> <directory> --corpus compiler <config> <manifest> --corpus repository <config> <manifest>, three alternating rounds per rule after other test processes completed.

Median complete process time includes checker loading. Both corpora have zero findings, so this measures traversal and checker queries rather than positive diagnostic throughput.

| Rule | Corpus | Native | Go | Native/Go |
| --- | --- | ---: | ---: | ---: |
| process-exit | compiler | 1.899782s | 0.306275s | 6.20x |
| process-exit | repository | 0.247895s | 0.122854s | 2.02x |
| blocking-streams | compiler | 1.817927s | 0.322838s | 5.63x |
| blocking-streams | repository | 0.254818s | 0.130741s | 1.95x |

Exact serialized streams are gzip-compressed with SHA-256 and original lengths in validation-wave-30-process/streams.json. Logs and timing records accompany them. New authored Adamic source files use .a; test .ts and .d.ts module identities are symlinks to .a source. The private Go overlay exporter only reads production blocking test fixtures and does not change the cohere submodule.
