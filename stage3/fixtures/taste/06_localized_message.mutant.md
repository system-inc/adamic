Mutant evidence for taste fixture 06's stage-0 record in status.json.

What moved: #3xq2xzc's C3 lowers a record on the left of `&&` to its right side where the record is present, and to undefined where it's missing (internal/lower/expression.go). A record is an object, so it's falsy only when it's missing, and a bare record pointer is never null. Fixture 06 is tsc's getLocaleSpecificMessage (utilities.ts:8604): `localizedDiagnosticMessages && localizedDiagnosticMessages[message.key] || message.message`. It moves from NotYet at 6:12, "a logical record operand converted to string without a proven result representation", to Compiles. TestFixturesTaste holds its native output to the recorded Node stdout, `fallback`, `fallback`, `fallback`, `localized`. The fixture is unchanged.

Mutant: in combine (internal/lower/expression.go), the `&&` arm that returns the Conditional is removed, so a record left with a string result stops at the NotYet again.

Command: go test ./stage3/fixtures -run 'TestFixturesTaste' -count=1

Result: Loom ran the pairing the other way round, the change with the old NotYet record, on chg_acgv7rmd (87ffd655, Nov 1). It went red on exactly TestFixturesTaste/taste/06_localized_message.a/stage0: "gap changed: update status.json and check the native output; got {Outcome:Compiles}; recorded {Outcome:NotYet ...}". The mutant is that comparison inverted: the old lowering against this Compiles record reports the same gap from the other side. The same mutant reddens TestLogicalRecordOperand in internal/native, which the line runs on this change.
