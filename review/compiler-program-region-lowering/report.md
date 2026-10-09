Built: allocation-time Program membership, indexed shape selection, adoption before publication and member reuse exclusion for roadmap step 06.
Commits: compiler/program-region-lowering on runtime/step06-core a3f28d97; the delivered SHA accompanies the push.
Commands: focused lower/oracle controls, census selectors, build, internal vet, changed .a checks, mutants and Linux counts; results are recorded beside this report.
Mutants: acyclic leaf, cyclic-member removal, member reuse, field-only adoption size, shape-key collision and entry-pair publication are caught by independent tests.
Uncovered: canonical cache adoption (runtime 06c, #gpscteb), async/task boundaries and runtime-owned entry pairs remain explicit refusals; the original TypeScript compiler was censused, not lowered.

Dependency: the requested runtime/step06-core a3f28d97 on runtime's 63e0056f stack. No runtime files changed. No prototype runtime or graph-region fallback was merged.

Lowering discovers concrete checker identities and capture descriptors, then constructs executable IR with ProgramRegion on allocation creation. Both passes share the checker. A descriptor identity check refuses inconsistent replay. No reflective deep copy or post-pass allocation rewrite is used; metadata replay is limited to classes, functions and capture cells. Flag off uses one lowering and collects no Program candidates.

The shape index groups all/required property names and filters impossible assignment pairs. The checker still decides surviving relations. Property names are quoted in cache keys. Pairwise and indexed selectors produce byte-identical membership.csv: 1,685 records, 1,599 selected, 86 counted, zero unresolved container identities. Seventy-seven generic declarations remain counted until concrete allocation-site types are known. Selection measured 7.415796 s pairwise and 1.785695 s indexed; complete leaves measured 8.99 s and 3.42 s on the 4-CPU quota. Source: TypeScript 6.0.3 050880ce, independent inventory fcb7451a, 78 compiler source files.

All fourteen container identities resolve. K131's alpha-renamed generic value is not invented: its concrete Node key proves membership. K60 and K144 remain selected by conservative structural SCC selection; this costs retention until program exit. The earlier compiler reading is kept as a disagreement, not overwritten.

| ID | Container | Membership | Reason |
| --- | --- | --- | --- |
| K24 | `(Declaration | NodeWithTypeArguments | ArrayTypeNode | TupleTypeNode)[]` | in region | Concrete selected element, key or value. |
| K60 | `ProjectReference[]` | in region | Conservative structural SCC over-inclusion; earlier compiler reading said counted. |
| K67 | `{ name: __String; oldSymbol: Symbol; }[]` | in region | Concrete selected element, key or value. |
| K85 | `FilePreprocessingDiagnostics[]` | in region | Concrete selected element, key or value. |
| K86 | `LazyConfigDiagnostic[]` | in region | Concrete selected element, key or value. |
| K90 | `(TransformerFactory<SourceFile> | CustomTransformerFactory)[]` | counted | No selected element type. |
| K91 | `(CustomTransformerFactory | TransformerFactory<SourceFile | Bundle>)[]` | counted | No selected element type. |
| K109 | `(ImportEqualsDeclaration | ImportDeclaration | ExportDeclaration)[]` | in region | Concrete selected element, key or value. |
| K116 | `Map<Path, ResolvedRefAndSource>` | in region | Concrete selected element, key or value. |
| K117 | `Map<Path, ResolvedRefAndOutputDts>` | in region | Concrete selected element, key or value. |
| K118 | `Map<Path, Diagnostic[]>` | in region | Concrete selected element, key or value. |
| K130 | `Map<Node, TEntry>` | in region | Concrete selected element, key or value. |
| K131 | `Map<Node, TPrivateEntry>` | in region | Concrete selected element, key or value. |
| K144 | `IncrementalBuildInfoFilePendingEmit[]` | in region | Conservative structural SCC over-inclusion; earlier compiler reading said counted. |

Objects adopt using complete adamic_object_size, including optional presence, readiness and type tails. Class adoption precedes constructor execution, and object adoption precedes field initialization. Fresh cells adopt empty storage before holding children; environments adopt before constructing interior cell pointers; ordinary closures adopt before storing captured cells. Program values cannot be statement-arena candidates or reused output storage. The runtime uniqueness guard also excludes a member input from in-place reuse.

The six mirrors, million.a, optional storage, constructor, ownership, captured cells and mapper controls agree with Node in native and JavaScript. Native controls use plain builds, ASan/UBSan, leak checks and counters. Million reports 1,000,003 allocations, 3 frees and 1,000,000 Program members. Captured closure and cell report two allocations and two members. Mapper controls assert actual storage addresses: counted storage reuses, member storage allocates fresh. Strong-cycle fixtures explicitly retain the flag-off adamic/cycle-capable refusal.

The core runtime removed cache.index, while two emitter sites still referenced it. Those sites now use the existing adamic_slot_index helper. Existing flag-off counts are checked byte for byte against a3f28d97, excluding only eleven new explicit opt-in Program rows; count-preservation.log records the comparison. No stage 3 status records change.

Canonical nested closures are refused when selected: adoption must precede their cache insertion, owned by runtime 06c. Map/Set entry-pair materialization is conservatively refused in a program with members, since its runtime allocator publishes child pairs before compiler adoption could occur. Parallel work with any selected members is refused conservatively; async Program lowering is refused. Detached/factory allocations are not separately proven by this unit; membership remains the checker type and owning-field proof. These are boundaries, not passing pending acceptance claims.

Exact commands and evidence:

- GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh: passed on repeat, done 24.800 s. Timing lines: go 0.023, node 0.024, submodules 0.064, markdown 0.076, clang 0.185, build 24.668, deferred test binaries 24.771, warmed cache 24.772 s. nproc=5; cpu.max=400000 100000 gives four CPUs. The first replacement-workspace setup saw the partially reconstructed selector before its shape-index file existed; setup-first.log records the undefined-symbol failure. Setup was rerun after construction completed.
- go build ./cmd/adamic: build.log, exit 0.
- go vet ./internal/...: vet.log, exit 0.
- go test ./internal/lower -run '^TestProgramRegion' -count=1 -v: lower.log; census leaves run separately with ADAMIC_PROGRAM_CENSUS_ROOT=/tmp/adamic-program-region-typescript.
- go test ./internal/oracle -run '^TestProgramRegion' -count=1 -v: oracle.log.
- go test ./internal/lower -run '^TestProgramRegionCensusPairwise$' -count=1 -v: census-pairwise.log.
- go test ./internal/lower -run '^TestProgramRegionCensusMembership$' -count=1 -v: census-indexed.log.
- go test ./internal/lower -run '^TestProgramRegionEntryPairsRefused$' -count=1 -v: entry-pair.log.
- go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts: counts.log, passed in 122.290 s. This is the existing unchanged count recorder, not a newly added test leaf.
- ./adamic c on eleven changed .a files: four check flag off, seven expected cycle-capable refusals; flag-on oracle controls pass for all eleven. a-check.log records each.
- python3 review/compiler-program-region-lowering/run-mutants.py: mutants.log plus each .patch and failing-test log. All sources restored in finally. No compilable Go evidence exists under review/.
- Integration's exact committed-tree lane-check command: lane-checks.log. Two inherited generated corpus files needed only a redundant trailing blank line removed by gofmt, recorded in inherited-formatting.patch. No test leaf changed in those files.

| Mutant | Independent test and failure |
| --- | --- |
| Mark acyclic leaf | TestProgramRegionOwnership: regions 3, want 2; Node, sanitizers and leaks remain clean before the count assertion. |
| Drop cyclic root | TestProgramRegionOwnership: regions 1, want 2; Node, sanitizers and leaks remain clean before the count assertion. |
| Admit member mapper reuse | TestProgramRegionMemberMapperStorage: exit 70, Program member array storage was reused, Node disagreement under ASan/UBSan. |
| Field-only adoption size | TestProgramRegionOptionalStorage: ASan heap-buffer-overflow and Node disagreement; the dedicated FieldOnlyAdoptionMutant test independently checks the same fault. |
| Unquoted shape key | TestProgramRegionShapeKeySeparators: distinct property shapes collide. |
| Remove entry-pair boundary | TestProgramRegionEntryPairsRefused: wanted refusal, got nil. |

Every added top-level test calls t.Parallel first. Observed seconds, including each test's own setup:

| Test | Seconds |
| --- | --- |
| TestProgramRegionCapturedCell | 0.96 |
| TestProgramRegionCensusMembership | 3.42 |
| TestProgramRegionCensusPairwise | 8.99 |
| TestProgramRegionConcreteContainers | 0.07 |
| TestProgramRegionConstructor | 0.76 |
| TestProgramRegionCountedMapperStorage | 0.34 |
| TestProgramRegionEntryPairsRefused | 0.03 |
| TestProgramRegionFieldOnlyAdoptionMutant | 0.58 |
| TestProgramRegionGraphParent | 0.75 |
| TestProgramRegionGraphRelations | 0.92 |
| TestProgramRegionGraphSymbols | 0.85 |
| TestProgramRegionMemberMapperStorage | 0.83 |
| TestProgramRegionMembership | 0.07 |
| TestProgramRegionMillion | 2.51 |
| TestProgramRegionOptionalStorage | 0.63 |
| TestProgramRegionOwnership | 0.85 |
| TestProgramRegionSCCOwningEdges | 0.00 |
| TestProgramRegionShapeKeySeparators | 0.00 |
| TestProgramRegionTaskBoundary | 0.07 |
| TestProgramRegionViewCycleIsNotOwnership | 0.00 |
| TestProgramRegionWeakParent | 0.85 |
| TestProgramRegionWeakRelations | 0.94 |
| TestProgramRegionWeakSymbols | 0.88 |

Protected-file hooks: internal/lower/lower.go lines 20-55 add explicit options, checker-sharing discovery and allocation construction; its async branch remains outside the Program contract; lines 114-131 apply descriptor metadata and carry discovery state. internal/native/emit.go, internal/native/native.go and internal/oracle/oracle_test.go are untouched. No runtime files changed.

The execution environment was replaced immediately before the first delivery push. Local commits 2a557e7d and 686f1c10 and their evidence were inaccessible and never pushed. The unit was reconstructed on the same requested base, and the checks and mutants above were rerun in the replacement workspace. No old local SHA is presented as a delivered commit. The conservative choice was to rebuild rather than publish an unverified reconstruction. No full package test or full gate was run.

The first counts run failed because pinned @types/node 25.3.3 was absent in the replacement workspace. npm ci --prefix stage3/api installed the lockfile dependencies; the repeated Linux count recorder passed, adding eleven rows and changing none. counts-first.log and node-types-setup.log retain this evidence.

Integration lane output: lane checks 5.5 s: gofmt and tools on 254 Go files, t.Parallel on 33 test packages; no t.Parallel analyzer on this tree; vet 33 packages. All added top-level tests call t.Parallel first; this runtime base does not carry the analyzer executable.
