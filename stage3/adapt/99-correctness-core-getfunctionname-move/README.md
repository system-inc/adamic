# getFunctionName move, stopped checkpoint

Status: STOPPED, NOT LANDABLE. This commit stacks on the stopped cast-message
draft d718a08df495c4d28da3228279ecf549d8dd010d, itself on road (b) cut
2a95448b8fadbcd30549a148bd28da5f8b75f907. No function has been moved by this
branch. adapt.cjs deliberately refuses before writes, with the stop reason.

## The common section's stop rule applies

"If the replay shows it refusing because of your edit, stop and report that."

An isolated scratch probe removes the exported helper from Debug, appends a
module-local declaration to deprecate.ts with its body byte-identical, and
changes the existing call to the local name. The probe is not a type-checked
candidate implementation or a completed full apply. Its purpose is to test
the earlier guards before changing them.

| Earlier adapter | Before replay | Moved probe replay | Result |
| --- | --- | --- | --- |
| 45-regex-captures | exit 0 | exit 1 | New refusal: debug:getFunctionName: want one node, got 0 |
| 41-explicit-any-remaining | exit 1, existing rule 31 | exit 1, rule 6 | Outcome changes to missing or duplicate reviewed site: 6 |
| 30-indexed-reads | exit 1, existing core.ts:992:21 | same exit and reason | Existing earlier refusal masks its debug capture site |

45 does not pin the complete function text. It requires a uniquely named
getFunctionName implementation in debug.ts, checks the regexp exec initializer,
and proves capture 1 mandatory. Its unchanged replay breaks because of the
move, so the user-supplied common section requires stopping here. No change
to 45 or 41 is authorized past that stop in this checkpoint.

41's rule 6 is the exact `return (func as any).name;` to `return func.name;`
rewrite scoped to debug.ts. Its final-tree replay finds the latter at the old
home before the move, then cannot find either spelling there after the move.
Its earlier existing rule-31 refusal is recorded, not misreported as success.
30 also pins the match[1] expression in debug.ts, but its pre-existing core
refusal occurs first in both probe runs. That is not a new passing result.

## Road (b)'s authorized guard amendment

The user expressly ruled that road (b)'s condition-4 whole-helper guard follows
the function to deprecate.ts, pinning the same body byte for byte and refusing
a drifted body on replay. The requested move directory must sort after road
(b). This amendment is pending, not implemented: the earlier 45 replay stop
must be resolved first. Road (b)'s README records the same status; its existing
adapt.cjs and pins.json guard are unchanged by this checkpoint. The user
explicitly permitted that amendment; the blocker is the separate earlier 45
pin and changed 41 replay, not a request to reapprove road (b)'s guard.

## Required proof item status

- Both full applies and full paired replay: not run for head 2; the targeted
  before/after replay probe is the reason for stopping.
- Stock find-references and grep proving no Debug.getFunctionName reader or
  re-export: not run on a completed moved tree; no such tree is claimed.
- adamic over 79 roots and proof none imports deprecatedCompat: not run.
- Adapted tree build adding zero diagnostics: not run.
- Every normal JavaScript hunk and both public declarations: not built/compared.
- Debug census, 355:26 gone and no additions: not run on a completed move.
- Six deprecate calls retaining names, Node witness: not run. Prior counted
  total is 6, from fcb70da5e2d899df060d6f2521e35824442c9951, not a new result.
- Mutant restoring Debug and bringing 355:26 back: not run.
- Road (b) guard drift refusal at the new home: pending, not passed.
- Ledger: records the ruling and this stopped state, not implementation success.

Adamic's ruled destination remains: deprecatedCompat ships in typescript.js,
not tsc, so when typescript.js comes home after 1.0.0, getFunctionName's
reflection (Function.prototype.toString) meets us there.

Raw logs and exact copied adapter/proof sources are in pin-stop-evidence.tar.gz.
All started measurement/apply processes were stopped or completed; no compiler
or census-tool implementation is modified.
