# Non-null check coverage

Base: origin/codex/non-null-check, 289d6870cdd624085cfd7aacc493a02aae6d2cb8.
Comparison: origin/main at 50045bd. The merge-base diff contains one commit,
Check non-null assertions before unwrapping nullable values.

The review read CLAUDE.md, the branch diff and commit message, its lowering,
refusal, test and oracle changes, and the existing non_null programs. The
existing testdata search found non-null expressions only in the nine non_null
programs introduced by the feature branch. The oracle uses an explicit fixture
list, not directory discovery. New passing fixtures have lowers=true and
checked=false. No notes program remains in that list.

## Conditions and representations

| Lowering case | Existing program before this work | Added evidence |
|---|---|---|
| Expression dispatch calls nonNull, operand evaluated once | non_null.a: next()! | calls, arrays, maps, chains, references, fields |
| Strip ir.Unwrap from a narrowed maybe-number or maybe-boolean | non_null.a: number; non_null_field.a | narrowed, fields |
| Strip ir.Defined from a narrowed reference | non_null_capture.a: captured string | narrowed: string; tuple_null: narrowed null match |
| Strip ir.Narrow from a narrowed mixed union | non_null.a: typeof number | union_narrow, narrowed: string and boolean |
| Restore declared representation of a property with a known symbol and type | non_null_field.a: number after clear() | fields: optional number/string/object, parentheses, side-effectful receiver; missing_field probes |
| Proven non-nullish, not narrowed from nullable storage, not a maybe pair: erase | non_null.a: plain string, 42, false | union_narrow: number/string union with no undefined |
| Erased union representation needs a narrower result | No program found | union_narrow: present(number|string) |
| Nullable or maybe operand: Coalesce with terminal panic | All existing failing non_null programs | all notes missing programs |
| Checked union result needs ir.Narrow | non_null.a: typeof number over number|string|undefined | union_narrow: both primitive members |
| MaybeNumber accepts present zero and NaN | non_null.a | calls, arrays, maps, fields, chains, narrowed |
| MaybeBoolean accepts present false | non_null.a | calls, arrays, maps, chains, narrowed |
| String reference accepts empty string | non_null.a | calls, arrays, maps, fields, chains, narrowed |
| Object reference | non_null.a: reference parameter | references: fresh result, fields, arrays, maps |
| Array reference | non_null.a: nullable match result | references: ordinary array; tuple_null |
| Map reference | No program found | references: map()!.get(...)! |
| Closure reference | No program found | references: closure()!() |
| Mixed union accepts zero, false, empty string, NaN | non_null.a | narrowed, union_narrow |
| Missing number, boolean, string, mixed union, null match | non_null_map/boolean/array/union/null.a | notes: array, map, result, narrowed, null |
| Missing object/array/map/closure references | No program found | notes: missing_*_reference, inside and outside try |
| Original expression text captured in panic | Existing failures and lowering unit test | notes: indexed call, keyed map call, optional chain, call, narrowed read |
| Load/type errors before constructing a check | No positive oracle program can exercise a diagnostic and also execute | blocked probes, described below |

## Operand and invocation forms

| Form | Existing program before this work | New program or notes |
|---|---|---|
| Field assertion, value present | No program found | coverage_non_null_fields.a |
| Array element, value present | non_null.a: match result's first string | coverage_non_null_arrays.a: zero, empty string, false, object |
| Map lookup, value present | No program found | coverage_non_null_maps.a: zero, empty string, false, object |
| Function result | non_null.a: number result | coverage_non_null_calls.a and references.a |
| Class method result | No program found | calls.a: source.read()! |
| Closure result | No program found | calls.a: result()! |
| Optional chain field/method result | No program found | chains.a, notes missing_chain |
| Narrowed variable | non_null.a and non_null_capture.a | narrowed.a, union_narrow.a, notes missing_narrowed |
| Narrowed property, parentheses | non_null_field.a without parentheses | fields.a |
| Tuple element | No program found | tuple_null.a |
| Present inside try/finally | No program found | references.a; finally prints calls=4 |
| Missing inside try/finally | non_null_catch.a: map only | Each of eleven notes failure forms has try and outside variants |
| Side-effectful array index and map key | No program found | arrays.a, maps.a, notes missing_array/map |
| Side-effectful optional-chain receiver | No program found | chains.a, notes missing_chain |
| Side-effectful field receiver | No program found | fields.a: box().value! |
| Side-effectful result returning a missing value | No program found | notes missing_result/null/*_reference; counter prints once |

Fixture names in abbreviated cells above are under internal/oracle/testdata;
notes failure forms are under notes/non-null-check. Each missing form is a
separate terminating program so one panic cannot hide later assertions.

## Limits

Six runnable source probes cannot produce native binaries on this branch:

- blocked_generic.a: T | undefined is instantiated by the generic caller, but
  the assertion's NonNullable<T> result is not lowered.
- blocked_brackets.a: string-keyed element access on a fixed object is NotYet.
- blocked_boolean_field.a: an optional boolean field is NotYet (slotless pair).
- blocked_scalar_null.a: number | null has no supported representation.
- blocked_null_undefined.a: RegExpMatchArray | null | undefined has no supported
  representation.
- blocked_optional_call.a: value?.() remains an unsupported optional call.

These are compile-time limitations, not executable output differences. Source
Node runs them; the observer retains Node output and exact build diagnostics.
The internal branches for absent symbols or unknown declared representations
cannot be forced independently by an accepted, well-typed property program.
They are fallback/diagnostic paths, not additional runtime features. This work
makes no claim of line or branch instrumentation coverage or of every possible
union combination.

## Mutation proof

One line, internal/lower/non_null.go:52, was changed temporarily. The Coalesce
operand was wrapped in an ir.Conditional with a false BooleanConstant, the
original value as WhenTrue, and fit(ir.Undefined{}, value.Type()) as WhenNot.
It compiled successfully and made coverage_non_null_maps.a fail by output/exit
comparison: source Node printed 0 1, empty string 2, false 3, map 4 and exited 0;
native and generated JavaScript printed nothing, panicked at numbers.get(key())!,
and exited 70. No clang warning or compile error killed this mutant. The
original line was restored, and no compiler change is included in the commit.

## Toolchain

The completed bash cloud/setup.sh run printed:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (1s)
setup: build cache warm (166s)
setup: done in 166s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

nproc printed 5. The environment file is /workspace/adamic-tools/env.sh.
Go 1.27.1, clang 20.1.8, Node v24.19.0. The first setup attempt overlapped the
branch switch and failed its warm build; setup was rerun on the final branch.
An initial scratch JavaScript invocation omitted oracle/node.mjs and could not
resolve the adamic runtime; every reported observation was rerun with that
loader. Neither exploratory failure is counted as a program disagreement.
