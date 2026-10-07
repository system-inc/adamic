Built: classified all eight Array stops and reduced them to a shared compiler witness; compiler unchanged.
Commits: base f05aec3cee91d2ce3bfced6d341865933ec28da1; this report and evidence are the only branch changes.
Commands and outputs: two full Array runs both 126 pass, 0 fail, 2338 refused, 8 crashed, 610 skipped; eight original Node runs pass.
Mutants: none; no new implementation or check was introduced, so no mutant catcher is claimed.
Not covered: compiler repair, new passing tests, full repository gate, or runtime/WASI changes.

## Finding

Observed: all eight programs generate C successfully but clang rejects that C before a binary exists.
They do not time out or stop in the Array runtime. In every case the first invalid line is the
array literal's `adamic_retain((0x0p+00))`: a double passed to `void *`.
The four lastIndexOf cases also emit `adamic_retain(true)`.

The common smallest tested example is:

```typescript
const object = {};
const array = [0, object];
```

`adamic c` exits 0. `adamic build` exits 1 with `passing 'double' to parameter of incompatible type 'void *'`.
Node exits 0 without output. The observable version `const object = {}; console.log(String([0, object].length));`
prints `2` on Node and has the same clang failure. Replacing 0 with true isolates the boolean variant:
clang rejects its integer-to-pointer conversion; Node still prints `2`.
A literal written directly as `[0, {}]` is a different control: it is refused as `an array of number | {} yet`.
Keeping the empty object in a binding is essential to reproduce this stop.

Observed type declarations from `adamic types`:

```text
variable target: {}
variable numbers: {}[]
variable booleans: {}[]
```

The general compiler path explains the output:

- `internal/lower/expression.go:30`, `representation`, maps the checker's `{}` to `ir.Object`.
  TypeScript's `{}` admits non-nullish primitives; it does not prove a heap object.
- `internal/lower/object.go:179`, `arrayLiteral`, obtains `ir.Object` for these elements and calls
  `fit` on each primitive. `internal/lower/expression.go:615`, `fit`, leaves number-to-object and
  boolean-to-object unchanged. The resulting literal has a reference element type with primitive values.
- `internal/native/emit_expressions.go:546` chooses reference slots from the literal's element type
  and wraps each element in `adamic_retain`, including the unchanged primitive.

Classification: **compiler, specifically general representation/array-literal lowering**.
This is outside `internal/lower/library_array*.go` and the Array runtime. Removing the search call
entirely preserves the failure, so neither the fromIndex conversion nor search runtime causes it.
Per the unit instruction, no compiler code was changed. No library workaround was added.

## Each of the eight

Paths are under `built-ins/Array/prototype/`. Every row is the same compiler failure above and is
reproduced by the common two-declaration example. The reduced search-specific witnesses omit the
assertion harness and the second assertion. Define `const target = {};`, then evaluate the expression
shown through `console.log(String(expression));`. All eight generate C (exit 0), fail native build
(exit 1), and run on Node (exit 0) with the output below.

| Program | Reduced expression | Node output | Cause |
| --- | --- | ---: | --- |
| indexOf/15.4.4.14-5-10.js | `[0, target, 2].indexOf(target, 2)` | -1 | compiler |
| indexOf/15.4.4.14-5-11.js | `[0, target, 2].indexOf(target, -1)` | -1 | compiler |
| indexOf/15.4.4.14-5-31.js | `[0, target, 2].indexOf(target, 2.5)` | -1 | compiler |
| indexOf/15.4.4.14-5-32.js | `[0, target, 2].indexOf(target, -1.5)` | -1 | compiler |
| lastIndexOf/15.4.4.15-5-10.js | `[0, target, true].lastIndexOf(target, 1.5)` | 1 | compiler |
| lastIndexOf/15.4.4.15-5-11.js | `[0, target, true].lastIndexOf(target, -2.5)` | 1 | compiler |
| lastIndexOf/15.4.4.15-5-31.js | `[0, target, true].lastIndexOf(target, 1.5)` | 1 | compiler |
| lastIndexOf/15.4.4.15-5-32.js | `[0, target, true].lastIndexOf(target, -2.5)` | 1 | compiler |

All eight unmodified test262 programs independently pass on Node using the official `sta.js` and
`assert.js` harness files. This verifies both assertions, not only the reduced witnesses.

## Before and after

Test262 is c8c798898646638cd0c24879f8e0374e847e7d74. Both runs used the compiler at f05aec3,
ordinary result caches, adaptation on, four jobs, and Node v24.19.0. Each command completed with exit 0:

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
go run ./cmd/adamic-test262 -adapt -jobs 4 -json -test262 /tmp/adamic-test262-corpus built-ins/Array > /tmp/adamic-array-eight-before.json 2> /tmp/adamic-array-eight-before.log
go run ./cmd/adamic-test262 -adapt -jobs 4 -json -test262 /tmp/adamic-test262-corpus built-ins/Array > /tmp/adamic-array-eight-after.json 2> /tmp/adamic-array-eight-after.log
```

| Run | Pass | Fail / disagreements | Refused | Crashed | Skipped | Total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Before | 126 | 0 | 2338 | 8 | 610 | 3082 |
| After investigation | 126 | 0 | 2338 | 8 | 610 | 3082 |

The complete filter JSON objects are identical, including the passing paths and all eight crash
reasons. Newly passing tests: none. The eight failures are not claimed as Node agreements.
Full per-directory tables and JSON are preserved in `evidence/test262-{before,after}.{log,json}.gz`.

## Setup and probe commands

`nproc` printed 5, with cgroup cpu.max `400000 100000`. Setup succeeded, first on the environment's
initial work branch and then again after switching to the requested base. Timing lines:

| Step | Initial setup seconds | Base setup seconds |
| --- | ---: | ---: |
| Go ready | 0.077 | 0.020 |
| clang ready | 0.476 | 0.154 |
| Node ready | 0.062 | 0.020 |
| markdown dependencies ready | 0.892 | 0.060 |
| submodules ready | 0.104 | 3.173 |
| WASI SDK ready | 4.230 | 3.189 |
| Go build ready | 40.738 | 236.444 |
| test binaries deferred | 40.817 | 236.537 |
| build cache warm | 40.818 | 236.539 |
| done | 40.842 | 236.586 |

Go is 1.27.1, clang 20.1.8, WASI SDK 27, Node v24.19.0. Exact setup commands:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh --wasi-sdk > /tmp/adamic-array-eight-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api > /tmp/adamic-array-eight-npm.log 2>&1
# After checking out f05aec3:
bash cloud/setup.sh --wasi-sdk > /tmp/adamic-array-eight-setup-base.log 2>&1
```

`npm ci` exited 0 and added 3 packages in 678ms. Neither setup failed. The first attempted baseline,
between branch checkout and the second setup, failed with:

```text
internal/lower/load_time_reads.go:10:2: no required module provides package github.com/system-inc/cohere/rule_runner; to add it:
    go get github.com/system-inc/cohere/rule_runner
```

Workaround: rerun the authorized setup to restore the base's pinned cohere 7945d102 and TypeScript
d92d9bfe submodules. No go.mod change or `go get` was needed. The successful baseline replaced that
preliminary log. The initial `adamic build` probe omitted `-o`, exited 2 with usage, and was corrected;
only corrected builds are claimed. The expression-statement reduction was refused and replaced by
the two-declaration witness above.

Corrected probe commands, with setup's environment sourced:

```sh
go build -o /tmp/adamic-array-eight-probes/adamic ./cmd/adamic > /tmp/adamic-array-eight-probes/compiler-build.log 2>&1
/tmp/adamic-array-eight-probes/adamic c <probe.a> > <probe.a.c> 2> <probe.a.c.log>
/tmp/adamic-array-eight-probes/adamic build <probe.a> -o <probe.a.bin> > <probe.a.build.log> 2>&1
node --input-type=module < <probe.a> > <probe.a.node.log> 2>&1
/tmp/adamic-array-eight-probes/adamic types /tmp/adamic-array-eight-probes/types.a > /tmp/adamic-array-eight-probes/types.log 2>&1
```

The eight original Node runs used Python `subprocess.run(['node', '--input-type=module'], input=...)`
with `sta.js + assert.js + original test source`; stdout, stderr and exit codes were written to
`/tmp/adamic-array-eight-node-originals.log`. All exit 0 with empty stdout and stderr.
Sources, generated C, diagnostics, corrected build statuses and Node outputs are preserved in
`evidence/probes.log.gz`. The original eight generated C programs are in `evidence/original-c.log.gz`.

Assumption: the instruction to leave compiler causes unchanged includes general lowering outside
library_array*.go. This classification follows the no-search reproducer and emitted C.
There are no new libc calls, runtime changes, oracle fixtures, counts changes, or checks to mutate.
The full repository gate and wasm execution were not run for this evidence-only change.
