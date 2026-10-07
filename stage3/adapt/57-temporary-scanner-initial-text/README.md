Temporary: comes out when non-null initialization and optional scanner inputs land

Plan published before implementation. Slice scanner.ts only. Replace precisely
the first var text = textInitial! with var text: string = undefined! and its
first setText(text, start, length) with setText(textInitial, start, length).
The intervening statements declare locals; none reads text. setText assigns
newText || "" before any scanner call. Keep this adaptation pending a ruling
and implementation for general non-null initializers.

languageVersion is optional: guard its two >= comparisons with !== undefined,
which preserves JavaScript's false result for omitted versions. The shebang
regex assertion is required, guaranteed by the initial #! test; retain it.
Other required-value ! sites retain their checks. Literal undefined! resets
remain uninitialized and require compiler read-before-assignment support.
Validate the token stream and complete upstream baseline before claiming done.

Validation: four edits, idempotent. The initial proof guard incorrectly treated
the word text in comments as a read; that failed attempt did not apply edits.
The corrected guard ignores comments and rejects a planted real intervening
read. The actual rewritten slice matches all 509,014 tokens and the original
SHA-256. Combined 55-57 full baseline: 106,367 pass, zero differences/failures/
pending, 336.493 seconds. No snapshots accepted.

Decisions: absent languageVersion legitimately selects pre-ES2015 Unicode maps,
so explicit undefined guards preserve its original comparisons. A successful
shebang test guarantees a match of the same non-global /^#!.*/ regex; its ! is
a required-value check, retained. No formatting Debug members were added.

Native limitation: non-null-check 6ae58a0 admits required expression checks,
but literal undefined! has checker type never and cannot lower even in the
minimal assignment-before-read probe. The exact required text initializer is
retained; dropping it compiles in a control but changes the requested spelling.
The full slice currently first refuses tokenValue's definite-assignment ! at
scanner.ts:498:19. Neither that refusal nor the initializer lowering is closed
by this adaptation. No native scanner or ownership proof is claimed.
