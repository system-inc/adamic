# Paired latent census: folded branded intersection

Both ledgers completed over all **79 files**, exit 0, in simultaneous first-error/no-output mode on the same adapted tree. The meter's existing `report.py.lowering_summary` passed coverage and distinct-count validation on both. Each ledger contains 4,544 units: 4,032 attempted, 498 deferred generic declarations and 14 checker-skipped bodies. The timing populations match all 4,032 attempted units exactly.

| Distinct sites | A | B |
|---|---:|---:|
| NotYet | 555 | 554 |
| Refused | 2652 | 2652 |
| error | 0 | 0 |
| panic | 0 | 0 |
| SkippedDependency | 2 | 2 |

**Zero A-lowers → B-stops units; zero A-lowers units absent or otherwise non-lowering on B.** The exhaustive lists are [stops](a-lowers-b-stops.json) and [absent or non-lowering](a-lowers-b-absent-or-stopped.json). This pair measures B against head 1; it does not substitute main for A.

Only **nullParenthesizerRules**, `src/compiler/factory/parenthesizerRules.ts:675:1`, changes first stop, remaining NotYet on both:

- A: **a checked predicate target LeftHandSideExpression over a generic argument TIn & ({} | null), read on the true side as Expression & LeftHandSideExpression**, at `src/compiler/core.ts:1799:32`.
- B: **a template interpolating an object, an array, a map, a function or undefined**, at `src/compiler/core.ts:1801:59` (cast's failure-message template).

The predicate row is the **only changed reason row**, **1 → 0**; its wording explicitly mentions a predicate and a true-side read. No view, branded-supertype, intersection, or write-through-view reason row changes. The template row stays **2 → 2 distinct sites**: B's newly reached stop is already represented in A by another caller. There is no new/growing reason row.

[Changed reason counts and up to five locations on each side](changed-reasons.md), [all unit movements](unit-changes.md), and [every A lowering-phase site absent from B](a-lowering-sites-absent-b.md) give the exhaustive lists. The latter contains exactly the predicate site at `core.ts:1799:32`; there are no other removed A lowering sites. JSON companions retain full diagnostic text and all stops.

Reading all compiler sources with the TypeScript AST found **78 direct cast/tryCast calls in 22 top-level meter units**. [All 22 A/B outcomes](caller-outcomes.md) include nullParenthesizerRules; it is the only caller unit that moves. `getTsConfigObjectLiteralExpression` and `identifierToKeywordKind` lower on both. `createTypeChecker` is checker-skipped on both (its first diagnostic is TS2322 at `checker.ts:1683:9`), so its body has no measured lowering outcome. The other 19 caller units stop on both. [The source inventory](direct-callers.json) preserves every call's location and source text.

## Time and box

Box: nproc **5**, cgroup CPU quota **4 cores** (`400000 100000`), cgroup memory **16 GiB** (host MemTotal approximately 17.6 GiB). Both launches began within 0.004 seconds. Each had a 5,400-second timeout with a 60-second kill grace; neither reached it.

| Measurement | A | B | B−A | B/A |
|---|---:|---:|---:|---:|
| Summed unit wall seconds | 981.694305 | 971.800246 | -9.894059 | 0.989921 |
| Summed unit CPU seconds | 1048.981864 | 1031.210375 | -17.771489 | 0.983058 |
| Whole-run wall seconds | 1779.622594 | 1772.418276 | -7.204318 | 0.995952 |
| Peak RSS (KiB) | 6220244 | 6473348 | +253104 | 1.040690 |

The meter includes registration, lowering and panic recovery per unit; whole-run time also includes loading, refusal scanning and output. [Top 20 units by B−A wall seconds, with ratios](timing-top20.md); [every unit (TSV)](unit-times.tsv); [every unit including CPU (JSON)](unit-times.json). nullParenthesizerRules takes 0.226951 → 0.482862 seconds (+0.255910; 2.128×).

Noise controls: **91 import declarations** execute no receiver read, destructure or predicate test, so cannot exercise this intersection rule; they still pay registration cost. Median absolute paired difference **0.011644s**, p95 **0.069616s**, maximum **0.134413s**. Median B/A **0.976**; 5th–95th percentile **0.673–1.825**. [Every control](timing-controls.md), [noise summary](timing-noise.json). B's summed unit wall time is 1.01% lower in this pair; one concurrent pair with this variation does not establish a causal speedup or a precise rule cost.

## Reproduction

- A: `3aec8230abb53ac93b92b24f5e3ae916f9cd36a8`.
- B: `5e5f57d3b530264cb09eba5712a8a362375aede2` (parent A).
- Main: `50503bc6abc22da12e9f9a0caad28b7f4bb1b20d`; A's stage3 equals main's exactly.
- TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`; adapted once from A, with all 79 input hashes in [source-manifest.json](source-manifest.json).
- Binary A SHA256: `04c72fc31d9fc1a9e3e0b26e1a154465d5d2078892a99b1106e7ad3057339004`.
- Binary B SHA256: `1c184a82c8964821b06aeafb789f2df88f0601f19f145f7e2d9ecf9f78761843`.

Exact commands: [preparation](preparation-commands.sh), [prescribed meter builds](build-censuses.sh), [simultaneous runs](run-censuses.sh), [delta and validation](compare.sh). Both detached worktrees shared setup's pinned submodules through a scratch symlink; no internal/ source was edited. Only this evidence directory is committed.

Setup completed in 358.241 seconds. The initial abbreviated A fetch could not resolve a remote ref; fetching B resolved A's full SHA, which was then fetched explicitly. The image lacked `/usr/bin/time`, so an initial wrapper launch exited before starting either census; archived initial-launch logs show that. [measure.py](measure.py) then recorded each actual run's wall time and Linux wait4 peak RSS, matching the resource definition of time's maximum RSS. The measured commands and UTC start/end times are in [A resources](a-resource.json) and [B resources](b-resource.json). Detailed AST progress files are not emitted by first-error mode; complete ledgers and unit times are present.

Raw ledgers and timing streams are archived as `a.jsonl.gz`, `b.jsonl.gz`, `a-times.jsonl.gz`, `b-times.jsonl.gz` (gzip CRC and raw hashes checked). To rerun compare.sh after archiving, first decompress those four streams with `gzip -dk`; analyze.py also reads gzip streams directly. [validation.json](validation.json), [census-delta.json](census-delta.json), [provenance.json](provenance.json), logs and SHA256SUMS retain the evidence.
