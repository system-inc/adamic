Built independent lowering chain slice 4 on main alone, serving steps 16, 20 and 32 and the chain slicing work.
Member commits: 113d76a5, a018fc10, 64b77e64, 19ca1506, 704edd9c; support: fd6f0ef6, 162d3d21, 4f30d11d; final pushed SHA is in the handoff.
Green: 292 lower tests in 12 shards; own Node/backends/sanitizers; 1,814 admission witnesses; readers; counts 50.921s; lane checks 4.3s.
Every one of 46 mutants is caught by its named assertion, Node stdout or sanitizer; exact names and catchers are in mutants.md and mutants-summary.json.
Not covered: full gate, whole-package suites, full native tsc execution and opt-in compiler census; two exploratory review witnesses retain identical main/slice panics.

| Member | Source SHA | Status | Dependency evidence | Tasks closed by this slice proof |
| --- | --- | --- | --- | --- |
| generic-body-relations | 2cc109e5 | kept | Checker shim operations plus main's instantiate.go:20, class.go:783 and invariance.go:309; no outside-chain symbol. | #js89dcw |
| generics-scout-main | 48391dcf | kept | Fixture-only compiler coverage. step16BodyRefusals at internal/oracle/step16_generics_test.go:33 depends on generic-body-relations within this slice. | Step 16; no additional named task |
| iteration-main | 518eca83 | kept | IteratorMethod/IteratorField belong to this member at internal/ir/iteration.go:5 and :15; main supplies object/call/exception machinery. Generator outcome registration uses generators-main within this slice. | #p6afa8g, #mtm9tke |
| project-references-main | e2aa3750 | kept | Uses main's loader/checker and its own project helpers. Probe imports only internal/load; declared in cloud/fast-gate/compiler-dependencies.json:22. | #acxgkff, step 32 |
| generators-main | 6f0ebd15 | kept | Own GeneratorTypes/GeneratorFrame at internal/ir/generator.go:5 and :35; iterationLocal/memberFunction existed on main. Combined iterator dispatch uses iteration-main within this slice. | #qk8rztp, step 20 |

The slice contains each member's compiler, test and fixture net against the plan's refined fork point. Historical scout documentation and measurement tools are under source-generics-scout, where their six retained measurement mutants and four controls were rerun. Historical frozen census outputs are not a new census run. No cohere code was copied. The checker gitlink, go.mod and go.sum remain identical to main. Shared source files were reconciled by owning hunks; no outside-chain branch was merged.

Dependency observations: all five own patches applied to main. The composed implementation builds on main and its named fixtures run. The new IR fields are supplied by iteration and generator members in this slice; the generic checker identity/mapping helpers already exist on main. Scout's exact refusal expectations and iteration's generator outcome records are within-slice interactions. These are observations over the extracted implementation, not a claim that merge order proves independence.

The generator/iteration conflict retained dynamic custom-protocol dispatch and separate generator sent/yield storage. The generator plan retains its concrete yield checker type for destructuring. aea87960 contributes only the three ordinary indexed-read agreements. 91b0148f contributes only project-owned loading selection and mixed-file controls. 2ca18b1b contributes only the project probe dependency entry. Scout outcome pins come from the composed chain's own step-16 test. The 2b2b1095 maybeBind assertion is a checked-any interaction absent on main: the original witness already gets a generic-body refusal here, so the original member test is retained.

The reader guard now checks oracle tests as well as lower/native/flow/fresh. It rejected the new cached-next mutant operand read and eight older oracle call operands before explicit compiler ownership/reasons were added. readers-oracle-audit.log is its observed failing control; readers-owned.log is green. The guard's top-level runtime is 0.64s after cache warming. Every newly imported or touched top-level test has passing observations in test-seconds.json; the largest observed ordinary member leaf is TestGeneratorChecker at 22.68s. Counts is the existing larger aggregate, passing in 50.921s. Each lower shard is capped by timeout 90 and go test -timeout 85s.

Commands run from the repository root unless noted, with source /workspace/adamic-tools/env.sh, GOPROXY='https://proxy.golang.org|direct', and ADAMIC_GOCACHE_OFF=1 after the first setup/cache contention:

```text
export GOPROXY='https://proxy.golang.org|direct'; timeout 240 bash cloud/setup.sh
First attempt: exit 124 during Go dependency warming; no done line.
Retry: exit 0. Go 0.089s; Node 0.076s; markdown 0.236s; submodules 0.261s; clang 0.362s; shared cache off 0.368s; build 61.841s; warm 61.973s; done 61.999s.
nproc: 5; cpu.max: 400000 100000. Go 1.27.1, Node 24.19.0, clang 20.1.8.

timeout 600 python3 review/compiler/chain-slice-4/lower-shards.py
292 tests; 12 go test ./internal/lower -run <named 25-test regex> -count=1 -timeout 85s -v shards.
All green; maximum 19.256s. Shard 3 rerun after npm ci installed pinned Node declarations.

go test ./internal/lower ./internal/load ./internal/flow ./internal/fresh -run '^Test(GenericBody|Generator|AsyncGenerator|Project|Mixed)' -count=1 -timeout 85s -v
Member-analysis controls green except initial missing @types/node; repaired-controls.log reruns that loader leaf and the main-specific maybeBind leaf green.

go test ./internal/oracle -run '^Test(ElementAccess|GenericBody|Step16|Generator|Step20|IterationDispatch|LoweringChainMixed)' -count=1 -timeout 85s -v
PASS, 15.268s; oracle-members-final.log. Includes semantic IR mutants, ownership sanitizer mutants and generator fixtures.

/tmp/slice4/oracle.test -test.run '^TestNativeAgreesWithNode$/stage3/fixtures/(generics|iteration|iteration-dispatch)' -test.timeout 85s -test.v
/tmp/slice4/oracle.test -test.run '^TestNativeAgreesWithNode$/internal/oracle/testdata/generic_body' -test.timeout 85s -test.v
/tmp/slice4/oracle.test -test.run '^TestNativeAgreesWithNode$/stage3/project-references-source' -test.timeout 85s -test.v
All pass from internal/oracle with ADAMIC_GATE_UNCACHED=1. Logs: fixtures-stage3, fixtures-generic-body and fixtures-project.
An earlier combined slash filter selected zero fixtures; it is not credited. An earlier direct binary call used the wrong cwd; it is not credited.

go test ./internal/oracle -run '^TestClassWrongOutput|^TestNativeAgreesWithNode$/internal/oracle/testdata/class_wrong_output_refused' -count=1 -timeout 85s -v
PASS, 1.305s. The two newly admitted review copies also pass TestClassWrongOutput107 under a test-only path overlay, 0.667s.

PROJECT_LOADER_TSC=/workspace/adamic/stage3/api/node_modules/typescript/lib/tsc.js timeout 240 python3 stage3/project-references-source/verify.py
PROJECT_LOADER_TSC=/workspace/adamic/stage3/api/node_modules/typescript/lib/tsc.js timeout 240 python3 stage3/project-references-source/verify_types.py
Both pass: stock TypeScript 6.0.3/Node, reference traversal and ambient declaration union. project-stock.log and project-types-stock.log retain observations and intended mutant catches.

timeout 600 python3 review/compiler/chain-slice-4/admission.py
Load/Lower-only main overlay and slice binaries; 60-witness commands each capped at 85s. 1,814 witnesses, 55 deltas, every delta attributed in admission-delta.json.
No tracked main accepted witness loses admission. The new type-lie witness is accepted by main and refused here. Two existing exploratory review witnesses panic identically: fxspptb_iterators_derived_symbol.a and fxspptb_iterators_override_source.a, Unhandled case in Node.Text: *ast.ComputedPropertyName.
A first sweep stopped on that panic; the resumed harness records each panic as an observation and only resumes complete matching shards.

go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 85s -v
PASS, 0.648s. The newly expanded oracle reader audit's failing control is retained separately.

go test ./internal/oracle -run '^TestCountsAreRecorded$' -parallel 8 -count=1 -timeout 85s -args -update-counts
PASS, 50.921s. One earlier contended run timed out and wrote no table. All 1,066 rows are attributed; 55 new rows and four changed existing rows. No successful regeneration was repeated.

 git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
PASS: lane checks 4.3 s: gofmt and tools on 71 Go files, t.Parallel on 6 test packages; a-check 13 .a files; vet 6 packages.
A single-branch clone initially lacked the named remote refs; explicit refspec fetch fixed that. The first actual analyzer run found three missing t.Parallel annotations; all three were fixed, their focused tests passed, and the lane rerun is green.
```

Mutant commands are retained in generic-mutants.py, iteration-mutants.py, generator-mutants.py, loader-mutants.py, their overlay JSON, scout-measurement.json and mutants-summary.json. Each Go leaf uses a 60s or 85s test timeout and an outer runner hard limit. Go-overlay source evidence is .go.txt. The six mapper mutants historically retired by the scout are not reinstated or credited as checks on absent implementation. The retained array-result mutant passes its independent Node comparison check; the intentionally wrong snapshot and indexed-read-old-stop overlays fail their named assertions. No build failure is counted as a kill.

Counts attribution: regexp.a gains one retain/release pair and regexp_tree.ts gains four because generator-compatible iterator result reads own boxed fields. user_iterators.a moves 663/663/455/904/64/0 to 1050/1050/677/994/97/0; user_iterators_rest_tdz.a moves 5/0/5/3/5/0 to 11/3/7/3/8/0. Iteration runtime dispatch caches closures and reads represented result fields, and close lookup takes runtime ownership. Both changes are held by their source Node/backend/sanitizer oracles; every added row is assigned to its registering member in counts-attribution.md. The unchanged 1,007 main rows are explicitly attributed to main in the JSON.

No member was dropped. Integration still owns main; task closure in the table is the local completed unit, not an external task-system mutation or a claim that main already contains this branch. The five source member commits are separate; one extra commit completed omitted project fixture staging, and shared repair/count/reader evidence follows in support commits. No history was rewritten and no PR was opened.
