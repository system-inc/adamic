# Theme.Add

`theme_add.a` exports the Go Add transition as `themeAdd`, an ordered store data
model and an explicit `ThemeOperations` dependency interface. Empty error text
means success; failures preserve CSSSyntaxError.Error(), including `0: `.
Clear-all, namespace-clear and delete remain separately owned callbacks. They must
mutate the supplied state with their full Go semantics before returning.

The Go oracle invokes the pinned original Add decision body. A scratch overlay
wraps only its three dependency calls to record their actual resulting states.
The Adamic driver replays these separately owned transitions and compares Add's
own validation, precedence, map contents, order, dead count, retained prefix and
requested dependency name/arguments. It does not reimplement those dependencies.
This is isolated helper parity, not cross-worker integration or whole-rule parity.

The suite captures actual Add arguments while all six consuming rule fixture
suites execute with Tailwind 4.3.3. Every distinct captured argument is replayed
against five controlled store states. Additional controls cover clear directives,
default precedence, delete/re-add effects, overwrites, Unicode and exact errors.
The driver takes no expected Add decision or final state from its input: only
pre-state and separately owned dependency states. Go expected output is compared
as full bytes, after source Node, emitted JavaScript and sanitized native each
finish with exit 0 and empty stderr. Mutants use independent isolated witnesses.

```
source /workspace/adamic-tools/env.sh
npm install --prefix /tmp/adamic-helper-wave1-03-tailwind --ignore-scripts --no-audit --no-fund tailwindcss@4.3.3
go test ./stage1/cohere/lint/helpers/slot_wave1_03 -count=1 -v -timeout=20m > /tmp/helper03-tests.log 2>&1
```

Validation refuses an empty consumer capture. Tests replace only the consumer
fixture's external package search path in a scratch overlay; no source worktree
or shared harness is edited. Source fixtures and actual Go rule expectations are
unchanged. See REPORT.md for measured coverage and exclusions.

## FrameworkStaticReading

`framework_static_reading.a` is the second helper. Supply the Go-generated
registration table and a dependency that composes declaration conversion with
property sorting. It preserves presence, one-call-only behavior and the complete
order/count/nil-order result. STATIC_REPORT.md and static-readiness.json list all
six consumers, measured parity and explicit integration limits.

The same package test command runs both helpers. `validate-static.py --scratch
<directory>` runs the second helper alone, with actual runtime name capture, every
registered utility, additional controls and three compiling semantic mutants.
