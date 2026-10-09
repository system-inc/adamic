Added exact in-place a-check headers to four checker-error controls.
Gate.aCheck checked every committed speculative .a file: five files, zero mismatches.
Four type errors matched TS1101 or TS1064; one NotYet passed without a header.
Eight removed-header and wrong-code predicate mutants were rejected; every header was restored.
Nested depths 0, 1, 2 and sibling 0 still pass; no-stubs output remains exact on the control.

The unchanged gate implementation is pinned to `3bf2489e2a5e09ab59d2926b1c0ff374dfa1fa45` from origin/devtools/fast-gate. Its actual a-check step built the ordinary compiler and compiled each fixture in place. Header mutants exercised that unchanged predicate against captured real compiler diagnostics, as documented by stage3/a-check-headers. Source bodies below all four added headers are byte-identical to e46fa73d; hashes and every capture are in [header-a-check.json](evidence/header-a-check.json).

Commands run, with stdout/stderr written to named logs:

```sh
source /workspace/adamic-tools/env.sh
git show origin/devtools/fast-gate:cloud/fast-gate/run.py > /tmp/speculative-a-check/gate.py
python3 /tmp/speculative-a-check/run.py > /tmp/speculative-a-check/run.log 2>&1
python3 stage3/census/speculative/audit.py /tmp/speculative-census-final /tmp/speculative-header-control > /tmp/speculative-header-control.log 2>&1
```

The exact harness is preserved as [header-check-harness.py](evidence/header-check-harness.py); copy it to /tmp/speculative-a-check/run.py to reproduce. No whole package or full gate was run. This commit fixes the reported a-check failure only; speculative raw totals and unverified-depth status remain the partial results from e46fa73d. Depth tables are not ready and are not included.
