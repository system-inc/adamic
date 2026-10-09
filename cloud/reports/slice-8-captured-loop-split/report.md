# Slice 8 captured-loop split build finding

Fixture: `internal/oracle/testdata/async_loop_body_capture.a`.
Fixture SHA-256: `5482df9d19a67866926888e01da3f7eaf060b6d7eae9e7984a82b34f5d488851`.
It captures a different `held` string on each awaiting while-loop iteration and
prints `item0 item1 item2` through three saved closures.

## Baseline results

| Compiler sources | Result from a built compiler under a 90-second kill |
| --- | --- |
| Slice 8 after the final two merges, `8a168605` | Forced split sanitizer build rejects the conflicting declaration below. |
| Before take-day, `51dda2d2` | The same fixture rejects the exact same symbol and declarations. |
| Fetched `origin/main`, `946a8f095a7fa419a92117406314b7b3d44630f0` | The exact fixture is refused at line 1: main has no async implementation, so no C is emitted. |
| Main synchronous control | Remove only `async`, `Promise<void>`, and the two awaits. Forced split sanitizer build succeeds and its output matches Node: `item0 item1 item2`. |
| Fixed slice 8 sources | Original awaiting fixture passes the forced split oracle against Node. A separately built forced split sanitizer native binary also matches Node on the original TypeScript source. |

The symbol conflict predates take-day's nine merges. It is an inherited bug in
the slice branch's stable splitter, not evidence of an async-iterables regression.
Main uses the older splitter; it does not reproduce this failure for the
supported synchronous closure control. The exact async case cannot test main's
native backend because main refuses async. No compiler changes were applied to
main or the pre-take-day baseline.

## Exact declarations

Symbol:

```text
adamic_function_async_loop_body_capture_a_run_anonymous_1_57b0c042c3c9f382836430fbbd45506c
```

Both current and `51dda2d2` emit this prototype and definition header:

```c
static adamic_code_function adamic_function_async_loop_body_capture_a_run_anonymous_1_57b0c042c3c9f382836430fbbd45506c;
static adamic_value adamic_function_async_loop_body_capture_a_run_anonymous_1_57b0c042c3c9f382836430fbbd45506c(adamic_closure *self, adamic_value *arguments) {
```

The runtime defines the alias as a function type, not an object type:

```c
typedef adamic_value adamic_code_function(adamic_closure *self, adamic_value *arguments);
```

Consequently the two source declarations are compatible C. The stable splitter
incorrectly treated the alias prototype as a state declaration and compared
these two unequal strings after renaming symbols:

```c
extern adamic_code_function adamic_unit_adamic_function_async_loop_body_capture_a_run_anonymous_1_57b0c042c3c9f382836430fbbd45506c;
adamic_value adamic_unit_adamic_function_async_loop_body_capture_a_run_anonymous_1_57b0c042c3c9f382836430fbbd45506c(adamic_closure *self, adamic_value *arguments);
```

It rejected the mismatch before invoking clang.

## Fix and proof

`units_stable.go` recognizes the emitter's two function-type aliases,
`adamic_code_function` and `adamic_counted_code_function`, and expands them to
the same signatures used by their definitions. Each closure prototype is
emitted once in the shared `units.h`; individual units include that header
without emitting another closure prototype. Each function still has one
owned definition. Declaration ABI guards remain in place and use the
canonical signature, including the counted ABI's `size_t argument_count`.

`TestSplitClosurePrototypesUseSharedHeader` checks one shared declaration,
no private redeclarations, and one definition for both ABIs, and rejects a
count-ABI mismatch. `TestStableUnitInsertion` passes. The existing
`TestUnitDeclarationDisagreement` still catches its valid-C mutant at link time.

The original fixture passes
`TestNativeAgreesWithNode/internal/oracle/testdata/async_loop_body_capture.a`
from a built oracle binary with `ADAMIC_NATIVE_SPLIT=1` and a 90-second kill.
The explicit forced split sanitizer artifact and Node both print:

```text
item0 item1 item2
```

## Additional runtime integration repair

The initial oracle invocation was blocked earlier by runtime compilation:
`library_object.c` retained two checked-reflection descriptors with no initializer
for the new `adamic_shape.kinds` member. This error is preserved in
`current-oracle.log`; the direct compiler reproduction in `current-split.log`
is the requested symbol conflict.

Commit `43eb1e93` completes field-kind metadata for the retained checked
reflection descriptors, string iterator state, and Node host descriptors.
Reflection entries distinguish number, boolean, and reference storage;
Node crypto's numeric finalized slot remains numeric. No object layout,
`adamic_object_size`, regional allocation, slot readiness/type bytes,
environment kind, or cell redirect behavior was changed.

The built native test binary passes `TestShapeKindsDistinguishScalarLayouts`
and `TestCountedShapeKindsCatchRuntimeMutant`; the latter checks that counted
runtime compilation succeeds and an inconsistent-kind mutant still aborts.

## Reproduction commands and isolation

Build each compiler outside the 90-second execution budget:

```sh
timeout -k 5 600 go build -o /tmp/adamic-gate/<revision>-adamic ./cmd/adamic
```

In the corresponding checkout, run the original fixture:

```sh
ADAMIC_NATIVE_SPLIT=1 timeout -k 5 90 /tmp/adamic-gate/<revision>-adamic build internal/oracle/testdata/async_loop_body_capture.a -o /tmp/captured-loop --sanitize
```

The pre-take-day and main checkouts are isolated detached worktrees.
Both pin the same cohere commit `7945d102a6c18dd36adf9114a758ce646e8b2359`
as slice 8. They share that dependency through a symlink; their binary builds
use `-buildvcs=false` to disable VCS stamping, without changing compiler logic.
Main lacks this fixture, so its exact-source test copies only the fixture into
the same relative location. The synchronous control is a separate input.

For oracle execution, compile `go test -c ./internal/oracle` outside the budget,
then run from `internal/oracle`:

```sh
ADAMIC_NATIVE_SPLIT=1 timeout -k 5 90 /tmp/adamic-gate/slice8-fixed-oracle.test -test.run '^TestNativeAgreesWithNode$/internal/oracle/testdata/async_loop_body_capture\.a$' -test.v -test.parallel 1
```

The direct Node comparisons use `node --experimental-strip-types` on copies
of the original fixture and the synchronous control with a `.ts` extension.
Native execution and Node execution each have a 30-second hard limit.

Each merge and both repairs passed `go build ./...`, `go vet ./...`
(600-second limits) and repository-wide `gofmt -l`. Counts remain at the
current-side version and are not regenerated by this investigation.
