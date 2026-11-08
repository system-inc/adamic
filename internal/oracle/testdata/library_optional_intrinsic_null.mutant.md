# Null library intrinsic guard witness

At unmodified fetched origin/main, 93fa14ff0f189f13be19aec4f82974d4f663e4e8,
`function items(value: number[] | null): number[] | null { return value; }`
followed by `console.log(`${items(null)?.pop()}`)` is admitted.
Node: stdout `undefined\n`, stderr empty, exit 0.
Native with sanitizers: stdout `undefined\n`, stderr empty, exit 0.
JavaScript backend: stdout empty, stderr
`adamic: panic: TypeError: Cannot read properties of null (reading 'pop')\n`, exit 70.
The JavaScript run uses the repository's oracle/node.mjs runtime loader.

Nullable string receivers (`string | null` and `string | null | undefined`)
for trim and toUpperCase are refused by optionalCall in internal/lower/expression.go
with `a call through ?. (an optional call)`. builtin's string receiver recognition
requires ir.String; these nullable string unions use ir.Union. A constant-null
local additionally narrows to never during TypeScript checking. No string
admission is widened by this change. The IsNull arm fixes admitted nullable
library receivers, not only Stats.

Control: TestNativeAgreesWithNode's library_optional_intrinsic_null.a fixture
(native and JavaScript comparisons),
also held to Node by the native and WASI fixture families. Its stdout is
`7\nundefined\nundefined\n3\n` with empty stderr and exit 0.

Mutant: restore optionalIntrinsic's condition to Not(IsUndefined(savedReceiver)),
retaining the receiver temporary and call. The null pop at source line 4 then
fails on JavaScript at stdout line 2: only `7\n` is printed, then the null pop
TypeError exits 70. Native remains correct because this nullable array's null
is represented as missing. TestLibraryOptionalIntrinsicNullGuardMutant pins that
JavaScript disagreement by changing only the guard back to IsUndefined.

The controlled production-source mutant was also run, not only the IR mutant:
`TestNativeAgreesWithNode_054/internal/oracle/testdata/library_optional_intrinsic_null.a`
failed with Node exit 0 and all four lines versus JavaScript exit 70, stdout
`7\n`, and stderr `adamic: panic: TypeError: Cannot read properties of null (reading 'pop')\n`.
The source was restored byte for byte afterward.
