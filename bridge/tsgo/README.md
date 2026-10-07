# typescript-go in native Adamic

Task k4fm1vf, milestone 5. The pinned checker is compiled into a Go C archive.
Adamic calls it in process through the public ABI in `tsgo.h`. No subprocess or
JSON protocol sits between a query and the checker.

From the repository root, after `bash cloud/setup.sh` and sourcing its tool environment:

```sh
go build -buildmode=c-archive -o /tmp/tsgo.a ./bridge/tsgo/archive
go build -o /tmp/adamic ./cmd/adamic
/tmp/adamic build bridge/tsgo/testdata/queries.a -o /tmp/queries --tsgo /tmp/tsgo.a
```

An Adamic program imports these declarations from `adamic`:

```ts
import { tsgoProgram, tsgoQuery, tsgoRelease } from 'adamic';

function run(): void {
    const opened = tsgoProgram('/path/tsconfig.json', ['/path/source.ts']);
    if (opened.kind === 'Error') { console.error(opened.message); return; }
    const answer = tsgoQuery(opened.value, '/path/source.ts', 42);
    if (answer.kind === 'Error') { console.error(answer.message); }
    else console.log(`${answer.value.nodeKind} ${answer.value.symbolName}: ${answer.value.type}`);
    const released = tsgoRelease(opened.value);
    if (released.kind === 'Error') console.error(released.message);
}
run();
```

All five bridge calls return a discriminated result. These declarations are
exported from `adamic`:

```typescript
export type TSGoError = { readonly kind: 'Error'; readonly message: string };
export type TSGoResult<T> = { readonly kind: 'Ok'; readonly value: T } | TSGoError;
```

`tsgoProgram` returns `TSGoResult<number>`, `tsgoQuery` returns
`TSGoResult<{ readonly nodeKind: number; readonly symbolName: string; readonly type: string }>`,
and `tsgoInspect` and `tsgoTypeParts` return `TSGoResult<string>`.
`tsgoRelease` returns `{ readonly kind: 'Ok' } | TSGoError`.
This is a breaking change from the scalar/void declarations: callers must narrow
`kind` and unwrap `value`. There are no implicit unwraps or error-to-panic helpers.
The harness owns its Answer/refusal type and its refusal wire format.

A checker failure returns `Error` with the C error buffer's text decoded from
its explicit UTF-8 byte length. Strings are copied into Adamic's counted storage
before C buffers are freed. Returned values remain valid after releasing their
checker handle. Both the result wrapper and any object payload support statement
regions. Invalid numeric arguments and NUL-bearing strings return errors in
Adamic's own words, identical in native and the Node adapter. A failed boundary
call without an error buffer returns `tsgo: invalid argument or out of memory`.
Optional timing clock failure does not panic or change the checker result.

Native build requires `--tsgo <archive>`; unlinked native calls and the ordinary
C command are refused. The dedicated native renderer replaces five reserved
compiler-generated bodies and their region variants. Ordinary rendering of
opted-in IR produces an explicit `Error` result if the renderer is unavailable.
Library calls are direct; taking them as function values remains unsupported.
The bridge is target-refused for wasm32-wasi, and runtime sources compile without
its unavailable Go/POSIX dependencies on that target.

JavaScript uses an explicitly selected Node-API adapter linked to the same C
archive. It does not reimplement the checker or replay calls in subprocesses.
On Linux:

```sh
clang -std=c11 -Wall -Wextra -Werror -pedantic -fPIC -shared -Ibridge/tsgo \
  bridge/tsgo/node/node.c /tmp/tsgo.a -lpthread -ldl -lm -o /tmp/tsgo.node
/tmp/adamic js bridge/tsgo/testdata/errors.a > /tmp/errors.mjs
ADAMIC_TSGO_NODE=/tmp/tsgo.node node --disable-warning=ExperimentalWarning \
  oracle/node.mjs /tmp/errors.mjs /absolute/tsconfig.json /absolute/source.ts
```

On macOS, replace `-ldl` with `-undefined dynamic_lookup`. Only Linux is gated
here. Without `ADAMIC_TSGO_NODE`, a JavaScript call returns
`Error` with `tsgo: checker library unavailable in JavaScript`; an adapter-load
failure returns `tsgo: cannot load JavaScript checker adapter`.
The small stable Node-API declaration subset is copied from Node v24.19.0;
its upstream MIT license is retained beside it.

No bridge error calls `adamic_panic`. Global Adamic allocation/stack exhaustion
can still panic while allocating a result or input buffer: those failures cannot
allocate an error value with the current allocator. Fatal Go runtime failures
(such as Go heap exhaustion) are outside the recoverable checker panic contract.
Recoverable Go checker panics already become C error buffers, which now become
`TSGoError` values. This unit does not change the C ABI or checker decisions.

For the first type-aware rule, `tsgoTypeParts(program, file, byteStart,
byteEnd, nodeKind)` selects an exact node by its trivia-inclusive byte span and
kind name, including identifier tokens. Its successful value contains constrained
type union parts as flags and TypeToString frames documented in `tsgo.h`.
Adamic decides the rule predicate. See [the type-aware lint driver](../../stage1/cohere/typeaware/README.md).

## C contract

Use `bridge/tsgo/tsgo.h`, rather than cgo's generated implementation header.
The public ABI contains program creation, query, release, output-buffer free and
query-result free. Strings carry UTF-8 byte lengths and have no terminator.
The header states every input's and output's owner. Input views are copied in C
before Go reads them; output buffers are exact-size C allocations. No Go pointer
crosses the ABI. Calls use monotonic integer IDs, are serialized, and reject zero,
stale and twice-released handles. The caller releases each handle once and frees
all output buffers, including error buffers. IDs never exceed JavaScript's exact
integer range, so the native Adamic `number` representation loses no bits.

A nonempty file list replaces the tsconfig's roots and is resolved relative to
the config directory. An empty list uses its roots. Query paths are resolved
relative to the process working directory. Positions are zero-based UTF-8 byte
offsets, including positions inside multibyte characters; EOF is refused. The
node selection is typescript-go's `GetNodeAtPosition(..., false)`, without JSDoc.
Node ranges include leading trivia. Kind is the pinned checker's numeric enum;
symbol names use `SymbolName`, with internal name markers escaped as `__`, and
types use the checker's default `TypeToString` exactly, including its normal
truncation. Semantic errors in source are allowed, as they are for a lint checker;
configuration errors, missing roots, absent query files and invalid positions
are errors. Content mappers are explicitly unsupported.

Adamic values keep their ordinary reference counting and regions. The external
checker library necessarily contains Go's runtime, allocator and collector.
Releasing a handle drops the program's Go roots; Go decides when to reclaim its
heap. It does not unload the Go runtime. ASan/LSan validate the C boundary and
Adamic allocations, not reclamation of the checker's Go heap. This is an external
checker bridge until stage 3, not a collector added to Adamic's own memory model.

## Oracle, sanitizers and mutants

The test logs must be redirected to files:

```sh
go test -v -count=1 -timeout 30m ./bridge/tsgo > /tmp/tsgo-smoke.log 2>&1
```

The default corpus queries every byte of the UTF-8 sample file, including an
accented identifier and a non-BMP string. For the requested compiler corpus,
fetch TypeScript 6.0.3 separately from the typescript-go submodule:

```sh
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/tsgo-typescript
ADAMIC_TSGO_CORPUS=/tmp/tsgo-typescript go test -v -count=1 -timeout 30m ./bridge/tsgo > /tmp/tsgo-corpus.log 2>&1
```

The pinned TypeScript commit is `050880ce59e30b356b686bd3144efe24f875ebc8`.
The corpus uses `checker.ts`, `parser.ts`, `types.ts`, and `utilities.ts`, with
400 evenly spaced byte positions in each. `oracle/main.go` independently loads
its program and calls the shim directly. It imports no bridge implementation.
Both runners write kind, symbol byte length, type byte length, symbol and type;
the complete output is compared byte for byte. The native query loop is compiled
from `testdata/queries.a` by stage 0, not handwritten C. A separate
`testdata/region.a` probe reads only string lengths, proves its wrapper and payload were made
in one statement region with `--count`, and compares those lengths with the Go oracle.

The tests build both an ordinary archive and one with an instrumented C boundary:

```sh
CC=clang CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all' \
  go build -buildmode=c-archive -o /tmp/tsgo-asan.a ./bridge/tsgo/archive
/tmp/adamic build bridge/tsgo/testdata/queries.a -o /tmp/queries-asan \
  --tsgo /tmp/tsgo-asan.a --sanitize
```

`testdata/api.c` verifies 100 repeated queries, exact-length inputs, owned outputs
surviving release, cleared frees, invalid positions, zero/stale handles, double
release and distinct subsequent handles. Mutants are built through Go overlays;
the checkout is never mutated. The input-length mutant is a deliberately invalid
caller view over an exact-size heap buffer. The other mutants change production
code: output length plus one, a release that keeps the handle live, a type queried
from the source-file node instead of the selected position, removal of the link
opt-in guard, omission of the output free, and heap allocation in the region entry.
The tests require the relevant sanitizer report, assertion or oracle failure,
so a compilation failure does not count as catching a mutant.

`ADAMIC_TSGO_TIMING=1` makes native release print `load_ns`, accumulated `query_ns`
and `queries`, plus `first_query_ns`, to stderr. Native uses a monotonic clock; Go uses `time.Since`.
The optimized native runner and ordinary Go oracle run in three interleaved
rounds. Timing excludes output formatting and printing, includes native C copies
and conversion to Adamic values, and includes lazy checker work in the query
that first asks for it. Create time is program loading plus checker-pool initialization, not a full semantic check.
These are API-call timings, not whole-process startup or complete lint timings.

Only Linux amd64 with Go 1.27.1 and clang 20.1.8 was measured here. C archive
support works on that platform. Windows, macOS, c-shared, concurrent callers,
allocation-failure injection, full semantic diagnostic retrieval and compiling
cohere's complete lint engine are not covered by this unit.

## Six-rule facts

The additive `tsgo_inspect` API exposes only the checker questions needed by the
next five native rules. See [facts.md](facts.md) for framing, identities and
ownership, and [the six-rule report](../../stage1/cohere/typeaware/SIX_RULE_REPORT.md)
for complete finding agreement and separate load/query/run measurements.
`run_ns` now measures wall time between successful create and the start of release;
existing `query_ns` continues to measure accumulated native adapter intervals.

## Errors-as-values proof

`TestErrorsAsValues` builds the real Node adapter, runs `testdata/errors.a` as
source on Node and as emitted JavaScript, and compares both with sanitized native
output. The caller refuses malformed inputs, a missing root, invalid spans,
query EOF, released handles and double release, then continues. Successful
creation, query, inspection, type-parts and release are exercised too; their
values survive release. Native error buffers are freed on every path.

Two production overlays must compile and run before being caught by comparison
with the unmodified source on Node: a C error path that panics, and a failure
reported as an `Ok` carrying the error text. Native target entrypoints also refuse
WASI before resolving or linking an archive; the ordinary WASI runtime archive
remains buildable.
