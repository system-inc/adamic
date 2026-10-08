Memoized namespace reachability by complete recursive component; module traversal no longer expands call paths.
Implementation commits: d27c99ebac77a9ddc8754daaa37760ca2801256b and 585927fceb3fd07361e0d9dfedbde7e890ba60a1; final feature delivery SHA is in the response.
Final 24-level witness: C in 51.999 ms, native build in 157.599 ms; Node and native print true, exit 0, empty stderr.
Mutants: missing memo, partial cycle union, missing early-enum reach, reaching-as-unreaching, missing runtime readiness and erased const-enum readiness were each caught.
Debug read meter: 55 to 0; complete native Debug/parser proof and the full repository gate remain unachieved.

## Reachability and complexity

The old active-path map prevented recursion but discarded completed functions.
Two calls per level walked every path of a shared DAG. The new graph records
known calls and callback arguments and scans each body once. Tarjan's stack
identifies strongly connected components. A completed component receives the
union of all its direct reads and outgoing component reach sets, and all its
members share that result. No active component's partial set is used as a proof.

Cached sets contain namespace declarations and ordinary namespace-scoped enums,
independent of the module evaluation point. Calls compare the set with the
currently initialized declarations. Unresolved targets retain runtime checks;
conservative traversal still includes calls under `if (false)`.
Graph construction is O(body syntax + call edges). Reach-set union cost additionally
depends on the number of distinct reachable declarations; this is not a claim of
linear storage for an unbounded number of namespaces.

The [exact 24-level input](../../internal/oracle/testdata/namespaces_call_graph.a)
is byte-identical to parser proof 33110b86's native-namespace-call-graph.a.
[Timing commands and observations](call-graph-evidence/timings.json) include
6,819 emitted C bytes. Initial measurements under concurrent Go compilation were
281.200 ms for C and 507.457 ms for build. The final restored-code measurement
is [51.999 ms for C and 157.599 ms for build](call-graph-evidence/timings-delivery.json),
with native stdout `true` again. These observations exclude building the Go
compiler binary. `TestNamespaceCallGraphLinearWork` counts body expansions:
13 at depth 12 and 25 at depth 24. Querying every already visited function again
must add zero expansions. This deterministic pin is independent of CPU load.

## Recursive components held to Node

[The accepted cycle](../../internal/oracle/testdata/namespaces_call_cycle.a) reads
live Debug state after initialization and prints `true:true` then `false:false`.
Native with sanitizers, generated JavaScript and source Node agree.
The unit test separately checks that each of three mutually recursive functions
reaches both First and Second, including reads discovered after the back edge.

The premature-cycle program is exactly:

```typescript
function one(n:number):boolean {return n > 0 ? two(n-1) : Debug.ready;}
function two(n:number):boolean {return n > 0 ? one(n-1) : Debug.ready;}
console.log(`${two(3)}`);
namespace Debug {export let ready=true;}
```

Node exits 70, stdout empty, with `adamic: panic: TypeError: Cannot read properties of undefined (reading 'ready')`.
The compiler exits 1 at reaching_cycle.a:2:59:
`stage 0 can't lower a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration yet`.
The refusal is sound: this invocation actually reaches the missing namespace.
[Exact source, diagnostics and safe-cycle observations](call-graph-evidence/cycle-results.json)
retain both programs.

## The three broader integration failures

Parser and IncrementalParser previously required a bodyless-function NotYet even
in an integration that implements checked overloads. Their normalized shapes
contain overload declarations followed by implementations. The tests now compare
against a concrete module-level overloaded function with the same declaration
pattern: if that lowers, both namespace shapes must lower; if it returns the
canonical bodyless-function NotYet, both must retain that exact boundary.
An unrelated module failure fails the test. On this branch both still return
bodyless-function NotYet, so no declaration-shape count is promoted.
The supplied scratch compiler itself was not rebuilt here.

The early-enum regression remains a refusal, rather than weakening its expectation.
Namespace graph reachability now also retains regular namespace enum reads and
tracks their declaration evaluation. The standalone namespace preflight test
bypasses module enum analysis and must still reject:

```typescript
namespace N {function read():number{return E.A;} const x=read(); enum E {A}}
```

This preserves the namespace boundary when integration replaces the module-level
initialization analysis. All three broader tests pass. No overload policy or
protected compiler orchestration file was changed.

## Mutants actually run

| Mutant | Catcher |
| --- | --- |
| Reuse active functions only, discard completed memo entries | Linear-work pin: 8,191 body walks, expected 13; Go test fails |
| Give each cycle member only its own direct reads | Cycle-union pin: one reachable namespace, expected two |
| Drop namespace enum reads from function reach sets | Independent namespace preflight loses the early-enum refusal |
| Treat known function calls as unresolved | Direct/helper/cycle static refusal pin fails |
| Remove runtime namespace readiness | Node/native exit mismatch on the unknown-function fixture |
| Erase the const-enum namespace readiness hook | Node/native exit mismatch on the unknown-enum fixture |

The runner is [call-graph-mutant.py](call-graph-mutant.py); the existing reachability
runner was also rerun. Mutants restore sources in finally blocks and reject Go
build failures as a catcher. Logs are committed under call-graph-evidence.
The eight existing semantic mutants also pass: wrong scoped function, scoped
constant, namespace enum, body order, exported state, Debug initialization,
returned assignment and factory binding are caught by Node stdout with clean
native execution. The three state mutants lose assignment, hoisting or a ready
check; the first two are caught by Node comparison and the last by checked
JavaScript exit comparison.

## Validation, matrix and setup

```sh
bash cloud/setup.sh > /tmp/namespaces-graph-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go test ./internal/lower -count=1 > /tmp/namespaces-graph-lower-final.log 2>&1
go test ./internal/lower -run 'TestNamespace|TestTscNamespaceDeclarationShapes' -v -count=1 > /tmp/namespaces-graph-focused-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode.*|TestNamespace' -count=1 -timeout 30m > /tmp/namespaces-graph-oracle.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/(internal/oracle/testdata/namespaces|stage3/fixtures/namespaces|stage3/namespaces/shapes)|TestNamespace' -count=1 -timeout 30m > /tmp/namespaces-graph-oracle-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/namespaces-graph-counts.log 2>&1
python3 stage3/namespaces/call-graph-mutant.py > /tmp/namespaces-graph-mutant-final.log 2>&1
python3 stage3/namespaces/reachability-mutants.py > /tmp/namespaces-graph-old-mutants.log 2>&1
go vet ./internal/lower ./internal/oracle > /tmp/namespaces-graph-vet.log 2>&1
```

Observed passes: complete lower package 35.082 s; focused namespace/shape tests
4.322 s; uncached native differential fixtures plus namespace mutants 177.346 s;
final namespace oracle 4.321 s; counts 32.915 s. Vet, formatting and diff checks
are clean. Full repository gate, all oracle test families and WASI were not run.

All twelve unchanged original namespace slices were rebuilt under ASan/UBSan
with leak checks and checked JavaScript against source Node. Matrix:
**6 Compiles, 3 NotYet, 3 Refused, 0 Checker**, no changed rows.
The two new oracle count rows are respectively allocations/frees/retains/releases/
peak/regions `1/1/0/1/1/0` and `2/2/0/2/1/0`; all existing rows are unchanged.

Setup succeeds: Go ready 0.073 s, Node 0.074 s, markdown dependencies 0.149 s,
submodules 0.172 s, clang 0.290 s, Go build 43.798 s, warm cache 44.096 s,
total 44.203 s. nproc=5, CPU quota=4. Go 1.27.1, Node 24.19.0, clang 20.1.8.
Refreshed origin/main ce0750f28ef3943057f1f852b3ae5d93e6c5d644 is already an
ancestor; `git merge --no-edit origin/main` reports `Already up to date`.
Only codex/namespaces-tsc was pushed. No main or area branch was modified.


## Third item: the 55 Debug reads

The paired meter at `stage3/meter/runs/20261007T205008Z.iT8wjb` reports 55
`NotYet: reading Debug` sites in both main and area. The branch measurement
on the completed pinned adapted source is **0**, using the same distinct
(kind, where, reason, text) definition across all 79 source files.
Qualified Debug reads already use the branch's namespace singleton bindings
and readiness helpers; no namespace container was manufactured to clear this
family. The existing post-initialization Debug probe and imported Debug fixture
pass Node/native/JavaScript. The unresolved-before-initialization fixtures stop
with Node's TypeError and preserve argument/RHS side effects. Proven premature
known calls retain the explicitly pinned NotYet boundary.

Measured totals are 1,417 NotYet, 4,668 Refused, three SkippedDependency events
and three recovered panics. Forty-eight function bodies are skipped for checker
diagnostics. Thirteen distinct NotYet sites are located within debug.ts; these
are other families, not remaining `reading Debug` observations. They cover
unknown/never/type-parameter values, a generic function value, a bodyless
signature, two enums inside functions, a callable namespace merge, and four
namespace-object observations. They are preserved in
[the complete count and remaining-site ledger](call-graph-evidence/debug-meter.json).
Clearing this reason does not mean 55 complete programs now compile.

The remaining complete Debug file was attempted with the production native
compiler. It exits 1 before lowering: its imported program has checker errors,
starting at builder.ts:1246:69 (`Path | undefined` passed as `string`). The full
build log is retained; no native Debug file or complete parser output is claimed.
The unchanged twelve original namespace fixture builds remain 6/3/3.
The 55-site blocker family is cleared; wider Debug/class/merge/generic and
checker repairs remain with their respective units.

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/namespaces-debug-overlay > /tmp/namespaces-debug-overlay.log 2>&1
gofmt -w /tmp/namespaces-debug-overlay/*.go
go build -buildvcs=false -overlay=/tmp/namespaces-debug-overlay/overlay.json -o /tmp/namespaces-debug-census ./stage3/census/latent/tool > /tmp/namespaces-debug-build.log 2>&1
bash stage3/apply.sh /tmp/namespaces-debug-adapted > /tmp/namespaces-debug-apply.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/namespaces-debug-census /tmp/namespaces-debug-adapted/src/compiler /tmp/namespaces-debug-latent-final.jsonl > /tmp/namespaces-debug-latent-final.log 2>&1
python3 stage3/census/latent/audit.py /tmp/namespaces-debug-census > /tmp/namespaces-debug-audit.log 2>&1
/tmp/namespaces-graph-matrix-final/adamic build /tmp/namespaces-debug-adapted/src/compiler/debug.ts -o /tmp/namespaces-debug-whole > /tmp/namespaces-debug-whole-build.log 2>&1
```

The final corpus measurement takes 177.711 s and verifies every source hash
before and after the run. The initial overlapping preparation measurement is
not used. Ordinary production loading and usable IR are disabled in this census
binary, with both output guards enabled. Audit passes continuation, refusal
coverage, body skipping, signature eligibility and same-line ranges, and catches
the body-scope and misattribution mutants. Compressed raw findings, source hashes,
audit output and the production build failure are committed with this report.
No checker option or compiler-source overlay is delivered to production.


Final delivery after removing the obsolete module active-path map: the full
lower package passes in 24.395 s, the uncached namespace oracle passes in
2.646 s, and vet is clean. The final twelve-fixture matrix row records code
commit 585927fc and the same 6/3/3 outcomes. Both new count rows were recorded;
no previous row changed. The meter binary was built at d27c99eb; the cleanup
only changes namespace initialization preflight, which the latent driver omits.
All code changes have been rechecked with production lowering and Node.
