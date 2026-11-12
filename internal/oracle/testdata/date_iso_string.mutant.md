# Date ISO string witnesses

`date_iso_string.a` is registered as lowering in the ordinary oracle corpus. Its
fixed rows are compared with Node by native, sanitized native, JavaScript and
WASI and WASI-shim readers. The generated counts use the repository's `-update-counts` mode.
The clock-dependent tsc expression has a separate lowering-only pin;
`new Date(0)` with the same regex replacement is in the byte oracle.

`TestDateISOStringRuntimeMutantsAreCaught` embeds a modified copy of the shipping
`runtime/date.c` under a private function name, leaving the shipping archive
linked. It compares sanitized execution with Node and requires exit 0 and empty
stderr before judging stdout, so a sanitizer error or panic cannot count as a catch.

* `expanded_year`: replace the four-digit year guard by `if (1)`. Both stdout
  lines 6 (`+010000`) and 7 (`+275760`) must differ from Node.
* `floored_day`: remove `if (within < 0) { days--; within += 86400000; }`.
  Stdout lines 3 (before epoch), 4 (the TimeClip input `-1.5`), and 11
  (year -1, one millisecond after midnight) must differ from Node. The original
  year-minus-one midnight row is retained at line 10; an exact midnight cannot
  witness a negative remainder, hence the additional fractional row.
* `milliseconds`: replace `within % 1000` by zero. Stdout lines 12 (`.123`) and
  13 (`.999`) must differ from Node.
* `invalid_date`: omit `adamic_throw_kind` and return an owned `invalid` string.
  Stdout line 15 must differ from Node's `RangeError: Invalid time value`.

The catch reports are at `date_iso_string_test.go:68`; removing a mutant must
instead reach the fatal comparison at line 61 or the required-row check at 66.

The non-integral `new Date(-1.5)` row must print
`fractional-negative=1969-12-31T23:59:59.999Z` at stdout line 4. Optional Date
receivers are exercised by both `toISOString` and `getTime`: lines 16-17 print
`undefined`, while lines 18-19 print the fixed ISO text and `1`. The shared
`optionalIntrinsic` receiver binding and nullish guard carry these library calls;
the Date host arm does not add another optional-chain implementation.

`TestDateISOStringMethodMutantIsCaught` uses `buildcache.GoTestMutated` to remove
the Date-arm allowlist. Its child must fail
`TestDateISOStringOtherMethodStopsByName` at `date_iso_string_test.go:19`, with
`Date method boundary must retain named NotYet`. The child has no oracle/toolchain
control. The exact `toUTCString` NotYet is unchanged from main.

Commands (from the source checkout, with its toolchain environment):

```
go test ./internal/oracle -run '^TestDateISOStringRuntimeMutantsAreCaught$' -count=1 -v
go test ./internal/lower -run '^TestDateISOStringMethodMutantIsCaught$' -count=1 -v
```

Runtime reader audit: `date.c` owns no shapes or static storage and adds no IR or
lowering field. `fields.go`/`TestRuntimeFieldLayoutsAreIncluded`,
`TestRuntimeStaticsAreListed` (storage.txt), and `TestShapeProducerInventory`
require no new shape or mutable-storage row. The storage inventory does require
the explicit `runtime-file:date.c` audit row, even for this storage-free file. GCC's explicit translation-unit
inventory has a Date producer and consumer. The regexp witness archive is the
transitive symbol closure of its existing roots; no root calls Date, so date.c
is not added to that partial archive. Probe symbol renames are derived from the
C declarations; the formatter has one global symbol and no private static names.
Run the whole native package, including the regexp/probe readers and GCC audit,
and the strict WASI products. No new skip site is introduced.
