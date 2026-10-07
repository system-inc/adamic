Lowering wall median: 18.013 -> 1.255 s; freshness CPU median: 16.22 -> 0.45 s on markdownblocks list_probe.ts.
Built: persistent immutable facts, shared snapshots, changed-fact joins and incremental escape closure without changing freshness proofs.
Commits: implementation e6b74f955b3fca3a39266b35e4031cb09807d98b; finder baseline 9efb67e63a946016103199eb427c96040e66f482; current main merge 53766c7deb012595f40553df7c85579ff53dd5c4.
Verification: 1,106 identical lowering decisions and 4,436 identical write proofs; full fresh/lower packages and filtered Node oracle pass; cache mutant changes ten refusals to acceptance.
Limits: Linux, filtered native oracle, finite pinned corpus; recency still scans the heap; no full repository gate or macOS run.

## Measurements and cause

Three alternating, otherwise isolated profiled runs time Lower only, excluding Load and profile shutdown. Old walls: 18.821, 18.013, 17.982 s; new: 1.078, 1.255, 1.720 s. Old cumulative fresh.ProveWrites CPU: 17.07, 16.22, 15.72 s; new: 0.45, 0.40, 0.67 s. The fixture SHA256 is 541e36fdb51601fec0ce350cb54568a012e3dc69d62ede5b3257aad1b35d45a9. All runs accept it. CPU figures are sampled cumulative profile time, distinct from wall time. Raw profiles, complete cumulative listings, timings and provenance are in [freshness-speed evidence](../internal/lower/cycletools/evidence/freshness-speed/).

The profile shows repeated copying and joining of accumulated state, followed by scanning the whole heap to close escape facts. Old round two spends 11.02 CPU seconds in state.copy, 3.86 in join and 1.73 in closeEscapes. New round one spends 0.01, 0.19 and 0.03 respectively. Copying growing maps at successive instructions produces superlinear work. Facts now use a persistent treap: snapshots share roots; writes copy changed paths and mutable leaf fields; joins visit changed facts and skip shared subtrees. Escape closure starts from new exposure and changed edges on exposed objects. The pre-call escaped snapshot shares its root. Recency renaming still scans facts, so this is not a claim of universal linear complexity.

## Decision and proof equivalence

The comparison executes old and final freshness against the same current lowering engine. Temporary Go overlays print every completed ProveWrites result; production diagnostics are unchanged. The comparator checks full lowering diagnostics and every write's Kind, Site, Name, Function, Proven and Why. The complete manifest contains 561 oracle fixtures, 319 stage 1 programs, 54 census reproducers, 16 load cases, 77 pristine census programs and 79 adapted census programs. The upstream census revision is 050880ce59e30b356b686bd3144efe24f875ebc8. Input hashes and category inventory are recorded.

There are zero differences across 1,106 inputs: 739 accept, 180 checker errors, 111 NotYet and 76 Refused. Freshness runs in 90 programs and produces 4,436 identical write proofs; other programs stop before it or do not invoke it. The compressed ledger preserves both proof vectors and decision hashes; full diagnostics were compared before compression. Four old/new arena trace pairs are byte-identical: arena and parent mutant on production lowering, and both with the separately queued arena-correction overlay. Production's preexisting arena decision is preserved; this speed unit does not claim to land that separate correction.

Reproduction uses `python3 internal/lower/cycletools/freshness.py 9efb67e63a946016103199eb427c96040e66f482 SCRATCH`, then `python3 internal/lower/cycletools/compare.py SCRATCH/old SCRATCH/new CENSUS OUTPUT --manifest MANIFEST --freshness --jobs 4`. The evidence inputs.json supplies the pinned manifest; pristine/adapted census roots must be materialized at their recorded paths. Temporary overlays are instrumentation only.

## Mutant and tests

The cache mutant replaces shared-node identity with mere key presence when deciding whether a fact changed. The full corpus decision diff catches ten Refused -> accepted results. A separately executed proof comparison over 43 freshness refusal probes catches the same ten decision differences and ten proof-vector differences (66 total writes). Caught probes include call_grandchild, call_escaped_parent, call_leaks_parent, call_owner, caught, global_then_function, next_iteration, older_instance, one_call and wraps_argument. The minimal mutation and ledgers are committed evidence.

`go test -buildvcs=false ./internal/fresh -count=1 -timeout=30m` passes in 122.739 s. Full lower passes in 105.071 s. The initial combined log contains an accidental duplicate test-field compilation error in fresh; that temporary edit was reverted, and the complete fresh rerun supersedes it. Persistent-facts property tests compare snapshots and branch joins/escape closure against ordinary maps and the full closure algorithm.

`ADAMIC_GATE_UNCACHED=1 go test -buildvcs=false ./internal/oracle -run 'TestFreshWriteProbesStayRefused|TestNativeAgreesWithNode/internal/oracle/testdata/(fresh|closure|regexp_cycle)' -count=1 -v -timeout=30m` passes in 70.835 s, with 34 Node and 50 native observations. Vet passes; formatting output is empty. All test output goes to files.

Toolchain: GOPROXY=https://proxy.golang.org|direct; source /workspace/adamic-tools/env.sh. Setup timings: Node 0.046 s, Go 0.051, markdown skip 0.011, ready 0.114, submodules 0.123, clang 0.399, build 49.942, deferred 50.169, cache 50.171, done 50.232. nproc=5, CPU quota=4. No cohere source was copied. The full markdown layout runtime oracle was not rerun in this freshness unit; its lowering and proofs are covered by the complete comparison.
