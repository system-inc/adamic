# Full checker.ts census

Each accepted attempt uses LATENT_FULL=1 and LATENT_ONLY_FILE=the completed tree's checker.ts, a 90-minute wall cap, and the common 30-minute child-output idle cap. The earlier preview-contaminated baseline attempt is excluded; all 746 baseline source hashes were verified after restoration before the accepted attempt started.

## before

Completed: false. Child exit: -9. Seconds: 2928.867. Stop reason: None.

The baseline census received SIGKILL while the shared cgroup recorded an OOM kill. No complete baseline checker finding stream was emitted.
Memory limit: 17179869184 bytes.

Last emitted unit: /workspace/or-predicate-before/src/compiler/checker.ts:30917:5.
No complete finding stream was produced. Missing owner findings must not be interpreted as zero findings.

## after

Completed: false. Child exit: running. Seconds: 0. Stop reason: None.

Last emitted unit: /workspace/or-predicate-after/src/compiler/checker.ts:30535:5.
No complete finding stream was produced. Missing owner findings must not be interpreted as zero findings.

The direct table includes every event with a finding or source-site location in each owner. The attributed table additionally includes every event raised while lowering an owner or its nested unit whose physical finding lies elsewhere. Boundary events remain labeled Boundary. All full raw JSONL, unit timings, exact sources and process results are retained in PROOF-EVIDENCE.tar.gz.

The standalone original/mutant witnesses restore the binder refusal, but both adapted arities remain blocked on main by the exact R1 checked-claim NotYet recorded in README.md. A timed-out checker census does not waive that failed acceptance condition.
