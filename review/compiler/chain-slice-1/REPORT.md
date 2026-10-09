Built independent c-portability slice 1 on main 5e33a17b toward lowering integration #wj4pmt1; the complete slicing unit is unfinished.
Code commit: e6cc9bef21012d6b1cbedf2ebf6685cb20edf526, source member ebad12cf143890e9972b7cfda4dfdf0c4b106484.
Proofs pass: 14 lowering shards, 274 passed/2 existing skips; portability 4.215s; uncached Node oracle 4.694s; reader guard 17.599s; counts 61.739s; scoped vet exit 0.
Mutants caught: plain signed bytes by 4,200 GCC overflow warnings; static number and boolean compound literals by one GCC pedantic warning each, with -Werror absent.
Not covered: complete true dependency graph, certified later slice order, complete red-to-member repair attribution, all waiting-task attribution, WASI, MSVC, physical ARM, or full gate.

## Delivered change

One member commit applies the source's net changes from 56ee1ddc, 230a996d
and 099a7aaf once. Long literal byte arrays use unsigned char, the runtime string
initializer casts the address to its immutable const-char view, and optional
scalar globals use C initializer lists. The chain's later repairs do not modify
this portability implementation. No unlanded topic branch is merged.

The source tests and long_unicode_literals.a fixture are preserved. Clang and
GCC compile and execute both emitted backends against source Node. ASan, UBSan
and leak detection are clean. The coverage test observes all 115 valid high
UTF-8 bytes in long literals. Native test leaf seconds are recorded in
leaf-seconds.json; the maximum is 4.14s, including its runtime build.

The admission proof compares all load/lower/IR/JavaScript sources, the checker
submodule and Go dependency inputs byte-for-byte with main. They are identical.
Native emission follows admission, so the frontend/lowering admission delta is
empty for every input. This is a source-equality proof, not a sampled corpus
claim. C compiler portability is the changed output property. See
admission-delta.json and admission-proof.py.

The complete successful counts writer ran once. Only the fixture row is added:
allocations/frees/retains/releases/peak/regions = 32/32/17/51/5/0. Removing that
one line makes counts.md byte-identical to main's table. No existing row moves,
changes or disappears. counts-attribution.json records that check.

## Chain inventory and remaining work

[PLAN.md](../chain-slices/PLAN.md) records 28 members from the actual merge
parents, their source SHAs, applicability observations and repair owners.
It includes optional-chain-after-call, spread-shorthand and optional-calls,
which were merged through the earlier lowering-a/b envelopes. The complete
commit and merge messages are saved beside the plan.

The first branch-net census was refined by removing stacked ancestors from
step24, per-backend stops, search-shrink, optional-presence and eep-presence.
The tsgo cache check now isolates internal/native/tsgo.go. Exact ranges and
conflict diagnostics are saved in own-net-patch-checks.json. Optional-calls
still needs its inherited lowering-a changes extracted. Counts and historical
evidence are excluded from code applicability checks and require their own
attribution during extraction.

Only slice 1 has proven independence. The later groups in the plan are proposed
extraction groups, not certified landing slices. Branch ancestry, shared files
and clean patch application do not establish a minimal semantic dependency
graph. The complete graph requires examining the remaining members' final
merged hunks and their repair/fixture edges; that work was not completed here.
The e628f527 repair families and lint repair are identified, but the full
16-red one-to-one attribution remains unfinished. No waiting task in a later
group is claimed closed. This partial result must not be reported as completing
#wj4pmt1's entire chain-slicing unit.

## Commands and observations

All test output went to the files here. Every process had an outer limit;
Go test commands had an internal limit. No whole package test or full gate ran.
Every test shell sourced /workspace/adamic-tools/env.sh. Warm proofs used
GOMAXPROCS=4 and unset GOCACHEPROG after cold shared-cache setup problems.

```text
export GOPROXY='https://proxy.golang.org|direct'
timeout 240 bash cloud/setup.sh
  exit 124 during cold dependency compilation; setup-initial.log
ADAMIC_GOCACHE_OFF=1 GOMAXPROCS=4 GOFLAGS='-buildvcs=false -trimpath -p=4' timeout 600 bash cloud/setup.sh
  first retry exit 1 in go list; its list.log and package output were empty
  final retry exit 0; setup.log
source /workspace/adamic-tools/env.sh
timeout 120 npm ci --prefix stage3/api
  exit 0; installed pinned declarations, node-types.log
python3 review/compiler/chain-slice-1/admission-proof.py
  PASS exact frontend/lowering equality
python3 review/compiler/chain-slice-1/run-lower-shards.py
  14 exact disjoint selectors, 276 terminal verdicts, 274 pass, 2 skip
  every shard uses go test ./internal/lower -run '^(...names...)$' -count=1 -json -timeout 90s
  lower-shards.json, lower-results.json, lower-union.json and lower-*.jsonl
 timeout 180 go test ./internal/native -run 'TestConstantPortability|TestLongStringBytesMutant|TestStatic.*InitializerMutant|TestLongUnicode' -count=1 -v -timeout 90s
  PASS 4.215s; portability.log
 timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout 90s
  PASS 17.599s; readers.log
ADAMIC_GATE_UNCACHED=1 timeout 180 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/long_unicode_literals.a$' -count=1 -v -timeout 90s
  PASS 4.694s; 3 native and 2 Node cache misses; oracle.log
 timeout 650 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -v -timeout 10m -args -update-counts
  PASS 61.739s; counts.log
 timeout 120 go vet ./internal/native ./internal/oracle
  exit 0, empty output; vet.log
```

The unchanged existing counts sweep exceeds the new-leaf 60s grain and was
bounded separately. New member test leaves are all under 60s. Existing lowering
skips are TestOriginalCycleLedger and TestOptionalWideningCensus, plus the
inherited nested mixed-union pending case. No skip was added or weakened.

Initial portability, lower enumeration, reader, oracle and counts attempts
reached outer compile limits without producing test verdicts. Their empty
initial logs are retained and are not credited as passes. The second setup
failure returned 1 with empty list.log/packages.json, so its cause is not
established. Abandoned scratch Go build directories consumed most of /tmp;
only named directories from completed timeout processes were removed. The
next setup succeeded. These are infrastructure observations, not semantic reds.

Final setup timings: Node 0.017s; Go 0.020s; markdown ready 0.058s;
submodules 0.064s; clang 0.141s; shared cache off 0.143s; build 53.737s;
test binaries deferred 53.933s; cache warm 53.939s; done 54.026s.
nproc=5; cpu.max=400000 100000 (4 CPUs). Go 1.27.1, clang 20.1.8,
Node 24.19.0. The environment file is /workspace/adamic-tools/env.sh.

## Lane checks and delivery

The required command runs on committed delivery inputs before the single push:

```text
git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
```

lane.log and lane-status.log record its result. The final answer reports the
pushed head and actual lane output. The code commit is one member on main;
additional commits contain only the plan and review evidence. No PR is opened.
