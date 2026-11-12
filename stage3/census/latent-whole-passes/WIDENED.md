# Widened-read attribution on current main

Main: **`529705a1e6346ed7c5babaf9b4449e97e8abef38`**. It contains #z5d0jye step 1 **`b4884718248c0498e5fca5b3acdd9dd8322520ea`** and the view-entry fix **`9ae9f9888b29ef3108952246ce755a707c722c2f`**, `chg_p17rtqz1fksx39vh60errayw4r`. The original REPORT base a93ac0b94 already contained both. The same worktree now merges current main without rewriting the prior report branch. Compiler sources match this main's lower tree; only scratch overlays add observations.

**The 243 are census state-leak artifacts, not overload-check entries.** All 265 historical widened-read units reproduce on this main's own adapted checker-clean tsc entry. None has an overload-check helper. The 243 `.kind`/`Node` failures have **no recorded viewSources entry at all**.

## Split of all 265

| Requested entry category | Units | What the trace establishes |
|---|---:|---|
| (a) overload check call | 0 | No overloadCheckCalls helper in any of the 265 products. Per overloaded function: none; skipPartiallyEmittedExpressions, skipParentheses and getThisContainer each contribute zero to this set. |
| (b) CheckedCast | 8 | Recorded cast entries; source expressions and targets listed per unit. |
| (c) recorded predicate narrowing | 14 | Recorded predicate entries; all source functions and locations listed per unit. |
| (d) orphan checked-field registration from a discarded predicate preflight | 243 | All are the reported `.kind` through Node words. The registry survives, but its view records do not. There is no IR view entry to name. Each discarded registration's actual TypePredicate source location, function and target is named in the ledger. |
| Total | 265 | |

By *intended registration origin* the predicate family is 257 (14 recorded + 243 orphaned); it is wrong to present the latter 243 as actual flow entries. `followup/evidence/entry-functions.csv` gives counts per source function (nonexclusive when a unit has several entries); `followup/evidence/widened-ledger.json` preserves every one of the 265 units, exact stop/read location, all recorded and discarded entry sources, types, source-function names, unknown-flow reasons and control outcome. There are no overloaded-function counts to apportion: category (a) is empty.

The pass intentionally merges real view entries into sentinel sites and weighs all `viewLie`s jointly. Multiple source entries in a unit are reported as a set, not as a fabricated unique producer. Unknown flow is distinguished from a reached entry. All 8 cast units and 14 recorded-predicate units remain after isolation; this does not certify them as whole-tsc production regressions.

## Cause and discriminating control

`stage3/census/latent/lower.go.txt`'s `latentDeferredStop` says it probes without changing unit state, but uses `probe := *l`. Only `probe.refusal` is reset. `probe.result == l.result`, so `refuseNode` → predicate registration → `viewSchema` writes CheckedFields into the real unit IR. Appending to `probe.viewSources` changes the copy's slice header; the original unit's header stays empty. The trace records the shared result pointer and the discarded registrations, rather than inferring an overload from a checked field name.

At `checkViewWidenedReads`, the keyed field-name registry is nonempty, receiver flow is Unknown, and `viewVouched` on an empty lies list returns false. Therefore `inside && !viewVouched(...)` refuses a raw `.kind` read with no view entry. This is the direct path for all 243. Source names are TypePredicate AST nodes, including registrations whose narrows flag was never finalized.

A scratch-only control replaces that shallow preflight copy with `*l.latentSnapshot()`. It changes no production check and keeps the same 265 unit selection, body-lowering path and tail order. **All 243 disappear; exactly the other 22 remain at the same read sites.** All 251 logged preflight-probe events switch from shared to separate result pointers. The first attempted control patch failed to match formatted whitespace and was rejected as invalid; only the asserted, corrected snapshot control (`isolated3`) supplies these results. No census/production source fix is committed here.

## Production whole-program answer, held to Node

The whole probe calls real `load.Load` and **production `LowerWithOptions(..., Options{})`**. Its overlay only logs tail entry and widened-pass entry/failure; it does not disable, skip or replace a pass, and it emits no native program. This uses the fixture alternative in the task, not a claim that every compiler root was swept.

| Whole root | Node output | Production whole | Scoped prototype | Isolated-preflight control |
|---|---|---|---|---|
| predicate.ts: `isIdentifier(node: Node): node is Identifier`, body reads node.kind | true | accepted; tail and widened pass reached | same `.kind`/`Node` NotYet at 3:64 | accepted |
| widened.ts: Node-preserving narrow branded overload, result read through Node | 1 | accepted; tail and widened pass reached | accepted | not needed |
| literal-overload.ts: Broad.kind:number → narrow overload returning NarrowNode → Node.kind:1&#124;2 | 1 | NotYet at 11:21 | identical reason and location at 11:21 | not needed |

The last fixture's exact reason is:

> a read of field kind through type Node, which may hold a checked view's value

Its production trace identifies **tag_overload_check**, source call 9:17, source Broad, target NarrowNode, known entry flow (site 1, Unknown=false), and an incompatible held field type. Thus a genuine overload-entry widened read **does** stop in whole-program lowering with the same words. The preceding predicate fixture demonstrates that the 243-shaped scoped failure can instead be an artifact: production accepts it, and isolating the census preflight removes it. Do not undo the overload view-entry fix to silence this census artifact.

The actual tsc entry was also tried in production. It stops earlier at `sys.ts:1654:13`, `NotYet: node:inspector.Session.post`, before any tail marker. I do not infer that no other tsc root can reach the tail; the decisive comparisons above are complete, small whole programs. The failed broader-object overload pilot stopped before the widened pass and was not used as the qualifying fixture.

## Evidence and checks

All final children exit 0 with their actual status retained; each has a 1,700-second bound. Exit 0 means the measurement completed, not that lowering accepted. Independent whole-file refusal scans are omitted on both scoped variants, as in REPORT.md; production whole-program runs retain their real preflights. The historical 265 source-unit locations exactly match the current 265 measured locations. The two scoped streams have complete relation trailers without overflow. `followup/report.py` asserts exhaustive unit identity, the split, all 243 removals, the unchanged 22, whole-tail markers and Node outputs before rendering this report.

Raw streams, status records, provenance, source-function metadata and fixture source are under `followup/evidence/`. Go observation code is `.go.txt`. The companion [PIN.md](PIN.md) records full-mode generator results separately, including any body wall; it is not folded into these counts.
