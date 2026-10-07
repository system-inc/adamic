Built: refreshed twelve existing ports and required external correctness inputs; no new claims.
Commits: tested source 9cd819bc2, green evidence 1ec895f2e; final audit commit follows on the same worker branch.
Commands and outputs: 18 owned steps and 16 external root checks pass; 100 passing test events, zero skips/failures; 627-ref audit has no eligible rules.
Mutants: rule/listener/dispatch/graph/fact/data/ownership and external comparator witnesses are listed in REQUIRED_INPUTS_REPORT.md and its archived events.
Not covered: parked React IR/SSA/capture analyses, four production bridge routes, shared checker context, own emitted JavaScript and the full repository gate.

Latest rebase and rerun: [LATEST_LANDING_REPORT.md](LATEST_LANDING_REPORT.md).

Fresh all-head selection covers 197 ranked names, 627 origin refs and 33 distinct
Markdown claim blobs. Original production ports, main/base native sources and
rule descriptors, and every origin claim are excluded. No eligible rule remains.
No new wave-07 reservation was made.

Both current main 39638d9e and integration d65a8f93 remain ancestors of the green
worker branch after the final fetch. Only codex/typeaware-wave-07 is pushed.
This final commit contains only audit proof and documentation; tested code is
unchanged. Required-input suite receipts explicitly report zero skips and zero
failed events. Source/proof integrity verified all 2,113 archived streams.

[REQUIRED_INPUTS_REPORT.md](REQUIRED_INPUTS_REPORT.md) records exact commands,
versions, test filter, setup timing, every mutant detector and remaining blockers.
Setup passed in 94s with nproc 5. Quiet compiler JSX medians are 1.989-2.084s
native versus 0.290-0.313s Go; repository medians are 0.339-0.344s versus
0.149-0.174s. Every timed output also matches Go byte for byte.
