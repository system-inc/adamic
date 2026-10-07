Rebased all 68 retained helpers; no new helper claimed or built in this landing unit.
Previous publication a70d0c52; tested merge dc66cf97; final rebased source c8d7f443; publication SHA is named in the final response.
All 24 actual helper packages PASS uncached; vet/format, seven Node input probes and a repeated statement-order oracle PASS; setup 28.060s, nproc 5.
All 290 distinct compiling semantic mutants caught again; every fresh witness is in evidence/landing-mutants.json.
Not covered: full repository gate, seventeen required stage 1 external-input checks, whole-rule findings or regex matching engine. No skipped check was credited.

## Bases and scope

Main advanced from b6b1538b to 4e0bfda50a19c705a1aac0d9932e08483806d61c; area advanced from d3a37422 to bb2ece564842c4b2f909b9f75c27e74c2efa4f29. Because the owned branch was no longer landing-ready, this unit made no reservation. All existing claims were already ported, tested and pushed. Only codex/lint-helpers-05 has been published by this worker, and only that branch is pushed here. Main and all area branches are integration-owned.

All 94 commits rebased cleanly onto main. The newer area adds witness options and recovery-row classification to the shared lint harness; it was merged onto the owned branch without conflicts or shared-file edits. The resulting tested source is dc66cf97ac36e91d66d097b7614d60944b22e41c. Upstream changed native emission, borrowing, string runtime, compiler-unit construction and oracle fixtures, so this landing reran every retained helper package and semantic mutant. The entire helper subtree remained byte-identical: Git tree 31fe64ad4dd85129ea8cdbb2bffdd5c28ce62d05. No helper, corpus, test or assertion was changed for this landing.

During the gate, main advanced again to 71d7e491b3c9724f7a0e2ee754592149e7f9790b. That merge changes 413 paths, all under stage3/. The final rebase used --rebase-merges to preserve the lint-area ancestry, yielding c8d7f443b7b20ffde5d0543a5e9664d43833d396. Both current main and current area are ancestors. All 3527 Git object entries under cmd, internal, oracle, cohere, dependencies, stage1 and compiler configuration/library paths are identical between the tested and final sources. Thus the fresh complete helper gate applies to the final source; it was not repeated solely for unrelated stage3 files. A final filtered helper oracle, static checks and all seven Node input probes were rerun on that source. Exact source/object identity is recorded in evidence/landing-input-identity.json and landing-final-input-identity.json.

Original batch-24 reservation fa546b14 was published before source a70d0c52. The final rebase maps them to e8efd273 and 832e1cd4 respectively; the final report commit follows them. Readiness remains 68 helpers, 365 prerequisite occurrences across 73 consumers and 50 helper-ready rules under the frozen adapter assumptions. No new rule or diagnostic integration is claimed.

## Commands and observed results

Every build/test shell sources /workspace/adamic-tools/env.sh; output goes directly to logs, never through a pipe. Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc reports 5, cgroup quota four cores, 17.6 GB.

```
bash cloud/setup.sh > /tmp/lint05-batch24-landing-setup.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -p 1 -count=1 -v -timeout=30m ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/slot05/... > /tmp/lint05-batch24-landing-helpers.log 2>&1
go vet ./... > /tmp/lint05-batch24-landing-published-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05 > /tmp/lint05-batch24-landing-published-format.log
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch15 -run '^TestSlot05StatementOrder$' -count=1 -v -timeout=10m > /tmp/lint05-batch24-landing-final-order.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch24-landing-final-oracle.log 2>&1
```

The complete helper command exited zero: all 24 packages PASS, no failures and no skipped tests. Individual package durations sum to 2540.017 seconds; this is not a wall-time or speedup claim. All original Go/Node/native comparisons and the newer emitted-JavaScript comparisons run as specified by each retained package. Native baselines and semantic variants run under ASan/UBSan with the existing success and stderr checks. Every variant must compile and execute successfully before a Go-result or dependency-trace mismatch counts. No warning, refusal, panic, nonzero exit or sanitizer-only failure earns semantic credit.

All 290 distinct variants were freshly caught: four inherited option/message variants and 286 slot-05 variants. The extra final statement-order rerun repeats one of those variants; it is not counted as a 291st. Its count-preserving reverse visits produce mutant `2|stmt:2;stmt:1;` versus Go `2|stmt:1;stmt:2;`, output line 27, byte 7. Final filtered helper PASS 3.090s. Vet and formatting exit zero with empty logs. Seven uncached Node input probes PASS 1.164s, zero hits and seven misses; the newly upstream empty-path fixture is included. The first post-runtime Node probe run also passed, 7.939s.

Every package and duration is in landing-package-results.json. Every exact fresh mutant trace/result witness, owning test and package is in landing-mutants.json; the complete gate log is retained. Existing per-package reports define their bounded corpus, dependencies and variant source replacements. Inputs were not trimmed and guards were not relaxed. Go's unspecified case-group ordering remains an external input where documented by those packages.

Setup timing lines:

```
setup: node ready (0.022s)
setup: go ready (0.026s)
setup: submodules ready (0.062s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.164s)
setup: markdown dependencies installed (npm ci, integrity verified); step-duration=1.701s
setup: markdown dependencies ready (1.763s)
setup: go build ready (27.802s)
setup: test binaries deferred (use --warm-tests) (28.031s)
setup: build cache warm (28.032s)
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (28.060s)
```

The complete build-flags/environment line is retained in landing-setup.log. Deferred setup test-binary warming is not a skipped correctness check; every selected package was built and executed in the uncached gate.

## Limits and publication

The full repository gate and its seventeen required stage 1 external-input checks were not run. None was skipped, relaxed, deleted or reported passing. This unit uses the bounded worker gate over all owned helper packages plus the filtered external oracle. It does not add a regex matcher, new rule listener, finding model or option adapter, or change shared registration/harness/compiler source. Upstream changes, including the leak-check work, were taken intact. All 68 claimed helpers are complete, and there is no extra reservation.

Publication rewrites only the owned branch with an exact force-with-lease against previous remote a70d0c52e130c5bd469be8a137165710cb11f313, as required by the user's explicit rebase instruction. No main/area push and no pull request. Current-base ancestry and remote SHA are checked before and after that publication.

The complete raw Go helper log is stored losslessly as evidence/landing-helpers.log.gz. One witness prints an empty Go value with a trailing space; gzip preserves that exact evidence without a text-diff whitespace change. Its uncompressed SHA-256 is recorded in landing-package-results.json. The raw /tmp log is unchanged.
