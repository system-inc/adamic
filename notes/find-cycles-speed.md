Built: cached cycle edges and checker-identical literal relations; median fixture wall 15.63 s -> 10.20 s, findCycles CPU 14.83 s -> 9.28 s.
Commits: implementation 5b064cdcb7b93be1b28ed30f2322526095f0141c; report/evidence 414fae44; pinned/current main 48c05d091f0a43c31cbe051b1d6578d99eeedf19.
Checks: 1,106 complete decisions identical; lower package, filtered ownership oracles, exact markdown layout fixture, vet and formatting passed.
Mutant: dropping cached field edges changed 34 refusals into acceptance; the same complete decision comparator exited 1. Four arena/parent trace diffs were empty.
Limits: freshness analysis unchanged and still dominant; no full repository gate, no runtime graph work, and no new guarantee that all-distinct shape graphs are linear.

## Cause and change

The developer-tools cumulative profile at 9caa98e3 correctly separates two costs: reaches/assignability (5.63 CPU s) and fresh.ProveWrites (8.84 CPU s, including 6.00 s in state.copy). The latter runs once lazily; repeated freshness invocation was not the cause here.

The fixture has 1,558 program shapes. About 1,485 are repetitions of the regexp instruction record: kind, out, alternate, assertion and ranges. Distinct literal expressions produce distinct checker types, so the old search compares their shape pairs through distinct checker relation keys. It also rebuilds outgoing edges for each reach query. Caching outgoing edges alone barely helped: the expensive first-time shape matrix remained.

The final finder caches each node's target-independent outgoing edges. For plain literal records only, it also interns relation keys with identical type/object flags, member order, member flags, readonly status and exact member-type pointers. The existing cohere checker shim additionally confirms type identity. This uses the submodule API, without copying checker code. Classes, references, signatures, indexes, methods, accessors and inherited spread members retain their original identities. The regression tests exercise differing member types, spreads, methods and readonly literals against the checker.

Outside tracing, breadth-first search visits each such literal equivalence class once. Equal records have the same member-type edges and the same shape relations; their differing omitted self-shape edge adds no reachability. Captured cells retain their distinct identities. Target matching, slot order, freshness checks and refusal construction are unchanged.

Tracing uses the original nodes and breadth-first traversal, including each original shape edge. The trace helper was restored from the previous memoize unit because the pinned base lacks it. The old reference for this experiment is pinned main's finder with that identical helper, available at 0aaf96d95d980f8b1ab9b47e52b0732de5332eb6:internal/lower/cycles.go. No arena semantic fix was ported to production.

Scratch counters on the final fixture report:

```
shapes=1558 expanded=55 interned=1573 representatives=88 relation_pairs=2070
```

The preceding edge/relation-cache version expanded 1,538 nodes. The final visitation avoids those duplicate literal expansions. Unique structural families can still require pairwise checking, and the fresh-state copying cost remains.

## Measurements

The list_probe.ts source is unchanged between pinned main and the evidence branch. Its SHA-256 is 541e36fdb51601fec0ce350cb54568a012e3dc69d62ede5b3257aad1b35d45a9.

Three isolated, interleaved pairs ran after all validation jobs finished, in order before1, after1, after2, before2, before3, after3. Wall below measures load.Load plus lower.Lower, matching the fixture helper's scope. It excludes compiler build, emission, clang and profile shutdown. CPU profiling covers lower.Lower; findCycles CPU is inclusive sampled CPU, not wall time.

| Round | Before wall s | After wall s | Before finder CPU s | After finder CPU s |
|---|---:|---:|---:|---:|
| 1 | 15.6345 | 10.1986 | 14.83 | 9.28 |
| 2 | 16.8878 | 15.8195 | 15.68 | 14.08 |
| 3 | 14.3479 | 9.3110 | 13.78 | 8.73 |
| Median | 15.6345 | 10.1986 | 14.83 | 9.28 |

Median lower.Lower alone: 15.4477 -> 10.0657 s. Median reaches CPU: 5.91 -> 0.01 s, at 10 ms sampling resolution. Median fresh.ProveWrites CPU: 8.88 -> 9.23 s. The after round 2 outlier spends 14.04 CPU s in freshness, versus 9.23 and 8.69 in the other after rounds. That implementation is unchanged; these observations do not establish why its time varied. Do not interpret the medians as a fixed speedup for every run.

A separate unprofiled pair measured child peak RSS with Python resource.getrusage: 447,140 -> 210,044 KiB (about 437 -> 205 MiB). GNU time is absent (`/usr/bin/time: No such file or directory`), so that measurement used Python instead. The setup Go distribution omits the pprof executable; `go build -o /tmp/find-cycles-speed/pprof cmd/pprof` built its standard reader.

All six raw profiles, unrounded measurements and cumulative reports are in [evidence](../internal/lower/cycletools/evidence/performance.json). Exploratory and overlapping runs were excluded from the table.

## Decision and trace proof

[compare.py](../internal/lower/cycletools/compare.py) checks complete decision strings, including exact refusal locations and fixes. Every source file is an entry with its imports, using production load.Load and lower.Lower. It compares acceptance, Refused, NotYet and checker diagnostics. The final run reused the recorded old results only after checking each input's SHA-256; the old finder had already run on every input.

| Corpus | Entries | Differences |
|---|---:|---:|
| Oracle .a/.ts files, including dependency entries | 561 | 0 |
| All stage 1 .a/.ts files, including gaps | 319 | 0 |
| Pristine census compiler sources | 77 | 0 |
| Adapted census compiler sources | 79 | 0 |
| Census repros | 54 | 0 |
| Load fixtures | 16 | 0 |
| Total | 1,106 | 0 |

The census pin is TypeScript v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8. The adapted tree applied every adapter in this checkout, including generated diagnostics, host-errors and readonly-views. Both compilers consumed the same bytes. These are ordinary production checker options; rejected input was not forced into lowering.

There are 739 accepted entries and 46 cycle-capable refusals, exercising findCycles on 785 entries. The other 321 stop earlier and provide frontend decision parity, not graph-edge coverage. Totals are 180 checker stops, 111 NotYet, 76 Refused and 739 accepted.

The committed compressed ledgers retain input hashes, both decision hashes and exact non-checker decisions. Complete checker strings were compared in the run; their repeated text is omitted from the committed ledger and retained locally in /tmp/find-cycles-speed/quotient-comparison/decisions.jsonl. The [summary](../internal/lower/cycletools/evidence/quotient-comparison-summary.json) reports zero differences.

The cache mutant returns before storing any edge labelled `field ...`. The comparator ran it on the same 1,106 hash-verified inputs and exited 1: **34 changed decisions, all Refused -> accepted**, comprising 32 oracle entries and two stage 1 entries. For example fresh_refused/returns_argument.a closes node.nodes through a returned argument; the baseline refuses Node[] at line 4 and its push at line 11, while the mutant accepts it. No compile warning or backend failure caught this mutant: the decision diff itself did.

Arena and parent traces were byte-identical in four comparisons: the pinned gate and an equally patched scratch pair applying the previous never[] arena fix. The pinned gate's existing false arena refusal is deliberately preserved. In the scratch fixed pair, the arena is accepted, its trace is empty, and the parent mutant is refused with identical full diagnostic text. Its trace is:

```
cycle reach: CollapseCopyNode -- slot contents --> CollapseCopyNode[]
  CollapseCopyNode[] -- collection type argument --> CollapseCopyNode
  CollapseCopyNode -- related holder types --> CollapseCopyNode
cycle reach: CollapseCopyNode[] -- slot contents --> CollapseCopyNode
  CollapseCopyNode -- field parent --> CollapseCopyNode[]
  CollapseCopyNode[] -- related holder types --> CollapseCopyNode[]
```

The source is the previous numeric_handle_arena.a fixture; the mutant adds parent: CollapseCopyNode[] and initializes it with arena. All eight original/final trace files are committed in evidence.

## Commands, setup and limits

All tests wrote directly to logs. The final successful checks were:

```
go test ./internal/lower -count=1 -timeout 30m
ok internal/lower 56.462s
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestFreshWriteProbesStayRefused|TestNativeAgreesWithNode/internal/oracle/testdata/(fresh|closures|regexp_cycle)' -count=1 -timeout 30m
ok internal/oracle 12.566s
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/markdownblocks -run '^TestMarkdownWhitespaceLayout$' -count=1 -timeout 30m
ok stage1/cohere/markdownblocks 1063.797s
go vet ./...
no output, exit 0
gofmt -l cmd internal
no output, exit 0
git diff --check
no output, exit 0
```

The exact markdown fixture compares external Go, source Node, the JavaScript backend, native and the original document implementation, and runs its existing controls. Its total test duration overlapped corpus verification and is not a before/after performance comparison. The timing shell was held until that check and the mutant diff finished.

The reference binary was built with a Go overlay replacing cycles.go with the trace-only reference. The final binary uses the committed finder. Measurement commands were:

```
go build -overlay /tmp/find-cycles-speed/old-overlay.json -o /tmp/find-cycles-speed/old ./internal/lower/cycletools
go build -o /tmp/find-cycles-speed/quotient ./internal/lower/cycletools
/tmp/find-cycles-speed/old -profile /tmp/find-cycles-speed/measured-before-1.pprof stage1/cohere/markdownblocks/testdata/list_probe.ts
/tmp/find-cycles-speed/quotient -profile /tmp/find-cycles-speed/measured-after-1.pprof stage1/cohere/markdownblocks/testdata/list_probe.ts
python3 internal/lower/cycletools/compare.py /tmp/find-cycles-speed/old /tmp/find-cycles-speed/quotient /tmp/find-cycles-speed/typescript /tmp/find-cycles-speed/quotient-comparison --manifest /tmp/find-cycles-speed/all-manifest.json --baseline /tmp/find-cycles-speed/baseline-all.jsonl
```

Their outputs and the repeated rounds were redirected into /tmp/find-cycles-speed logs; final evidence is committed under internal/lower/cycletools/evidence. inputs.json records every entry. The corpus helper also supports fresh old/new execution without --baseline.

Toolchain setup first exported GOPROXY='https://proxy.golang.org|direct', ran bash cloud/setup.sh and sourced /workspace/adamic-tools/env.sh. nproc=5, CPU quota=4, Go 1.27.1, Node 24.19.0, clang 20.1.8. Setup succeeded:

```
setup: go ready (0.037s)
setup: node ready (0.058s)
setup: submodules ready (0.122s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.012s
setup: markdown dependencies ready (0.167s)
setup: clang ready (0.263s)
setup: go build ready (35.183s)
setup: test binaries deferred (use --warm-tests) (35.384s)
setup: build cache warm (35.386s)
setup: done (35.442s)
```

Current main was explicitly fetched and merged: Already up to date. The full repository gate and complete oracle package were not run under the worker-gate exception. No protected compiler/native files, runtime files, oracle_test.go or cohere sources were edited. No PR was opened. This unit does not change cycle policy, admit memoize, fix native undefined readiness or implement graph regions.
