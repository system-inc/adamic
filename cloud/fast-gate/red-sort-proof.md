# Red-sort proof

Run with the configured Go toolchain:

```
timeout 120s python3 -m unittest discover -s cloud/fast-gate -p run_test.py -k PlantedRedSortProof -v
```

`PlantedRedSortProof` builds a temporary Go module with two packages. Main's
`legacy.TestMainRed` really fails an assertion; `answer.TestAnswer` initially
passes. The candidate changes `Answer()` from 1 to 2. Both records come from
`go test -count=1 -json -timeout 90s ./...` and production `Gate.finish()`, including
real dependency-content hashing. Main's finished record is published as a real
Git tree at `gate-logs/<main12>/20261009T081700Z/full-main` in a local origin.
The candidate uses the production remote lookup and sorter, with no mocked main
record or candidate sort.

Expected and observed: one main red, one candidate red, no infra or stale units,
and `candidate reds: 1` in the final status. The all-mains mutant changes the
candidate count to zero. Running the same planted proof against it produces
exactly one assertion failure and zero errors; the mutation test requires this.
The two proof tests passed on October 9, 2026.

The fixture proves the verdict path. It does not claim a live gated code landing
on main. Watcher scheduling and integration's landing/trailer use remain separate.

## Published interface

Both `fast.json` and `full.json` carry `tools_fingerprint`, the exact SHA-1 of
watcher `boxTools()`'s non-recursive `git ls-tree HEAD -- <boxSide>` bytes.
Ledger rows carry `input_hash`, `input_paths`, and their declared `inputs`;
an unavailable hash is null with `input_hash_error`, never matching evidence.
`phase_units` also records aggregate phases and individually reported commands.

`red_sort` has `infra`, `mains`, `candidate`, and `stale` lists. Each entry names
its unit, input hash and reason. It also publishes `candidate_reds`,
`no_main_record`, `main_sha`, `main_record`, `tools_fingerprint`, `version`,
and `status`. Lookup errors additionally carry `lookup_error`.

The status suffix is `candidate reds: N` when a finished matching-tools record
exists, or `no main record on these tools`. Stale units append
`stale: main has since fixed this, recut, don't land`. Red and void verdicts
retain their original color; zero candidate reds alone is not a green rerun.
