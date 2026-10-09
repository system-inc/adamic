Built: the Program container selector now rejects acyclic payloads and phantom primitive brands, advancing step 06 and preserving step 24 point 4.
Commits: delivery branch compiler/region-selector-acyclic on 52252d5a8ab0e0f977b98b7826bde79420643a99; delivery SHA is in the final response.
Commands: both censuses, Program lower/oracle/native controls, ownership, leaks, TestCallTargetReaders and changed-package vet pass; counts and lane results follow below.
Mutants: re-inclusion, payload over-inclusion, cyclic-root removal, phantom-brand ownership and ignored method captures are caught by independent assertions.
Uncovered: original tsc lowering, full package/full gate runs, non-Linux hosts and the inherited host/optional-field pending lowering probes.

The requested exact runtime base takes precedence over the generic main-start instruction.
No other worker's unlanded changes were merged. Runtime's 23a4f2f0 report was read as evidence.
The selector keeps the structural SCC result for ordinary objects. A selected container payload
must additionally have an owning path to recursive storage, or opaque generic, unknown or
callable storage that cannot be proven acyclic. Primitive intersections remain scalar storage,
including numeric IDs whose phantom brand contains any. Interface methods and accessors
remain conservative because they can hide captures. This is a memory-retention change.

The pinned declaration census contains 1,685 records. Both indexed and pairwise selection
produce byte-identical expectations: 1,590 members, 95 counted, zero unresolved container
identities and 77 waiting uninstantiated declarations. Previously it was 1,599/86.
Exactly these nine containers leave; every D record is unchanged:

| Record | Type now counted |
| --- | --- |
| K7 | `FileReference[]` |
| K41 | `Map<Path, SharedExtendedConfigFileWatcher<T>>` |
| K56 | `CommentDirective[]` |
| K58 | `AmdDependency[]` |
| K59 | `PluginImport[]` |
| K60 | `ProjectReference[]` |
| K111 | `Map<Path, Set<ResolutionWithFailedLookupLocations>>` |
| K112 | `Set<ResolutionWithFailedLookupLocations>` |
| K144 | `IncrementalBuildInfoFilePendingEmit[]` |

K7, K56, K58, K59 and K60 have scalar record fields. K111 and K112 transitively own
string arrays, boolean flags and Set<Path>, where Path is a branded string. K144 is
numeric serialization storage, including branded IDs and numeric tuples. K41 is a
special declaration-only case: its branded Path key no longer supplies false membership;
SharedExtendedConfigFileWatcher<T> is uninstantiated, so it remains pending concrete
allocation-site evidence. This does not claim its eventual instantiation is acyclic.
Concrete Node keys in K130/K131 remain admitted. Opaque watcher methods preserve the
other watcher containers rather than treating their hidden captures as absent.

Node and NodeArray<Node> remain selected at all four allocation sites in the tsc
projection. Its oracle counts stay 11/4/19/26/11/7, and ownership stays 21/19/10/25/12/2
(allocations/frees/retains/releases/peak/members). Linux ASan/UBSan, LeakSanitizer,
Node, JavaScript backend, plain native and counted controls pass. The counted mapper
remains 11/11/5/12/8/0; member storage remains 12/7/4/15/10/5.

Each mutation is restored in finally; patches and full failing output are retained.
No warning, build error or timeout counts as a kill:

| Mutant | Catcher |
| --- | --- |
| Re-include all nine containers | TestProgramRegionCensusMembership: K60/K144 independent readings fail and membership.csv differs. |
| Remove payload qualification | TestProgramRegionAcyclicContainerPayloads: member=true, want false. |
| Drop cyclic root allocation, original ruling fault | TestProgramRegionOwnership: regions 1, want 2; semantic/sanitizer/leak controls run before this count assertion. |
| Treat numeric phantom brand as owned | Payload control: pending member=true, want false. |
| Ignore interface method captures | Payload control: methods member=false, want true. |

The inherited runtime faults also pass their catching tests: member-release-frees
(heap-use-after-free), weak-not-forgotten, cell-weak-not-forgotten and short-adoption.
The existing extra/missing-member IR controls remain clean under Node/sanitizers/leaks
and detect their changed independent two-member census.

Every new or touched leaf stays below 60 seconds with setup included: the new payload
control is 0.11s; restored indexed census 8.03s; pairwise 15.84s. The unchanged Node
projection is 0.18s. Full focused runs report lower 18.251s, oracle 30.154s; native
63.429s and counts 198.562s. TestCallTargetReaders passed in 40.44s.

Setup's initial bounded cache-warming attempt did not reach done. The early test
commands reached external exit 124 while building dependencies, and were rerun.
Setup retry passed with timing lines: go ready 0.051s, node ready 0.080s,
submodules ready 0.136s, markdown dependencies ready 0.158s, clang ready 0.353s,
go build ready 81.638s, test binaries deferred 81.832s, build cache warm 81.835s,
done 81.879s. nproc=5, cpu.max=400000 100000. Environment:
/workspace/adamic-tools/env.sh; Go 1.27.1, Node 24.19.0, clang 20.1.8.
The locked stage3/api @types/node dependency was installed with npm ci after the
first host check found it missing. The host non-null assertion and optional delete
probes still skip on this base; no passing coverage is claimed for them.

Exact commands, all output redirected to log files, from the repository root:

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout 300 bash cloud/setup.sh
timeout 600 bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
export GOMAXPROCS=4
# Census source: git clone --depth 1 --branch v6.0.3, pinned 050880ce59e30b356b686bd3144efe24f875ebc8.
# Diagnostics: node <source>/scripts/processDiagnosticMessages.mjs <source>/src/compiler/diagnosticMessages.json
export ADAMIC_PROGRAM_CENSUS_ROOT=/workspace/scratch/selector-typescript
timeout 90 go test ./internal/lower -run '^TestProgramRegionCensusMembership$' -count=1 -timeout 90s -v -args -update-program-membership
# Each package separately, with ADAMIC_GATE_UNCACHED=1:
timeout --kill-after=2s 90 go test ./internal/lower -run '^TestProgramRegion' -count=1 -timeout 90s -v
timeout --kill-after=2s 90 go test ./internal/oracle -run '^TestProgramRegion' -count=1 -timeout 90s -v
timeout --kill-after=2s 90 go test ./internal/native -run '^TestProgramRegion' -count=1 -timeout 90s -v
timeout --kill-after=2s 90 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s -v
timeout --kill-after=2s 450 python3 review/compiler/region-selector-acyclic/run-mutants.py
timeout --kill-after=2s 90 go test ./internal/lower -run '^TestProgramRegion(AcyclicContainerPayloads|CensusMembership|CensusPairwise|TscNodeMemberSites)$' -count=1 -timeout 90s -v
timeout --kill-after=2s 600 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 600s -v
timeout --kill-after=2s 90 go vet ./internal/lower
gofmt -l internal/lower/program_region_selection.go internal/lower/program_region_selection_test.go internal/lower/program_region_census_test.go
git diff --check
```

Recorded counts pass in 198.562s (native cache hits=1, misses=1213).
internal/oracle/counts.md is byte-for-byte identical to the pinned base. No new
oracle fixture was added, so there are zero added or changed rows to attribute
and no -update-counts rewrite is needed. The new witness is a checker-only temp
.a file inside a focused Go test. Changed-package vet, formatting and diff checks
pass. Full evidence is stored as deterministic .log.gz files beside this report.

