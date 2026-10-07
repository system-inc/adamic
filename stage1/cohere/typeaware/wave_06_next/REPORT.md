Built: child-process error listeners, response status checks, and independent awaits in loops, all native Adamic.
Commits: original wave completed in 027adc366b9f2fbd379df517779a2dd2dfb6ee82; continuation claim b57fbb83 preceded implementation and push.
Checks: 84 controls, 77 compiler roots and 287 repository roots compared byte for byte with Go; ASan/UBSan and released-handle checks passed.
Mutants: each rule's diagnostic identifier mutation and a wrong resolved declaration mutation exited normally and failed the Go byte comparison.
Limits: default rule configuration and these finite source populations; no full repository gate or exhaustive upstream fixture matrix.

The three rules are `nexus/correctness-require-child-process-error-listener`, `nexus/correctness-require-response-status-check`, and `nexus/performance-no-independent-await-in-loop`. Their files and native control-flow helpers live here. All new Adamic sources use `.a`. The response graph is an owned adaptation of the original wave's return graph; the original remains unchanged.

The claim followed a fetch of all origin heads and inspection of 327 references, unique claim documents, and implemented names on main and the bridge base. The combined volume ranking contained 197 checker-dependent rules; 25 of the base's 26 ports participate in that ranking. Process-exit, race-timeout and blocking-stream candidates became claimed during the final refresh and were skipped. The selected three had zero findings in both recorded corpora. No additional rules were claimed after these three.

`bridge/tsgo/checker/wave06_declarations.go` supplies raw resolved-signature and symbol declaration metadata, alias resolution, declaration parents and function body text. `wave06_declarations.a` decodes it. The only registration edit is one fallback line in this worker's earlier `promised_shape.go`. Existing shared dispatch, registration generator and harness files are unchanged. All judgments execute in native Adamic. The final validation uses the normally registered C archive, without a private dispatcher overlay. The Go oracle overlay only builds an owned driver against unmodified production Go rules.

Run from the repository with the configured toolchain:

```
source /workspace/adamic-tools/env.sh
TMPDIR=/workspace python3 stage1/cohere/typeaware/wave_06_next/validate.py --scratch /workspace/wave-06-next-registered --compiler /workspace/wave-06-typescript > /tmp/wave-06-next-registered.log 2>&1
```

| Population | Roots | Findings | Identical bytes |
| --- | ---: | ---: | ---: |
| Controls | 84 | 42 | 28864 |
| TypeScript compiler | 77 | 0 | 5318 |
| Repository | 287 | 0 | 18485 |

Each rule reports 14 positive control findings. Controls exercise resolved versus shadowed producers, listener and escape handling, status checks across branches and finally blocks, ordered loop effects, tainted exits and indexed-loop headers. Matching zero-volume corpora alone would not establish implementation behavior. Full diagnostic records include findings, fixes and suggestions; these three Go rules produce no fixes or suggestions. Canonical ordering retains duplicate findings.

The compiler checkout is TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`; compiler and repository manifests reuse the original frozen populations. Cohere is pinned at `715ba94f3608a6500086b1076ce5cb7e51b836db`, typescript-go at `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`. Setup was reused from the original wave: 77 seconds, `nproc` 5, CPU quota 4, Go 1.27.1, clang 20.1.8 and Node 24.19.0.

All three populations also match under ASan/UBSan, with empty sanitizer stderr. A diagnostic-ID mutation per rule preserves successful execution and findings counts but changes serialized bytes; only the independent comparison rejects it. A bridge mutation substitutes the call node for its resolved declaration and is rejected by the same comparison. Inspecting the new question after releasing its program produces the expected panic exit 70.

| Timed fresh process | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 4.885982 | 1.721635 | 2.838 |
| Repository | 0.611213 | 0.272065 | 2.247 |

These are single warm-filesystem wall times for the three-rule suite, including loading and parsing, excluding builds. Timing runs executed before concurrent regression checks. Full run commands, exits and elapsed times are in `validation/runs.json`; complete compressed stdout/stderr, input manifests, hashes and positive counts are in `validation/`. This evidence supports the tested populations, not a claim of universal semantic equivalence or a performance improvement.

Final regression results are recorded alongside this report. The full root test gate and all upstream fixtures were not run. No shared-harness gap remains for these three rules in the owned native runner.

Final checks passed:

- `TMPDIR=/workspace go test ./bridge/tsgo/... -count=1 -timeout=15m -v`
- `ADAMIC_WAVE06_ARTIFACTS=/workspace/wave-06-original-regression TMPDIR=/workspace go test ./stage1/cohere/typeaware -run '^TestWave06(AgreementAndMutants|PinnedFlags)$' -count=1 -timeout=15m -v`
- `TMPDIR=/workspace go vet ./...`
- `gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware` (empty output), and `git diff --check`.

The original-rule regression compared 166 controls and 129 findings (36936 identical bytes, including its sanitizer run). Its three rule mutants, promised-type mutant and released-registry mutant were all rejected by their checks. Complete logs are in `validation/regression-*.txt`.
