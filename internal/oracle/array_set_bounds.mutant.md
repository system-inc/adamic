# An array element write past the end still panics once the write is inline: a mutant (#2ccxtyr)

The mutant is one edit, applied alone, then reverted. Run
`go test ./internal/oracle -run '^TestStage3EnumSparseArrayBoundary$' -count=1 -v`.
Not run yet: written on a machine that doesn't build Go or C, so "goes red" below is what the code says, not what a
run printed.

The test's fixture, stage3/fixtures/enums/07_regex_array_map.a, writes `result[value] = name` into an empty array of
a numeric enum, so its stop is adamic_array_set's bounds check on the scalar path, the path fannkuchredux's writes
take, and the test holds the sanitized native run to exactly `adamic: panic: index 1 is outside an array of length 0`
and exit 70. The runtime's key reads adamic.h's bytes, so the edited header is a fresh build, not a cached result.

## M1. The inline write drops its bounds check

File: internal/native_runtime/c/adamic.h, `adamic_array_set`.

Old:

```c
	adamic_value *slot = adamic_array_at(array, index);
	if (slot == NULL) {
		adamic_array_set_outside(array, index);
	}
	if (array->references) {
```

New:

```c
	adamic_value *slot = adamic_array_at(array, index);
	if (array->references) {
```

Expected: TestStage3EnumSparseArrayBoundary goes red on its native leg with `native: exit codes differ`. adamic_array_at
answers NULL for index 1 of an empty array, the store goes through that NULL slot, and the sanitized run stops on an
AddressSanitizer report, exit 1, where the test wants the panic's exit 70. Whatever the optimizer makes of the store
through NULL, no run of the mutant prints the panic's message, so the leg can't stay green. The JavaScript leg stays
green: its own bounds check is untouched.
