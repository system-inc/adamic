# Backend witness handoff

The implementation is stock-TypeScript-valid, but the required successful
JavaScript/native runtime agreement is not available on the frozen main compiler.
No compiler source was changed to admit it.

The exact original four overloads and implementation reproduce the expected
binder refusal for both arities. Restoring those declarations and original calls
in the mutants restores the same refusal on both backends:

```
overload 1 of or type parameter P cannot satisfy implementation parameter T
```

Keeping the complete adapted declaration group fails first on the retained rest
set, even when the program calls orPredicate2 or orPredicate3:

```
stage 0 can't lower overload 1 of or with a non-array rest parameter yet
```

The companion witness removes only that independently blocked or set and keeps
the exact helper declarations. It still fails for each arity, on each backend:

```
stage 0 can't lower a checked predicate claim of R1, over R1, which this body is lowered without knowing yet
```

For the two-predicate companion, the error is main.ts:16:17, the f1(p) call in
its returned closure. For three predicates it is main.ts:20:17. Adding explicit
<N, A, B> type arguments to the two-predicate call gives the same result. A
program containing the helpers without calling them lowers successfully.
These controls distinguish the instantiated generic checked-claim stop from
the full group's independent rest-overload stop.

The reported checked-claim wording is emitted by
internal/lower/predicates_branded.go in brandedCall (line 182 on this base). This report
does not propose relaxing that check or changing the binder invariant.

PROOF-EVIDENCE.tar.gz contains witness-before, witness-after, witness-helpers,
witness-mutant and witness-rest-* directories. Each has its source, Node output,
compiler stdout/stderr and exact exit codes in witness-backends.json. The core
text provenance records verify the original/adapted copied declaration groups.
The relevant commands are:

```sh
adamic js main.ts
adamic build main.ts -o native --sanitize
node node.mjs
```

All Node programs produce their expected results. Both original and mutant
negative controls pass their refusal checks. The adapted positive acceptance
condition fails: neither backend runs the instantiated helper program.
