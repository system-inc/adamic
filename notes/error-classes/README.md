2 programs differ; 19 new oracle programs agree. Branch: `coverage/error-classes`, cut from `origin/codex/error-classes-counts` at `4627ced8bfb40135800f1280bd409a98c47bb783`.

The compiler is unchanged. The disagreements below are isolated reproductions, excluded from the oracle fixture list. The `unsupported/` programs are compile-time boundary probes, also excluded. All new program files end in `.a`.

The source oracle uses Node 24.19.0 through `oracle/node.mjs`, independently stripping the source's types. Successful fixtures were compared with native ASan/UBSan builds, release builds and the JavaScript backend, and checked with LeakSanitizer. The separate CLI builds below were also executed and compared, including their stderr and exit status.

## Disagreements

`cause_null.a`, in full:

```ts
try { throw new Error('null cause', { cause: null }); }
catch (error) {
    if (error instanceof Error) {
        const cause = error.cause;
        console.log(`${typeof cause} ${cause === null} ${cause === undefined}`);
    }
}
```

| Runner | Complete stdout |
| --- | --- |
| Node source | `object true false\n` |
| Native, sanitized and release | `undefined false true\n` |
| JavaScript backend | `object false false\n` |

All exit 0 with empty stderr. The branch's `internal/lower/error_classes.go:203` boxes the cause with `fit(value, ir.Union)`. Native emits null as `NULL` (`internal/native/emit_expressions.go:29`), and the union conversion has no separate null tag (`internal/native/union.go:31`); `adamic_union_typeof` therefore interprets it as undefined. Separately, `internal/lower/expression.go:648` sets `AlwaysFalse` on the null comparison because the checker's `unknown` does not explicitly include null. That explains the JavaScript backend's false null test as well. These are inferred causes supported by inspecting the emitted C and JavaScript, not compiler fixes.

`prototype_null.a`, in full:

```ts
try { console.log(Error.prototype.toString.call(null)); } catch (error) { if (error instanceof TypeError) console.log(error.toString()); }
```

| Runner | Complete stdout |
| --- | --- |
| Node source | `TypeError: Method Error.prototype.toString called on incompatible receiver null\n` |
| Native, sanitized and release | `Error\n` |
| JavaScript backend | `Error\n` |

All exit 0 with empty stderr. Suspected line: `internal/lower/error_classes.go:243`. Null has the Object representation, so the non-object TypeError branch is skipped. The object branch then finds no name/message symbols and returns the default `Error`. It needs to distinguish a null receiver from an actual object before the generic field logic.

Both isolated programs were temporarily registered in a scratch Go test file and run through `TestNativeAgreesWithNode/notes/error-classes`; both failed the stdout comparison, with both backends shown above. The scratch registration was removed before counts and the full gate. Their CLI builds also reproduced the differences.

## Cases checked against existing programs

Names in the last column beginning with `coverage_error_` are new root fixtures under `internal/oracle/testdata/`. Each row distinguishes observable cases rather than claiming every possible combination of inputs and call chains. Existing names also refer to programs under `internal/oracle/testdata/`, including the two subdirectories explicitly named.

| Branch case or condition | Existing program using it before this work | New program or remaining boundary |
| --- | --- | --- |
| Error, TypeError and RangeError construction, throw and nominal ancestry | `error_classes.a`, `error_checks.a` | `coverage_error_kinds.a` also stores and throws them through a constrained generic |
| Explicit ReferenceError, SyntaxError, URIError and EvalError, positive and negative instanceof tests | ReferenceError only as a generated TDZ failure in `error_checks.a`; no explicit four-kind suite | `coverage_error_kinds.a` |
| Built-in constant message/no cause fast path versus dynamic constructor path | `error_classes.a` | `coverage_error_argument_order.a`, `coverage_error_optional_messages.a` |
| Message omitted, explicitly undefined, empty or computed; optional string parameter | `error_classes.a`, `4ddd17f_opt_3.a` for base Error | `coverage_error_optional_messages.a` for every built-in, default and explicit subclass constructors |
| Options omitted, explicit undefined, empty literal, explicit undefined cause | Omitted in `error_classes.a`; no complete options suite | `coverage_error_optional_messages.a` |
| Error subclasses, deep default constructors, explicit super and subclass fields | `error_classes.a` | Optional super and missing messages in `coverage_error_optional_messages.a`; URIError and SyntaxError subclass causes in `coverage_error_causes.a` |
| Default name; name/message mutation; empty-name, empty-message and neither-empty toString paths | `error_classes.a` | Both empty generic fields in `coverage_error_prototype.a` |
| Generic prototype call on real Error, object, absent name/message and explicit string-or-undefined fields | `error_classes.a` | Calls through RangeError and URIError prototypes in `coverage_error_prototype.a` |
| Prototype scalar coercion for numeric/boolean name and message | `error_classes.a` has name 1/message true | false/zero, NaN/Infinity in `coverage_error_prototype.a` |
| Prototype incompatible number/string/undefined receiver | `error_classes.a` | Boolean receiver in `coverage_error_prototype.a`; null differs in `notes/error-classes/prototype_null.a` |
| String, number, boolean and nested Error cause; property assignment and shorthand syntax | `error_classes.a` | Thrown causes, true, NaN, concrete subclass and a const cause alias in `coverage_error_causes.a` |
| Object, array, function, Map and undefined causes, including heap-built contents | No | `coverage_error_causes.a`; null differs in `notes/error-classes/cause_null.a` |
| Constructor message before cause; an argument throw prevents later argument evaluation | No error-constructor argument-order program | `coverage_error_argument_order.a` |
| Stored/returned/passed Error, throw through a named function and caught raw rethrow | `error_classes.a`, `exceptions.a` | Constrained generic in `coverage_error_kinds.a`; identity through generated-error rethrow in `coverage_error_nested_rethrow.a` |
| Interface method propagates a generated error | `9984394_lib_dispatch.a` uses toFixed | Explicit EvalError/URIError overrides in `coverage_error_interface_throw.a`; pad guard in `coverage_error_padding_dispatch.a` |
| Callback invocation propagates an explicit Error, through direct/indirect calls, map/filter/find/some/forEach/reduce/Array.from/Map/Set/sort | `closures_throw.a`, `closures_throw_uncaught.a` | Nominal generated range failures in `coverage_error_callback_library.a` |
| Additional every/findIndex/findLast/findLastIndex/flatMap/toSorted callback failure paths | No complete nominal library-failure suite | `coverage_error_callback_library.a` |
| Nested try, catch without binding, finally, return/break/continue overriding pending completion | `exceptions.a`, `finally_leaves.a` | A generated RangeError overridden by finally's return in `coverage_error_catch_finally.a` |
| Catch throws and still runs its own finally | `finally_leaves.a`, `catchability-limits/d96d304_stack_catch_finally.a` | Catch's padStart failure in `coverage_error_catch_finally.a`; cause wrapper after rethrow in `coverage_error_nested_rethrow.a` |
| toFixed safe constant path, dynamic low/high rejection, truncation and NaN-to-zero | Valid cases in `number_edges.a`; 101 and prototype 101 in `error_classes.a`/`error_checks.a` | -Infinity, negative fraction, negative zero, NaN, positive fraction, 100.9, 101, Infinity in `coverage_error_number_bounds.a` |
| toFixed rejects bad digits even for NaN/infinite receiver | No caught nonfinite receiver suite | `coverage_error_number_bounds.a` |
| toPrecision/toExponential bounds, optional argument, NaN digits, nonfinite receiver bypass | `number_formats.a`, `error_classes.a` | Both infinite digit bounds, fractional boundaries, -Infinity receiver in `coverage_error_number_bounds.a` |
| Number.toString radix 2..36, NaN-to-zero, truncation; no nonfinite receiver bypass | `error_classes.a`, `radixes.a`, `radix_range.a` | 36.9, infinite arguments and invalid radix on each nonfinite receiver in `coverage_error_number_bounds.a` |
| Number.prototype format .call guards | toFixed alone in `error_checks.a`; valid calls in `library_math_number_prototype.a` | Every formatter, low/high/valid calls in `coverage_error_number_prototypes.a` |
| Repeat constant shortcuts, negative count, fraction truncation, NaN and positive Infinity | `error_classes.a`, `string_limits.a` | Negative Infinity, both infinities on empty text, huge finite empty count in `coverage_error_repeat_unicode.a` |
| Repeat length product uses UTF-16 units, rejecting oversized result before allocation | ASCII overflow in `error_classes.a`, `catchability-limits/d96d304_try_repeat.a` | Supplementary/lone-surrogate inputs in `coverage_error_repeat_unicode.a` |
| Repeat prototype .call | `error_checks.a` | Negative Infinity on Unicode in `coverage_error_repeat_unicode.a` |
| Normalize valid constant form and invalid form guard | `normalize.a`, `error_classes.a` | Dynamic valid/invalid form with constant text and prototype .call in `coverage_error_normalize_forms.a` |
| padStart/padEnd result-size guard; empty fill bypass; constant safe-size shortcut | `string_limits.a`, `catchability-limits/d96d304_try_pad.a` | Dynamic Infinity, empty fill, NaN, negative/fractional size through interface; callback and prototype .call in `coverage_error_padding_dispatch.a` |
| Protected concatenation length guard, including try-finally without catch | `catchability-limits/d96d304_try_finally_concat.a` | Already covered, no duplicate allocation of huge strings |
| Recursive stack failure, shallow success, cleanup of live strings and catch's finally | `catchability-limits/d96d304_try_stack.a`, `d96d304_stack_cleanup.a`, `d96d304_stack_catch_finally.a` | Already covered by registered fixtures and explicit runtime tests |
| Array.with out-of-range nominal RangeError; empty array and frozen unrelated object | `error_checks.a`, `library_array_with.a`, `57f2d04_with_frozen.a` | Reference elements, negative fractional/NaN/infinite indexes, source preservation and replacement evaluation in `coverage_error_array_with.a` |
| TDZ read/write, assignment RHS order, failing read skips sibling operand, callback target readiness | `error_checks.a` | Early interface call and later successful call through finally in `coverage_error_tdz_interface.a` |
| Invalidated undefined property read/write and null property read | `error_classes.a`, `error_checks.a`, `047cb0d_narrowed_in_try.a`, `9984394_defined_in_try.a` | Already covered |
| Range precision: immutable constants, field facts, arithmetic and bounded ascending/descending counters | Existing valid formats in `number_formats.a`, `number_edges.a`; lower unit tests also assert precision | Unknown parameter store, counter written in body and descending success in `coverage_error_bounds_mutation.a` |
| Presence/readiness refinement preserves reference ownership and keeps checks after invalidating calls | `reuse_narrowed.a`, `reuse_lent_global.a`, `error_checks.a` | `coverage_error_tdz_interface.a`, `coverage_error_bounds_mutation.a`; existing counts unchanged |
| Uncaught error exit 70, stdout flush, empty name/message and Unicode diagnostics | `error_classes_uncaught.a`, `error_classes_uncaught_name.a`, `error_classes_uncaught_empty.a` | URIError, optional message parameter and lone-surrogate replacement in `coverage_error_uncaught_surrogate.a` |
| Raw runtime String.fromCodePoint failure outside try, including spread | Fractional direct code point in `from_code_point_fails.a` | NaN in spread in `coverage_error_codepoint_uncaught.a` |
| Unsupported runtime calls reached directly, through callbacks or interface methods | `catchability-refused/9984394_lib_codepoint.a`, `9984394_lib_dispatch_codepoint.a` | Concrete boundaries below |

## Cases without a runnable Adamic program

Each listed probe was run from source on Node and attempted with `go run ./cmd/adamic build`. Compile-time refusal is not counted as an output disagreement. All 28 files below are in `unsupported/`, and none is in the oracle fixture list. Exact compiler diagnostics are in `unsupported-observations.json`.

| Probe files (all `.a`) | Why no accepted program was possible |
| --- | --- |
| `codepoint`, `codepoint_spread` | String.fromCodePoint's bad direct/spread arguments cannot unwind to catch. The branch deliberately refuses the try; the existing interface probe checks the same boundary. |
| `array_length`, `array_from_length` | Invalid new Array/Array.from lengths cannot unwind to catch. Array.with is catchable and covered. |
| `json_parse` | JSON.parse is explicitly Refused because the result's type cannot be proven from text. |
| `json_callback` | JSON.stringify replacer functions are NotYet; their visited values/holders need proven callback types. |
| `normalize_dynamic` | A dynamic receiver's normalization expansion cannot unwind; the form-guard helper is deliberately not marked fully guarded. Dynamic form over constant text is supported and tested. |
| `reduce_right` | reduceRight is not implemented; lowering refuses the inherited method access. |
| `error_without_new` | Reading/calling SyntaxError as a constructor function without new is not lowered. |
| `error_coercion`, `error_extra_arguments` | The checker rejects non-string message and more than two arguments before the lowerer's corresponding branches can be reached by valid source. |
| `error_stack_read`, `error_stack_write` | Native frames do not have a JavaScript source stack. |
| `cause_write`, `cause_unknown`, `cause_cycle` | Mutable causes and erased unknown causes are not proven for ownership/cycles. A concrete cause closing a Holder/Error cycle is correctly Refused by the cycle finder. |
| `options_variable`, `options_spread` | Options must be a literal exposing concrete cause type, without spread/computed/other fields. |
| `error_spread`, `error_structural`, `error_reflection` | Native shapes do not represent Error's inherited/nonenumerable property descriptors. |
| `error_override` | Synthetic built-in toString override ABI is not yet checked. |
| `prototype_array`, `prototype_function`, `error_generic_field` | Generic lookup/coercion for those receiver/field representations is not lowered. |
| `prototype_optional_field`, `prototype_optional_receiver`, `prototype_undefined_fields` | Absent optional properties, possibly undefined object receivers and erased field representations are NotYet. |

An initial interface probe combined a captured interface receiver with a callback. The conservative cycle finder refused the capture; the final interface program calls through the interface directly, while separate accepted callback programs cover propagation. An initial callback suite included reduceRight; it was separated into the boundary probe above. No compiler refusal or oracle expectation was weakened.

## Dependency proof

Changed exactly one line of the branch implementation, `internal/lower/error_classes.go:80`: the coalescing fallback `l.constant("")` became `l.constant("mutant")`. Ran the new `coverage_error_optional_messages.a` oracle program. The compiler and native binary built successfully; Node began with `Error||undefined`, native and the JavaScript backend began with `Error: mutant|mutant|undefined`. The oracle failed on stdout. The file was restored byte for byte in a Python `finally` block, and the same test passed uncached afterward. This is a semantic mutant, not a compiler warning or a build failure.

## Commands and observations

Toolchain setup:

```sh
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
```

`/workspace/adamic-tools/env.sh` is the path setup printed/created in this environment. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0. `nproc` printed `5`.

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (88s)
setup: done in 88s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Inspection: `cat CLAUDE.md`, `README.md`, `docs/0.1.md`, `docs/memory.md`; `git fetch origin main codex/error-classes-counts`, then explicit remote ref fetch; `git log --format='%h %s%n%b' origin/main..origin/codex/error-classes-counts`; `git diff origin/main...origin/codex/error-classes-counts`; reads of the changed implementation/tests/reports; `rg` over existing testdata; `git switch -c coverage/error-classes origin/codex/error-classes-counts`.

The first focused oracle run found the original null-cause disagreement and three refusals. The second passed all 14 then-finalized fixtures. A later combined subtest regexp selected zero tests; its empty result was discarded and replaced by the two correctly scoped commands below. The final focused run passed all 19 programs. The notes run intentionally failed for the two isolated programs.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/coverage_error_' -count=1 -v > /tmp/adamic-error-coverage/final-oracle.log 2>&1
# With temporary notes registration, removed afterward:
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/notes/error-classes' -count=1 -v > /tmp/adamic-error-coverage/notes-oracle.log 2>&1
# One-line mutant, then restored and rerun:
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/coverage_error_optional_messages.a$' -count=1 -v > /tmp/adamic-error-coverage/mutant.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/coverage_error_optional_messages.a$' -count=1 -v > /tmp/adamic-error-coverage/restored.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/adamic-error-coverage/counts.log 2>&1
python3 notes/error-classes/run_programs.py > /tmp/adamic-error-coverage/verified-cli.log 2>&1
gofmt -l cmd internal > /tmp/adamic-error-coverage/gofmt.log
go vet ./... > /tmp/adamic-error-coverage/vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./... > /tmp/adamic-error-coverage/gate.log 2>&1
git diff --check
```

`run_programs.py` runs the following for each of the 19 new oracle sources, two disagreements and 28 unsupported probes, writing outputs and statuses under `/tmp/adamic-error-coverage`:

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs <absolute-file.a>
go run ./cmd/adamic build <file.a> -o /tmp/adamic-error-coverage/<stem>
/tmp/adamic-error-coverage/<stem>  # every successfully built program
go run ./cmd/adamic js <file.a>  # saved to /tmp/adamic-error-coverage/<stem>.mjs
node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/adamic-error-coverage/<stem>.mjs
```

The script asserts that all oracle programs agree, both notes programs disagree, and all unsupported programs fail to build. It does not run absent binaries or absent JavaScript output. Intermediate CLI runs used the same commands through the scratch `run.py`; the type-correct final JSON callback was also built/run individually before the saved verification script ran.

Counts regeneration passed in 12.187 seconds. Exactly 19 rows were added; all prior rows are byte-identical. Formatting and vet produced no diagnostics. The full uncached gate exited 1: every package passed except `stage1/cohere/json`, where `TestDocumentedStageZeroGaps/repeatInTry.ts` still expects repeat inside try to be refused. Lowering succeeded (`gap changed: <nil>; update GAPS.md`). The compiler and stage-1 files are unchanged in this coverage branch. A targeted untouched-base reproduction is recorded below.


Supplementary emitter inspection:

```sh
go run ./cmd/adamic c notes/error-classes/cause_null.a > /tmp/adamic-error-coverage/cause_null.c
go run ./cmd/adamic c notes/error-classes/prototype_null.a > /tmp/adamic-error-coverage/prototype_null.c
```

Supplementary Cohere attempt: built the pinned submodule with `go build -o /tmp/adamic-error-coverage/cohere ./command/cohere` from `cohere/`, and inspected `--help`. Created byte-identical `.a` copies and a scoped tsconfig under `/tmp/adamic-error-coverage/cohere-scope`, then ran `/tmp/adamic-error-coverage/cohere --directory /tmp/adamic-error-coverage/cohere-scope --no-fix` with all 19 fixture basenames as explicit arguments. It exited 1 with "nothing to check: none of the named paths is in the program" and explained that `.a` is not a TypeScript or JavaScript extension. This is an unavailable lint check, not a passing result. No `.ts` program was created, and no source was modified by Cohere.


Full-gate baseline check:

```sh
git worktree add --detach /tmp/adamic-error-coverage/baseline origin/codex/error-classes-counts
rmdir /tmp/adamic-error-coverage/baseline/cohere
ln -s /workspace/adamic/cohere /tmp/adamic-error-coverage/baseline/cohere
# From /tmp/adamic-error-coverage/baseline, with the same toolchain environment:
go test ./stage1/cohere/json -run '^TestDocumentedStageZeroGaps/repeatInTry.ts$' -count=1 -v > /tmp/adamic-error-coverage/baseline-gap.log 2>&1
```

An initial attempt to address this package by absolute path from the primary checkout was rejected by Go's workspace/module boundary before running tests; the command above runs from the baseline module instead.

The untouched base at `4627ced8bfb40135800f1280bd409a98c47bb783` reproduced exactly the same `gaps_test.go:35: gap changed: <nil>; update GAPS.md` failure (0.034 seconds). This confirms the full-gate failure predates the coverage additions.
