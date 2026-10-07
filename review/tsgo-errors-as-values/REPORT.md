Built: all five checker bridge calls return Ok/Error values; real Node adapter proves both backends.
Commits: claims 5db6a30d022d7f703a79953d431d0399b64a8734; implementation cfc27004721facddaefb1bf38f8b1521c72cd5ab.
Gates: touched packages, flow, bridge/checker, whole oracle with WASI, Linux counts, WASI archive and vet passed.
Mutants: panic instead of Error and Ok containing error text compiled and ran; Node comparison caught both.
Limits: global allocator/stack exhaustion and fatal Go runtime failures remain fatal; macOS/Windows and harness migration not tested.

# RuleContext bridge errors as values

Task #45rq89s, design review #k4fm1vf section 5. Branch
`codex/tsgo-errors-as-values`, based on
`48c05d091f0a43c31cbe051b1d6578d99eeedf19`. The task owner supplied section 5
verbatim after the early claims push. Its Answer/refusal type and wire line
belong to typeaware-08, not this unit.

## Contract

The prelude exports:

```typescript
export type TSGoError = { readonly kind: 'Error'; readonly message: string };
export type TSGoResult<T> = { readonly kind: 'Ok'; readonly value: T } | TSGoError;
export type TSGoQuery = {
    readonly nodeKind: number;
    readonly symbolName: string;
    readonly type: string;
};
```

| Call | Result |
|---|---|
| tsgoProgram | TSGoResult<number> |
| tsgoQuery | TSGoResult<TSGoQuery> |
| tsgoInspect | TSGoResult<string> |
| tsgoTypeParts | TSGoResult<string> |
| tsgoRelease | { readonly kind: 'Ok' } \| TSGoError |

Assumption, recorded in the implementation commit: successful payloads use
`value`; successful release has no payload. The user explicitly approved the
claimed `TSGoError` shape. All callers must narrow `kind` before unwrapping.

C error buffers are decoded with their explicit UTF-8 byte lengths, copied to
Adamic-owned strings, and freed. Partially produced query/string buffers are
also freed on failure. A boundary failure without a message has the uniform
fallback `tsgo: invalid argument or out of memory`. Invalid numeric inputs and
NUL-bearing strings return uniform Adamic messages, not panics. No
`adamic_panic` or `ir.Panic` remains in the native adapter or its lowering.
Optional profiling-clock failure no longer aborts a checker call.

Every result-producing native entrypoint supports an ordinary heap result and
a statement-region result. Query's nested payload and its wrapper both live in
the supplied region. Unlinked ordinary rendering returns a named Error value;
unlinked native compilation still refuses the call.

JavaScript emission calls an explicit Node-API adapter linked to the same Go C
archive. Node executes the real checker in process; it does not simulate bridge
failures, replay requests, or run checker subprocesses. `ADAMIC_TSGO_NODE` selects
the adapter at runtime. Missing or unloadable adapters produce named Error
values. Minimal stable ABI declarations come from Node v24.19.0, with the
upstream MIT notice retained. Build instructions are in bridge/tsgo/README.md.

WASI has no checker bridge. Both library build entrypoints refuse that target
before resolving an archive; CLI refusal is held as well. The runtime's checker
implementation is guarded by `ADAMIC_TSGO && !ADAMIC_TARGET_WASI`. Its ordinary
WASI archive therefore compiles without Go/POSIX checker dependencies. The
Darwin feature macro was retained beside the POSIX macro.

## Observed setup

`GOPROXY='https://proxy.golang.org|direct'` was set before setup and Go commands.
`bash cloud/setup.sh --wasi-sdk > /tmp/tsgo-values-setup.log 2>&1` exited 0.
Each tool shell sourced `/workspace/adamic-tools/env.sh`.

```text
Node v24.19.0; node ready 0.022s
Go go1.27.1 linux/amd64; go ready 0.025s
submodules ready 0.072s
markdown dependencies ready 0.076s
clang 20.1.8 ready 0.241s
WASI SDK ready 0.279s
go build ready 34.813s
build cache warm 35.040s
setup done 35.070s
nproc=5; cgroup cpu.max=400000 100000
```

The existing Go build cache subsequently filled the workspace disk. Removing
that regenerable cache with `go clean -cache` freed about 27 GB; builds resumed.
No source or test evidence was removed.

## Fixture and mutant observations

`bridge/tsgo/testdata/errors.a` exercises missing-root creation, NUL-bearing
config/root/path/kind/question arguments, negative/fractional/NaN/infinite
numbers, invalid inspection and type-parts spans, out-of-range queries, stale
query/inspection/type-parts handles and double release. It also exercises
successful creation, query, inspection, type-parts and release, then reads
retained query/string values after release. The caller prints refusals and
continues; it does not catch or invoke a process panic.

Source on Node, emitted JavaScript on Node, and sanitized native output were
identical: **1,062 stdout bytes, empty stderr, exit 0**, ending with
`caller continued`. All use the real checker archive. Linux native validation
explicitly enabled LeakSanitizer; macOS does not enable that unsupported leg.
The source and emitted-JavaScript runners are checked for empty stderr too.

The two required mutants changed production native adapter source through Go
overlays. Neither a compile failure nor an unrelated crash counts as detection:

| Mutant | Observed execution and detector |
|---|---|
| C error path panics | Compiled; missing-root call exited 70 with `adamic: panic: no file at ...`; unmodified Node continued and exited 0 |
| Error becomes Ok carrying its text | Compiled; exited normally and printed `unexpected Ok open missing`; unmodified Node printed its refusal |

The existing bridge mutants were rerun too:

| Mutant | Detector |
|---|---|
| Input view length plus one | ASan heap-buffer-overflow |
| Output string length plus one | ASan heap-buffer-overflow |
| Released handle retained live | C stale-handle assertion |
| Query type taken from source-file node | Independent Go oracle mismatch at byte 6 |
| Link opt-in guard removed | Lowering refusal test |
| C output free omitted | LeakSanitizer |
| Region wrapper allocated on heap | LeakSanitizer |

The query corpus compared all 162 sample-file byte positions, including Unicode,
with the independent direct-Go checker oracle: 3,261 identical output bytes.
The counted region probe agreed with Go on answer `6` and reported:

```text
allocations 12 frees 10 retains 3 releases 10 peak 10 regions 2
```

The two region values are the result wrapper and query payload; heap frees plus
region values account for every allocation. The old scalar bridge fixtures were
migrated to explicit result handling.

## Gate commands and outputs

Every command wrote test output to a log file; no test output was piped.

```sh
go test -count=1 -timeout 30m ./internal/lower ./internal/native ./internal/flow ./internal/javascript ./internal/load ./cmd/adamic > /tmp/tsgo-values-packages-final.log 2>&1
```

Exit 0: lower 87.307s, native 305.792s, flow 198.853s, load 2.781s,
cmd/adamic 27.549s. JavaScript has no package tests; its backend is exercised by
the bridge proof and whole oracle. The native run includes `TestTSGoRefusesWASI`
and compiles every ordinary runtime source into the WASI archive.

```sh
go test -v -count=1 -timeout 30m ./bridge/tsgo/... > /tmp/tsgo-values-bridge-final.log 2>&1
```

Exit 0: bridge 287.999s; checker 0.515s; archive/cost/oracle/spec have no package
tests. Existing bridge mutants and the new required mutants passed their
expectations. After tightening stderr parity, the final error proof was rerun:

```sh
go test -v -count=1 -timeout 30m ./bridge/tsgo -run '^TestErrorsAsValues$' > /tmp/tsgo-values-errors-final.log 2>&1
```

Exit 0, 19.317s; both required mutants compiled, executed, and were caught again.

```sh
ADAMIC_ORACLE_WASI=1 go test -count=1 -timeout 40m ./internal/oracle > /tmp/tsgo-values-oracle.log 2>&1
go test -count=1 -timeout 30m ./internal/oracle -run '^TestCountsAreRecorded$' -args -update-counts > /tmp/tsgo-values-counts.log 2>&1
```

Both exited 0: whole oracle including WASI 654.522s; Linux counts 63.798s.
Regenerated counts.md has no content diff. The last adapter C edits after the
oracle started only renamed checker-only shape arrays to satisfy the existing
layout proof; ordinary/native-WASI oracle builds preprocess that implementation
out. Final native and bridge runs include those edits.

```sh
go test -count=1 ./internal/native -run 'TestRuntimeFieldLayoutsAreIncluded|TestTSGoRefusesWASI' > /tmp/tsgo-values-layout-wasi.log 2>&1
go vet ./internal/lower ./internal/native ./internal/flow ./internal/javascript ./internal/load ./cmd/adamic ./bridge/tsgo/... > /tmp/tsgo-values-vet.log 2>&1
git diff --check
```

Targeted layout/WASI run exited 0 in 5.192s. Vet exited 0 with empty output;
diff checking passed.

An initial fixture guard incorrectly expected an identifier symbol at a position
where GetNodeAtPosition chooses the initializer; the guard was corrected to the
actual checker answer, while the independent output comparison was retained.
The first package run caught duplicate C shape-array names that confused its
existing layout checker. The first region run caught its argument escaping via
an error-message print. Names were made unique and the pure region control now
returns a visible negative sentinel on unexpected refusal. Both gates were
rerun successfully; the actual error fixture always prints and continues.

## Limits and ownership

No recoverable bridge failure is still routed to a process panic. Shared Adamic
allocator or stack exhaustion can still panic while creating the result/input
buffer; there is no fallible allocator with which to construct that Error.
Fatal Go runtime failures, notably heap exhaustion, cannot be recovered by
Go's checker-panic recovery. These are global runtime limits, not returned C
checker statuses. Allocation-failure injection is not covered.

Only Linux amd64 was run. macOS adapter linking uses `-undefined dynamic_lookup`
and omits Linux LeakSanitizer options; it was not executed. Windows is not
covered. JavaScript needs an explicitly built adapter; the absence of an adapter
is a refusal, not a successful checker answer. The external Go runtime still
owns its own heap and collector.

Existing stage1/cohere/typeaware drivers use the old scalar bridge contract;
typeaware-08 owns their migration and harness refusal handling. Those stage1
suites were not run or silently adapted to a panic helper. This unit claims the
requested package/bridge/oracle/flow/counts/WASI gates, not `go test ./...` or
completion of the shared harness.

No main push, force-push or rebase was performed. Only the requested task branch
was pushed; no pull request was opened. The standing landing rule is acknowledged:
when explicitly asked to land, use the designated own landing branch and give
its verified SHA; integration alone moves main.
