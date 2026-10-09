Stopped incompatible Weak method and callback views; routed narrowed union reads through the existing liveness check for #63pvx2b.
Compiler commit: 2b13a3d1bab431f2bceec803ddf048cb57cd6735; counts commit: e60c7fa51; latest main merged: 7506c6e9a.
Build, focused vet/lowering checks, uncached Node proofs in both backends with native sanitizers, and counts regeneration pass.
Three revert mutants caught: method views restore wrong output and a retain fault; union liveness restores a null-pointer fault; callbacks restore six closure faults.
Not covered: general Weak representation adapters, other differential programs, physical platforms beyond this Linux box, or the full repository gate.

## Findings and outcomes

The top finding on main f54b8bd32 is a silent wrong result. The supplied
`9984394_view_method_parameter.a` prints **`t1 none` and exits 0** in the native
-O2 build; source Node prints **`t1 b1` and exits 0**. The sanitized build faults.
Its complete source is pinned under `internal/oracle/testdata/weak_miscompiles/`.

All ten supplied sources are preserved verbatim. Every source Node run exits 0
with empty stderr. [results.json](results.json) records their observations before
and after the changes and asserts that each Node observation is unchanged.
The main audit logged outcomes; its PASS labels mean the observations were
collected, not that the backends agreed. The final tests assert the outcomes.

| Program | Node stdout (newlines escaped) | Main native -O2 | Final native and JavaScript outcome |
|---|---|---|---|
| 9984394_view_method_parameter | `t1 b1\n` | `t1 none\n`, exit 0 | Both stop with NotYet at 19:27: StrongTaker viewed as WeakTaker changes Weak storage. |
| 9984394_view_method | `true b1\n` | Signal; sanitized fault in adamic_retain | Both stop with NotYet at 19:29: StrongGetter viewed as WeakGetter changes Weak storage. |
| 4ddd17f_weak_union_narrowed | `before\nafter: nn\n` | Signal; sanitized null-pointer read | Native release and sanitized builds print `before\n`, then the existing liveness panic, exit 70. JavaScript matches Node. |
| 4ddd17f_weak_single_narrowed | `before\nafter: nn\n` | Existing liveness panic, exit 70 | Same native panic and unchanged Node/JavaScript observation. Control. |
| 9984394_narrowed_each | `a1\nb1\n` | Signal; sanitized closure heap-buffer-overflow | Both stop with NotYet at forEach, 10:2. |
| 9984394_narrowed_find | `b1\nb1\nb1\n` | Signal; sanitized closure heap-buffer-overflow | Both stop with NotYet at find, 10:16. |
| 9984394_narrowed_map | `a1,b1\n` | Signal; sanitized closure heap-buffer-overflow | Both stop with NotYet at map, 10:14. |
| 9984394_narrowed_reduce | `4\n` | Signal; sanitized closure heap-buffer-overflow | Both stop with NotYet at reduce, 10:16. |
| 9984394_narrowed_some | `true\n` | Signal; sanitized closure heap-buffer-overflow | Both stop with NotYet at some, 10:15. |
| 9984394_narrowed_sort | `true\n` | Signal; sanitized closure heap-buffer-overflow | Both stop with NotYet at sort, 10:2. |

The native panic is exactly:

```text
adamic: panic: a weak reference was read after what it pointed to was freed
```

This is the existing ruled Weak lifetime difference from Node tracing, not a new
language choice. A terminal panic keeps what it held when it stopped; those runs
do not undergo a completion leak check.

## Narrow checks

Method views compare the actual parameter and result representations. A Weak
handle slot cannot become a plain nullable target slot without a conversion.
Only Weak-versus-non-Weak differences trigger this added method check. The
existing path-bearing NotYet names the two viewed types.

Main's d05a1bf9 boxed-union filter stop did not catch any of the six Weak programs.
The new stop applies at a callback call over an array storing Weak handles when
the callback parameter expects a target. It examines the element parameter of
visiting/mapping callbacks, the second parameter of a reducer, and both element
parameters of sort. The diagnostic names the method and the missing
handle-to-target conversion, with the fix: use a loop with an explicit narrowed
copy. Expression and discarded-statement calls use the same guard.

The initial filter-level stop also blocked the existing, correct weak_narrowed.a
fixture. The guard was narrowed to the callback call. Indexing, identity checks,
length, and for-of over filtered Weak arrays still compile and match Node. The
red regression and counts logs are preserved as initial evidence; final runs pass.

A union of branded Weak targets remains represented as Weak after the checker
removes undefined. Its read now derives the presence check from absence in the
checker type, and uses the same WeakTarget liveness check as a single target.
No native emitter or runtime file was changed.

Three new compiling controls hold matching Weak method signatures, a boolean
filter followed by explicit narrowing, and a union Weak with its strong owner
still held. Source Node, backend JavaScript, native -O2, ASan/UBSan and completion
leak checks agree. Both main merges were clean and brought only stage1 tests.
No dependency branch was merged. No stage3 fixture status record changed.

## Mutants

`python3 internal/lower/testdata/run-weak-differential-mutants.py` reverts each
core change independently and restores the source in a finally block. Each
mutant's Go test exits 1 and recovers the intended old runtime behavior; compiler
warnings or build failures do not count as catches. The runner exits 0 only when
all three are caught. See [mutants.json](mutants.json) and the three mutant logs.

| Revert | Catcher |
|---|---|
| Skip Weak method slot comparison | Both view tests admit the program. Native parameter test again prints `t1 none` with exit 0; the getter faults in retain. Node/JavaScript retain their original outputs. |
| Omit union presence/liveness fact | Union test faults on a null-pointer read instead of producing the pinned native liveness message. Node/JavaScript still print `before` and `after: nn`. |
| Return no error for incompatible Weak callback slots | All six callback tests admit the program and fault under ASan in the closure. Each also fails its pinned NotYet and native exit assertion. Node/JavaScript retain their original outputs. |

## Test grain

All thirteen new test leaves are separate top-level Test functions, with
`t.Parallel()` as their first statement. Seconds below include their setup on
this box with a 4-CPU quota. No new leaf reaches 60 seconds.

| TestWeakDifferential leaf | Seconds |
|---|---:|
| BooleanFilter | 1.84 |
| Each | 0.92 |
| Find | 0.89 |
| HeldUnion | 1.57 |
| Map | 0.92 |
| MatchingMethods | 1.36 |
| Reduce | 0.55 |
| SingleNarrowed | 1.31 |
| Some | 0.64 |
| Sort | 0.54 |
| UnionNarrowed | 1.61 |
| ViewMethod | 0.64 |
| ViewMethodParameter | 0.49 |

## Commands and outputs

Every test writes to its own log. No whole package or full repository test gate
was run. The focused regression selection is recorded exactly below.

```text
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
  setup exit 0; nproc=5; cgroup cpu.max=400000 100000
go build ./cmd/adamic
  exit 0
go vet ./internal/lower ./internal/oracle
  exit 0, no diagnostics
go test ./internal/lower -run 'TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat|TestAViewThatCantWriteIsNotRefused|TestAMethodReadAsAValueIsRefused' -count=1 -v -timeout 5m
  pass, 1.982s
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestWeakDifferential|^TestWeakReadsUndefinedOnceFreed$|^TestPredicateMiscompileRefusals$|TestNativeAgreesWithNode/internal/oracle/testdata/(class_as_interface|class_inheritance_interface|optional_class_method|weak_parent|weak_narrowed|host_array_predicate|weak_miscompiles|predicate_refusals)' -count=1 -v -timeout 10m
  pass, 9.715s; native misses=58, Node misses=63
python3 internal/lower/testdata/run-weak-differential-mutants.py
  all three caught, exit 0
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 15m -args -update-counts
  pass, 69.314s for the existing aggregate count writer
```

The new stops occur in shared lowering before either backend receives a program;
there is no native binary to sanitize for a stopped source. The liveness programs
and all compiling controls run in both backends, including sanitized native builds.
Every reverting mutant also runs the admitted programs through both backends.

Setup timing lines: Go ready 0.026s; Node ready 0.031s; markdown dependency step
0.009s and ready 0.089s; submodules ready 0.103s; clang ready 0.195s; build ready
37.914s; test binaries deferred 38.063s; cache warm 38.065s; done 38.097s.
Go 1.27.1, Node 24.19.0, clang 20.1.8. No setup failure or workaround.

## Counts and lane checks

Linux regeneration adds five rows and changes no existing runtime or predicate
row. Allocations/frees/retains/releases/peak/regions:

| New fixture | Counts | Reason |
|---|---|---|
| matching_methods.a | 11/11/13/21/7/0 | Matching Weak parameters and results finish with all allocations freed. |
| boolean_filter.a | 12/12/22/28/9/0 | Safe boolean filter and explicit Weak narrowing finish with all allocations freed. |
| held_union.a | 5/5/13/17/5/0 | Target remains strongly owned through the narrowed union read. |
| 4ddd17f_weak_union_narrowed.a | 4/2/5/6/4/0 | Counts stop at the native liveness panic. |
| 4ddd17f_weak_single_narrowed.a | 4/2/4/6/4/0 | Same terminal check; one fewer read-retain than the union shape. |

Refused shapes have no runtime counts. The two deliberate Weak lifetime
observations use the count writer's additional-fixture hook and their own oracle
leaves, following the existing Weak divergence convention.

After committing, integration's required command passed:

```text
git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
lane checks 1.2 s: gofmt and tools on 4 Go files, t.Parallel on 1 test packages; vet 1 packages
```

The latest merged tip passed the own-fixture uncached rerun in 1.917s (native misses=22, Node misses=24). The final committed tip is checked again before the single push. Complete logs are in [logs.tar.gz](logs.tar.gz). Focused vet also
covers the production lowering package. All evidence is under this review path;
there are no compilable Go sources under review/. This unit serves #63pvx2b;
no numbered roadmap step was supplied, so none is inferred.
