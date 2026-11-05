# 99r compiler census

Both are fresh ordinary per-file censuses of the same complete after-tree, built
with the meter's `go build -o BINARY ./stage3/census/tool` command. There is no
measurement overlay, refusal suppression, partial-root substitution or backend
change in either table. Both completed with exit 0. "Refused" and "NotYet" are
recorded compiler outcomes, not successful native compilation.

The tables below follow exactly the 73 roots whose first stop in 99q was
core.ts:736:36. Each root is counted once and groups use the exact reason.
All 79 TypeScript roots were independently checked and lowered in both runs.
The 83 raw rows include three non-TypeScript files and one whole-program row;
those extra rows are retained and excluded from the 79-root results.

## main 82b0c076395636800764ed1af7ada569681fc2ae plus 99l, 99m, 99q and 99r

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 73 | Refused | `src/compiler/core.ts:736:36` | a cast the runtime can't check |

## main 82b0c076395636800764ed1af7ada569681fc2ae plus 99l, 99m, 99q and 99r, plus diagnostics/brand-mints-2 b834a6a4f4e5ca92ae326cbe8f4f2fe9c2e82f4a

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 73 | Refused | `src/compiler/core.ts:736:36` | a cast the runtime can't check |

## Other six roots, identical in both runs

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 2 | NotYet | `src/compiler/commandLineParser.ts:293:35` (1); `src/compiler/diagnosticInformationMap.generated.ts:13:45` (1) | reading an enum before its runtime initialization; move the call after the enum declaration |
| 2 | accepted | - | - |
| 1 | NotYet | `src/compiler/debug.ts:243:59` | a generic instantiation whose type argument for T, constrained to Node, can't be read |
| 1 | NotYet | `src/compiler/semver.ts:71:9` | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |

The two accepted roots are corePublic.ts and hostErrors.ts. The whole-program row
retains the ordinary single-entry limit for 79 roots; it is not a per-root result.

## Provenance and scope

The source is the full new-stack apply of base b729c7e0. comparison.json verifies
all 746 files, permits only the pinned core/checker owners to differ, and proves
core JavaScript and both public declarations byte-identical. source-manifest.json
freezes every after-tree source file; the manifest was checked again after these
runs. Both compiler binaries read this exact same tree without rewriting it.

The first binary is unmodified main at 82b0c076. The stack adds only adaptations;
its production compiler sources equal that main exactly. The second was built
from the clean detached scratch merge ec09e5346dfc07b8a911a716690a8d8b0779e689:
main plus stack snapshot a8a00ea9805805e4e7be5eb0a5648e2d53253ef1, merged with
exactly diagnostics/brand-mints-2 at b834a6a4. No newer diagnostics commit or local
compiler repair was substituted. The snapshot pins and emitted source are the
same as the final adapter; later contract hardening only strengthens refusal of
moved owners and does not change this source. The final full-apply source check
also compares against the frozen manifest.

compiler-provenance.json records both binary SHA-256 values, exact commits,
Go build information, configured `-buildvcs=false -trimpath`, and resolution of
the lower package to the scratch merge's own source directory. VCS stamps are
not embedded under those configured flags, so source identities and binary
hashes are recorded independently. The two ordinary runs ran concurrently;
their timings are not a performance comparison.

## The pinned brand branch does not move the measured stop

All 73 roots still stop at the same core.ts:736:36 cast on the clean merge. The
actual after-tree there is `emptyArray as unknown as SortedReadonlyArray<T>`.
This is a measured limitation, not an inferred next stop or an assertion that
this stack has successfully unblocked those roots. The four core spellings,
source admission and runtime oracle proofs remain valid; the compiler was not
changed to make the table look successful.

A separate checker-clean control, brand-control.ts, also has the ordinary cast
refusal with both binaries. The pinned branch's existing
TestBrandMintsOfTscsSortedArrays fixture fails before lowering with TS2322 and
TS2345 on unchecked indexed values. brand-fixture.log.gz retains that red result.
These controls are supplemental, outside the 79-root tables. They do not establish
that the pinned mint feature is working on this merged compiler. No speculative
cause is substituted for the exact recorded refusal.

## Reproduce the ordinary tables

```sh
# From the exact compiler checkout for each label, with the configured tool PATH:
go build -o BINARY ./stage3/census/tool
BINARY /workspace/scratch/99r/after/src/compiler OUTPUT.jsonl > OUTPUT.log 2>&1
```

census-main.jsonl.gz and census-brand.jsonl.gz retain all raw rows. Their logs
retain each root's outcome. census-results.json retains all 79 normalized root
results for each exact stack label and the excluded raw rows. No single-cast
lowering success, full checker native output, or erased relativeComplement body
proof is claimed by these ordinary first-stop measurements.
