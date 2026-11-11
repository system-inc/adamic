# Mutants for a record spread from a source that may be missing (#bzrdem0)

Each is one edit to records.go, run on a house box, then put back.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| The missing arm dropped | `result.Spread = ir.Coalesce{Value: result.Spread, Fallback: ir.RecordLiteral{Element: of}, Of: ir.Record}` | `result.Spread = ir.Coalesce{Value: result.Spread, Fallback: ir.RecordLiteral{Element: of}, Of: ir.Record}.Value` | TestRecordSpreadFromMissingAgreesWithNode, undefined_args.ts and undefined_args.a natively and sanitized, where the spread walks undefined as a record; null_args and tracing_write_stack_event the same. JavaScript stays green, since its own spread of undefined copies nothing |
