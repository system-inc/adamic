# Mutants for the bind of an absent method (#1k4z29h)

- **The emitter's check dropped.** In internal/native/emit_expressions.go's boundMethod, remove the
  `e.checkDefinedWhen(...)` line before `return makeValue()`. TestBindOfAbsentMethodStops goes red on
  spread_method.ts, spread_field.ts and fixed_argument.ts: each now ends on closure.c's backstop,
  "adamic: panic: runtime invariant broken: a bound method has no code to call", not on "method weigh is absent". It ends with no
  AddressSanitizer report, which is the backstop holding. fixed_argument.ts also prints "note" first, since the
  fixed argument now runs before the stop.
- **Both checks dropped,** the emitter's and closure.c's three compares. The same rows end on an
  AddressSanitizer SEGV, main's behavior before #1k4z29h.
