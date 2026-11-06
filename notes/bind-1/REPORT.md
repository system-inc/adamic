# Bind scout (#kpp0f8s)

Status: stopped before implementation. This report records source evidence, not a
completed bind lowering or a completed cycle analysis.

Base: `49e49ba1de332b3a7c99c01f2364f50e53889fa5` on `apple/bind-1`.
The only intended tracked change is this report. The report commit is the head;
`git rev-parse apple/bind-1` identifies it without a self-referential hash.

## Why I stopped

The pinned source is TypeScript v6.0.3, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`, from `stage3/source.json`.
`stage3/meter/twice-daily.sh:39` obtains this source for `stage3/apply.py`.
The 46 compiler callers are `maybeBind(X, X.m)`. The helper at
`src/compiler/core.ts:1522` takes `fn: ((this: T, ...args: A) => R) | undefined`
and at line 1523 returns `fn?.bind(obj)`. There are 43 non-writeFile callers
and three writeFile callers. The compiler directory has just one existing
direct optional method bind, `program.ts:1950`, outside these 46 callers.
A TypeScript 6.0.3 syntax-tree walk independently confirmed these counts.
The two binder.ts calls belong to a different local function and are excluded.

The requested exception admits a declared method read only as the direct
receiver of `.bind`. It cannot admit such a method argument of
`maybeBind(X, X.m)` while preserving that exception exactly. There is a current
adaptation caveat: 97-cli-axolotl-b and 97-cli-axolotl-d turn many optional method
declarations into optional properties of indexed method types, for example
`m?: { m(): string; }["m"] | undefined`. A read of such a property is not the
same checker symbol as a declared method. I did not establish which of all 43
reads still carry method symbols after the complete adaptation pipeline.

Independently of that distinction, the helper body has a function-value
optional bind, rather than the admitted method-receiver optional bind shape.
A scratch reduction using that adapted property spelling gets past the method
read and stops at the helper's `fn?.bind(obj)`, with the optional-call NotYet.
Node prints `ok` for the same reduction.
The task preserves existing signature refusals. I found no stage3 adaptation
that rewrites these callers: the adaptation mentions concern optional property
types and explicit-any evidence, not expansion of maybeBind calls.

Consequently, implementing only the two specified forms would not unblock the
46 callers as written. I stopped to surface this scope mismatch before adding
an unreviewed helper special case, broadening unbound-method, or changing the
pinned compiler source. A follow-up can explicitly choose direct bind support
alone, or authorize a stage3 caller expansion alongside it. The latter still
needs the complete memory, refusal, oracle and mutant work in the task.

This is not a claim that most callers fail the cycle rule. None of the 46
immediate destinations below is syntactically a write onto the input receiver.
Several store into other hosts. Different identifiers do not prove disjoint
objects or that the receiver cannot reach the destination. I did not complete
that whole-program reachability analysis, did not assign speculative cycle
refusals, and added no refusal pins. Step 0 is therefore partial.

## The 46 destinations

All locations in this table are in the pinned `src/compiler/` source. Stored
means the immediate destination, including a field of an object later returned.
Fallback expressions can select an existing callback instead of making a bound
value. Passing to an emit method does not assert that its implementation cannot
retain the value.

| Site | Receiver and member | Destination |
|---|---|---|
| builder.ts:1727 | `host.writeFile` | passed to state.program.emitBuildInfo |
| builder.ts:1815 | `host.writeFile` | passed to affected.emitBuildInfo |
| builder.ts:1914 | `host.writeFile` | returned from getWriteFileCallback |
| checker.ts:54264 | `host.getSymlinkCache` | stored in returned ModuleSpecifierResolutionHost literal |
| checker.ts:54275 | `host.getGlobalTypingsCacheLocation` | stored in returned ModuleSpecifierResolutionHost literal |
| program.ts:494 | `system.createHash` | stored in compilerHost, returned from createCompilerHostWorker |
| program.ts:1956 | `host.getGlobalTypingsCacheLocation` | stored in program, returned from createProgram |
| program.ts:2626 | `host.realpath` | stored in EmitHost returned from getEmitHost |
| program.ts:2632 | `host.createHash` | stored in EmitHost returned from getEmitHost |
| program.ts:2634 | `host.trace` | stored in EmitHost returned from getEmitHost |
| program.ts:5114 | `directoryStructureHost.directoryExists` | stored in returned ParseConfigFileHost literal |
| program.ts:5115 | `directoryStructureHost.getDirectories` | stored in returned ParseConfigFileHost literal |
| program.ts:5116 | `directoryStructureHost.realpath` | stored in returned ParseConfigFileHost literal |
| tsbuildPublic.ts:302 | `system.now` | stored in host.now; host was made from system and is returned |
| tsbuildPublic.ts:438 | `host.resolveModuleNameLiterals` | stored in compilerHost.resolveModuleNameLiterals; compilerHost is stored in state |
| tsbuildPublic.ts:439 | `host.resolveTypeReferenceDirectiveReferences` | stored in compilerHost.resolveTypeReferenceDirectiveReferences; compilerHost is stored in state |
| tsbuildPublic.ts:440 | `host.resolveLibrary` | stored in compilerHost.resolveLibrary; compilerHost is stored in state |
| tsbuildPublic.ts:441 | `host.resolveModuleNames` | stored in compilerHost.resolveModuleNames; compilerHost is stored in state |
| tsbuildPublic.ts:442 | `host.resolveTypeReferenceDirectives` | stored in compilerHost.resolveTypeReferenceDirectives; compilerHost is stored in state |
| tsbuildPublic.ts:443 | `host.getModuleResolutionCache` | stored in compilerHost.getModuleResolutionCache; compilerHost is stored in state |
| tsbuildPublic.ts:500 | `host.trace` | stored in state.write; state also stores host and compilerHost |
| watch.ts:677 | `system.watchFile` | stored in returned WatchHost literal |
| watch.ts:678 | `system.watchDirectory` | stored in returned WatchHost literal |
| watch.ts:679 | `system.setTimeout` | stored in returned WatchHost literal |
| watch.ts:680 | `system.clearTimeout` | stored in returned WatchHost literal |
| watch.ts:761 | `host.getDefaultLibLocation` | stored in compilerHost, returned from createCompilerHostFromProgramHost |
| watch.ts:774 | `host.trace` | stored in compilerHost, returned from createCompilerHostFromProgramHost |
| watch.ts:775 | `directoryStructureHost.directoryExists` | stored in compilerHost, returned from createCompilerHostFromProgramHost |
| watch.ts:776 | `directoryStructureHost.getDirectories` | stored in compilerHost, returned from createCompilerHostFromProgramHost |
| watch.ts:777 | `host.realpath` | stored in compilerHost, returned from createCompilerHostFromProgramHost |
| watch.ts:778 | `host.getEnvironmentVariable` | stored in compilerHost, returned from createCompilerHostFromProgramHost |
| watch.ts:779 | `host.createHash` | stored in compilerHost, returned from createCompilerHostFromProgramHost |
| watch.ts:780 | `host.readDirectory` | stored in compilerHost, returned from createCompilerHostFromProgramHost |
| watch.ts:856 | `system.realpath` | stored in returned ProgramHost literal |
| watch.ts:857 | `system.getEnvironmentVariable` | stored in returned ProgramHost literal |
| watch.ts:861 | `system.createHash` | stored in returned ProgramHost literal |
| watch.ts:864 | `system.now` | stored in returned ProgramHost literal |
| watchPublic.ts:125 | `system.createHash` | stored in host.createHash; host was made from system and is returned |
| watchPublic.ts:506 | `host.useSourceOfProjectReferenceRedirect` | stored in compilerHost.useSourceOfProjectReferenceRedirect |
| watchPublic.ts:529 | `host.resolveModuleNameLiterals` | stored in compilerHost.resolveModuleNameLiterals |
| watchPublic.ts:530 | `host.resolveModuleNames` | stored in compilerHost.resolveModuleNames |
| watchPublic.ts:534 | `host.resolveTypeReferenceDirectiveReferences` | stored in compilerHost.resolveTypeReferenceDirectiveReferences |
| watchPublic.ts:535 | `host.resolveTypeReferenceDirectives` | stored in compilerHost.resolveTypeReferenceDirectives |
| watchPublic.ts:543 | `host.getModuleResolutionCache` | stored in compilerHost.getModuleResolutionCache |
| watchPublic.ts:549 | `host.hasInvalidatedResolutions` | stored in local customHasInvalidatedResolutions; nested closures use it |
| watchPublic.ts:552 | `host.hasInvalidatedLibResolutions` | stored in local customHasInvalidLibResolutions; nested closures use it |

## Existing stops inspected

No form was newly admitted and no refusal reason was changed.

- `internal/lower/refusals.go:282`: a method read as a value loses its object,
  with `(unbound-method)` in the fix. This remains applicable to declared
  interface-method arguments of maybeBind. A scratch reduction confirms it;
  this is not a whole-tsc run.
- `internal/lower/functions.go:190`: a function or method with a this parameter
  followed by another parameter (dynamic receivers are not implemented).
  This check is on declarations passing through signature(); the helper's
  callable type is not itself proof that this exact declaration check fires.
- `internal/lower/expression.go:1581`: a call through ?. (an optional call).
- `internal/lower/prototype.go:96`: inherited library member bind read as an
  own field (unbound-method). This is the generic prototype read diagnostic.

I read libraryMethodCall and libraryMethodBind in
`internal/lower/library_method_values.go`. Its captured-cell factory and
closureRecords registration are the relevant implementation precedent. No new
IR node, heap kind, side table, or allocation path was added.

## Unfinished work

All implementation and fixture work is unperformed: general function binding,
receiver snapshots, fixed arguments, identities, optional chains, escape and
throw ownership, bound length/name/new/rebind stops, runtime-key and typeof
stops, individual cycle pins, native/javascript/wasi differential fixtures,
and all four requested mutants. There is no mutant catcher or output line to
report. I did not mutate an unchanged compiler merely to produce a failure.

| Requested mutant | Status | Catching test and output |
|---|---|---|
| Evaluate fixed arguments at each call | not planted; bind lowering absent | none |
| Widen the same receiver to paths | not planted; new receiver rule absent | none |
| Drop the bound closure from closureRecords | not planted; new closure absent | none |
| Drop the optional bind nullish test | not planted; new optional bind lowering absent | none |

No new corpus fixture, test name, shape, compiler package or IR case was added.
Corpus walkers, skip census, product order, shape inventory and parallel census
therefore receive no new compiler or fixture input from this report.

## Commands and validation

All Go commands source `/workspace/adamic-tools/env.sh`. Setup and builds set
`GOPROXY=https://proxy.golang.org|direct`. Output is saved to files, never piped
into head or tail.

| Command | Result |
|---|---|
| `git fetch origin` | exit 0 |
| `git checkout -b apple/bind-1 origin/main` | exit 0, branch checked out |
| `git rev-parse HEAD` | base recorded above |
| `git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/apple-bind-typescript` | exit 0, pin verified by rev-parse |
| Python regex inventory over pinned src/compiler | 46 matching two-argument calls, three writeFile |
| Node TypeScript 6.0.3 syntax-tree inventory | exit 0; 46 matching calls, three writeFile, 43 others, one direct optional method bind |
| `git diff --check` | exit 0 before staging |

Inspection used `cat`, `sed`, `rg`, `rg --files`, `nl`, `wc`, `ls`, `pwd`,
`command -v`, `du`, `ps`, `free`, and `git status`, `git branch`, `git diff`,
`git rev-parse`, `git show`, `git config` against the instructions, language/memory docs, ruling,
setup/meter/apply scripts, source pin, compiler lowerers, fixture harness,
counts docs, gate tests and upstream call contexts. These were read-only.
An initial `cat CLAUDE.md` from /workspace failed; it was then read from
/workspace/adamic before setup or changes. Some path probes named files absent
on this base, including stage3/adapt/00-setup/apply.sh,
internal/lower/function_values.go, internal/oracle/array_union_test.go,
internal/oracle/counts.md and cloud/fast-gate/README.md. The current counts
layout is documented in internal/oracle/testdata/counts/README.md.
The first environment execution attempt reported no usable environment after
startup; retry succeeded. Runtime/network status and the non-secret policy
were read under the cloud runtime skill before network work.

Scratch probes, with source files only under `/tmp/apple-bind-probes`:

| Command | Result |
|---|---|
| `go run ./cmd/adamic c /tmp/apple-bind-probes/helper.ts` | exit 1, unbound-method at line 6 |
| `go run ./cmd/adamic c /tmp/apple-bind-probes/signature.ts` | exit 1, this-parameter signature stop at line 1 |
| `node /tmp/apple-bind-probes/helper.ts` | exit 0, `ok` |
| `node /tmp/apple-bind-probes/signature.ts` | exit 0, `ab` |
| `go run ./cmd/adamic c /tmp/apple-bind-probes/property.ts` | exit 1, optional-call stop at line 3 |
| `node /tmp/apple-bind-probes/property.ts` | exit 0, `ok` |

The first reduction declares Host.m as a method, defines a typed maybeBind with
`return fn?.bind(obj)`, then calls `maybeBind(host, host.m)`. The refusal is:

```text
adamic: /tmp/apple-bind-probes/helper.ts:6:31: Adamic 0.1 refuses a method read as a value (m would lose its object, and this with it); call it in an arrow that keeps the object: () => host.m() (unbound-method)
```

The second declares `function f(this: { name: string }, text: string): string`
and calls `f.bind({ name: "a" }, "b")`. The refusal is:

```text
adamic: /tmp/apple-bind-probes/signature.ts:1:12: stage 0 can't lower a function or method with a this parameter followed by another parameter (dynamic receivers are not implemented) yet
```

The property reduction changes Host.m to the current adaptation's
`m?: { m(): string; }["m"] | undefined` spelling and otherwise keeps the helper
probe's source. Its refusal is:

```text
adamic: /tmp/apple-bind-probes/property.ts:3:12: stage 0 can't lower a call through ?. (an optional call) yet
```

These are observations of the baseline, not new pins or mutant results. Logs
are `/tmp/apple-bind-{helper,signature,property,helper-node,signature-node,property-node}.log`.
The reductions do not substitute for running the complete adapted tsc program.
Read-only process/resource checks also used `df` and `/sys/fs/cgroup` CPU and
memory limits during the cold build. No resource configuration was changed.

Baseline checks completed so far:

| Command | Result and log |
|---|---|
| `bash cloud/setup.sh` | exit 0; 1859.649 seconds; `/tmp/apple-bind-setup.log` |
| `go build ./... && go vet ./...` | exit 0; `/tmp/apple-bind-build-vet.log` |
| `go test ./internal/lower/ ./internal/javascript/ ./internal/ir/ ./internal/fresh/ ./internal/flow/ ./internal/split/` | exit 1; lower, flow and split each timed out after ten minutes; javascript, ir and fresh passed; `/tmp/apple-bind-packages.log` |
| `go test -parallel 2 ./internal/native/` | exit 1; timed out after ten minutes with TestProduct_SplitTSGoSplit running; `/tmp/apple-bind-native.log` |
| Python report generation/edits and `git add notes/bind-1/REPORT.md` | exit 0 |
| `git diff --cached --check` | exit 0 |

The timeout output line in each failed package is
`panic: test timed out after 10m0s`. The native package's pending product was
`TestProduct_SplitTSGoSplit (5m59s)`. Lower's pending tests included
TestArrayNarrowingBooleanCallbackAgrees, TestArrayNarrowingPreservedUnionAgrees,
TestAsyncGapsNameTheMissingPiece, TestClassFeaturesAccessorCaptureCycle and
TestDecoderGetterBesideCheckedNodeViewAgrees.

Setup warming, build/vet and initial package/probe commands overlapped during
this cold setup. Setup reported a final one-minute load of 31.80 on a four-CPU
quota. These are partial baseline attempts, not a final isolated passing gate.
I did not prepare the full product waves or rerun the timeouts with longer
limits after stopping before implementation. No compiler regression is claimed
from those timeouts; no compiler source was changed.

Final baseline results:

| Command | Result and log |
|---|---|
| `go test ./internal/oracle/ -timeout 30m` | exit 1; timed out after thirty minutes, with earlier failures; `/tmp/apple-bind-oracle.log` |
| `go test ./cmd/adamic-gate/ ./internal/compilercensus/ ./internal/gatesample/ ./cloud/` | exit 0, all four packages passed; `/tmp/apple-bind-gate-rules.log` |
| `go test ./internal/oracle/ -run TestCountsAreRecorded -count=1 -timeout 30m` | exit 1; missing upstream input and invalid empty measurement rows; `/tmp/apple-bind-counts.log` |
| `go run ./internal/compilercensus/cmd/dependencies` | exit 0; `/tmp/apple-bind-compiler-dependencies.json`, stderr in `/tmp/apple-bind-census.log` |
| Python exact and two-way comparison with `cloud/fast-gate/compiler-dependencies.json` | exit 1, differs; `/tmp/apple-bind-census-compare.log` |
| Read-only post-timeout process checks | no live orphan fixture process identified; one defunct clang process remained |

Oracle failure examples:

- TestAssignmentProofWiderTarget and its mutant: `Cannot find module 'typescript'`.
  The tool copy exists, but bare module resolution for that harness was not
  configured. The syntax-tree scout loaded the tool copy by absolute path.
- TestStdoutColumnsRefusesWASI: `native: wasm32-wasi requires WASI_SYSROOT`.
- TestFamiliesCatchAPlantedCase_Stage0EmitsDeterministically and
  TestFamiliesCatchAPlantedCase_CountsAreRecorded: their probe was killed at
  ninety seconds.
- The package ended with `panic: test timed out after 30m0s`. Pending tests
  included TestTheJobserverNeverReachesTheEmittedCOrClang (11m26s), stack-input
  tests, and TestDefaultTaggedSourceViewsDefaultBoxedWrite.

Counts failures include `ADAMIC_TYPESCRIPT_SOURCE must name the pinned upstream
TypeScript checkout; Program sources never fetch` and
`invalid runtime count row for internal/oracle/testdata/async_counts/async_counts_captured.a: ""`.
I omitted ADAMIC_TYPESCRIPT_SOURCE from those baseline invocations even though
the scout clone was available. I did not finish gate preparation or rerun the
whole suite after deciding to stop before implementation. These failures are
not evidence of a bind change. No count records were regenerated or changed.
Read-only checks of two affected committed record paths confirmed they are
nonempty on both the base and working tree; the empty measurements are not a
claim that their committed files are empty.

The census comparison found 34 recorded package entries and 100 measured ones:
zero expected-only package keys, 66 actual-only keys, and six shared entries
with different inputs. Those six are internal/boundedrun, internal/ccjobserver,
internal/compilercensus, internal/leakcheck, internal/lowered and stage3/lane.
Both input-set directions were compared, including the expected-only
stage3/lane/adaptedtree input of stage3/lane. The complete differences are in
/tmp/apple-bind-census-compare.log. I did not change the gate table to hide this
baseline discrepancy.

No new fixture, count row or compiler dependency was added by this head.
The only committed file is this report. Commit and push outcomes, and the
resulting head hash, are reported separately after the report is committed.
No Loom command was run.
