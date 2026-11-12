# Mutants for a record spread from a source that may be missing (#bzrdem0)

Each is one edit to records.go, run on a house box, then put back.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| The missing arm dropped | `result.Spread = ir.Coalesce{Value: result.Spread, Fallback: ir.RecordLiteral{Element: of}, Of: ir.Record}` | `result.Spread = ir.Coalesce{Value: result.Spread, Fallback: ir.RecordLiteral{Element: of}, Of: ir.Record}.Value` | TestRecordSpreadFromMissingAgreesWithNode, undefined_args.ts, undefined_args.a, null_args.ts and null_args.a natively and sanitized, where the spread walks a missing record as a record. JavaScript stays green, since its own spread of undefined copies nothing. tracing_write_stack_event.ts stays green: the spread still lowers, and the pinned stop is at its later JSON.stringify |
| The source judged as its own type, as main judges it | `present = l.checker.GetNonNullableType(own)` | `present = own` | TestRecordSpreadFromMissingAgreesWithNode, tracing_write_stack_event.ts, which stops at the spread with main's text instead of reaching the pinned JSON.stringify stop; undefined_args and null_args stop the same way |
