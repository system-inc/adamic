# Stop report

No new commit or push. The local branch is typescript/core-runtime-checks at
b729c7e07010fe6db34c6afc1d11c8ca63a86ae3. Changes are confined to this new
directory and stage3/upstream/LEDGER.md. The adapter is a reviewable implementation
of the truthful skipWhile result and three splices, not a completed delivery.

Delivery stops for these concrete reasons:

1. The cast-free dequeue draft type-checks but changes observable Node behavior
   for sparse initial arrays. Original dequeue returns undefined; the draft
   throws its occupancy error. It is not built, as required by the site 3 rule.
2. The required compiler-red filterMutate mutant is not established by the
   meter. Its generic owner remains deferred, with instantiated=false.
3. Current main measures the original clear length assignment as NotYet and
   implements that classification in class.go:471. The task requires Refused,
   never NotYet. Relabeling this result would be false; changing the production
   compiler would exceed the permitted commit scope.

Completed evidence:

- Read CW005, CW006, CW057 and CW074 from the requested sort branch.
- AST findReferences over adapted src finds one call, injectClassTypeMetadata
  at transformers/ts.ts:1089. It uses filter(..., isModifier) for its own narrowing.
- Both 79-root Adamic types runs exit 0. The accepted run uses a complete copy
  of the fully adapted baseline src with only the four pinned owner edits.
- Node witnesses for all splice edits and skipWhile's false exclusion type pass.
- The owner contract passes, including 29 refused fixtures and a zero-write,
  byte-and-mtime-preserving second adapter run.
- Stock TypeScript 6.0.3 parses the owners and type-checks the cast-free draft.
- skipWhile is absent from both shipped public declarations by AST inspection.
- The baseline full apply and baseline fresh compiler build pass.
- generate_shards --check passes with the prescribed tool environment.
- Both complete site census runs and the restored mutant census finish, with
  zero checker diagnostics; coverage and actual classifications are in CENSUS.md.

Unfinished or failed proof obligations:

- The first unchanged lane Go run fails at 55-second Python wrapper deadlines
  under concurrent load. Its log is retained. The retry was terminated at stop;
  no passing lane Go result is claimed.
- The edited full apply was terminated at stop. There is no full edited apply
  success or fresh before/after public declaration byte comparison claim.
- The full lane oracle and test identity comparison were not run.
- The ordinary 79-root per-file census is incomplete; only completed rows are
  retained and explicitly labeled partial.
- The restored filterMutate mutant is NOT_PROVEN_RED, not a passing proof.

README.md contains the full dequeue draft, caller audit and splice outputs.
No issue or other message has been sent upstream. No other adaptation is edited.
