# sameMap: four sites and the public boundary

Base: origin/main 89ac4a8c1de0b02d95965be72f7f8bf1c92433d2. The isolated
candidate checkout starts exactly there and adds only adaptation 76. The proof
branch merges main normally; no pushed history is rebased or rewritten.

The honest sameMap overloads return T|U, respecting each overload's mutability
and undefined case. Both array casts disappear; slice's local result is (T|U)[].
This proposal is still NOT activated. Its JavaScript erases byte-identically,
but the following coupled source contracts cannot all pass the checker under
the permitted edits. No public declaration, runtime statement or hiding cast is
changed by the active adapter.

| Site | Truthful local edit | Why it is not activated |
|---|---|---|
| builder.ts:564:5 | Private array helper returns (DiagnosticMessageChain\|T)[]\|undefined. | This fixes its own TS2322 but exposes the three next assignments below. Activating it alone breaks the compiler build. |
| builder.ts:545:13 | Mode-mismatch result uses a private recursive chain whose next permits original T and converted chains. | Public DiagnosticMessageChain.next cannot hold reusable chains. Widening the private converter clears this assignment but exposes public callers. |
| builder.ts:551:13 | The same private recursive result for the module-not-found branch. | Same public next boundary; there is no proven narrowing of arbitrary reusable T into DiagnosticMessageChain. |
| builder.ts:555:106 | The reconstructed result uses the private recursive next type, and returns chain without the two existing DiagnosticMessageChain assertions. | The merged object then remains a private chain, not a public diagnostic. Claiming a public type would hide the union again. |

Trials, all with stock TypeScript 6.0.3 against the current adapted main tree:

1. Honest sameMap: exactly builder.ts:564 TS2322.
2. Widen the private array helper: exactly the three next-field errors.
3. Widen the converter to a private recursive chain and remove its chain casts:
   those three clear, but repopulateDiagnostics at 526 and the public
   DiagnosticRelatedInformation.messageText at original 608 reject the result.
4. Add private overloads separating already diagnostic input from reusable
   input: the 526 error clears, but the public messageText assignment remains
   (observed line 622 after the private alias/overloads are inserted).
5. Widen that private related-information caller as well: the errors move to
   convertToDiagnostics' public Diagnostic result and relatedInformation field
   (observed 589 and 595; original 575 and 581).

The fourth trial is a checker experiment, not a landed proof of those overloads.
At the remaining boundary, an honest reusable-input result still includes chains
without messageText, category and code. Changing the private caller's return to
a wider type propagates that mismatch into the public Diagnostic contract;
changing that public contract is explicitly outside this unit. A narrower
result needs a proved ownership/purity invariant or a runtime conversion/check.
No gate-passing truthful type-only repair was found within the allowed scope.
This is the stopping point, not a claim that every possible source proof has
been exhausted.

same-map-chain-boundary.a is a cut-down type model. sameMap's implementation and
the private array helper's return statement are copied from tsc; the converter,
head creation and three result functions are explicit model stubs. A typed
getter mutates an earlier array element to a reusable entry while mapping.
Node prints 564:false, 545:false, 551:false, 555:false: each resulting first child
lacks messageText. This demonstrates the general signature's alias problem;
it does NOT establish that TypeScript's own builder callers create getters.
The original numeric-alias witness in same-map-proposal.a remains as well.

check-chain-boundary.cjs typechecks the honest fixture. At each of the four
sites, assigning its result to the unchanged public type rejects with TS2322.
Its per-site mutant reinstates a narrow result signature and a hiding assertion
ONLY in the virtual mutant. That lying program admits the public assignment;
the negative contract check catches it. All four mutants erase to exactly the
same JavaScript, showing why output identity alone cannot prove these types.
No mutant assertion is written into adapted compiler sources.

Evidence is in evidence/same-map/: raw checker logs, type-only draft patches,
erasure hashes, Node fixture output, per-site mutant report and both full lanes.
The gate results cover the active finite-key adaptation with sameMap unchanged;
they do NOT certify the rejected sameMap draft. Public API bytes and emitted
JavaScript are compared against the current-main build, without extra sanctions.

Final active gate: both lanes PASS, 106366 passing, one identical sanctioned
API failure, zero pending. Public typescript.d.ts and tsserverlibrary.d.ts bytes
match main, as do all ten JavaScript artifacts and the API baseline diff.
32 local lane tests pass. The four new contract mutants and the existing 76
mutants pass. These results do not activate or certify the rejected union draft.
