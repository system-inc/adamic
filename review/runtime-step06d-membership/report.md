Built: the missing prototype membership-mutant tests and an explicit Linux counted mapper proof; no runtime or compiler changes.
Commits: runtime/step06d-membership from 325aded7e5d481d67119dcb1a5cb9d28e05f2e8a; delivery SHA accompanies the push.
Commands: ownership and mapper fixtures held to Node, ASan/UBSan, Linux leaks and balanced counts; both census selectors and native/lower/oracle vet pass.
Mutants: extra member, missing member, actual member storage reuse and early member release are caught; compiler's three fault plants independently fail their tests.
Uncovered: full package gates, non-Linux hosts, graph-region fallback and original tsc lowering; K60/K144 records and expectations remain unchanged.

## Contract and scope

Read the nine-point October 8 ruling in d23a4fdc:docs/memory.md, Program region
section, as the available source of the 06a contract. Membership is inferred from
concrete allocation-site types; marked members survive until one-shot CLI teardown.
Non-graph strings, scalar maps and transient arrays stay counted. Over-inclusion
costs retention, under-inclusion must remain memory-safe or be refused.

The landing already contains program_region_ownership.a and
program_region_map_storage.a under internal/oracle/testdata/program_region,
plain/sanitized/JavaScript/Node ownership controls, and the counted/member mapper
address invariant with its embedded reuse mutant. Those fixtures are unchanged.
The prototype's reflective extra-member/missing-member proof was missing as a Go
test, so it is ported to two parallel top-level tests, using the landing path and
its existing lowering/check helpers. Only test IR is altered; production lowering
is never post-processed. A fixed two-member baseline makes the independent census
explicit. A missing member is counted in this landing, not a claim about the
prototype's graph-region fallback.

The prototype cycles_decision_test.go drops observable backlinks to check Node
semantics; those are not the membership ruling faults. Do not duplicate them as
membership tests. The landing already carries the reusable storage mutant from
program_region_map_storage_test.go and the early-release runtime fault in
TestProgramRegionCoreMutants/member-release-frees. Both were actually run.

The mapper leak helper runs a counted build only on macOS; Linux previously had
only LeakSanitizer in that test. The mapper helper now separately runs a counted
binary on every platform, compares Node behavior after removing only the anchored
count-instrumentation line, and checks allocations = frees + members. This is the
only change to an existing test. No census, status or count-table expectation is
rewritten.

## Independent fault runs

The compiler's run-mutants.py first three fault definitions were run unchanged,
except directing evidence into scratch and bounding every go test by both an
external 90s timeout and -timeout=90s. Its temporary source changes are restored
in finally. Every faulty program compiles; no compiler warning is accepted as a
catch. The corresponding patches and test output remain in the scratch directory.

| Fault | Catcher and observation |
| --- | --- |
| Compiler marks CountedLeaf as a member | TestProgramRegionOwnership fails: regions 3, want 2. Node and sanitizers were clean before the independent count assertion. |
| Compiler omits cyclic root membership | TestProgramRegionOwnership fails: regions 1, want 2. Node and sanitizers were clean before the independent count assertion. |
| Compiler admits member mapper reuse | TestProgramRegionMemberMapperStorage fails: exit 70, Program member array storage was reused, versus successful Node. |
| Prototype extra-member IR plant | TestProgramRegionExtraMemberMutant verifies the wrong census is 3; semantic output, sanitizer and leaks remain clean. |
| Prototype missing-member IR plant | TestProgramRegionMissingMemberMutant verifies the wrong census is 1; semantic output, sanitizer and leaks remain clean. |
| Runtime release frees a marked member | Existing TestProgramRegionCoreMutants/member-release-frees compiles its changed runtime and catches heap-use-after-free with ASan. |
| Embedded mapper guard forced true | Existing TestProgramRegionMemberMapperStorage catches actual source/output address equality, not a planner prediction. |

Final ownership count columns are allocations/frees/retains/releases/peak/members:

| Source/control | Counts | Balanced |
| --- | --- | --- |
| program_region_ownership.a | 21/19/10/25/12/2 | 21 = 19 + 2 |
| extra leaf member mutant | 21/18/10/26/12/3 | 21 = 18 + 3 |
| missing root member mutant | 21/20/10/24/12/1 | 21 = 20 + 1 |
| mapper, ordinary counted storage | 11/11/5/12/8/0 | 11 = 11 + 0 |
| mapper, member storage | 12/7/4/15/10/5 | 12 = 7 + 5 |

Both existing Program fixture rows match these results. They were not regenerated
or edited. The count-preserving fault cases demonstrate why Node agreement alone
cannot establish membership: each wrong allocation policy retains correct output.

## K60 decision: counted

Record: fcb7451a8239723fd903ae5998f54d2a86715e1f:docs/stage3-tsc-cycles.md, K60:
`ProjectReference[]`, element ProjectReference, witness
`view SourceFile -> SourceFile.amdDependencies -> readonly AmdDependency[] ~ ProjectReference[]`.
The original record is category (c), inference unproved. The landing membership.csv
selects it; programCompilerContainerReading in program_region_census_test.go:328
records counted. Both are left unchanged.

Source observations at TypeScript pin 050880ce59e30b356b686bd3144efe24f875ebc8:

- types.ts:7367-7376: ProjectReference holds path, originalPath, prepend and circular.
  Its only values are strings/booleans, not Program, SourceFile or another reference.
- commandLineParser.ts:3198-3212: getProjectReferences allocates the array and fresh
  scalar records from config, normalizing paths. The circular boolean is a flag;
  it is not a pointer to a circular dependency graph.
- program.ts:1517 receives that array; program.ts:2672-2674 returns the same array
  from the Program getter. ResolvedProjectReference is a different type, carrying
  sourceFile/commandLine/references at types.ts:4957-4960 and program.ts:3999-4069.

Node on TypeScript 6.0.3 actually parses a circular-flag config, observes the
normalized scalar records, verifies the Program getter preserves the input array
identity and recursively rejects any owning cycle in those entries. This says
nothing about when V8 chooses to collect unreachable memory.

Lifetime inference: the caller/parsed command line and Program keep counted
ownership of this array while needed. Keeping it strongly as a counted child of
Program is sufficient; the scalar records have no graph backlink that requires a
Program header. When the last owner goes, ordinary counts can release them. The
SourceFile/AmdDependency structural-view witness is not a source store creating
an owning cycle through these config objects. Counted is correct for these sites;
the selector's current member result is conservative over-inclusion and retention
cost, not observed premature release.

## K144 decision: counted

Record: the same independent census, K144:
`IncrementalBuildInfoFilePendingEmit[]`, element IncrementalBuildInfoFilePendingEmit,
witness `[fileId: IncrementalBuildInfoFileId] -> [fileId: IncrementalBuildInfoFileId] ~ IncrementalBuildInfoFilePendingEmit[]`.
The original record is category (c), inference unproved. The landing membership.csv
selects it; the independent compiler reading expects counted. Neither changes.

Source observations:

- builder.ts:1093 defines a number or one-/two-number tuple, not compiler graph nodes.
- builder.ts:1339-1355 builds fresh pending-emit serialization storage from the
  live builder's Map<Path, BuilderFileEmit>, encoding a numeric ID and emit kind.
- builder.ts:1360-1377 attaches the array to the returned build-info DTO.
- builder.ts:2239-2242 reads only numeric emit flags; builder.ts:2309-2318 restores
  serialized entries into a new map. These data do not hold the live builder.

Node exercises actual getBuildInfo on an incremental builder for full, declaration
and JavaScript pending emits. It observes [1], [[1]] and [[1,1]], checks that every
stored payload is numeric, rejects owning cycles, verifies exact JSON round trips
and observes a fresh array with identical values on a second getBuildInfo call.

Lifetime inference: this is transient scalar serialization output, owned by the
returned build-info object until its consumer is done. It is not the live AST's
EmitNode metadata referred to by the Program-graph rule. Ordinary counted ownership
can keep an externally retained DTO alive and free a discarded serialization copy;
there is no owning field cycle to break here. The tuple/array structural-view
witness does not establish a live graph edge. Counted is correct for these sites;
current inference safely over-includes them and can retain each temporary copy
until CLI exit. This assessment is recorded here for the compiler; no selector or
expectation is changed by runtime.

## Reproduction and limits

Setup passed in 46.995s: Go 0.022, Node 0.018, submodules 0.066, markdown 0.067,
clang 0.161, build 46.739, warm cache 46.963. nproc=5, cgroup quota=4 CPUs.
Linux, Go 1.27.1, Node 24.19.0, clang 20.1.8. TypeScript 6.0.3 source is pinned
at 050880ce; the independent inventory is pinned at fcb7451a. Generate diagnostics
with the pinned scripts/processDiagnosticMessages.mjs before the census.

The initial ownership attempt failed before building because the workspace was
full. The task's scratch checkout and generated runtime cache were moved to /tmp;
the previous generated cache was preserved via its original-path symlink. Final
runs use XDG_CACHE_HOME=/tmp/step06d-cache and the existing Go cache. This is an
environment repair, not a test or runtime result. The first attempt to compare the
counted mapper's raw stderr to Node also caught its instrumentation line; the final
comparison strips only the exact anchored count line and independently checks it.

Every test invocation used timeout --kill-after=2s 90s and -timeout=90s. No test
reached that limit. Final oracle controls use ADAMIC_GATE_UNCACHED=1, GOMAXPROCS=4:

| Test | Seconds |
| --- | ---: |
| TestProgramRegionOwnership | 0.65 |
| TestProgramRegionExtraMemberMutant | 1.27 |
| TestProgramRegionMissingMemberMutant | 1.24 |
| TestProgramRegionCountedMapperStorage | 0.56 |
| TestProgramRegionMemberMapperStorage | 0.74 |
| TestProgramRegionCoreMutants/member-release-frees | 6.68 |
| TestProgramRegionCensusMembership | 3.54 |
| TestProgramRegionCensusPairwise | 8.88 |

Both selectors reproduce all 1,685 unchanged rows: 1,599 members, 86 counted,
zero unresolved container identities, 77 uninstantiated declarations waiting.
Indexed selection 1.795503s; pairwise selection 7.158020s. Both explicitly log the
K60 and K144 member-versus-counted disagreement. This is a declaration census,
not an assertion that original tsc compiles on this branch.

Common clang flags: -std=c11 -Wall -Wextra -Werror -Wcast-function-type-strict
-pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function
-Wno-unused-parameter -Wno-self-assign -ffp-contract=off
-fno-optimize-sibling-calls -pthread. Program mode adds -DADAMIC_PROGRAM_REGION.
Sanitized builds add -O1 -g -fsanitize=address,undefined
-fno-sanitize-recover=all. Counted builds add -O2 -DADAMIC_COUNT; plain uses -O2.
The runtime mutant is sanitized and counted. No performance improvement is claimed.

Commands: go test ./internal/oracle -run '^<individual test>$' -count=1 -timeout=90s -v;
go test ./internal/native -run '^TestProgramRegionCoreMutants$/^member-release-frees$'
-count=1 -timeout=90s -v; census tests similarly in ./internal/lower with
ADAMIC_PROGRAM_CENSUS_ROOT=/tmp/step06d-typescript; node
review/runtime-step06d-membership/node-records.cjs; go vet ./internal/oracle
./internal/native ./internal/lower. Setup builds ./...; formatting and diff checks
are clean. Each command's complete output is a log in
/workspace/scratch/step06d-membership. No full gate, WASI, watch/service lifetime,
parallel publication or graph-region fallback is asserted.
