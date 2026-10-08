Built: broad-object admission to both original tagged union targets; FunctionLikeDeclaration.parameters reads.
Commit: this checkpoint on codex/views-mixed-unions-2, based on 6858d196.
Checks: original union oracle PASS 43.488s; cast admission/refusal lower tests PASS 0.891s.
Mutants: bypass each union tag guard, valid native and JavaScript runs violate the pinned exit-70 refusal.
Uncovered: original class members read remains a named Map fallback refusal; array metadata producers and element dispatch are not certified here.

The target imports are the complete 78-file declarations from pristine TypeScript
050880ce. The oracle verifies every declaration hash. All seven function arms
and both class arms enter directly from a broad object with a numeric kind.
Every valid tag is held to Node in release native, sanitizer native and JavaScript.
A tag-80 value prints true on Node and stops with the exact named union-cast
exit-70 diagnostic on Adamic. Both executed tag bypass mutants instead print true.

Function parameters array reads pass for all seven arms. A numeric parameters
payload prints undefined on Node and stops with the pinned named array-field
refusal in both Adamic backends. Class members currently refuses at its read
because the global fallback sees an unrelated Map descriptor elsewhere in the
complete type graph. This is not reported as a certified array pair. The next
checkpoint scopes that fallback using the existing shared allocation graph.

Command: ADAMIC_BRAND_ORIGINAL_DECLS=/tmp/views-brand2-original-declarations
ADAMIC_GATE_UNCACHED=1 go test -overlay /tmp/views-union-step1/overlay.json
./internal/oracle -run '^TestCheckedViewOriginalUnionTarget(Tags|s)$' -count=1 -v.
The overlay restores the previous lazy fallback while testing this admission
commit independently of the in-progress scoped fallback change. It disables
no refusal. Raw outputs are in logs/admission.log and logs/lower.log.

Original brand candidate counts remain 32 certified pairs / 549 reads of 35 /
552. Two name pairs / two reads remain ours; the intersected Identifier pair /
one read belongs to lane 7. The original mixed primitive column has 31 pairs /
141 reads pending, including those three brand entries; the non-brand share is
28 / 138. Dictionary selection has a separate two-pair / eleven-read inventory.
Dates remain October 9 for our branded family and October 13 for mixed primitives.
