| Site | Road | Target gone? | New / gone / changed findings | typescript.d.ts diff lines, adapted / pristine | tsserverlibrary.d.ts diff lines | Calls added per lane run |
|---|---|---|---|---|---|---|
| 3 | Local readonly view | Yes | 0 / 1 / 0 | 0 / 0 | 0 | **0** |
| 3 | Public System.writeFile property | Yes | 0 / 1 / 0 | 2 / 2 | 0 | **0**, erased |
| 10 | Local fresh forwarding host | Yes | 0 / 1 / 0 | 0 / 0 | 0 | **0**, entry unexecuted |
| 10 | Public ModuleResolutionHost.directoryExists arrow property | No | 0 / 0 / 0 | 6 / 2 | 0 | **0**, erased |
| 10 | Public System.directoryExists property, alternate interpretation | No; refusal changes | 0 / 0 / 1 | 2 / 2 | 0 | **0**, erased |

Diff lines count removed plus added lines, excluding context/headers. Findings are unique `(kind, where, reason, text)` entries; changed findings are paired at the same projected source location rather than counted again as new/gone. Each road is measured independently against the requested main, not cumulatively. Both possible site 10 public declarations are priced because adapted ModuleResolutionHost.directoryExists is already a property using an indexed method type.

Only this report and measurement evidence are committed. No adaptation or production compiler change is included.

## Site 3: preserve the captured value

The owner is `patchWriteFileEnsuringDirectory`, adapted sys.ts:1383. It reads sys.writeFile once at 1385:31, immediately reassigns sys.writeFile, then calls the captured function through `originalWriteFile.call(sys, path, data, writeByteOrderMark)`. The captured function is not identity-compared or otherwise exposed. Its receiver is explicitly sys.

The [local diff](local3-sys.ts.diff) is:

```ts
const writeFileView: { readonly writeFile: System["writeFile"]; } = sys;
const originalWriteFile = writeFileView.writeFile;
```

The remaining code is unchanged. The view is the same object, not a copy. The getter is read once with the same receiver before the same reassignment; the exact function object, including its own .call property, is retained. It adds one local alias assignment per capture, no function call or wrapper, and no function identity change. The lane executes 2,200 captures and 4,627 callback invocations in both control and candidate. Thus the measured extra work is 2,200 local reference assignments and **zero added calls**. [The witness](readonly-view-witness.json) covers receiver-sensitive calls, a getter, later reassignment, identity, and an own .call override.

A naive arrow that reads live sys.writeFile recurses into the replacement. Bind captures the right receiver/function at the right moment, and its census also removes the target, but creates a different function and changes .call lookup. A function with an own .call gives different results: [counterexample](own-call-witness.json). Bind is retained only as a rejected trial, not the value-identical road.

## Site 10: isolate the writable host view

The [local diff](local10-sys.ts.diff) passes a fresh host:

```ts
resolveJSModule(moduleName, baseDir, {
    fileExists: (path: string) => nodeSystem.fileExists(path),
    readFile: (path: string) => nodeSystem.readFile(path),
    directoryExists: (path: string) => nodeSystem.directoryExists(path),
    realpath: nodeSystem.realpath ? (path: string) => nodeSystem.realpath!(path) : undefined,
    getCurrentDirectory: () => nodeSystem.getCurrentDirectory(),
    getDirectories: (path: string) => nodeSystem.getDirectories(path),
    useCaseSensitiveFileNames: nodeSystem.useCaseSensitiveFileNames,
});
```

The current adapted realpath slot admits explicit undefined. Each arrow retains nodeSystem as the receiver. A writable slot on this new host does not alias System's method slot. The temporary host and its callback identities differ; nodeSystem's members are not reassigned. Each forwarder invocation would add one call. Construction would create one object and five or six arrow values per require entry. Those allocations are not counted as calls.

**The lane never enters this nodeSystem.require resolution path.** Its entry counter and all six forwarding counters are zero in control and candidate. The price of zero calls is therefore specific to this lane, not a zero-overhead claim for arbitrary module loading. No general equivalence is claimed for external mutation/getters or reflection on a substituted host; the full lane comparison below is the measured observation boundary.

The discarded partial-realpath-spread form removed the target but added a refusal about optional trace hidden by a structural source. Its raw stream is retained separately; it is not this road.

## Public declaration roads

All experiments use upstream's locked npm ci followed by node_modules/.bin/hereby dts. Separate pristine-upstream and adapted controls are retained. Every actual stock tsc 6.0.3 CLI check passes the same **79 compiler roots**, with strict, exactOptionalPropertyTypes, noUncheckedIndexedAccess, ES2024/ESNext/Bundler, no emit, and pinned Node ambient types. Exact configs and commands are adjacent to the report.

| Public change | Exact pristine diff | Exact adapted diff |
|---|---|---|
| System.writeFile method → `(path, data, writeByteOrderMark?) => void` property | [typescript.d.ts](site3-typescript.d.ts.diff) | [typescript.d.ts](declaration3-adapted-typescript.d.ts.diff) |
| ModuleResolutionHost.directoryExists → `directoryExists?: ((directoryName: string) => boolean) \| undefined` | [typescript.d.ts](site10-typescript.d.ts.diff) | [typescript.d.ts](declaration10-adapted-typescript.d.ts.diff) |
| System.directoryExists → `directoryExists: (path: string) => boolean` | [typescript.d.ts](declaration10-system-upstream-typescript.d.ts.diff) | [typescript.d.ts](declaration10-system-adapted-typescript.d.ts.diff) |

All corresponding tsserverlibrary.d.ts diffs are empty files. That shim reexports TypeScript, so byte identity does not insulate its consumers from the changed typescript.d.ts API. [Stock erasure](declaration-erasure.json) gives byte-identical JavaScript for each changed source file: no executable call is added by these declaration changes. Declaration variants were measured with dts, stock tsc and the full census; their zero inserted-call price follows erasure, rather than an additional oracle suite per type-only variant.

The System.writeFile property removes site 3. The two site 10 declaration changes do not solve the mutable-view issue:

- Adapted ModuleResolutionHost.directoryExists starts as `directoryExists?: { directoryExists(directoryName: string): boolean; }["directoryExists"] | undefined`. Replacing its indexed method type with an arrow leaves the target refusal unchanged.
- Making System.directoryExists a property changes the refusal to: `a value of type System seen as ModuleResolutionHost, which can write ((directoryName: string) => boolean) | undefined where (path: string) => boolean is read`. The optional view can still write undefined over a required function. The refusal remains at 1612:77.

Public arrow-property declarations also change declaration shape and method variance. Passing these roots does not establish compatibility for all external consumers. At least the lane's public-API baseline case necessarily observes the changed generated declaration text; the exact additional hunks are above. No public type change is needed by either local road.

## Full sys.ts census

The compiler source is adamic **8e3a784aa5a601f57342c07b68f145fe56252c21**. The private detached worktree initialized recursive shallow submodules. Upstream v6.0.3 is **050880ce59e30b356b686bd3144efe24f875ebc8**. stage3/api npm ci and unmodified stage3/apply.sh succeeded with ADAMIC_TYPESCRIPT_SOURCE pointing to that checkout. Temp directories, Go/module caches and npm caches were private scratch. [Provenance](provenance.json) records pins and tool identities.

Every full-file run uses the same unmodified main census generator/binary and:

```text
LATENT_FULL=1
LATENT_ASSERT_NO_OUTPUT=1
CENSUS_MODULE_FORMAT=CommonJSBundle
LATENT_ONLY_FILE=<private tree>/src/compiler/sys.ts
latent <private tree>/src/compiler <raw.jsonl>
```

All headers report `diagnostics: []`, `checker_rejected: false`, full mode and CommonJSBundle. The tool's historical generic “measured on a checker-rejected program” label does not override those headers. Both no-output guards run. This is a refusal/lowering census, not successful native compilation.

| Input | NotYet | Refused | Unique total | Exit | Census seconds |
|---|---:|---:|---:|---:|---:|
| Main control | 41 | 18 | 59 | 0 | 350.052 |
| Site 3 readonly view | 41 | 17 | 58 | 0 | 107.407 |
| Site 3 public property | 41 | 17 | 58 | 0 | 165.422 |
| Site 10 fresh host | 41 | 17 | 58 | 0 | 138.431 |
| Site 10 public ModuleResolutionHost arrow | 41 | 18 | 59 | 0 | 106.565 |
| Site 10 public System property | 41 | 18 | 59 | 0 | 117.364 |

Elapsed times include differing machine load and are not performance comparisons. [Finding deltas](finding-deltas.json) preserve every new/gone/changed entry. [Full projected comparison](evidence/comparison.json.gz), raw streams, timing ledgers and [run flags](census-runs.json) are retained. Coordinate projection aligns unchanged source lines; raw observations retain actual locations, units, phases and attempt context. Boundary recorder events are excluded from unique compiler-finding totals.

## Lane observations and call counts

The unmodified stage3/oracle/run.sh ran on fresh counting copies, all runners, eight workers, no test filter, Node v24.19.0, sequentially without concurrent compiler/census work. [Observer diffs](observer-control.diff) and [executed helper](helpers/lane-pricing.py.txt) show the instrumentation. Counters append synchronously from every worker; observer calls are excluded from road overhead. Counts cover the oracle's install/build/test run. [Fixture audits](fixture-audit-control.json) verify all 32 tracked files under nested node_modules/built directories.

| Run | Passing | Failing | Pending | Site 3 captures | Site 3 callback calls | Site 10 resolution entries | Forwarder calls added |
|---|---:|---:|---:|---:|---:|---:|---:|
| Control | 106,366 | 1 | 0 | 2,200 | 4,627 | 0 | 0 |
| Readonly local view | 106,366 | 1 | 0 | 2,200 | 4,627 | 0 | 0 |
| Fresh local host | 106,366 | 1 | 0 | 2,200 | 4,627 | 0 | 0 |

Every sole failure is `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`; every sole baseline diff is api/typescript.d.ts. Install/build exit 0 and tests/oracle exit 1, as required by the lane's existing exception. The three baseline.diff files are **byte-identical**. [Both generated declaration files](local-declarations-identical.json) are **byte-identical** across the local roads and control. There is no new lane assertion or baseline observation for either local road. The public declaration roads necessarily change the API artifact described above.

[Counts and full phase commands](lane-counts.json), raw per-worker counter streams and full oracle reports/logs are retained in evidence/. Earlier fixture-incomplete or resource-contended runs are explicitly [disqualified](DISCARDED-RUNS.md); none supplies a headline count.

## Readers and integrity

[READERS.md](READERS.md) lists all 43 checker-resolved references to the three public members under consideration: 27 System.writeFile, eight ModuleResolutionHost.directoryExists and eight System.directoryExists, including writes. It also lists three implicit spread readers in harness/test code. [The full 526-entry named inventory](readers-all-public.json) retains overrides, homonyms and unresolved receivers across all 712 adapted src TypeScript/TSX roots. [Computed/rest/reflection audit](readers-dynamic.json) and [all literal mentions](evidence/readers-all-text.txt.gz) supplement it, including type references and test strings.

The local capture adds only the shown readonly-view read; the local host adds the shown forwarding reads. Existing public member declarations and values remain unchanged by these local edits. Captures/restores elsewhere in program/watch/harness helpers therefore retain their original function values. The declaration roads erase at runtime but change the public type/API artifact. The full oracle establishes the local lane observation boundary; it does not expand site 10 coverage beyond its measured zero entries.

[Source integrity](source-integrity.json) verifies that the baseline source stayed byte-identical and each private experiment changes only its intended sys.ts or types.ts. No production source is changed in this branch. Helpers are retained as audit text with this run's fixed scratch paths; the private diffs and pinned setup identify the experiment inputs. [SHA256SUMS](SHA256SUMS) covers the final report/evidence files.
