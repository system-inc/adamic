# Adaptation 41 timer-head amendment

Adamic task b5w3ycg, Nov 7 21:47:55Z. Base origin/main is `0cb816a0ec439f63542b3f83bb13941fd79f8f41`. Upstream is TypeScript 6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`. The prior deletion measurement is `typescript/core-sys-timers` at `0a5f1eb71c8c179926893b81eb90207a799ec590`, `stage3/adapt/41-explicit-any-remaining/sys-timers/REPORT.md`.

| Head 1 proof item | Result |
| --- | --- |
| Full amended apply, all adapters in order | Exit 0 |
| Source scope, all 746 src files | Only sys.ts differs, by exactly the two ambient declaration deletions |
| Actual full-apply tree equals the early census candidate | Byte-identical across all src files |
| adamic types, all 79 sorted compiler roots together | Main, amended tree, and both source mutants exit 0 with zero diagnostics |
| Stock 6.0.3 ordinary compiler build | Main and amended tree exit 0 |
| All nine emitted compiler JavaScript files | Byte-identical |
| typescript.d.ts, typescript.internal.d.ts, tsserverlibrary.d.ts | Byte-identical |
| Isolated 41 replay at its own stage | Zero edits, zero writes, bytes and mtimes preserved |
| LF/CRLF | Same reconstruction; both declaration lines deleted completely |
| Half-applied and owner-drift guards | Eight no-write refusals; unrelated edit preserved |
| sys.ts census | 123 main walls -> 120 amended walls; zero checker diagnostics |
| Ambient sites | Both ambient function refusals and both no-body observations disappear |
| Expected new #995spev site | Candidate sys.ts:1499:13, NotYet `reading setTimeout`, phase lowering; expected, not a failure |
| Restored-declaration mutant | Restores both ambient sites and the entire 123-wall baseline set |
| Additional full-finished-tree replay experiment | Original main and amended 41 both stop at existing rule 31; no unrelated guard changed |

## Small deletion form in 41

41 did not support an empty replacement: its old split(after) logic could not recognize a deleted owner. Rules 13 and 14 now have empty after strings, an explicit declaration name and the shared deleteGroup node-global-timer-heads. Their before strings remain the exact pristine upstream declarations, as required at 41's own position in the apply order. The adapter validates both parsed top-level declaration shapes before planning any disk write. Both exact originals are deleted as complete lines; both absent is idempotent; one absent is refused. Duplicate, drifted and moved owners are refused. The previously narrowed 41 heads are not accepted as an unreviewed pristine state. The two existing proof reconstruction helpers understand the same complete-line deletion form.

The full-tree replay experiment is explicitly separate from own-stage idempotence. Later adapters change the SourceMapSource text pinned by existing rule 31. The original main adapter also refuses that finished tree at rule 31. This amendment does not weaken that unrelated guard. The isolated before-41 tree is extracted from the real full-apply stage refs, applies successfully, and then replays with zero edits/writes.

## Census and mutants

The current-main compiler and standard guarded latent overlay are rebuilt. Physical scratch Go evidence files end in .go.txt. LATENT_FULL=1 and LATENT_ONLY_FILE=sys.ts select sys.ts while loading all 79 roots; LATENT_ASSERT_NO_OUTPUT=1 remains enabled. No speculative or replay flags are enabled. Distinct walls use (main-mapped location, kind, exact reason, phase); boundary events are not compiler stops. The legacy checker-rejected label is retained although these programs have zero checker diagnostics.

Deleting lines 57-58 of fully adapted main shifts subsequent coordinates by two. The two ambient Refused sites and two NotYet no-body sites are gone. The one new lowering NotYet is `reading setTimeout`, at amended 1499:13 (main 1501:13), the expected first-class value read owned by Runtime #995spev. The source mutant restores the exact main declarations, reproducing the entire main wall set. No site is suppressed or treated as successful lowering.

## Head 2 stopped at the required gate

No 99-correctness-core-sys-timer-slot directory, API sanctions, upstream ledger line, shard regeneration or head-2 branch is built. The exact ruled C-noinfer-b source counterfactual was tested first: System's two methods and the two private pollScheduled annotations are the only four edits against the declaration-deleted source. Its 79-root adamic types command exits 0. The census removes main 384:79 and the private any site, and the any-slot-restoration mutant brings 384:79 back.

The census adds the seven named expected sites plus one unlisted site:

`src/compiler/sys.ts:1519:36` (actual declaration-deleted sys.ts:1517:36), NotYet / refusal_scan: `a member typed any (require()module) related to {} at the initializer site`.

Thus the observed new-site count is 8, not the required 7. The explicit head2-gate.py child exits 1. This is an exposed initializer relation, not a claim of a runtime regression. Accepting it, changing lowering or changing another owner requires more than the requested ruling. Per the stop instruction, head 2 stops here. Its full apply, own build, public API/shard regeneration and full lane are not run. Nothing about that incomplete second head is represented as passing.

census-comparison.json retains every kept source ledger, gone/new rows, the unlisted row, and both mutant results. source-scope.json retains all source hashes. artifacts.json retains all nine JS hashes and declaration hashes. evidence.tar.gz retains actual child exits, raw streams, timings, logs and exact sys.ts source bytes. Checkpoint `14f97d1e1620840a31bdc5e8ef9d56cdb7db95f8` was pushed separately so the final amendment branch remains one commit directly on the pinned main.

## Reproduction

Source /workspace/adamic-tools/env.sh, use the pinned ADAMIC_TYPESCRIPT_SOURCE, install the pinned stage3/api parser, and run a full apply from pristine TypeScript. The retained run.py logs the child's own exit and bounds every piece at 30 minutes. contract.cjs takes the baseline applied git product and a new output directory. The early source/census drivers and comparison script retain the exact frozen scratch paths. finish-head1.py records the artifact proof; replay-control.py records the pre-existing full-tree rule-31 limitation rather than weakening it.

No compiler or lowering implementation is changed. The expected #995spev site is a named dependency for the held first head, not a proof failure. Head 2 needs a further ruling for the additional initializer site before construction can continue.
