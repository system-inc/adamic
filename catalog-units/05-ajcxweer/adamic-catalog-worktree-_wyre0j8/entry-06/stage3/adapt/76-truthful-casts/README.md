Permanent type-only proof of the assertion cache's finite keys.

The implemented edit replaces Debug's broad MatchingKeys alias with the six
literal names that shouldAssertFunction actually receives. The cache starts
empty, stays private, and has exactly two writes: the helper writes its literal
name, and setAssertionLevel clears an already enumerated key. getOwnKeys only
collects owned enumerable keys. The existing key-array assertion therefore has
a concrete six-key provenance; 76 does not add another narrowing assertion.
Its adapter rejects any new caller key, writer, escape, or changed enumerator
before writing. Runtime statements, order and enumeration stay unchanged.
The native backend may still need checked finite-key array views; this source
proof does not claim the remaining array assertion is already admitted.

sameMap is not silently repaired. Its truthful type is readonly (T | U)[], with
both array casts removed and a (T | U)[] slice result. A scratch trial exposes
builder.ts:564 TS2322. Widening that private helper reveals three further
assignments to DiagnosticMessageChain.next. The existing public diagnostic
contract cannot accept reusable chain entries. No compensating cast is allowed.
The proposal, checker output and counterexample are evidence, not an applied
source change. This part is blocked pending a truthful builder contract or an
approved runtime repair. The unchanged sameMap remains on the blocker list.

Validation: stage3/apply.sh and the complete stage3/oracle through stage3/lane
pass identically to main: 106,366 passing, the same one sanctioned API baseline
failure, zero pending. Both API diffs and all ten built JavaScript artifacts are
byte-identical. Local lane tests: 32 pass. No new public API lines are sanctioned.

Run `NODE_PATH=<stock API node_modules> node stage3/adapt/76-truthful-casts/check.cjs
<adapted-tree>`: idempotence, new caller-key mutant and changed-enumerator mutant
pass; both mutants are rejected. The sameMap proposal typechecks independently,
rejects a string-only result assignment, and catches the old-signature mutant
that incorrectly accepts it. Its function JavaScript is byte-identical.
Full and sliced extended parser dumps remain 36,429,231 bytes, SHA256
686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
See [unit evidence](../../drivers/parser/evidence/front32/README.md) for commands,
full checker logs and both proposal patches. The failed first contract check
used the already widened scratch draft as the supposed old-signature source;
rerunning on the production tree correctly tests and catches that mutant.
