Mutant evidence for taste fixture 13's stage-0 record in status.json.

What moved: H2 (ruled Oct 18, #x2gt39p) refuses a read of a value typed exactly void, in .a and .ts. Fixture 13 (tsc's moduleNameResolver.ts:610) binds `reportDiagnostic(...)`, a call through `(diag: Diagnostic) => void`, and prints `String(result)`. A function that returns a value can stand behind that void type, and Adamic releases what a void call returns (#f418fbz), so the read would print undefined where Node prints the value. Its first stop moves from Compiles to the refusal at the read, 7:20: "Adamic 0.1 refuses a read of a void value, as an argument". The fixture is unchanged. Being refused, it leaves the oracle's taste list and its runtime count record.

Mutant: in voidReadRole (internal/lower/void_reads.go), the call-argument case returns "" instead of "an argument", so a void value passed to a function is admitted again.

Command: go test ./stage3/fixtures -run 'TestFixturesTaste' -count=1
Result with the mutant, on Pupa at 8e61a6e6 (Oct 27): exit 1, red on exactly TestFixturesTaste/taste/13_void_callback.a/stage0 ("gap changed: update status.json and check the native output"). The same mutant reddens TestVoidReadsAreRefused/passed.a and passed.ts in internal/lower. Restored, both pass.
