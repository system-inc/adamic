# Batch 6: adaptation 44 on current main

Rebased 5b9012a9 onto c4c59914. The patch-table conflict retains main's
42/43 rows and adds 44; the measured total is 79 files, 5278 additions and
5239 removals. No compiler, upstream source, API reference or gate code changes.

All commands source /workspace/adamic-tools/env.sh, use four CPU affinity,
and write output directly to logs. The complete lanes have a 900-second
process deadline; focused Node proofs have a 90-second deadline.

- bash /workspace/batch6-main/stage3/lane/run.sh /workspace/batch6-baseline:
  PASS, 541.675 seconds.
- bash stage3/lane/run.sh /workspace/batch6-44:
  PASS, 513.678 seconds. Both lanes have 106366 passing, one identical known
  Public APIs failure, zero pending, and only api/typescript.d.ts differing.
- node stage3/adapt/44-config-generics/proof.cjs /workspace/batch6-baseline
  /workspace/batch6-44 /workspace/batch6-44-proof: PASS. Twelve reviewed edits;
  ten built JavaScript and two public API artifacts identical; baseline diff
  bytes identical; exact source population and LF/idempotent checks pass.
- node stage3/adapt/44-config-generics/check-types.cjs
  /workspace/batch6-44/adapted-tree /workspace/batch6-44-final-types.json: PASS.
  Zero consumer diagnostics, four wrong-return diagnostics, one narrower-default
  subtype diagnostic, nine diagnostics for the declined raw-field generic.

Eighteen mutants caught: thirteen reviewed-source/predicate guards, one
wrong-return checker mutant, one falsy-input runtime mutant, two real artifact
bytes, and one lane-count mutant. adapt44-proof.json retains fresh receipts.
The first checker run overlapped the baseline and hit its 90-second deadline;
an isolated rerun passed in 38.106 seconds, and the final candidate rerun passed.
No whole Go package or full Adamic gate ran; no new fixture or native claim.

Setup: Go 0.028s, Node 0.031s, markdown 0.100s, submodules 0.101s,
clang 0.200s, Go build 74.564s, deferred binaries 74.710s, cache 74.711s,
total 74.750s. nproc=5, cgroup quota=4 CPUs; measurements use CPUs 0-3.
