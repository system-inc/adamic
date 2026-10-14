import json,pathlib,re
p=pathlib.Path(__file__).resolve().parent
before=json.loads((p/'before.json').read_text())
recheck=json.loads((p/'before-recheck.json').read_text())
before=[row for row in before if row['test']!=recheck[0]['test']]+recheck
after=json.loads((p/'after.json').read_text())
bu=json.loads((p/'units-before.json').read_text());au=json.loads((p/'units-after.json').read_text())
old=sum(x['core_seconds'] for x in bu);new=sum(x['core_seconds'] for x in au)
s=f'''Built cmd/adamic references to shared native-runtime archive products, with a supplier that never compiles inside behavior tests.
Base: 27d661a838681d3144808e108e3e61e32e940a69; submission is the single commit on compiler/cmd-adamic-archives.
Measured five original units: {old:.3f} -> {new:.3f} core-seconds, saving {old-new:.3f} ({100*(old-new)/old:.1f}%).
Mutants: removing the sanitized declaration fails its original build leaf; dropping the version key part fails the new key guard.
Not covered: full gate, other whole packages, macOS, traced store publication, or runtime piece 4 merged locally.

All runtime C and all existing tests and their assertions are unchanged. No fixture was added, so counts.md needs no refresh. Only cmd/adamic's test package changes; no other package imports that main package's test sources. No new package needs a compiler-dependencies.json row. library.go's existing RuntimeRequestFor, SupplyRuntimes and Build APIs express every request; no API extension.

Feature and options audit

Every linked archive uses the base feature set `[]`: no ADAMIC_CLOSURE_CONVENTION, ADAMIC_CANONICAL_CLOSURES, ADAMIC_CLOSURE_RECEIVERS, ADAMIC_REGEXP_REPLACE_CALLBACK or ADAMIC_NODE_HOST. features.py ran the existing driver in C mode on every file-based witness; features.json and features-*.c.txt preserve the evidence. The anonymous throw witness is the unchanged source in TestWASIRequestThrows; its successful guarded build additionally verifies its base feature set.

| Consumers | Features | Options passed by behavior tests | Shared product options |
|---|---|---|---|
| Five TestExplainChecksOutput build leaves, eight TestNonNullExplainChecks build leaves, TestRequestNativeStaysCommand | [] | Sanitize: true | Sanitize: true |
| TestWASIRequest/request.a and request_alias.a | [] | Target: wasm32-wasi, Request: true, Count: true | Target: wasm32-wasi, Count: true |
| TestWASIRequestThrows | [] | Target: wasm32-wasi, Request: true | Target: wasm32-wasi |

Unlisted options are zero. The other cmd/adamic tests do not link an archive: C/JS rendering and handler selection only lower programs; target/usage guards refuse before runtime compilation.

Reuse and piece 4

The declaration uses the exact existing oracle identity: Name native-runtime, the same four declared Files, RuntimeRequest.Portable as Flags, and buildcache.Tool reports for Compiler and Archiver. shared-key.log verifies the two Inputs recipes are identical. Request changes WASI link flags only, so clearing it for declaration matching preserves the ordinary WASI archive key. The local TestProduct_CommandRuntime functions are discovery references to these identities, with the existing request.Build(directory, 4) recipe, not a second archive namespace. Sanitized and uncounted WASI reuse existing oracle products. Counted WASI adds the required combination under that same identity scheme; oracle does not currently declare that combination.

Inspected runtime/loom-c2-runtime directly. Piece 4 changes runtimeProductGroups from 3 to 6 and adds NativeRuntimes product tests numbered 3 through 5 for each option. It partitions complete feature-set archives, still compiling four C units at once within each archive. It changes neither RuntimeRequest's key nor runtimeInputs nor the build recipe. The base set remains shard 0. Thus **nothing in this declaration changes when piece 4 lands**: existing oracle keys are reused, and the counted-WASI reference remains required. Whichever lands second needs no shard-number adaptation.

Measurements

These are single observations on the four-CPU cgroup, not statistical estimates. nproc = 5; cpu.max = 400000 100000. Go 1.27.1, native clang 20.1.8, WASI SDK clang 20.1.8, Node 24.19.0. measure.py bounds each leaf by timeout 60 and -test.timeout=55s, captures output directly to logs, and records elapsed seconds and child user+system CPU seconds. A baseline leaf's cold run has a fresh runtime cache and its warm run reuses that cache. The after run uses products built ahead and ADAMIC_BUILD_CACHE=read. Go test binary compilation is outside these measurements in both versions.

The first baseline leaf initially overlapped the after binary's Go compilation. before-recheck.json replaces only that leaf's cold and warm observations, taken without competing compilation; both original logs remain. units.py measures the original top-level units with their original XDG isolation, including C/JS assertions and the shared archive between the two WASIRequest leaves. Its total, rather than the sum of isolated cold leaves, is the savings above.

| Runtime-linking leaf | Before cold s | Before warm s | After ahead s |
|---|---:|---:|---:|
'''
for row in after:
 test=row['test']; cold=next(x for x in before if x['test']==test and x['temperature']=='cold');warm=next(x for x in before if x['test']==test and x['temperature']=='warm')
 s+=f"| {test} | {cold['seconds']:.3f} | {warm['seconds']:.3f} | {row['seconds']:.3f} |\n"
s+='\n| Original unit | Before s | After s | Before core-s | After core-s |\n|---|---:|---:|---:|---:|\n'
for b,a in zip(bu,au):s+=f"| {b['test']} | {b['seconds']:.3f} | {a['seconds']:.3f} | {b['core_seconds']:.3f} | {a['core_seconds']:.3f} |\n"
s+='''
The three new product tests each build one complete archive. Settled traced runs took: sanitized 6.10 s, WASI 5.15 s, counted WASI 5.32 s (test times including request setup). TestCommandRuntimeProductKeyHoldsEveryPart took 1.19 s in its green verification. Each is below 60 s. Tracer process startup/settlement additionally appears in the complete *-traced.log files. Archive production moves work ahead; the core-seconds saving above is behavior-test CPU time, not a claim that archive production costs zero.

Guard and mutants

commandRuntime runs in every test process and every CLI driver subprocess. Requests outside commandRuntimes fail before any cache lookup, naming features and options. Its Get miss callback returns an error and never calls Build, even outside runner read mode. Only TestProduct_CommandRuntime functions call request.Build. The empty-cache negative control fails TestRequestNativeStaysCommand with runtime product not built ahead, without a C compile.

mutants.py independently changed real declarations and restored the source after each run:

* undeclared-sanitized.patch removes the sanitized product row. The unchanged TestExplainChecksOutput/overload_some_empty/build fails with no runtime product declared for features [] with Sanitize:true. The prebuilt cache remains present. Exit 1.
* missing-key-part.patch drops Portable[0] from Inputs. TestCommandRuntimeProductKeyHoldsEveryPart fails naming adamic-runtime-v6. Exit 1. No clang warning or source-program refusal catches either mutant.

Exact validation

Every invocation wrote its output to a file, without piping the test process. The scripts in this directory preserve selections and environment. Invocations used the setup env file and ADAMIC_BUILD_STORE=off:

* go test -c -o /tmp/adamic-before.test ./cmd/adamic, initially bounded at 180 s (expired), then at 300 s (passed); after binary: bounded at 90 s, passed. GOCACHEPROG was explicitly set to the setup-installed hook for builds and mutant go test commands after the interrupted no-hook setup attempt.
* features.py: timeout 90 overall, timeout 55 per driver.
* measure.py /tmp/adamic-before.test before: timeout 600 overall, 60/55 s per leaf. One isolated before-recheck as explained above.
* units.py /tmp/adamic-before.test before: timeout 240 overall, 60/55 s per unit.
* Each TestProduct_CommandRuntime function separately: timeout 60 /tmp/adamic-traced -- /tmp/adamic-after.test -test.timeout=55s -test.v -test.run=^NAME$. All three settled, zero refused in the final runs.
* measure.py /tmp/adamic-after.test after: timeout 120 overall, 60/55 s per leaf, ADAMIC_BUILD_CACHE=read. All 17 pass.
* units.py /tmp/adamic-after.test after: timeout 90 overall, 60/55 s per unit, ADAMIC_BUILD_CACHE=read. All five pass with their original assertions.
* remaining.py: timeout 120 overall, 60/55 s per test. All other cmd/adamic Test functions pass separately; remaining.json lists each selection and seconds. No whole-package test invocation or full gate.
* mutants.py: timeout 180 overall; each go test ./cmd/adamic -run SELECTED -count=1 -timeout 60s bounded by timeout 90. Both killed and restored.
* Required IR reader check and changed-package vet: see ir-readers.log and vet.log. Final clean diff and formatting checks: see final-checks.log.

Setup failures and recovery

The initial timeout 240 bash cloud/setup.sh stopped while fetching the nested TypeScript submodule. setup-first.log ends at that clone. A depth-one fetch of its exact pinned commit d92d9bfee114c80be2c375d72edae966176e3a4f, then checkout, repaired it (submodule-repair.log). No submodule revision was changed.

The retry with --wasi-sdk reached toolchain readiness but its 240 s overall bound expired during Go warming (exit 124). Timing lines: Node 0.048 s, Go 0.056 s, submodules 0.118 s, markdown 0.159 s, clang 0.399 s, WASI SDK 5.397 s, shared cache 13.612 s. A later no-hook warming attempt was stopped with SIGTERM (143) to avoid competing compilation during measurements: Node 0.243 s, Go 0.253 s, submodules 0.535 s, markdown 0.795 s, clang 1.575 s, WASI SDK 2.079 s, cache off 2.087 s. Full setup completion is not claimed. Sourced /workspace/adamic-tools/env.sh for every toolchain shell; explicitly reused its installed Go cache hook to complete the targeted binaries and checks.

Two methodology failures were corrected before the final run: ordinary untraced local products are not readable in runner read mode; the traced builder then refused an untracked new Go file. Staging the new file before tracing fixed the latter. trace-untracked-refusal.log preserves that refusal. The final products have settled read sets and all after measurements use runner read mode.

Conservative assumption: local product Test functions reference the shared native-runtime keys for discovery and may build a missing shared key ahead; they do not introduce distinct cmd archive products. A developer must run products first on a cold cache. No archive may be built by the behavior-test supplier, including in ordinary local mode.
'''
(p/'REPORT.md').write_text(s)
print(f'{old:.6f} -> {new:.6f}; saved {old-new:.6f} core-seconds')
