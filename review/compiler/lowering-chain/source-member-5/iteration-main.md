Rebuilt inherited readonly-array iteration from adfb8d3d on main 7a10c877 for step 20 (#yvgst44).
One net-change commit on compiler/iteration-main; no old source base or extra dependencies carried.
Build, vet, changed-file a-check, stage1 probes, stage3 fixtures, iteration oracles and regenerated counts pass.
All 17 source-branch mutants fail at their intended checks.
General Iterable dispatch stays pending (#85genag); generic tsc signatures still stop before its loops.

## Meaning and conflicts

Actual arrays seen through interfaces inheriting library ReadonlyArray now iterate, including generic inheritance and branded array views. Live backing arrays, weak slots and nested mutable invariance are preserved. Writes to built-in Symbol.iterator, aliases and prototype methods are refused. No general structural dispatch was added.

Applied only the net diff dcdbb9098f77f30ad41790c56df1bd63ad462b63..adfb8d3ddde6effe850c53e4638812ab3f43557c, excluding counts.md, regenerated locally. Source census/replay archives remain historical evidence, not measurements of rebuilt main. Historical claims of a ruled general protocol are superseded for this unit by the user's instruction to leave #85genag pending.

| Conflict | Resolution |
| --- | --- |
| internal/lower/expression.go, sameKeeping | Kept main's sameCallable closure ABI, library callback exception and censusNeverRest checks. Added inherited readonly-array compatibility and prevented generic container matching from bypassing view checks. |
| internal/lower/object.go, elementType | Kept main's nested empty-array contextual and proven-element handling. Normalize inherited ReadonlyArray contexts after the empty fallback and before contextual storage. |
| internal/lower/refusals.go | Retained both indexed-write and descriptor enumeration refusals; added intrinsic iterator write refusal before library refusal handling. |

Six intentionally refused .a fixtures received a-check headers without program or Node-output changes.

## Validation

Environment /workspace/adamic-tools/env.sh sourced. Every test wrote a log. Fresh Node runs used ADAMIC_GATE_UNCACHED=1.

| Command | Result |
| --- | --- |
| go build ./... | pass |
| go vet ./internal/... | pass |
| shared Gate.aCheck on 17 changed .a files against origin/main | pass: 11 checked, 6 expected refusals |
| go test ./stage1/... -run 'Gap\|Gaps\|Probes' -count=1 -timeout 15m | pass |
| go test ./stage3/fixtures -count=1 -timeout 15m | pass |
| go test ./stage3/fixtures -run '^TestFixtures$/host/24' -count=1 -v | pass: fresh Node, native, stage0 |
| go test ./internal/oracle -run '^TestNativeAgreesWithNode$/stage3/fixtures/iteration/\|^TestStep20' -count=1 -timeout 10m -v | pass, rerun after headers |
| go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 15m -args -update-counts | pass |

Six accepted witnesses: collections.a, strings.a, object_iteration.a, array_view_stress.a, test262_array_views.a, array_view_weak.a. Each agrees with fresh Node in both backends, native release, ASan/UBSan and leak checking. All 17 exact outcome probes pass, plus inline intrinsic guard probes.

Five retain NotYet: accessor_binding, iterable_view, object_binding, object_binding_stress, user_forwarding. Six retain refusal: generator, delegated_generator, intrinsic_iterator_write, intrinsic_iterator_alias, intrinsic_iterator_prototype, array_view_variance.

Host/24 already records Compiles on main. Fresh Node and native agree with recorded stdout `false\nfalse\ntrue\nfalse\naBc_09.TS\n`, empty stderr and exit 0. The source's old red does not reproduce. No stage3 status records changed, all Node observation bytes remain unchanged, and no Compiles fixture regressed.

## Counts

Only six new rows added; no existing row moved. Each is newly registered accepted iteration fixture, with generated ownership counts and no manual edits.

| Fixture | Allocations | Frees | Retains | Releases | Peak | In regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| collections.a | 35 | 35 | 24 | 47 | 11 | 0 |
| strings.a | 20 | 20 | 14 | 27 | 9 | 0 |
| object_iteration.a | 5 | 5 | 1 | 6 | 3 | 0 |
| array_view_stress.a | 34 | 34 | 50 | 72 | 16 | 0 |
| test262_array_views.a | 11 | 11 | 25 | 38 | 11 | 0 |
| array_view_weak.a | 5 | 5 | 8 | 13 | 4 | 0 |

## Mutants

Commands: python3 stage3/fixtures/iteration/check_array_view_mutants.py /tmp/iteration-main/mutants; python3 stage3/fixtures/iteration/check_snapshot_mutant.py. All fail intentionally at the expected checks:

| Mutant | Catcher |
| --- | --- |
| snapshot-view | stdout mismatch against Node for live array iteration |
| strong-weak-slot | AddressSanitizer: SEGV for weak slots |
| allow-intrinsic-write | missing intrinsic iterator refusal |
| widen-mutable-element | missing nested mutable slot refusal |
| wrong exact baseline reason | exact outcome contract catches invented object-iteration reason |

Three positive historical evidence audits pass and their twelve mutants fail. Commands: census.py docs/step-20-iteration/base-census.jsonl.gz /tmp/iteration-main/audit-census.json --compiler dcdbb9098f77f30ad41790c56df1bd63ad462b63; hidden_audit.py; audit_retirements.py docs/step-20-iteration/array-view-replays.json.gz. Each ran positive and once per listed --mutant.

| Mutant | Catcher |
| --- | --- |
| census-roots | root deduplication changed |
| census-units | unit deduplication changed |
| census-witnesses | witness selection changed |
| census-coverage | resolved source coverage changed |
| hidden-bytes | historical byte credit changed |
| hidden-roots | diagnostic root deduplication changed |
| hidden-witnesses | witness provenance changed |
| hidden-boundaries | boundary extraction changed |
| retirement-coverage | retirement coverage changed |
| retirement-retired | retirement classification changed |
| retirement-witness | retirement witness changed |
| retirement-declaration | declaration snapshot changed |

## Next stops and ruling

Fresh no-output selected-unit replay on pinned tsc 050880ce59e30b356b686bd3144efe24f875ebc8 with current main adaptations measures:

| Source | First stop |
| --- | --- |
| src/compiler/core.ts:82:17, firstDefinedIterator | NotYet: a function returning U or undefined |
| src/compiler/core.ts:93:17, reduceLeftIterator | NotYet: a function returning U |
| src/compiler/core.ts:1337:1, arrayFrom | Refused: overload result U[] cannot be served by implementation result (T or U)[] |

These observations are on a checker-rejected entry-root program, not a whole-tsc compilation claim. Requested loop signatures at 83:25, 97:29 and 1343:25 do not reproduce because signature checks stop before the bodies. General Iterable remains NotYet in direct iterable_view.a:4:25.

The old no-output measurement overlay initially panicked on main's private IR argumentFacts cache. Only the scratch overlay was adjusted to leave that derived cache empty when copying ir.Program; the cache recomputes from copied public IR. No production code was changed for this tool repair. Original failure, corrected overlay and fresh measurements are archived.

Ruling proposal sent for #85genag: preserve [Symbol.iterator], cache next once, distinguish yield and completion payloads, implement ECMA-262 IteratorClose with existing ownership and exception rules. General dispatch remains pending. Generic return/overload work, object destructuring iteration, erased iterator origins, generators and whole-tsc compilation are outside this rebuild. Original unadapted test262 files were not rerun; the adapted array-view witness was held to Node in both backends.

## Toolchain and evidence

GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh succeeded. nproc=5; Go 1.27.1, Node 24.19.0, clang 20.1.8. Setup timings: Node ready 0.029s; Go ready 0.029s; markdown step 0.009s and ready 0.088s; submodules 0.107s; clang 0.217s; build 46.678s; test binaries deferred 46.874s; cache warm 46.876s; done 46.910s.

An initial gofmt invocation before sourcing could not find gofmt; the sourced invocation succeeded before tests. Disk filled while writing the final report; only Go's rebuildable cache was cleared after validation. Logs remained safe in /tmp.

[Fresh logs, mutants and measurements](iteration-main-evidence/logs.tar.gz). Source historical logs remain under step-20-iteration/evidence.
