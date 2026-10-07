Two unknown-parameter sites examined; no unsound assertion was admitted.
Node prints wrong1 from a function typed number after Debug.type<number> on a string.
A concrete runtime-check source mutant stops instead; removing either compiler assertion-proof check fails its boundary test.
This group remains 2/2; the tracked total remains 13 before the enum group.
Debug.assert needs tagged unknown truthiness and a proven postcondition; Debug.type<T>'s empty body cannot prove its generic assertion.

Sites: debug.ts:213:28 (assert) and 365:29 (type). Exact cut, semantic mutant and
production diagnostic are in unknown_assertions.a, unknown_assertions.json and
unknown_assertions.results.json. The latent census observes parameter representation
errors while continuing through checker-rejected programs. Production is stricter:
asserts cond requires a boolean parameter, and an empty assertion body is refused
because normal return has not narrowed its argument. Representation support alone
cannot fix Debug.type. This is a necessary soundness refusal, not a harmless
missing storage slot. The narrow sound acceptance is a concrete checked narrowing,
or a non-narrowing type-only documentation function with a separately authorized
source adaptation. Neither policy was changed here.

TestDebugUnknownAssertionBoundaries independently pins both production refusals.
The boolean-parameter guard mutant instead reaches an unproved normal-return
refusal; the empty-postcondition mutant reaches unknown-representation NotYet.
Both tests fail specifically with Debug assertion boundary lost and no Go build
failure. The source mutant adds an actual number check, changing Node behavior
from wrong1 to a loud uncaught error. No native success is claimed for this group.
