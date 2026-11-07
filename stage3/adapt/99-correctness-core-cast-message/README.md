# cast failure message, stopped draft

Status: STOPPED, NOT LANDABLE. One draft commit on accepted road (b) cut
2a95448b8fadbcd30549a148bd28da5f8b75f907. Task b5w3ycg, Adamic's
Nov 7 22:06:52Z approval, Compiler accepts under the delegation.
Rule class: JavaScript-changing.

The chosen wording is `Invalid cast. The supplied value ${value} did not pass the test.`
The draft removes only the quoted predicate name and its getFunctionName call.
`${value}`, the condition, the return and Debug.fail call remain byte for byte.
Whole before/after cast owners and hashes are pinned in owners.json. A changed
condition refuses. An already-adapted owner performs zero edits.

## Why stopped

The required census proof asks for an existing cast site to disappear and for
restoring the original owner to bring it back. The unchanged full-mode census
records zero cast sites in all three runs. The unit is deferred_generic_instantiation,
instantiated=false. No body-lowering success is claimed. A mutant with the
original core.ts bytes still has zero findings in cast. This cannot demonstrate
the required mutant condition, so this draft must not be treated as proved.

The composed preflight starts from the caller measurement's verified tree,
main 3214dd46 plus accepted road (b), then applies the exact elementAt
amendment from 2a95448b. This matters: the pinned cut has the declaration,
whereas the retained caller-count tree has the earlier arrow. The first
preflight used that earlier tree; it is superseded and retained as diagnostic
history. Corrected original, edited and restoring-mutant censuses all have
zero checker diagnostics and identical distinct non-Boundary site sets.
On the corrected cut, cast is at core.ts:1786:1 and the message at 1789.

Both fresh full applies were started and then deliberately stopped. Their
children's actual exits are -9; the outer wrappers exit 1. stop-processes.json
records why and which descendants were killed. No full-apply success,
complete paired replay, normal build, public declaration comparison or
79-root types result is claimed.

## Count before editing

The user explicitly authorized citing the prior completed measurement:
fcb70da5e2d899df060d6f2521e35824442c9951,
stage3/census/getfunctionname/REPORT.md. It measures the same unchanged cast
body, before this edit, over all 398 lane units. It is not a new lane run on
this cut's elementAt declaration.

| Counted expression | Whole-lane helper entries | Control entries |
| --- | ---: | ---: |
| Debug.getFunctionName(test) in cast's failure message | 0 | 1 |

This does not say cast itself is never called. Its failure-message expression
has zero counted executions. The cited fixed observer's control independently
forces a cast failure and fires the counter. All 398 units pass their ordinary
guards: 106366 passes, one expected failure, zero skips; no untagged helper
entries. The six compatibility calls are separate and unchanged by this head.
The original/adapted failure text has no real-call counterexample because
that expression's observed execution count is zero.

## Ownership

The pre-write audit inspected 224 adapter/pin inputs on 2a95448b. There is no
active whole-cast owner or exact-message pin in 99q, 99r or the earlier
99-correctness-core adapters. 40 and 41 retain historical cast evidence;
that evidence is not an active whole-owner guard. Therefore this is the
user-ruled new directory. No earlier adaptation is amended in this head.
getFunctionName and debug.ts are byte-identical across the three preflight
trees. Road (b)'s existing helper guard is unchanged in this head.

## Proof item results

- Setup and locked API install: exit 0. Census tool built from 2a95448b: exit 0.
- Adapter contract: exact text-only edit, replay zero edits, condition drift refused.
- Node source witness: six failing value classes retain Error class, predicate
  call count, value interpolation and throw point. Four passing cases retain
  identity. This uses pinned source transpiled in scratch with the unchanged
  existing Debug.fail; it is not a completed normal-build witness.
- Both full applies: stopped, child exits -9, not passed.
- Full paired replay: not run.
- adamic types over 79 roots: not run.
- Normal emitted JavaScript hunks and public declarations: not built/compared.
- Core full-mode census: zero diagnostics, no added sites, but no original cast site.
- Restoring mutant: restores original bytes; required returning site absent.
- getFunctionName untouched: byte comparison passes.
- Ledger: added as a stopped, not-landable draft.

STOP.json, raw-stopped-proof.tar.gz and the source patch retain the evidence.
All logs are archived. The census overlay only renames physical scratch Go
files to .go.txt. No meter logic or production compiler file changes.

Head 2 independently hits the common section's stop rule: adapter 45 requires
getFunctionName in debug.ts and refuses after a byte-preserving move. That
move and the authorized relocation of road (b)'s guard are not implemented here.
