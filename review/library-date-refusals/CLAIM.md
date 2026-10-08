# Date refusals claim

Branch: codex/library-date-refusals, based on codex/library-date at f7cbc8c.

Own Date fixtures, Date-specific lowering/runtime files and their small dispatch
hooks, plus runner prelude/adaptation fixes for Date harness restrictions.
Do not edit cmd/adamic-test262/verdict.go or classify.go: another worker owns them.
Do not change the language to accept programs tsc rejects. Compare tsc diagnostics
on the adapted source before treating a refusal as an implementation target.

First reproduce and fix the library_date_days.a missing trace.txt failure after
merging cloud/grok-date-coverage at 39fb2e7. Then measure built-ins/Date before and
after, address valid refusals largest reason first, and hold new behavior against
Node with mutants. The earlier coercion diagnosis is superseded by the decision
that TS-invalid programs are correctly refused; the separate runner outcome is
owned by another worker.
