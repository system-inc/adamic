# Records coverage

Reviewed branch: `codex/records-lowering` at `7e2bc734ce77dfe432ce593025fdc066a5ef973c`.
Coverage branch starts at that commit. Current main is merged only into this branch.

There are 13 new admitted oracle programs. No admitted program has an output difference.
Programs in this folder are deliberately outside the oracle fixture list: eight expected
runtime stops, eight literal-key refusals, and probes of unsupported forms.
`observations.md` records Node output, native output or build diagnostic, and exit codes.

## Coverage inventory

Existing means an actual `.a` program in `internal/oracle/testdata`, not a Go unit test
or the runtime's C harness. New names below have the prefix `records_coverage_`.
Rows group related branches of the implementation, with their individual conditions listed.

| Case in branch code | Existing program | Added program or note |
| --- | --- | --- |
| Pure unrestricted mutable string signature, alias, interface, instantiated value parameter | operations, census | calls (type alias), inherited_signature (inherited index signature) |
| Contextually typed empty and nonempty record literals | operations, optional | order_edges (empty reflection and JSON), inherited_signature (return context) |
| String literal, computed string, shorthand, numeric computed property | operations (string and computed), census | key_forms (shorthand, computed number); numeric_literal_key.a records the failing uncomputed numeric spelling |
| Add, overwrite, overwrite keeps position, delete present and absent, string reinsertion | operations | order_edges adds numeric delete/reinsert and noncanonical boundaries; growth adds runtime-built keys and table compaction |
| Present and missing reads, present undefined distinguished from absence | operations, ownership, scalars | prototype_own, values, nested_arrays, coalesce_scalars |
| String, dynamic string, numeric and dot access; numeric conversion once | operations (string), census (dynamic) | key_forms (number, -0, fraction, negative, NaN, Infinity, dot); evaluation (one receiver/key/value evaluation) |
| Read and assignment used as values, discarded assignment/delete | operations | evaluation (assignment result), coalesce_scalars (coalescing result), key_forms (dot delete) |
| `in` presence including false/undefined, Object.hasOwn own hit/miss | operations, ownership | evaluation (left key before right receiver), key_forms (numeric key), prototype_own (own undefined and own string) |
| Keys, values, entries: empty, indices first, insertion order, overwrite, reinsert | operations | order_edges, key_forms, growth |
| Index classifier: empty, leading zero, nondigit, >10 bytes, UINT32_MAX boundary | operations covers 01 and both uint32 boundary keys | order_edges covers empty, -0, 1.0, 000, 1e2, long decimal, zero |
| Snapshot arrays and spread retain keys and values across mutation | ownership | growth (keys), nested_arrays (array values and entries); calls (copy returned from function) |
| Leading spread from a record, overwrite/add following fields, copy before later-field side effects | operations, ownership | evaluation, inherited_signature, nested_arrays |
| Spread fixed object into record or record into unannotated fixed result | None | spread_fixed_in.a, spread_fixed_out.a: NotYet |
| for-in declaration and assignment binding; deletion skips key, addition unvisited, overwrite current | for_in; census (parameter) | iteration_edges (let, self-delete, delete/re-add before visit, integer addition, continue, early return, empty); existing for_in has assignment and break |
| JSON own-key order, undefined omission, optional record, nested record | operations, ownership, optional | json_options (replacer key order, duplicates/missing, indentation, escapes, NUL/lone surrogate/emoji keys, nonfinite values), nested_arrays |
| Number and optional number value representations | operations, scalars | coalesce_scalars; coalesce_optional_number.a pins tagged optional refusal |
| String and optional string values | census, ownership | calls, coalesce_scalars, prototype_own, json_options |
| Boolean values; missing optional read representation | scalars | values (add/overwrite/delete/read); coalesce_scalars (false is present) |
| Object values and held reads/entries/copies | ownership, narrowed_reference | calls (class holder); existing ownership already tests runtime-built object values |
| Arrays and nested records as values | census (arrays), optional (nested optional records) | nested_arrays (array snapshots and nested boolean records) |
| Closure and Map values | None | values (call after lookup, function values enumeration, map mutation after lookup) |
| Passed and returned records, aliases, class fields | census (passed), optional (optional passed) | calls (return, argument, field, method), inherited_signature (return copy) |
| Lazy `??=`: absent, own undefined, present zero/false/empty string, fallback once | census (array absent/present and key once) | coalesce_scalars and evaluation; key_forms adds dot form |
| `||=` | None | or_assign.a: Refused, write the if |
| Literal prototype name read and in; dot/template forms statically guarded | No ordinary fixtures; Go test covers all 12 names | literal_read_*.a and literal_in_*.a for all four requested names; these are expected compile stops |
| Dynamic prototype names: missing read and in, own undefined/string hit, own-only hasOwn/spread/JSON | prototype_read (missing toString), prototype_in (missing constructor), operations (own __proto__) | dynamic_read_*.a, dynamic_in_*.a (expected runtime stops); prototype_own (all four names) |
| Runtime __proto__ assignment boundary; computed literal data definition | prototype_set, operations | prototype_own reinforces computed definition and own-only copying |
| Narrowed scalar/reference re-read after deletion | narrowed_number, narrowed_reference | Already covered; not duplicated |
| Fixed/record storage views, including nested views, arguments, returns and callbacks | No oracle program | fixed_view.a pins NotYet; lower/records_test.go already exercises nested and union views |
| Invariant mutable records and shallow spread values | No oracle program | invariance.a and spread_widening.a pin refusals; supported equal-storage aliases/copies exercised by calls and inherited_signature |
| Cycle-closing writes versus fresh writes through records | No oracle program | cycle.a pins refusal; fresh tests recursive object/record writes, held descendants, snapshot ownership and cleanup |
| Mixed/named signatures, readonly or numeric signature; unsupported slot representations | No oracle program | mixed_signature.a, readonly_signature.a, number_signature.a, optional_boolean.a, union_values.a, weak_values.a |
| Optional indexing, prototype-setting initializer, multiple spread | No oracle program | optional_index.a, prototype_initializer.a, multiple_spread.a |
| JSON object metadata and callable toJSON possibility | No oracle program | json_object_values.a and json_function_values.a: NotYet |

No supported program can exercise a branch requiring a checker-invalid argument count,
an ill-typed write, symbol-as-string key, or unsupported representation: the checker or
lowering stops first. Weak slots, readonly views, symbol/pattern/multiple signatures,
nullable records, aliased Object intrinsics and arbitrary generic record functions are
outside the admitted contract. This coverage does not claim those forms compile.

## Expected stops and unsupported forms

The dynamic read and in probes stop with native and JavaScript exit 70 and exactly
`adamic: panic: record member '<name>' is missing; records hold own keys only`.
Node's inherited result is intentional and is not counted as a difference.
Each literal read/in probe is refused and names its Object.prototype member.
Own prototype-name entries, including undefined, do agree in all three executions.

The requested `||=` and fixed-object spread conversions cannot be admitted oracle
programs on this branch. Numeric literal property names also fail lowering, while
computed numeric property names succeed. Union values and optional booleans have
unsupported slot representations; tagged optional-number `??=` is NotYet.
JSON of object-valued records needs complete metadata and function-valued records
may have a callable toJSON. The remaining note programs document the explicit
storage, signature, syntax and cycle boundaries. See the observed diagnostics.

## Mutation proof

Changed exactly one production line in `internal/native/runtime/record.c`, in
`compare_indices`, from `return a < b ? -1 : a > b ? 1 : 0;` to
`return a < b ? 1 : a > b ? -1 : 0;`.
The order_edges oracle failed with `stdout differs`: Node began
`0|1|3|4294967294|`, native began `4294967294|3|1|0|`.
Both exited 0 with empty stderr; the mutant compiled and passed sanitizers.
The original line was restored in a finally block and the entire new-program
oracle passed again. No production mutation is committed.

## Commands

Setup: `bash cloud/setup.sh`, then `source /workspace/adamic-tools/env.sh`.
Timing lines: Go ready 0s; clang ready 1s; Node ready 1s; submodules ready 1s;
build cache warm 272s; done in 272s. `nproc`: 5. cpu.max: 400000 100000.

The reproducible verifier invokes, for every new `.a`:

```sh
go run ./cmd/adamic build <file> -o <output>/<stem>
<output>/<stem>
node --disable-warning=ExperimentalWarning oracle/node.mjs <absolute-file>
go run ./cmd/adamic js <file>
node --disable-warning=ExperimentalWarning oracle/node.mjs <emitted-mjs>
```

Refused programs have no binary to run. The verifier checks admitted programs for
full stdout/stderr/exit agreement, dynamic probes for the exact paired backend
stop, and remaining notes for a compile stop. It records all observations in JSON.

```sh
python3 notes/records-lowering/verify.py /tmp/adamic-gate/records-final-all
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/records_coverage_' -count=1 -timeout 10m -v
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 10m -args -update-counts
# This command failed with the one-line mutant, then passed after restoration.
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/records_coverage_order_edges.a' -count=1 -timeout 10m -v
gofmt -l cmd internal
go vet ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...
git diff --check
```

All test output goes to files under `/tmp/adamic-gate`, then is read without piping
a live test process through head/tail. Earlier probe runs exposed console typing
mistakes, an incorrectly invoked JavaScript loader, and the numeric/union limits;
those observations were corrected or isolated before the final validation.

The final explicit sweep checked all 49 new programs: 13 admitted fixtures and
36 note probes. All matched their expected outcomes. The final record oracle
passed in 16.042s; count recording passed in 27.472s. Formatting and vet were clean.

The all-package gate was interrupted by an environment reconnect after 23 packages
had completed successfully. The remaining 21 packages were resumed with the same
uncached settings. Their exact command is recorded in `gate-command.sh`.

The resumed 30-minute gate timed out in `stage1/cohere/markdownblocks` while
`TestMarkdownTextSplitting` awaited a sanitizer build. It was retried alone:

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 60m ./stage1/cohere/markdownblocks
```

The remaining gate also timed out in `stage1/cohere/typeaware` during volume/mutation
checks. All other remaining packages passed or had no tests. The type-aware retry:

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 60m ./stage1/cohere/typeaware
```

Type-aware passed its retry in 1743.523s. Markdown completed in 3373.036s
with one failure: `TestMarkdownUnicodeWidths` could not load the scratch
`emoji-regex` dependency. All other Markdown tests had no failures. Installed
the pinned dependencies and reran just that failed test:

```sh
npm install --prefix /tmp/adamic-markdown-width --ignore-scripts --no-audit --no-fund emoji-regex@10.6.0 get-east-asian-width@1.6.0 narrow-emojis@0.0.3
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./stage1/cohere/markdownblocks -run '^TestMarkdownUnicodeWidths$'
```

The repaired Unicode-width test passed in 258.925s. Every test now has a
successful result across the original gate, completion run and targeted retries.
There was no uninterrupted all-package pass: the environment reconnect, two
30-minute timeouts and missing scratch npm dependencies are recorded above.
No compiler changes were needed for validation. Only the coverage branch is pushed.
