# Repeated union destructuring reports an extra TS2488

Date: October 8, 2026. Upstream issue prepared for Kirk under
@system_adamic's 18:05 ruling. Not filed externally.

Built: a minimal witness, direct CLI comparison and loader comparison.
Base: Adamic f4efdd2369311d1420aa53fdf5c1a55bdda811d4; no compiler changes.
Commands: reproduction passes 15 observations; probe builds; package vet passes.
Mutants: remove undefined, remove the first call, add string-to-number assignment;
each caught by the checker observation assertion. No lowering or full gate run.

## Minimal program

[repeated-destructure.a](repeated-destructure.a), exactly the witness in ledger
3f0926c0, column 3, D003, builder.ts:1292:61:

```typescript
declare function transform<T, U>(items: Iterable<T>, callback: (item: T) => U): U[];
declare const values: Iterable<[string, number] | undefined>;
// @ts-expect-error The first destructuring diagnoses the undefined tuple.
transform(values, ([key, value]) => value);
transform(values, ([key, value]) => value);
```

Stock tsc 6.0.3: exit 0, empty stdout and stderr.
Direct typescript-go: exit 2:

```text
repeated.ts(5,20): error TS2488: Type '[string, number] | undefined' must have a '[Symbol.iterator]()' method that returns an iterator.
```

Adamic loader: exit 1, same message at line 5, column 20. Only the path and
formatting differ. Raw, unnormalized outputs are in [evidence](evidence/reproduce.log).

Versions: Node 24.19.0, Go 1.27.1, typescript-go CLI 7.1.0-dev;
cohere 7945d102a6c18dd36adf9114a758ce646e8b2359;
its TypeScript submodule d92d9bfee114c80be2c375d72edae966176e3a4f.

## Cause and disposition

Observed in pinned typescript-go's
`tsc/internal/checker/checker.go:6470`, `getIterationTypesOfIterable`:
its cache is keyed by type id and iteration-use flags. A cached failed iteration
result is ignored when an error node is present; `noCache` is set and the worker
runs again. The union worker then defers a new `reportTypeNotIterableError` at
that site's node. This behavior is explicitly documented by the upstream comment.

Stock 6.0.3's installed `lib/typescript.js:88697` takes a different path for union
types: it returns a cached `noIterationTypes` immediately, before reporting a
new diagnostic. The first failed union iteration stores that sentinel. The
second occurrence therefore produces no further TS2488. These source references
were read in place; no dependency code is copied into this branch.

The independent CLI reproduces the same difference without Adamic's shim,
loader, prelude, collection overrides or virtual .a file system. Thus the
failure-cache diagnostic policy is inside typescript-go, not Adamic's diagnostic
aggregation. Per the ruling, stop here and send this compatibility issue upstream
through Kirk. No suppression, checker patch or source adaptation is proposed.
Both sites remain semantically invalid; tsc's missing repeated diagnostic does
not prove destructuring undefined safe.

## Reproduction

```sh
bash cloud/setup.sh > /tmp/ts2488-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm install --prefix /workspace/scratch/ts2488 --ignore-scripts --no-audit --no-fund typescript@6.0.3 > /tmp/ts2488-npm.log 2>&1
(cd cohere/TypeScript/tsc && go build -o /tmp/ts2488-tsgo ./cmd/tsc) > /tmp/ts2488-tsgo-build.log 2>&1
go build -o /tmp/ts2488-load ./stage3/issues/ts2488-destructuring > /tmp/ts2488-load-build.log 2>&1
NODE_PATH=/workspace/scratch/ts2488/node_modules TS2488_TSGO=/tmp/ts2488-tsgo TS2488_LOAD=/tmp/ts2488-load node stage3/issues/ts2488-destructuring/reproduce.cjs > /tmp/ts2488-reproduce.log 2>&1
go vet ./stage3/issues/ts2488-destructuring > /tmp/ts2488-vet.log 2>&1
```

The runner uses identical explicit CLI flags for both independent compilers:
strict, unchecked indexes, exact optionals, ES2024 target/lib, ESNext/Bundler,
forced modules, verbatim module syntax, importing TS extensions and no emit.
Temporary .ts CLI views are generated from the committed .a source. CLI views
live under TMPDIR; they are never repository programs. The loader runs separately
with its normal production options. Each observation checks count, diagnostic
code and exit status, and prints complete stdout/stderr.

| Source control | tsc TS2488 | direct Go TS2488 | loader TS2488 |
|---|---:|---:|---:|
| Witness with first error suppressed | 0 | 1 | 1 |
| Directive removed | 1 | 2 | 2 |
| First call and directive removed | 1 | 1 | 1 |
| Undefined member and directive removed | 0 | 0 | 0 |

Adding `const wrong: number = "wrong"` to the valid iterable control produces
one TS2322 under all three paths. This proves that the green control is checked.
The complete runner reports PASS for all 15 observations.

Three actual witness mutants were run independently, restoring the source in a
finally block: removing undefined plus the directive removes the expected Go
error; removing the first call plus directive introduces the expected-error-free
stock TS2488; adding the bad assignment introduces stock TS2322. Each causes the
runner's diagnostic assertion to fail, not a build warning. Logs:
`/tmp/ts2488-mutant-remove-undefined.log`,
`/tmp/ts2488-mutant-remove-first.log`,
`/tmp/ts2488-mutant-extra-assignment.log`.

Setup timing: node 0.025s, Go 0.027s, submodules 0.075s, markdown 0.078s,
clang 0.247s, Go build 32.223s, cache warm 32.416s, done 32.457s.
`nproc=5`; cgroup quota 4 CPUs. [Setup output](evidence/setup.log).
No failure or workaround. Probe builds and vet exit 0 with no output;
`git diff --check` also passes. No compiler package changes, full repository
oracle, whole-tree census rerun, native execution, shim suppression or upstream
fix is covered. Column 3 remains one known upstream diagnostic difference.
