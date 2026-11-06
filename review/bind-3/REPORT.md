# Partial bind lowering, #kpp0f8s

Fixed-arity declared-function binds and the ruled stable-identifier method heads are implemented within the limits below. A full green gate is not claimed.

Base: `5ae1e240168601060a873ae0204bf90828398bf3` (`origin/main`). Branch: `apple/bind-3`.
Implementation head: `52f91e326281cf01a2564a9baf0e53775146a9e9`. The final review/evidence commit follows this implementation commit.
The first three commits (form one, form two, then mutants) were pushed in the requested order before step-four broad validation. Four follow-up implementation fixes were pushed as audits and broad checks found gaps. No main push, force push, Loom submission or stage-three change.

Step zero was skipped under the final instruction. The supplied disposition is 46 `maybeBind(X, X.m)` callers: ten store outside X's reach; 36 close loops already closed by arrows. No new 46-site survey or independent table is claimed. The direct `program.ts:1950` object-literal shape is pinned.

[files.md](files.md) gives one line per changed source. [ir-switches.tsv](ir-switches.tsv) lists all 169 import-aware type switches found in cmd, internal and stage1, including aliased and unqualified cases. Stage1 has no type switches over this compiler IR. The switch audit distinguishes explicit bound-node handling from subset/generic conservative paths; that label alone is not a proof. The audit found and fixed suspension reordering and a missing split origin for the kept own closure.

## Implementation and limits

Declared fixed-arity function values with one call signature bind through a captured-cell factory, a fresh `ir.BoundClosure`, and a synthesized wrapper. The target, thisArg and fixed values are ordinary factory arguments, evaluated left to right before allocation and owned by captured cells. Fixed values precede call arguments and are borrowed on repeated calls. Arrows ignore the retained receiver. Default parameters are preserved.

`ir.BoundMethod` admits the direct `X.m.bind(X)` and `X.m?.bind(X)` heads, where X is one const identifier or an unassigned parameter. Native calls `adamic_object_callee` once before fixed arguments, retains the own closure arm or keeps the class method code entry, and stores the receiver. Optional empty lookup returns undefined before evaluating fixed arguments. Native also normalizes the shared null sentinel used by a function/null/undefined field, preserving the loosely nullish test. The nullable-union row covers null, undefined, skipped effects, a present value and replacement after binding. JavaScript reads the member once and uses a receiver-aware adapter bound with JavaScript's bind. Replacement, missing-then-added, escape and class-method rows agree with Node. A later audit found that static constructor values require the same separate method table as ordinary dispatch. The pre-fix `typeof Host` parameter probe lowered but panicked in the backend; the lookup now snapshots from that table too and its running pin passes JavaScript, native leak checks and wasi (15.795 seconds, `/tmp/bind-static-fixed.log`).

The runtime uses the ordinary closure heap kind. `adamic_closure` gains a method-entry field shared by all closures. The constructor and call entry live in `runtime/closure.c`; the common `free_children` releases the owning environment cells. There is no self-pointer or side table. The fixed target, receiver and fixed arguments are the closure's links through its registered Environment. `BoundMethod` does not allocate; the inner `BoundClosure` goes through both `graphAllocation` and `programAllocation`, receives graph type ids and uses GraphClosure adoption. Code entries are not retained references. Graph adoption can memcpy the struct. Six independent arrow/bound cycle witnesses compile, match Node and pass leak checking; dropping only bound links leaks the build-cache bound witness.

Explicit partial scope: multiple call signatures, rest signatures, possible actual argument-count readers, narrowing signatures and suspending method-bind operands remain named NotYet. A function-typed property with a different thisArg is only admitted when its own plain-literal closure arm is proved; arbitrary class-backed property views are not widened. Getter names are conservatively excluded across the source program. The rebinding stop follows direct binds and simple aliases; provenance through returns, fields or parameters is not tracked, so that named refusal is incomplete. Index-signature record mutation and writes to absent optional physical fields retain main's existing stops. The checked-view witness stops earlier at main's unchecked-cast rule. A real `fs.realpathSync.native` binding probe also stopped at its existing library-value restriction; that exploratory pin and speculative backend change were removed. The adapter currently forwards declared fixed slots; an unbounded argument vector and exact actual-count forwarding are not proven. Overload selection, predicate checks and indirect bound provenance need further work. None of these gaps is claimed implemented.

No `signature()` or function-body this rule was widened. The required function `show(this: Host)` assigned to `() => string` still stops at `internal/lower/expression.go:768`, `this outside a method`.

## Final boundary corrections

Inherited library signatures stay outside the new head, after the old libraryMethod allowlist. The RegExp probe had printed `/x/` on Node and panicked in backend JavaScript; `inherited_library.a` now stops by name. An opaque Function view of an escaped bound value prints 1 on Node; the existing graph/dynamic boundary now reports the read's source location instead of panicking, held by `opaque_metadata.ts`. This only records graphDynamicSite in functionObservation; it does not admit Function binding or new this bodies. The scanning reference now calls checkBoundArguments in the same position as production. The formerly failing pipeline/inventory focus passes in 0.019 seconds.

## Stops held by pins

Reasons are asserted independently in `internal/oracle/function_bind_test.go`; refusal fixtures live only in `testdata/bind_refused/` and name Node's answer in comments.

| Fixture | Expected reason / implementation location |
|---|---|
| `length.a`, `typeof_length.a` | bound length metadata; function_bind.go:196 |
| `name.a` | bound name metadata; function_bind.go:198 |
| `rebind.a` | bind of a bound function; function_bind.go:200 |
| `new.a` | bound construction; function_bind.go:205 |
| `dynamic_length.a`, `dynamic_name.a`, `dynamic_bind.a` | dynamic function descriptors; unknown.go:164 |
| `this_widening.a` | this outside a method; expression.go:768 |
| `different_receiver.a` | method bind with a different receiver; function_bind.go:319 |
| `call_receiver.a` | call as method receiver; function_bind.go:307 |
| `getter_path.a` (Language's program verbatim), `path.a` (no this parameter) | receiver path; function_bind.go:309 |
| `getter.a` | getter-backed method value; function_bind.go:344 |
| `assigned_receiver.a` | assigned receiver identifier; function_bind.go:324 |
| `index_receiver.a` | indexed receiver; function_bind.go:293 |
| `detached.a`, `method_write.a` | existing unbound-method rule; refusals.go:260 |
| `optional_tail.a` | existing call-through-optional-chain stop; expression.go:1593 |
| `async_fixed.a` | suspension inside method binding operands; flow/async_normalize.go:158 |
| `checked_receiver.ts` | existing unchecked-cast refusal, not evidence of a new admitted view |
| `inherited_library.a` | bind of an inherited library function; function_bind.go:302 |
| `opaque_metadata.ts` | existing dynamic-shape/graph boundary; graph_types.go:491, location recorded by function_rulings.go:108 |

The arrow and bound cycle rows are running fixtures, each asserted on its own, under the corrected graph-region ruling. No new cycle refusal was introduced.

## Mutants

All ten mutants (the five requested and five additional boundary checks) were caught by semantic checks, not compiler warnings. Each ran only its focused test, then restored sources byte for byte and checked the diff. [mutants.json](mutants.json) preserves exact output and test names. The first five ran against a clean committed baseline; later additions used their staged fix as the restoration baseline. The final two are run by `/tmp/bind-final-boundary-mutants.py`, with results in `/tmp/bind-final-boundary-mutants.log`.

| Mutant | Named catcher | Observed failure |
|---|---|---|
| Fixed expressions run on every call | `TestFunctionBind` | JavaScript stdout differs: built1 became built2/built3; counter became 3 |
| Same receiver widened to paths | `TestMethodBindRefusals/path` | want receiver-path stop, got nil |
| Bound links removed from cycle/membership walk | `TestMethodBind/bind_cycle_cache_bound.a` | `ERROR: LeakSanitizer: detected memory leaks` |
| Optional nullish test removed | `TestMethodBind/method_bind.a` | native stdout differs; absent false and effects ran |
| Lookup deferred until call | `TestMethodBind/method_bind.a` | JavaScript printed second instead of first |
| Suspension safety stop removed | `TestMethodBindRefusals/async_fixed` | want suspension stop, got nil |
| Static method table lookup removed | `TestMethodBind/method_bind.a` | JavaScript exit codes differ; `TypeError: fn is not a function` |
| Inherited library stop removed | `TestFunctionBindRefusals/inherited_library` | want inherited-library stop, got nil |
| Opaque metadata source site removed | `TestFunctionBindRefusals/opaque_metadata` | diagnostic nil-pointer panic |
| Null sentinel normalization removed | `TestMethodBind/method_bind.a` | native null arm false, effects 1; Node true, effects 0 |

## Commands and evidence

Commands below are grouped by repeated invocation rather than every navigation/read. All test output went to files. Nothing was tested on an untouched base. Setup used GOPROXY's pipe fallback and the pinned tools environment. Full suites were started only after both forms and the requested mutants had been committed and pushed.

- `bash cloud/setup.sh`, `git fetch origin`, `git checkout -b apple/bind-3 origin/main`, `git rev-parse HEAD`: completed. The first setup cache warm overlapped creation of a Go source and failed; rerun succeeded. `bash cloud/setup.sh --wasi-sdk` installed the pinned SDK. No stage-three adaptation was changed.
- `go build ./...`, `go vet ./...`: passed for both implementation steps. Final rerun before the static backend correction passed too; `/tmp/bind-build-final.log`, `/tmp/bind-vet-final.log`. The post-correction build and vet also passed (`/tmp/bind-build-static.log`, `/tmp/bind-vet-static.log`).
- `go test ./internal/lower/`: first full run timed out at ten minutes; it reported three new inventory omissions (fixed and focused green) and a missing TypeScript source prerequisite. A superseded run was terminated after sources changed. Logs: `/tmp/bind-lower-form1.log`, `/tmp/bind-lower.log`.
- Focused lower refusal order, local-field, graph-allocation and view-read inventory tests: passed. Logs `/tmp/bind-lower-inventories.log`, `/tmp/bind-inventories-form2.log`.
- `ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^Test(Function|Method)Bind' -count=1 -timeout 15m`: passed all eight running fixtures and refusal tests. Native sanitizers/leak checks, release methods, JavaScript and wasi are exercised. The pre-final audit restored run passed in 14.577 seconds (`/tmp/bind-audit-restored.log`); all new pins, including the two final boundaries, passed in 21.712 seconds (`/tmp/bind-final-boundaries.log`). A final post-mutant restored run is recorded below. Earlier focused iterations and exploratory failures are in `/tmp/bind-{focus,refusals,wasi,all}-{form1,form2}.log`, `/tmp/bind-async-refusal.log`, `/tmp/bind-view-refusal.log`, `/tmp/bind-realpath-callable.log`.
- `python3 /tmp/bind-final-boundary-mutants.py`: both final boundary mutants caught, restored and diff checked. Static-table removal was separately caught by TestMethodBind/method_bind.a (`/tmp/bind-static-mutant.log`).
- `python3 internal/native/testdata/run-bind-mutants.py`: all six mutants caught and sources restored; `/tmp/bind-mutants-final.log`, `/tmp/adamic-bind-mutants/results.json`. The first five-only run is `/tmp/bind-mutants.log`.
- `go test ./internal/oracle -run 'TestCountsAreRecorded_[0-9]+/internal/oracle/testdata/(function_bind[.]a|method_bind[.]a|bind_cycle_.*)' -count=1 -timeout 15m -args -update-counts`: passed (46.888 seconds); exactly eight new runtime records, no old rows regenerated. After adding the static pin, only `method_bind.a` was remeasured and passed in 14.942 seconds (`/tmp/bind-counts-static.log`). `/tmp/bind-counts-update.log`.
- Pinned TypeScript prerequisite: external clone `/tmp/bind-typescript-source`, verified `050880ce59e30b356b686bd3144efe24f875ebc8`, exported as ADAMIC_TYPESCRIPT_SOURCE. `/tmp/bind-typescript-setup.log`.
- Product preparation: `/tmp/bind-run-products.py` selects 413 products and transitive dependencies from `cloud/fast-gate/product-order.json`, waves 353/43/17, initially two package jobs, finally three bounded package jobs, and `-parallel 2`. Logs `/tmp/bind-products/` and `/tmp/bind-products-final.log`. First attempt stopped because cmd/adamic requires a unit build cache, fixed with ADAMIC_BUILD_CACHE_DIR=/tmp/bind-build-cache. A later attempt was stopped before the compiler audit edit so no product could be built from changing sources. The stable run was also stopped before the static-method backend correction; its logs are `/tmp/bind-products-before-static.log` and `/tmp/bind-products-before-static/`. No unfinished broad run is claimed green. A later workspace restart interrupted active runs; before it, oracle producers reported `no space left on device` on the 8.8 GiB `/tmp` mount. Removing only superseded Go-export entries and abandoned scratch from this task's private cache reclaimed 2,145,921,346 bytes; the failed log is `/tmp/bind-products-space-failure.log`. Product preparation was retried with the completed cache retained.
- Full consumers use `ADAMIC_GATE_UNCACHED=1`, `-count=1 -timeout 30m`, and `-skip '^TestProduct_'` only after their prerequisite products pass. Native additionally uses `-parallel 2`. Final results and exact package names are recorded below. Individual consumer commands were `go test -count=1 -timeout 30m -skip '^TestProduct_' ./internal/flow/`, then the same flags for `./internal/ir/`, `./internal/split/`, `./internal/lower/`, with `-parallel 2` for `./internal/native/` and `./internal/oracle/`. Fresh and JavaScript ran together without product skips. Logs are `/tmp/bind-{flow-recovered,ir-final,split-final,lower-final,native-final,oracle-final}.log` and `/tmp/bind-fresh-javascript-static.log`.
- `go run ./internal/compilercensus/cmd/dependencies`: succeeded. Compared with the hand map in both directions: 100 derived packages vs 34 listed, +66 packages/-0, +488 dependency edges/-1 (`stage3/lane` to `stage3/lane/adaptedtree`). Exact comparison exits one because the maps differ; all 34 declared packages are covered, as the census test requires. [dependency-diff.json](dependency-diff.json) preserves both directions. The broad hand map was not rewritten for bind.
- `go test -count=1 -timeout 30m -skip '^TestProduct_' ./cmd/adamic-gate/ ./cmd/adamic-gate/parallelcensus ./cmd/adamic/treewalk ./internal/compilercensus`: all passed, 5.596/8.736/6.765/8.087 seconds respectively. `/tmp/bind-gate-final.log`. Earlier census/rule run passed too, `/tmp/bind-census-rules.log`.
- `gofmt -l cmd internal`: clean (`/tmp/bind-gofmt.log`), final repeat `/tmp/bind-gofmt-final.log`.
- IR search and import-aware AST audit: 169 switches, 1376 `ir` case lines, cmd/internal/stage1 searched. Gob registration, split node inventory, lower local/graph/view inventories and the corpus fixture registration were updated. Parallel census accepted the new tests. No new product test name or lint rule was added, so no product-order or lint registry edit was needed. External `devtools/fast-gate` skip/rule tooling is absent from this checkout; the local gate/census equivalents were run. No external gate verdict is claimed.

## Final validation limits

The broad runs below predate the last four isolated corrections. IR ran before the static-method JavaScript correction; the later full fresh, JavaScript, flow, split, lower, native and oracle runs tested `99a3d4fa3`. They are not claimed as a full green gate on the final implementation.

| Check | Result / log |
|---|---|
| All prerequisite products | 413 passed; products.json and /tmp/bind-products-final.log |
| Full fresh | passed, 118.631 seconds; /tmp/bind-fresh-javascript-static.log |
| Full JavaScript | passed, 1.023 seconds; same log |
| Full IR | passed, 13.339 seconds; /tmp/bind-ir-final.log |
| Full flow | passed, 243.021 seconds; /tmp/bind-flow-recovered.log |
| Full split | passed, 203.111 seconds; /tmp/bind-split-final.log |
| Full lower | completed, 1314.000 seconds; sole failure TestScanningReferenceCallsEveryPass; fixed and focused green, /tmp/bind-final-lower-recovered.log |
| Full native, parallel 2 | timed out after 1800.078 seconds, no assertion-failure markers; /tmp/bind-native-final.log |
| Native timeout's two active tests | TestParserConstructionRuntimeMutantSynthesizedKey and TestRegExpBytecodeTest262Bucket22Of32 passed in focus, 14.960 seconds; /tmp/bind-native-timeout-focus.log |
| Whole oracle, parallel 2 | timed out after 1800.636 seconds and reported 12 missing-TypeScript module failures; /tmp/bind-oracle-final.log |
| All 12 assignment-proof failures | passed in focus, 14.336 seconds; /tmp/bind-assignment-proof-focus.log |

Oracle's active tests at its deadline were the thread-sanitized gob shard 028, RuntimeProductKeyHoldsEveryFeatureAndOption/requests and TheJobserverNeverReachesTheEmittedCOrClang/runs. Setup already had TypeScript 6.0.3 at /workspace/adamic-tools/typescript. The stock-typecheck helpers override NODE_PATH to an absent stage3/api/node_modules. A temporary repository-root node_modules/typescript symlink to setup's package resolved all 12 failures; it was removed afterwards. No stage-three source, caller or adaptation was touched.

The pre-nullish final build passed after setting TMPDIR=/workspace/bind-final-tmp (`/tmp/bind-build-recovered.log`); final vet passed (`/tmp/bind-vet-boundaries.log`). Earlier attempts of the last build and lower focus failed only while writing Go output to the full /tmp mount (`/tmp/bind-build-boundaries.log`, `/tmp/bind-final-lower-focus.log`). Lower pipeline and closed inventories then passed in 0.019 seconds. The final ten-mutant record includes exact named catchers and restored-source checks. Final restored bind checks, with uncached observations and wasi enabled, passed in 16.744 seconds (`/tmp/bind-final-restored.log`). Only the eight new count rows were remeasured on the final compiler and passed in 9.147 seconds (`/tmp/bind-counts-final.log`); no old row was regenerated. Final gate, parallel census, treewalk and compiler census tests passed in 2.370/3.549/2.717/3.313 seconds (`/tmp/bind-readers-final.log`). Formatting is clean and the refreshed switch inventory still has 169 sites.

Not completed: a full green native or whole-oracle run within the 30-minute bound on this four-CPU box, and a whole-suite repeat after the last isolated corrections. No faster-box or external gate result is claimed. The implementation boundaries above remain deliberate NotYet rather than invented behavior. Apple submits; no Loom command was run.

Final dependency generation matched the earlier derived map exactly (`/tmp/bind-dependencies-final.json`), so the saved two-way comparison remains current. Count-file census and predicate-table validation passed (`/tmp/bind-count-census-final.log`). No existing count record was changed.

The last nullable-union pin exposed a native mismatch: `/tmp/bind-nullish-before.log` printed false/one effect where Node printed true/zero. Normalizing the existing shared null sentinel before the optional guard fixes it without any runtime layout or heap-link change. All bind tests, sanitizers/leak checks, release methods and wasi passed in 15.736 seconds (`/tmp/bind-nullish-fixed.log`). Final build and vet passed (`/tmp/bind-build-nullish.log`, `/tmp/bind-vet-nullish.log`). Only the method count row was remeasured for that fixture addition, passing in 18.353 seconds (`/tmp/bind-counts-nullish.log`). All eight new rows balance allocations against frees plus region releases. The tenth focused mutant reproduces the semantic mismatch and was restored byte for byte with a clean working diff.

Final count-file census and predicate-table check passed again in 0.076 seconds (`/tmp/bind-count-census-last.log`); formatting remains clean (`/tmp/bind-gofmt-last.log`). The file inventory covers all 82 changed files without duplicates.
