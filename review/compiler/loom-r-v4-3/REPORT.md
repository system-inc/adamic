Built: views V4 carrying V3, with checker-proven primitive equality operands bypassing JavaScript adapter-root lookups.
Commits: one commit on exact dbdf977f5b8b0979d414ebe6f20e7aadd312e309, carrying 82a5dffd7654f3eb2cb4652a12b99401210f442a; final SHA is in the push report.
Commands and outputs: 315 focused entries plus three new unit checks, 48 focused reader entries, guards, vet, a-check and both complete JSON packages pass; timings below.
Mutants: all 19 carried source mutants, object-lookup omission, primitive-proof omission, removed header and wrong reason are caught.
Not covered: a full repository gate, ahra/Loom submission, or reproduction of the remote loaded Lunatic environment.

The branch starts at the requested exact base, despite origin/main advancing. The no-commit cherry-pick of 82a5dffd had no conflicts. This commit includes its entire change and the primitive optimization. No runtime C line or counts row differs from 82a5dffd. No new oracle fixture or runtime feature set was added.

Equality and inequality, plus Object.is, carry a two-operand checker certificate in IR before union boxing. Number, string, boolean, bigint, undefined and null, and unions solely of these, qualify. Scalar IR representations also bypass normalization for generated comparisons. Unknown, any, object, function and unions admitting them retain lookup. Unresolved or otherwise unproved types conservatively retain it. This does not add bigint lowering support.

The compiler deliberately refuses a callable erased to source object type. Therefore the object-typed regression exercises the real JavaScript emitter on object-typed IR operands and real adapter root metadata. Source-level callable identity and cached/no-stacking tests separately pass in native release, sanitizers and JavaScript. The object-lookup compiler mutant runs successfully on Node but produces false false instead of true true, so the identity assertion fails semantically; no compiler error kills it.

Full-corpus JavaScript measurements use the same 3,919-case pinned protocol, fresh processes, GOMAXPROCS=4 and alternating revision order. Each output is 300,372,755 bytes with source Node SHA-256 8571e190419a2e27c5796b3f1ff6485cfc824ddf161fcf48e9b3a55739ac614f, exit 0 and empty stderr. Source/corpus provenance and current compiler hashes are retained. Three runs per revision:

| Revision | Samples (s) | Median (s) |
| --- | --- | --- |
| dbdf977f5b | 30.169, 26.252, 26.785 | 26.785 |
| V4-3 | 26.400, 26.058, 25.786 | 26.058 |

V4-3/base median ratio: 0.9728. The prior V4 diagnostic was 25.465 s base versus 48.148 s V4. That earlier single comparison and the runtime-only causal ablation remain in previous-diagnostic-report.md; neither ablation is a compiler change here.

Both complete stage1/cohere/json package runs use -count=1 -json -parallel=4 -timeout=30m, outer limit 1860 s, GOMAXPROCS=4, sequentially with no competing validation work. The pinned corpus, Go oracle owners and lowered/native port products are settled first on both trees; the latter use -parallel=1. Commands and all raw JSON are in fixed/commands.jsonl and fixed/{base,v4}-*.jsonl.

| Revision | Package elapsed (s) | Pass / skip events | Failed or >=60 s tests |
| --- | --- | --- | --- |
| dbdf977f5b | 299.357 | 2121 / 5 | 0 |
| V4-3 | 303.516 | 2122 / 5 | 0 |

All 2,048 ordinary agreement shards pass on each revision. No shard crosses the 60 s budget on V4-3. fixed/shard-times.csv lists every test on both sides; fixed/{base,v4}-tests.json retain outputs and backend timings. There is no observed output mismatch to reduce. These local observations do not establish the exact remote timeout mechanism.

base slowest ordinary shard: TestPortMatchesGoCohere_1643, 13.590 s.
v4 slowest ordinary shard: TestPortMatchesGoCohere_1226, 15.010 s.

Focused validation runs exact top-level tests from fixed/selected-focused.json, using compiled test binaries, -test.count=1 -test.timeout=90s -test.parallel=1, outer 95 s per entry. Products run serially, other entries with two workers. Every TestV4 name is selected (fixed/v4-coverage-audit.json). All lane 5 boxed parameter/result and producer negatives run; all adapter identity/cache/no-stacking tests run. Native fixtures use release and ASan/UBSan/LeakSanitizer through internal/native. Source Node determines their expected behavior.
315 focused entries: 308 pass, 7 explicit skips; maximum 14.055 s. 53 reader packages compile, 48 selected reader tests finish (2 skips); maximum 24.669 s. Optional/WASI skips retain their exact reasons in logs.

New test seconds are in fixed/new-test-times.json. The primitive-union oracle fixture checks source Node, native release, sanitizers and JavaScript, including nullish unions, NaN and signed zero. The proof classifier separately checks bigint and mixed object/callable unions. The two JavaScript tests prove representation fallback and object identity.

Guards pass: TestCallTargetReaders (-timeout 90s), TestEveryMapAProgramHoldsIsSorted, TestRuntimeSetsAreEachUsed, TestNativeAgreementRowsNameRegisteredPrograms, TestEveryTestHasOneDescriptiveName, TestEveryTestIsParallelOrSaysWhy and TestEveryCCompilerRunsThroughNative. No guard or baseline was loosened. New tests have unique descriptive names and call t.Parallel first. The IR addition is a fixed boolean array, with no new map or call target. Existing native-agreement registrations and exact runtime witnesses remain intact. No compiler-consuming package, direct C compilation or build-directory-bearing product was added. go vet ran on internal/{flow,fresh,ir,javascript,load,lower,native,native/runtimeproducts,oracle} and stage1/cohere/json, with a final JavaScript vet after its test edit. Exact outputs are in fixed/*.log.

A-check invokes the candidate compiler on all 47 added/modified .a paths, including review witnesses, through unchanged pinned Gate.aCheck (fbac28c62493f27a788edc02a18bc8edb68de5da). Status counts: 40 checked, 1 NotYet, 5 refused, 1 TS7009 checker error. All six expected-error headers already match; no header or body bytes changed. Both removed-header and wrong-reason mutants fail the exact predicate. fixed/a-check/status-table.md is the per-file table; statuses.json and gate-baseline.json contain real current diagnostics and hashes.

Carried source mutants and exact catchers:

| Mutant | Catcher | Seconds |
| --- | --- | --- |
| escape-Argument | TestV4EscapeArgumentRefusal | 18.104 |
| escape-Result | TestV4EscapeResultRefusal | 6.672 |
| escape-Passed | TestV4EscapePassedRefusal | 7.051 |
| escape-Returned | TestV4EscapeReturnedRefusal | 6.100 |
| escape-MixedProducer | TestV4EscapeMixedProducerRefusal | 6.879 |
| unsupported-domain | TestV4DirectRecursiveRefusal | 22.860 |
| unsupported-producer | TestV4DirectProducerRefusal | 17.843 |
| surface-borrow | TestV4EscapeAdapterName | 13.866 |
| bind-name-borrow | TestV4EscapeAdapterBindNameOwnership | 14.701 |
| identity-root | TestV4EscapeAdapterIdentity | 16.565 |
| metadata-inline | TestPortElementMetadataEmission | 9.132 |
| metadata-omitted | TestPortElementMetadataEmission | 8.707 |
| native-order | TestV4DirectCalleeOrder | 13.102 |
| javascript-order | TestV4DirectCalleeOrder | 12.029 |
| construct-witness | TestV4EscapeAdapterWithoutConstructSignatureRefusal | 3.244 |
| boxed-witness | TestV4Lane5BoxedAdamicRefusal | 6.684 |
| factory-witness | TestV4Lane5FactoryWiderWriteRefusal | 6.323 |
| runtime-witness-mismatch | TestRuntimeSetsAreEachUsed | 9.715 |
| main-library-spread-omitted | TestLibrarySpreadRegExpUnion | 9.773 |
| object-typed lookup omission | TestJavaScriptObjectOperandAdapterIdentity | 1.542 |
| primitive proof omitted | TestV4PrimitiveUnionIdentityEmission | 18.149 |

The focused driver hit disk exhaustion while opening a new log after 258 completed entries. Generated binaries and bulk stdout artifacts were removed; their hashes and measurements remain. Only unfinished entries resumed, with complete pass logs recovered rather than rerun. No assertion, timeout or 60 s leaf was retried. fixed/disk-cleanup.json, focused-driver.log and focused-resume.log preserve the interruption and recovery. Generated test binaries and multi-GB bulk outputs are excluded from the commit.

Toolchain setup ran with GOPROXY=https://proxy.golang.org|direct, then sourced /workspace/adamic-tools/env.sh. Linux, clang 20.1.8, Go 1.27.1, Node 24.19.0; nproc=5, cgroup CPU quota=4 and memory limit=16 GiB. Setup timing lines:

```
setup: node ready (0.022s)
setup: go ready (0.023s)
setup: submodules ready (0.073s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.008s
setup: markdown dependencies ready (0.076s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.159s)
setup: shared cache ready (0.945s)
setup: go build ready (15.584s)
setup: test binaries deferred (use --warm-tests) (15.658s)
setup: build cache warm (15.660s)
setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.00 0.37 1.20 1/454 222594 load-after=0.77 0.52 1.23 1/459 223196
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (15.686s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.eQSbcV
```

The complete package elapsed is 299.357 s base and 303.516 s V4-3 (+1.39%). The per-shard backend elapsed sums, which overlap and are not package wall time or CPU time, are in fixed/backend-summary.json. The fresh full-corpus median is 2.72% lower on V4-3; the full-package sample is 1.39% higher.

The disk interruption and required sequential package comparisons took the first green push beyond the 30-minute target; no failing leaf or budget overrun was retried.

JavaScript load/lower/emit: base 0.619 s, V4-3 0.666 s. Emitted bytes: base 188,693, original V4 207,020, V4-3 195,710.
