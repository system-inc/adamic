# View guard roads ed3c0354 against 16429a30

Unlinked artifact names below are relative to `rerun-ed3c0354/`.

Task 3zex15c. Target: `ed3c0354e2d3f97663778e91aac7f100a47b4e18`, found on `memory/view-guard-roads-3`. The compiler was built from that exact checkout. The parent and its cached streams remain `16429a30`; the same parent-adapted TypeScript 6.0.3 tree is used on both sides.

## First-error result

0 raw guard findings and 0 distinct findings are added against the parent. 0 refusal observations occur across 0 units. There are 0 probe errors and 0 diagnostic errors. All 635 recorded views in 4058 attempted units have the same per-unit population as the parent. The complete normal inventory also retains 504 deferred generic units.

| First widened road | Parent observations | Head observations | New raw findings | New distinct findings |
| --- | ---: | ---: | ---: | ---: |
| elements | 0 | 0 | 0 | 0 |
| index values | 0 | 0 | 0 | 0 |
| union arms | 0 | 0 | 0 | 0 |
| base pair check | 0 | 0 | 0 | 0 |

Of the earlier 289 exact raw findings, **0 remain and 289 are gone**. There are **0 new findings outside those 289**. The independent view-location/source/target comparison finds 0 previously refusing views still refusing and 0 new refusing views; this avoids counting a changed first reason as a cleared view. Of the earlier 114 units with no ordinary finding, **0 still have a latent guard stop** and 114 have none. Current ordinary-finding-free guard-stopped units total 0.

The stage3/census/pair raw/distinct results are preserved in [delta.md](rerun-ed3c0354/delta.md) and delta.json. Parent-to-head cleared findings, moved findings and changed guard-first stops in already guard-blocked units are all zero because the parent guard had no refusals. The comparison to the earlier head is separate from those parent-to-current counts.

## Remaining and new refusals

[Every refusal and its source reading](rerun-ed3c0354/FINDINGS.md) is listed by road, with exact reason, source text, root types and native path diagnostics. Refusing observations containing an actual split crossing without literal-field separation: **0**. Refusing observations containing a checker-identical but pointer-distinct pair at the root or any path step: **0**. Separate machine-readable lists are split-crossing-cases.json and checker-identical-distinct-object-cases.json. These categories can overlap.

The identity fixture [identity.ts.txt](identity.ts.txt) records one view and zero real refusals at this head, as at the parent. The scratch mutant that refuses every pair emits one refusal on each side, proving both observers are active. The complete extended observer source is byte-identical on both sides and builds on both. All four fixture runs preserve the cached parent's ordinary bytes apart from checker-cache telemetry. See identity-result.json and raw/identity-*. The production lowering source was never modified by the measurement seam.

## Ordinary outcomes and measurement validity

The normal first-error output retains all **5767 ordinary finding observations**. Serialized finding arrays are byte-identical to the seam-free cached baseline (SHA-256 `da230b61cdc7d4e729e15cb6a0e71d78dc80eb895f734b00900f30cd9ea84463`). Every header, file, unit and finding byte outside the relation_report telemetry is identical (SHA-256 `9a0fb2a329003b458ee59f4962d4dbdb16013a3382017a247ffbe15fa188e4d9`). The excluded telemetry differs only in checker relation-cache sizes. There is no ordinary regression, changed ordinary first stop or new panic. normal-verification.json contains the assertions.

[The method](rerun-ed3c0354/METHOD.md) specifies exactly what the seam populates and what it does not model. This is **units only**: it does not model whole-program graph membership. A latent refusal does not establish a production whole-program stop. The only observer extension is post-refusal identity/literal-field metadata at the root and every real failure-path step, plus recognition of the new split-arm road label. The identical extended observer is also built on the parent and checked with the fixture control. The parent census has no refusals, so its cached stream needs no supplemental annotations.

## Full mode

No full-mode work was started for ed3c0354: the first-error pair has no new findings, so the requested conditional full-mode check does not apply. The archived 769b0e53 full-mode results describe only that earlier head.

## Time and peak memory

Each wrapper records the child's own exit code and wait4 peak RSS. Times include checker loading and census startup; the build row measures the go child and its compiler descendants. Parent census measurements are reused, not rerun. The 56.464-second parent diagnostic build overlapped the head census; these operational timings are not a controlled speed comparison.

| Run | Child exit | Elapsed seconds | Peak RSS KiB |
| --- | ---: | ---: | ---: |
| Head build | 0 | 35.499 | 1455640 |
| Head first-error tsc entry | 0 | 2781.883 | 4587344 |
| Identity fixture | 0 | 0.087 | 31996 |
| Identity forced-refusal control | 0 | 0.095 | 32060 |
| Parent diagnostic build | 0 | 56.464 | 1467404 |
| Parent identity fixture | 0 | 0.321 | 32852 |
| Parent identity forced-refusal control | 0 | 0.163 | 32372 |
| Reused parent first-error guard census | 0 | 4888.397 | 4624656 |
| Reused parent seam-free baseline | 0 | 4966.145 | 5101784 |

The superseded 8e57a877 first-error run was stopped on the target change after 273.022 seconds and 1168 timed units (child exit -15, peak RSS 1328768 KiB); there is no completed pair for it. No d4e6f92c or d3757b95 census ran. No superseded or current census process is left running after completion.

Raw streams and per-unit timings are beside this report in raw/; files over 1 MB are gzipped. Exact overlay sources use .go.txt names. analyze.py, write-report.py, provenance.json and the archived parent evidence make the comparison reproducible. The earlier completed census remains [REPORT-769b0e53.md](REPORT-769b0e53.md).
