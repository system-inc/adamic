# spread_accessor_reads.go mutants (#vjvcnnj head 3)

Each mutant is one edit to `spread_accessor_reads.go`, run against `TestSpreadPrototypeField`
(internal/oracle/spread_prototype_test.go), `TestSpreadAccessorNames`,
`TestSpreadAccessorReadChecksTheKind` and `TestSpreadAccessorReadKeepsMainsRead` (spread_accessor_reads_test.go).
M0, no mutant, is the control and is green on every row. The script is the job's command, mutants-h3.sh; M1 to M5 first ran at -3
(J-47mjstw), all six at -5 (J-3drk0pq), and read_both.ts, added after -5, under M0 and M1 alone (J-cg3yzd8).

| mutant | edit | red on |
|---|---|---|
| M1, the getter branch dropped | the read by name is `ir.DynamicProperty`, an own-slot read that never runs a getter | read_both.ts, whose Sized call prints `undefined undefined 0` on sanitized native where Node prints `number 20 20` (J-cg3yzd8); TestSpreadAccessorReadChecksTheKind and TestSpreadAccessorReadKeepsMainsRead, the read no longer Reflect's run-time key read. read_real.ts stays green: a real Sized alone meets no stop, so it keeps main's read on either road |
| M2, the slot kind check dropped | the helper's `If` that panics with `spreadAccessorReadMismatch` is left out | TestSpreadAccessorReadChecksTheKind alone, since no program reaches a slot of another kind today (its comment names the probes) |
| M3, acceptsUndefined's gate dropped | every read of a name in the set is read by name, observed or not | eleven rows of TestSpreadPrototypeField that want today's stop: read_parameter, read_local, read_return, read_string_argument, read_setter_plain, c13b and c13c lower instead, and read_plus, read_less, read_narrowed and written_after stop on another stop's words, a BinaryExpression with a number \| undefined and a number |
| M4, setter-only names left out | the name set takes a member only when its kind is a getter | read_setter_typeof.ts, which stops at #r8mazrh's read where Node prints undefined, and TestSpreadAccessorNames' setter-only row |
| M5, the wide gate | spreadAccessorStops answers true for every recorded read, so each is read by name | TestSpreadAccessorReadKeepsMainsRead, two run-time key reads and no ir.Property, the plain literal's weight read by name too |
| M6, the pass finds nothing | spreadAccessorReads returns before its walk, so every recorded read stays main's plain read | every honest row, each failing closed at compile time with today's words and never by lowering the plain read: read_typeof, read_undefined, read_nullish, read_template, read_real, read_flag, read_literal and read_both at spreadCopyRead's "a read of weight (or open) from { ...built }, a spread's copy", and read_setter_typeof at #r8mazrh's "which may be a spread's copy"; TestSpreadAccessorReadChecksTheKind and TestSpreadAccessorReadKeepsMainsRead stop there too |
