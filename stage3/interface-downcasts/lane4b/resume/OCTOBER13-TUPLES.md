Built: merged views-integration 432d4913 and certified four original tuple overlaps, seven candidate reads.
Commits: clean merge 4f4ff26deb0d156d53428a3a666652ee1c4d625b; certificate group committed immediately after this report.
Checks: original tuple controls PASS 27.024s; prior lane controls PASS 180.195s; scoped counts and mutant update PASS 24.074s; final verification PASS 23.118s.
Mutants: seven independent omissions caught in sanitized native, release native and JavaScript, 21 semantic counterfactuals; no crashes or compiler failures credited.
Uncovered: five pairs / 16 reads remain blocked; total certified 37 pairs / 165 candidate reads; production tsc reachability unmeasured.

Integration merged without conflicts. The canceled non-null compiler-area
branch is not an ancestor. Integration itself moves three existing non-null
fixtures to .ts byte-for-byte; this group creates no .ts source. The provenance
key for source-file-uninitialized follows that rename with the identical hash.
The ranked audit now checks all fixture and shared certificate hashes again.

Original source remains pinned to 050880ce59e30b356b686bd3144efe24f875ebc8.
Both declaration preparations emit the same original declaration corpus. The
tuple preparation validates that pin and the complete declarations before any
fixture executes. adopt-tuple-certificates.py compares every original read's
file, offsets and text against the tuple owner's candidate census, and matches
each certified owner's count. It pins both owner files by hash. This is shared
certificate credit within this lane's existing inventory, not additional global
completion. Standalone witnesses do not prove whole-tsc execution coverage.

| Pair | Reads | Fresh mutants |
| --- | ---: | --- |
| 97898.<element> | 4 | omit branded file-id check; omit signature-position union check |
| 97913.<element> | 1 | omit root member read check |
| 97931.forEach | 1 | weaken the two-position signature tuple to one position |
| 97934.outSignature | 1 | omit outer field check; replace tuple shape with object shape; omit nested signature check |

Every unmutated poisoned witness fails at the exact named read with exit 70 in
all three modes. Every mutant then exits zero with nonempty stdout and empty
stderr in all three modes. These outputs are counterfactuals, not claimed
Node-equivalent typed results: native unchecked boolean positions can appear as
undefined or zero. Source Node is independently checked before compilation.
Full positive owner controls cover string, tuple, scalar and undefined members,
with native leak checks. The four formerly refused frontier controls now also
match Node across all backends and pass leak checks. The historical blanket
tuple refusal assertion becomes a supported positional descriptor assertion;
unsupported tuple kinds still retain the compiler's guards.

Eleven scoped original count labels were added (four frontier positives and
seven poisoned witnesses). The existing self-contained fixes-tuple.a fixture
now has a registry entry and measured counts.md row. The mandated global
TestCountsAreRecorded update was attempted and failed in 60.144s with 39 fixture
failures outside this group. Their baseline status was not independently
established here, and they were not pursued. The focused updater touches only
owned rows, which are then verified without update-counts.

Commands, with tool environment sourced and all output redirected to saved logs:

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh
node stage3/interface-downcasts/tuples/prepare.cjs /tmp/lane4b-upstream /tmp/lane4b-tuple-declarations
ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS=/tmp/lane4b-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewObjectPrimitiveOriginalPairs$/^compiler-options-key-probe$' -count=1 -v -timeout=10m
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/lane4b-tuple-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewTuple(Original(SignaturePositions|RootIndex|SignatureForEach|OutSignature)|Lane4b(CastFrontier|OutSignature))$' -count=1 -v -timeout=15m
ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS=/tmp/lane4b-declarations ADAMIC_INTERSECTION_ORIGINAL_DECLS=/tmp/lane4b-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewObjectPrimitive(Source|OriginalPairs|CommentPairs|PackagePairs|OverlapCounts)$' -count=1 -v -timeout=15m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout=15m -args -update-counts
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/lane4b-tuple-declarations ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS=/tmp/lane4b-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewObjectPrimitive(TupleMutants|RemainingFrontiers|Fixes|FixCounts)$' -count=1 -v -timeout=10m -args -update-counts
ADAMIC_TUPLE_ORIGINAL_DECLS=/tmp/lane4b-tuple-declarations ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS=/tmp/lane4b-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewObjectPrimitive(TupleMutants|RemainingFrontiers|Fixes|FixCounts)$' -count=1 -v -timeout=10m
python3 stage3/interface-downcasts/lane4b/resume/adopt-tuple-certificates.py --update
python3 stage3/interface-downcasts/lane4b/resume/adopt-intersection-certificates.py
python3 stage3/interface-downcasts/lane4b/resume/rank-lazy-demand.py
```

Setup: Node 0.035s, Go 0.032s, markdown 0.092s, submodules 0.097s,
clang 0.239s, build 78.961s, cache 79.099s, done 79.136s; nproc 5,
CPU quota 4. Initial frontier recheck failed solely on four obsolete compile
refusal assertions; an initial mutant assertion said element instead of field
and was corrected to the observed exact diagnostic. Both logs are preserved.

Routing blockers, all rechecked after integration, most reads first:

- Lane 6 dictionary admission plus lane 4b member adaptation: CompilerOptions
  and BuildOptions dynamic keys, two pairs / 11 reads, both cast refusals.
- Lane 5 callable field union: EmitHelper.text, one pair / three reads.
- Lane 2 mixed array storage/consumer, followed by lane 4b member selection:
  bundle fileInfos and multi/bundle fileInfos forEach, two pairs / two reads.

Separately, direct array-to-tuple cast 97898 remains refused. Its four original
positional reads are certified through the admitted original carrier, so this
boundary is not double-counted as another remaining pair. No local refusal was
weakened. Stopping here because the remaining inventory needs those compiler
changes in views-integration, as requested.
