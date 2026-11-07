# or predicate adaptation: stopped on base conflict

Task 1w5bve2. No adapter, compiler change, or call-site rewrite is shipped.

Base: origin/main d15620712c459baf72b3d359c1829df8d0808fda.
Evidence report read in full: typescript/or-calls
701181917b36db38528257ff37da1a4e64aaac16,
stage3/census/or-calls/REPORT.md.

## Blocking proof conflict

The requested baseline and original-overload mutant must show the five
predicate parameter sites returning. The base tip itself adds the withdrawn
rest-packed exemption. Its commit subject says the five parameter sites of
or stop refusing. This is implemented in
internal/lower/predicates.go:470 (predicateParameter) through
 overloadImplementationParameter, and pinned by
internal/lower/predicates_overload_rest_test.go:235,
TestOrOverloadParametersClear. Its exact expected judgment is:

```
f1 trusted, f2 trusted, f1 trusted, f2 trusted, f3 trusted
```

The adjacent comment explicitly says that dropping the helper's rest branch
refuses all five as before. TestOrOverloadNextStop at line 157 instead pins:

```
overload 1 of or type parameter P cannot satisfy implementation parameter T
```

These are source inspections, not newly executed tests or census results.
Restoring only the TypeScript overloads and original calls cannot undo this
compiler rule. Removing it would change internal/lower files outside the
requested branch scope. Stop before implementing an adaptation whose mandated
mutant evidence cannot hold on this base. The prerequisite is a base on which
the withdrawn compiler exemption has been removed, or a revised proof contract.
No compiler rule or proof expectation was changed here.

## Preliminary inspections

Adaptation 49-honest-optionals pins lastResult! and edits general-rest
signatures, rather than pinning the whole or declaration group or the two
predicate overload declarations. Its inspected pins do not by themselves
require an amendment for deleting those two declarations.

The stock 6.0.3 parser was used on pinned upstream source fetched read-only to
identify enclosing owners for the nineteen reported predicate calls. Their
names and lexical ancestors were searched across every adapt.cjs and owners
or sites JSON. This is preliminary owner discovery, not the required stock
checker census on a completed adapted base. No claim of re-establishing the
nineteen resolved signatures is made. The whole checker import is pinned by
99-correctness-core-multimap/owners.json; any implementation must account for
that pin when introducing the new helper imports.

The original report records 2,186,858 closure calls across all 23 sites and
zero value differences. The nineteen predicate sites total 2,018,000 calls;
the four rest sites total 168,858. Only the predicate subtotal has the stated
boolean-only observation. In particular utilities.ts:9004 records non-boolean
values. The requested ledger attribution should preserve that distinction.
This report does not rerun or claim a new lane measurement.

## Work and proof status

- Fetch of main and both evidence branches: exit 0.
- Requested branch created directly at origin/main: exit 0.
- API npm ci: exit 0.
- Required cloud setup and pinned source clone: stopped with SIGTERM after
  the blocking base inspection, before completion. Their process groups were
  terminated. Logs remain in /workspace/scratch/or-predicate-*.log.
- Both full applies, paired replay, stock checker site census, Node witnesses,
  79-root Adamic types, stock adapted build, emitted JavaScript comparison,
  public/internal declaration comparison, ts-key enumeration audit, full core
  site census and mutant: not run. No passing results are claimed.
- No ledger ruling line added because the adaptation was not implemented.

One tracked file is added, this stop report. The branch is an evidence-only
checkpoint and must not be treated as a completed adaptation.
