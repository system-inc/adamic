Removed Int32-backed Uint16 lowering and both standalone runtime helpers at runtime review's request.
Kept class-constructor value lowering and both previously requested merges; final commit SHA is reported with the push.
The 70000 write fixture gives 4464 on source Node and pins typed array element type Uint16Array as Adamic NotYet.
The acceptance mutant maps Uint16 through Int32 and bypasses the constructor guard; the fixture fails because lowering succeeds instead of producing NotYet.
Original confirmed coverage is now 0/22 roots until area/runtime's real Uint16Array integration lands.

Removed internal/native/runtime/uint16_array.c and uint16_array.h, their include, the Uint16 IR kind, JS mapping, native Int32 mapping and conversion hooks. Restored the original typed-array unsupported guards and constructor lists. The runtime owner reports incorrect output from the temporary storage path; this withdrawal does not attempt to repair or compete with the runtime-owned adamic_typed_array_uint16 implementation.

new_expression_uint16_notyet.a constructs one element, writes 70000, and prints its value as a string. TestNewExpressionUint16RemainsNotYet independently requires source Node exit 0, stdout 4464 and empty stderr, then checks the exact typed-array element NotYet. Both the old Uint16 fixture and the new refusal fixture are registered as not lowering; the class-value fixture remains registered as lowering. update-counts removes the old runnable Uint16 row and does not record NotYet fixtures; the class-value row is unchanged.

Validation commands (output in evidence/drop-uint16-*):
- go test ./internal/oracle -run 'TestNewExpressionUint16RemainsNotYet|TestNativeAgreesWithNode/internal/oracle/testdata/new_expression_' -count=1 -timeout 10m
- python3 stage3/notyet-new-expression/run-uint16-notyet-mutant.py
- go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts
- go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -count=1 -timeout 30m

The focused run passed in 1.047s. The mutant was killed by "want Uint16Array element-type NotYet, got <nil>", not a compiler warning or load error. It uses a scratch Go overlay and never modifies production compiler files. Its runner replaces the obsolete conversion-mutant runner. The initial fixture incorrectly passed a number directly to console.log; changing the output to a string fixed its checker error before these results were recorded.

Package results: ir 27.730s, lower 49.019s, native 383.075s; JavaScript compiled with no test files. The initial parallel oracle process had already loaded the invalid console.log fixture, so it failed on that old input; the corrected fresh oracle run passed in 27.819s. Counts refresh passed in 64.735s. Both the failed initial output and successful final output are retained. No full repository gate ran.
