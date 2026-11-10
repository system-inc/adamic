# Mutant evidence for namespaces/10_tracing_escape.a's row moving from NotYet to Compiles (#67jk7wq)

The row moves because tracingEnabled, a program namespace read as a value (`tracing = tracingEnabled`), now
lowers to one object made at the start of its block (internal/lower/namespace_value.go).

## Mutant: the row left NotYet, with the lowering in (expected red)
Loom's run of chg_km1wxbyz3a9qch5mtx6c80334q (emit/namespace-value-4 cefe3798, the lowering without this row,
run future of 9ac2dc52, repro 5e7f131cea6447776424bfd30500a1946e2b4335de981c412b355a32a5e1ec02):
--- FAIL: TestFixturesNamespaces/namespaces/10_tracing_escape.a/stage0
    fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:Compiles What: Stdout:};
    recorded {Outcome:NotYet What:stage3/fixtures/namespaces/10_tracing_escape.a:38:11: stage 0 can't lower a namespace object
    used as a value; ...}
It was that run's only red unit.

## Fixed (expected pass)
The row reads Compiles, and its native subtest holds the program to the recorded Node output, `1,3\nparse\n`.
