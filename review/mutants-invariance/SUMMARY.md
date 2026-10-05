# Mutants against B2's invariance checks (claude/stage1-language-gaps-o3lbxx, 1e2ef21)

Stopped when the account's credit ran out: 21 of the 38 mutants in `mutants.py` ran, m22 to m38 didn't.

Each mutant is one change to internal/lower/invariance.go. "Branch tests" are the branch's own:
internal/lower and the invariance_readonly.a oracle fixture. "Probes" are the 85 round-twelve B2 probes in
review/round12/b2, each compile verdict compared with the unmutated baseline (`base.verdicts`). Per-mutant
lines are in `results.txt`.

## Survivors: pass the branch's tests and change no probe

| Mutant | What it breaks | Probe written to catch it | Unmutated verdict |
|---|---|---|---|
| m03 | no walk into a property's own type | `kill/k03_nested_readonly_field.a` | refused |
| m06 | a readonly tuple target treated as writable | `kill/k06_readonly_tuple_view.a` | compiles |
| m15 | canWrite doesn't look inside a readonly property | `kill/k15_constraint_readonly_nested.a` | refused |
| m16 | canWrite doesn't look inside a readonly container's elements | `kill/k16_constraint_readonly_array_elements.a` | refused |

The four probes were checked only on the unmutated compiler. None was run against its mutant before the
credit ran out, so they're written to catch these, not yet shown to.

## Caught only by my probes, not the branch's tests

- m10: a function's return type isn't compared. Caught by b43 and b44.
- m21: two Sets are never paired. Caught by b22.

The branch's tests would want a case for each.

## Can't be caught

- m12 is equivalent: it disables `narrowedFrom`'s `from == to`, which `widened` has already returned on.
- m19 is equivalent in practice: pairing containers of different kinds matters only for an intersection of two
  container kinds. I didn't write a probe for it.

## Caught by the branch's tests

m01, m02, m04, m05, m07, m08, m09, m11, m13, m14, m17, m18 and m20. Some probes caught them too:

- m01 changed five probes, including the four readonly-widening ones (b2_silent, b53, b54, b55).
- m04 changed 26, m05 changed 9, and m20 changed 2.

## Not run

m22 to m38:

- m22 to m31 and m36: viewSite's sites, one at a time (initializer, assignment, call and new arguments, return,
  array element, property, arrow body, parentheses).
- m32: literals no longer exempt.
- m33: class instances walked.
- m34: undefined kept in unions.
- m35: methods walked.
- m37: the visited set ignored.
- m38: container elements skipped.

`python3 drive.py <worktree> m22_site_initializer ...` resumes with them. Mutating the
file in place needs a worktree of its own, at 1e2ef21.
