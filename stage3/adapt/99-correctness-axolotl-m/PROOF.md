# 99m proofs

The only source change is `append<U>(result, v)` inside flatMap. Exact commands,
exits, and timings are in measurements.json; full logs are retained compressed.
Compiler SHAs, dependency pins, binary hashes, and build commands are in
compiler-provenance.json. The stock checker is TypeScript 6.0.3.

| Check | Result |
| --- | --- |
| Prescribed timeout-wrapped setup | Exit 0, 1698.823 seconds |
| Full apply from pristine pinned TypeScript, including product packaging | Exit 0 |
| Ordinary keyed apply replay into a fresh output | Exit 0; all source bytes identical |
| Direct adapter replay | Zero edits and writes; bytes and mtime preserved |
| Owner contract | Changed and partly adapted owners refused before any write; unrelated edits and LF/CRLF preserved |
| Stock TypeScript 6.0.3 site query | U inferred at append and addRange; zero core.ts diagnostics before and after |
| Adamic types, all sorted 79 roots | Zero diagnostics before and after; both exit 0 |
| Normal upstream compiler builds | Exit 0 before and after |
| Both public declaration files | typescript.d.ts and tsserverlibrary.d.ts byte-identical |
| Complete core.ts transpiles, comments kept and removed | Byte-identical |
| All emitted JavaScript from the normal compiler builds | Byte-identical |
| Entire adapted src tree | Only core.ts differs, by the one explicit type argument |
| Final census source snapshot versus full apply output | Every source byte identical |
| generate_shards.py --check | Exit 0 |
| Stage3 lane fixture Go tests, serial scheduling | Exit 1: two fixed 55-second API-generator wrapper deadlines |
| Underlying API-generator Python tests, run directly | Exit 0; all 22 tests pass in 204.344 seconds |
| Ordinary current-main per-file census | All 79 compiler roots measured |

The lane oracle is skipped under the requested type-only exception.
comparison.json records byte comparisons, SHA-256 hashes and sizes for both
public declarations, every emitted JavaScript file, and both complete core.ts
transpiles. All other src files are byte-identical, so the adapted source tree
has identical JavaScript as well.

The lane fixture command uses exactly 99l's selection and serial scheduling:

```sh
go test -count=1 -parallel=1 -skip 'TestShard_|TestProduct_' ./stage3/lane/
```

`ADAMIC_TYPESCRIPT_SOURCE` names the pinned upstream checkout. The first attempt
ran beside full apply and census. Three Python wrappers hit their 55-second
deadlines, and the cold product checkout hit its wrapper deadline. An idle retry
still hit the two 55-second API-generator wrapper deadlines.

The bytecode-cache-only retry also hit those two deadlines. The unchanged
command still failed with Node 24's standard bytecode cache warmed through
`NODE_COMPILE_CACHE=/workspace/scratch/flatmap-node-compile-cache`, and with
`--max-semi-space-size=128` appended to the inherited `NODE_OPTIONS`. All four
logs are retained. These settings change module compilation and garbage-collection
cost; tests, test selection, deadlines, and checker settings are unchanged.

The final idle runs fail only `TestLaneGenerateApiRejectionsPython` and
`TestLaneGenerateApiReconstructionPython` at their 55-second wrapper deadlines.
The unchanged underlying `test_generate_api` module passes all 22 tests when run
directly with `python3 -E -S -B -m unittest test_generate_api` in 204.344 seconds.
This does not replace the failed Go check. No compiler source or dependency pin
was changed.

The initial baseline apply completed every adapter but its product archive
failed with tar's file-changed check because the baseline compiler build was
started before packaging finished. That log is retained as
apply-before-packaging-failed.log.gz. The final full apply and keyed replay
completed successfully, with output mutation kept sequential to packaging.
The baseline normal compiler build itself passed.

The inherited complete 99l census is the before comparison for the 73 moving
roots. A current-main baseline rerun reproduced the exact line-409 refusal,
then was deliberately interrupted as redundant. Its partial output is retained
and is not presented as a complete census. The final census uses the unmodified
ordinary current-main tool. Its complete source snapshot is byte-identical to
the full apply output; the comparison is recorded in comparison.json. Current
main's compiler, CLI, census, dependency pins, and internal sources are identical
to the flatten base, as recorded in compiler-source-comparison.log.gz.

No overlays, refusal suppression, backend changes, or partial-root fallbacks
were used. CENSUS.md gives the final first-stop table and the stop selected for
future analysis. The one requested commit changes only this directory.

The ordinary meter's extra whole-program row reports `lower: stage 0 compiles a program from one entry file, got 79`.
That is its single-entry lowering limit. It does not affect the 79 independent per-file results
or the passing 79-root Adamic types check. The raw row is retained.
