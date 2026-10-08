# Debug local const enums

The two declaration shapes at debug.ts:879 and :893 now lower. Source Node,
native with sanitizers, and emitted JavaScript print:

```text
╭─╫:15:16
╭─╫:0:16
outerinner
```

The fixture keeps all BoxCharacter and Connection member initializers from
TypeScript 6.0.3; its formatter body is reduced. Direct reads must follow the
declaration in its lexical block. Ordinary local enum runtime objects and
deferred captures remain NotYet. Direct earlier reads are already rejected
by the checker as TS2450; no invalid source was admitted to test lowering.

Commands: `go test ./internal/lower -run 'TestDebugLocalConstEnums|TestEnum|TestNamespace' -count=1`;
`go test ./internal/oracle -run 'TestNativeAgreesWithNode/stage3/namespaces/debug-groups/local_const_enums' -count=1`;
`DEBUG_GROUP_COMPILER=/tmp/debug-groups-adamic python3 stage3/namespaces/debug-groups/run.py local_const_enums`.

Mutants: compiler changes glyph ─ to x, caught by clean stdout differences in
both backends; disabling the deferred-body check, caught by
TestDebugLocalConstEnums (`local enum boundary lost: <nil>`). Each temporary
compiler mutation was restored. The source semantic glyph mutant is recorded
in local_const_enums.results.json.

The original thirteen-site ledger has two fewer unsupported declaration shapes;
four namespace object observations and two unknown assertion shapes remain
preserved boundaries. A fresh latent census is still required for its exact
site count, because subsequent body failures can replace removed stops.
