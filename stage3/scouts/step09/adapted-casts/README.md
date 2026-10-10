# The 31 casts added by adaptations

Measured population: ledger 855bcfaa and classification ea1b2359. These are adapted-source coordinates, not stock coordinates. Reconstruction uses the ledger commit's main-45487a80 pipeline, adaptation 43 from 643639ea, and its documented timer/71 compositions. No current-main adaptation is merged into that measurement. The delivery branch starts at origin/main 031a1259.

All 31 casts are attributed to the immediate before/after adaptation that first introduced their exact AST text. Final UTF-16 start/end spans must match the pinned ledger; the independent outside-stock classifier must have the identical location/text population. Source hashes and introduction counts are in evidence/source-measurement.json.

No row is newly certified as an Adamic upcast. The upstream API accepts the 19 split/helper identity candidates, but its compiler project options and library differ from Adamic: exact optional-field admission and the sound regex split contract matter. The original Adamic castProof refusals remain recorded beside each source/target type. A printable type equality is not a dual-proof upcast.

## Decisions

| Decision | Sites |
| --- | ---: |
| construction proof or checked view | 1 |
| root correlation proof or checked view | 2 |
| truthful builtin specialization | 4 |
| generic write and construction proof | 2 |
| truthful internal declaration | 17 |
| truthful throwing host binding | 2 |
| correlated generic protocol proof | 3 |

There are 21 virtual type-only proposals: 15 helper call views, four split casts and two nonempty-array consumers. The measurement checks the full adapted closure for zero upstream semantic diagnostics, and byte-identical emitted JavaScript in every changed file. It removes erased assertion parentheses as well as each assertion where required to preserve emitted bytes. These checks do not claim native admission. In particular, the four split sites still need Adamic's capture-free builtin proof (or checked element views); globally changing regex split to string[] is false.

The remaining ten rows cannot be discharged by merely annotating the current expression. Completed allocations and object roots need field/root proof or checked refinement. Generic writable views need the original specialization's checked-write/construction proof; a wide base constraint does not restrict narrower T. Watcher calls need key/callback/flag correlation, which checking whether a value is a function cannot establish. The fs sites require a truthful throwing host binding; accepting undefined must preserve Node's TypeError rather than substitute a string.

## Every adapted location

| ID | Adapted file:line:column | Adaptation | Decision |
| --- | --- | --- | --- |
| 01 | src/compiler/checker.ts:14147:20 | stage3/adapt/41-explicit-any-remaining/ | construction proof or checked view |
| 02 | src/compiler/commandLineParser.ts:2461:117 | stage3/adapt/43-any-returns/ | root correlation proof or checked view |
| 03 | src/compiler/commandLineParser.ts:2466:108 | stage3/adapt/43-any-returns/ | root correlation proof or checked view |
| 04 | src/compiler/emitter.ts:4105:54 | stage3/adapt/45-regex-captures/ | truthful builtin specialization |
| 05 | src/compiler/emitter.ts:4976:46 | stage3/adapt/45-regex-captures/ | truthful builtin specialization |
| 06 | src/compiler/factory/nodeFactory.ts:1216:15 | stage3/adapt/60-temporary-factory-local-symbol/ | generic write and construction proof |
| 07 | src/compiler/factory/nodeFactory.ts:5496:15 | stage3/adapt/61-temporary-factory-type-expression/ | generic write and construction proof |
| 08 | src/compiler/moduleNameResolver.ts:670:66 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 09 | src/compiler/moduleNameResolver.ts:1625:90 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 10 | src/compiler/moduleNameResolver.ts:1644:94 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 11 | src/compiler/moduleNameResolver.ts:1668:68 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 12 | src/compiler/moduleNameResolver.ts:1993:43 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 13 | src/compiler/moduleNameResolver.ts:2008:58 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 14 | src/compiler/moduleNameResolver.ts:2136:60 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 15 | src/compiler/moduleNameResolver.ts:2468:54 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 16 | src/compiler/moduleNameResolver.ts:2528:86 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 17 | src/compiler/moduleNameResolver.ts:2529:88 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 18 | src/compiler/moduleNameResolver.ts:3066:62 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 19 | src/compiler/moduleNameResolver.ts:3081:66 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 20 | src/compiler/moduleNameResolver.ts:3159:95 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 21 | src/compiler/moduleNameResolver.ts:3192:98 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 22 | src/compiler/moduleNameResolver.ts:3336:58 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 23 | src/compiler/moduleSpecifiers.ts:590:76 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 24 | src/compiler/moduleSpecifiers.ts:1198:76 | stage3/adapt/32-indexed-reads-program/ | truthful internal declaration |
| 25 | src/compiler/parser.ts:10784:43 | stage3/adapt/45-regex-captures/ | truthful builtin specialization |
| 26 | src/compiler/semver.ts:287:64 | stage3/adapt/45-regex-captures/ | truthful builtin specialization |
| 27 | src/compiler/tracing.ts:338:17 | stage3/adapt/62-temporary-tracing-write/ | truthful throwing host binding |
| 28 | src/compiler/tracing.ts:357:13 | stage3/adapt/63-temporary-tracing-legend/ | truthful throwing host binding |
| 29 | src/compiler/watchUtilities.ts:736:27 | stage3/adapt/32-indexed-reads-program/ | correlated generic protocol proof |
| 30 | src/compiler/watchUtilities.ts:811:39 | stage3/adapt/32-indexed-reads-program/ | correlated generic protocol proof |
| 31 | src/compiler/watchUtilities.ts:818:21 | stage3/adapt/32-indexed-reads-program/ | correlated generic protocol proof |

## Exact type and smallest proposal by family

### stage3/adapt/41-explicit-any-remaining/: construction proof or checked view

Rule 30 changes the allocator to an incomplete AllocatedSignature and asserts completion after field assignments.

Truthful type: `AllocatedSignature & Pick<Signature, "parameters" | "minArgumentCount">`.

The constructor must remain AllocatedSignature. After parameters and minArgumentCount are assigned, the completed value has AllocatedSignature & Pick<Signature, "parameters" | "minArgumentCount"> (a Signature). A local annotation claiming Signature at allocation would lie. Require definite field initialization proof or a checked completion view; do not change constructor return to Signature.

The source optionalizes required fields; it is not assignable to Signature. The observed assignments establish completion semantically, but a source-only annotation does not carry a checker proof.

Checked completion/refinement view or an equivalent static construction/root proof; no unchecked assertion.

### stage3/adapt/43-any-returns/: root correlation proof or checked view

Concrete recursive recovery return replaced any; the config converter asserts its root conversion is an object.

Truthful type: `AdamicJsonRecoveryObject`.

An object-root overload can return AdamicJsonRecoveryObject when rootExpression: ObjectLiteralExpression | undefined and returnValue: true. Keep the general return AdamicJsonRecoveryValue | undefined. firstObject is an object; rootExpression is only Expression | undefined after the kind test, so the second site also needs a checker-recognized root proof or checked result view. Do not globally give convertToJson an object return.

Primitive and array roots legitimately return primitive/array recovery values. These two callers constrain roots to object or absent, and returnValue is true; the generic converter signature alone cannot prove that relationship.

Checked completion/refinement view or an equivalent static construction/root proof; no unchecked assertion.

### stage3/adapt/45-regex-captures/: truthful builtin specialization

Capture-free split separator produces a dense array of strings; avoid optional capture elements.

Truthful type: `string[]`.

Remove this redundant cast under the upstream String.split contract. Under Adamic the capture-free separator needs a builtin proof returning string[] at this site; do not globally narrow regex split, which can insert undefined captures.

Upstream resolved source and target are string[] and assignable. Adamic probe refused. Semantic density follows only for the reviewed capture-free separator; native builtin specialization remains required.

Capture-free builtin proof can eliminate the cast; otherwise an array element checked view is required under the sound split library.

### stage3/adapt/60-temporary-factory-local-symbol/: generic write and construction proof

Permit writing undefined to an optional generic field without claiming presence in the returned T.

Truthful type: `Omit<Mutable<T>, "localSymbol"> & { localSymbol: undefined }`.

After this write the truthful field is localSymbol: undefined; the local post-write shape is Omit<Mutable<T>, "localSymbol"> & { localSymbol: undefined }. It cannot be returned as every T: a T may require a Symbol. An undefined-admitting base constraint alone does not exclude that subtype. Check writes against the original specialization and prove initialization at callers, or change the private factory result protocol; a checked cast alone cannot make the returned T true.

The receiver view loses T-specific field restrictions. The wide view can write a value that a narrower T cannot hold; this is not an upcast proof.

A checked cast alone is insufficient: the original specialization, key protocol or host implementation must also be proved.

### stage3/adapt/61-temporary-factory-type-expression/: generic write and construction proof

Permit the optional typeExpression input to be written through the generic factory receiver.

Truthful type: `Omit<Mutable<T>, "typeExpression"> & { typeExpression: JSDocTypeExpression | undefined }`.

Truthful post-write shape: Omit<Mutable<T>, "typeExpression"> & { typeExpression: JSDocTypeExpression | undefined }. Widening the base constraint to admit undefined does not prove all T accept it: a T can require a particular expression. Require checked writes for the actual specialization plus a construction/caller proof, or change the private factory result protocol. Do not claim Mutable<T> without that proof.

Writing through the erased generic restriction needs more than checking property presence; the original specialization governs what values may be stored.

A checked cast alone is insufficient: the original specialization, key protocol or host implementation must also be proved.

### stage3/adapt/32-indexed-reads-program/: truthful internal declaration

Admit explicitly undefined directoryExists at the local helper call without editing the utilities owner.

Truthful type: `(directoryName: string, host: { directoryExists?: ((directoryName: string) => boolean) | undefined; }) => boolean`.

Change directoryProbablyExists owning host parameter to { directoryExists?: ((directoryName: string) => boolean) | undefined; }; remove this call view and its erased assertion parentheses. Its unchanged presence test handles explicit undefined. Cross-owner type-only change.

Upstream assignability is true, but exactOptionalPropertyTypes changes callable parameter variance. The ledger records Adamic refusing; this is not a proven Adamic upcast.

No cast needed after the verified truthful declaration proposal.

### stage3/adapt/62-temporary-tracing-write/: truthful throwing host binding

Allow JSON.stringify result string | undefined at a Node fs boundary that throws for undefined.

Truthful type: `{ writeSync(fd: number, data: string | undefined): number; }`.

The target callable signature is truthful only for a binding that accepts string | undefined and preserves Node ERR_INVALID_ARG_TYPE/TypeError for undefined. Add that typed throwing boundary contract; do not cast a string-only callable to one accepting undefined or replace undefined with a string. A callable checked cast cannot prove external function behavior. Native Node binding/exception support remains required.

Contravariant parameter widening is not a safe upcast; Node accepts the invocation and throws for undefined data. The source adaptation supplies no native implementation.

A checked cast alone is insufficient: the original specialization, key protocol or host implementation must also be proved.

### stage3/adapt/63-temporary-tracing-legend/: truthful throwing host binding

Allow JSON.stringify result string | undefined at a Node fs boundary that throws for undefined.

Truthful type: `{ writeFileSync(path: string, data: string | undefined): void; }`.

The target callable signature is truthful only for a binding that accepts string | undefined and preserves Node ERR_INVALID_ARG_TYPE/TypeError for undefined. Add that typed throwing boundary contract; do not cast a string-only callable to one accepting undefined or replace undefined with a string. A callable checked cast cannot prove external function behavior. Native Node binding/exception support remains required.

Contravariant parameter widening is not a safe upcast; Node accepts the invocation and throws for undefined data. The source adaptation supplies no native implementation.

A checked cast alone is insufficient: the original specialization, key protocol or host implementation must also be proved.

### stage3/adapt/32-indexed-reads-program/: correlated generic protocol proof

Preserve the key-specific watcher callback/flags/arguments protocol across union-typed .call receivers.

Truthful type: `WatchFactory<X, Y>[K]`.

Carry K extends keyof WatchFactory<X, Y> through the selected receiver WatchFactory<X, Y>[K] and its argument tuple Parameters<WatchFactory<X, Y>[K]>. The callback is Parameters<WatchFactory<X, Y>[K]>[1]; forwarded eventArgs are Parameters<Parameters<WatchFactory<X, Y>[K]>[1]>. Instantiate separately for watchFile and watchDirectory. These are exact correlated types, not a single function that accepts arbitrary callback/flag combinations. A typed local alone does not discharge the generic .call body; requires correlation proof or code specialization.

A union of key-specific callables cannot safely accept every combination of unioned parameters. Runtime callable/signature checking cannot recover the missing key correlation by itself.

A checked cast alone is insufficient: the original specialization, key protocol or host implementation must also be proved.

The JSON repeats the full exact cast expression, source type, target type, original probe, truthful type and explanation for every row. For JSON recovery, the aliases are recursive: AdamicJsonRecoveryValue = string | number | boolean | null | AdamicJsonRecoveryValue[] | AdamicJsonRecoveryObject; AdamicJsonRecoveryObject has [key: string]: AdamicJsonRecoveryValue | undefined. Object-only results apply to true returnValue and object/absent roots, not arbitrary JSON.

The post-write Omit shapes describe the field effect relative to the assumed base shape. They do not certify that a partially initialized base node already satisfies every other member of T, or that public factory contracts can be changed without review. The Signature completed shape retains the truthful partial constructor allocation, then makes parameters and minArgumentCount required; the constructor must not claim those fields exist before their writes.

## Fixtures, mutants and limits

These are three reduced witnesses, retaining the exact adapted cast/write/call expression. They do not copy full emitter/factory/tracing function implementations. Support types, driver setup, and the tracing catch are reduced or added to isolate observations. Source comments name adapted TypeScript 6.0.3 spans.

| Fixture | Node observation | Source mutant caught by independent golden | Current native result |
| --- | --- | --- | --- |
| 01_capture_free_split.a | Three newline forms produce ["a","b"]; empty input produces [""] | Optional capture separator introduces extra/undefined array elements | Refused unchecked cast |
| 02_generic_field_write.a | Identity retained; expression written then a present undefined field through the alias | Replace the expression write with undefined | Refused unchecked cast |
| 03_tracing_undefined_data.a | JSON object data writes; undefined data throws TypeError | Replace undefined with an empty string, allowing the second call to continue | Refused callable parameter widening |

The fs witness imports real node:fs with its full overload declarations; adaptation 65 instead gives the adapted tracing fs variable a six-operation local shape. Consequently its native refusal text mentions the real buffer overload, while the ledger source shape has the string overload. Both expose the same unsafe admission of undefined, and the exact diagnostics are retained separately.

All three source mutants still finish with exit 0 and empty stderr, and fail only the exact stdout golden. These are source semantic mutants, not removal of implemented native checks: all three native programs refuse before emission. Their truthful first-line refusal headers pass unchanged Gate.aCheck. Six additional mutants remove or replace each header; each fails that expected-result predicate. No native runtime, sanitizer result, or full tsc/scanner execution is claimed.

verify.py independently checks all 31 identities and measured owners, the 21 proposal records, Node observations, body hashes, and header results. --drop-row and --wrong-owner must fail the coverage and immediate-introduction assertions. measure.cjs --runtime-edit changes the actual virtual helper body; it must fail the emitted-byte guard.

## Reproduce

Fetch the ledger, casts, real-any and devtools/fast-gate refs. Add a detached worktree at 855bcfaa. Use new scratch paths and the upstream locked node_modules matching Node declarations 25.3.3. Every test command writes a log.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/adapted-casts-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund > /tmp/adapted-casts-api-install.log 2>&1
export NODE_PATH="$PWD/stage3/api/node_modules"
python3 /tmp/adapted-casts-ledger/stage3/step09-ledger/prepare.py /tmp/adapted-casts-measure --node-modules /path/to/upstream/node_modules --extra 643639ea:stage3/adapt/43-any-returns > /tmp/adapted-casts-prepare.log 2>&1
python3 stage3/scouts/step09/adapted-casts/classify.py > /tmp/adapted-casts-classify.log 2>&1
node stage3/scouts/step09/adapted-casts/measure.cjs /tmp/adapted-casts-measure > /tmp/adapted-casts-measure.log 2>&1
go build -o /tmp/adapted-casts-adamic ./cmd/adamic > /tmp/adapted-casts-build.log 2>&1
python3 stage3/scouts/step09/adapted-casts/observe.py /tmp/adapted-casts-adamic > /tmp/adapted-casts-observe.log 2>&1
git show fbac28c62493f27a788edc02a18bc8edb68de5da:cloud/fast-gate/run.py > /tmp/adapted-casts-gate.py
python3 stage3/scouts/step09/adapted-casts/a-check.py /tmp/adapted-casts-gate.py /tmp/adapted-casts-a-check > /tmp/adapted-casts-a-check.log 2>&1
python3 stage3/scouts/step09/adapted-casts/verify.py > /tmp/adapted-casts-verify.log 2>&1
```

Setup: Go ready 0.023s, Node 0.022s, submodules 0.058s, markdown 0.064s, clang 0.171s, Go build 46.801s, deferred test binaries 46.961s, warm cache 46.962s, done 46.995s. nproc 5; CPU quota 4. Go 1.27.1, clang 20.1.8, Node 24.19.0. Initial reconstruction stopped because 643639ea was not fetched; the retry used a new output after fetching it. The builtin fixture initially hit missing local Node declarations; installing the existing stage3/api lock resolved that setup dependency.

Adaptations and compiler source are unchanged. Internal/oracle counts are unchanged because these fixtures are scout inputs outside its compiled registry; local counts.md is refreshed. No whole package tests or full gate ran. Unchecked callable behavior, generic initialization and native split proofs remain compiler/host owner work.
