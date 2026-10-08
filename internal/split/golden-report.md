# Shapes golden: generated narrowed union helpers

Base: origin/area/platforms at e1ae0e25. No classifier or eligibility rule changes.

git log -S narrowed_union_member -- internal/lower identifies:

- 98e4747e9a944ace9f4f86850a2c212d11edfb46, Check union members before narrowed reads.
- 47d0b56e, the intervening revert.
- 9f10b92b059dba1f5c777f7d74dfd72591b98346, Restore narrowed union member checks.

The restored change is present on this base in internal/lower/expression.go.
A read of an ir.Union narrowed to one runtime representation now calls an
ordinary IR helper built by libraryArrayBuilder.finish. The helper checks typeof,
panics if the member no longer matches, and returns ir.Narrow of its argument.
This catches calls or captured writes that invalidate the checker's narrowing.
The typeof condition itself is an observation and does not generate a helper.

In shapes.a, both calls occur inside unionParameter:

| Origin | Argument | Result |
| --- | --- | --- |
| helper index 14: line 48, value in return result + value | boxed number or string union | number |
| helper index 15: line 50, value in return result + value.length | boxed number or string union | string |

The helpers have no declaration/name tokens of their own, so their Position is
zero. Their parameter and return boundary schemas are absent. They remain
ordinary callable functions in ir.Program, just as other compiler-generated
functions already reported by split. Reporting every function preserves the
analysis contract and keeps these call-graph nodes visible.

For each helper the exact decision is:

    narrowed_union_member javascript signature: missing boundary schema

Each helper is pure: all reads are of its argument, and panic is Wasm-safe.
Crossability fails first because one runtime parameter has no boundary schema.
Its ir.Union input would also be outside the crossing types, and it has no loop
or recursion. Neither helper is eligible or a crossing entry point. Thus the
golden is extended with both lines; the classifier is left unchanged.

## Validation

Required gate:

    go test ./internal/split/... ./cmd/adamic-split/... -count=1 > /tmp/workers-split-golden-tests.log 2>&1

This also executes each .a fixture against source Node, emitted JavaScript and
sanitized native. The same six refinement mutants are rerun individually and
restored; each is caught by an assertion failure, without a build failure:

| Mutant | Catch |
| --- | --- |
| Console accepted as pure | TestFixtures/decisions.a logger/handler table |
| Unknown IR expression accepted as pure | TestUnknownOperationIsImpure, named futureOperation reason |
| Panic made impure | TestFixtures/panic_decode.a |
| Coalescing panic made impure | TestFixtures/panic_decode.a |
| JSONDecode omitted from recognized operations | TestFixtures/panic_decode.a |
| JSONEncode omitted from recognized operations | TestFixtures/compute_json.a |

Logs: /tmp/workers-split-golden-mutants.log and
/tmp/workers-split-golden-mutant-NAME.log. IR inspection:
 /tmp/workers-split-golden-inspect.log.

## Compute Worker

    go run ./cmd/adamic-split workers/compute/handler.a

```text
json javascript signature: parameter 3 not crossable
error javascript signature: parameter 3 not crossable
invalid javascript signature: return not crossable
integer javascript too small to cross
skuValid wasm entry
primes wasm entry
numberAt javascript too small to cross
summarize wasm entry
closure javascript too small to cross
quote javascript signature: parameter 1 not crossable
asciiToken javascript too small to cross
topWords javascript signature: return not crossable
closure javascript too small to cross
handle javascript signature: parameter 1 not crossable
```

Both primes and summarize remain Wasm crossing entry points.

The required gate passed: internal/split 12.891s, cmd/adamic-split 0.044s.
Setup used GOPROXY=https://proxy.golang.org|direct and reported Node ready
in 0.030s, Go 0.039s, clang 0.223s, Markdown dependencies 1.141s,
submodules 3.754s, build cache warm 307.748s, total 307.780s; nproc=5.
The initial recursive history fetch was stopped; setup fetched only the pinned
checker commit. Setup log: /tmp/workers-split-golden-setup.log.

No full repository test gate, workerd run, or 600-request corpus rerun.
