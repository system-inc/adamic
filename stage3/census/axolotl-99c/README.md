# 99c validation evidence

Implementation: `820d57df8395a56359c8eec0c3e109b8d7d27226` on `typescript/axolotl-99c`.
Base main tip: `80c2d3282e44b527555d59690bb413cc883d2bf2`. This evidence branch starts at that same base.
The implementation commit adds only the three files in
`stage3/adapt/99-correctness-axolotl-c/`.

## Final results

- `summary.json`: final counts and implementation identity.
- `apply-after-final.log` and `apply-proof.json`: complete twelve-declaration apply
  and proof that its input equals the baseline final snapshot.
- `measurement.json`, `diagnostics-before.json`, `diagnostics-after.json` and
  `types-*.stderr`: all 79 compiler roots, 53 to 51 diagnostics, exactly D033 and
  D034 removed, zero added or changed chains.
- `mutants.json`, `mutants.log`, `types-mutant-*.stderr`: thirteen independent
  restoration mutants. Twelve edited properties each add one diagnostic;
  configFilePath, already handled by adaptation 20, adds two.
- `verify-post-oracle.json`: both edited files and all ten built JavaScript
  artifacts are byte-identical, exact full API reconstruction, no-write
  idempotence and cross-file refusal checks.
- `compiler-input-proof.json`: stock builds preserve all 79 compiler roots from
  the measured apply snapshots.
- `oracle-final/report.json`: final checked stock build and all upstream tests,
  106,367 passing, zero failing, zero pending, zero baseline differences.
- `api-final-guard.json`, `api-expected.json`, `api-99c-final.diff`: only the
  mechanically proved API reference is accepted. The incremental API diff is
  two replacement lines. No other reference baseline is edited.
- `own-keys-final.json`: real built before/after functions preserve thirteen
  present undefined slots, including parsed empty-config include and exclude.
- `setup.log`, `build-before.log`, `build-after-final.log`, and
  `oracle-final/build.log`: setup and checked builds, including build-tests.

## Intermediate probes

The first ten-member tree is `/workspace/scratch/after`; the final twelve-member
result is `/workspace/scratch/after-final`. Files named `verify.json` and
`verify-final.json` describe the earlier ten-member tree. The final verifier is
`verify-final.cjs`, with final results in `verify-complete.json` and
`verify-post-oracle.json`.

`measurement-ten-only.json` records the first attempt: D033 and D034 disappear,
but TSConfig.exclude is exposed at the enclosing literal. `exclude-only.json`
records the next probe: widening exclude alone exposes TSConfig.include. These
are actual present undefined values in the existing conditional spread. Both
internal siblings are included in the final adapter, without public API changes.

The raw ten-member oracle (`oracle/`) has one expected API snapshot failure and
no other baseline difference. Its accepted-snapshot rerun (`oracle-accepted/`)
and the final twelve-member oracle (`oracle-final/`) both pass all 106,367 tests.

`mutants-initial.log` preserves an initially overstrict evidence assertion that
expected one error when restoring configFilePath; it actually restores D033 and
its independent undefined assignment TS2412. The final mutant runner admits and
records exactly those two and restores every source between iterations.
`build-adamic-early.log` records a build attempted before setup had populated the
compiler submodule; the later `build-adamic.log` is the successful current-main
CLI build. No checker or adaptation workaround is applied for either probe.

## Reproduction

`commands.sh` records material commands; run them from Adamic's root. Baseline
apply must run from a checkout of the recorded base, before 99c exists. All
outputs are files. Use fresh scratch output and mutant directories when
replaying. `check-types.py` discovers the complete sorted root list and compares
full diagnostic multisets. `mutants.py` runs the real CLI for every restoration.
`verify-final.cjs` checks source and build bytes, API reconstruction and refusal
behavior. `accept-api.py` permits only exact generator output at the one API
reference path. The audit scripts and JSON retain named, contextual, indexed,
spread and destructuring references across the pinned upstream source.

Tool versions are in `go-version.txt`, `node-version.txt`, and
`typescript-version.txt`. No upstream checkout, installed dependency tree or
compiler binary is committed as evidence.
