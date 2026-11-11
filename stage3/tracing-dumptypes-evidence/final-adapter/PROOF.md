# dumpTypes proof for dp24r5m

Base: efe887763f26e96653546fae904d57539139f7ae, origin/main at setup.
TypeScript: 050880ce59e30b356b686bd3144efe24f875ebc8, version 6.0.3.
Node: v24.19.0, Linux x64.


## Ownership

40 and 41 retain these reads because the fields belong to Type subtypes; their
classification explicitly declines a uniform unknown replacement. 62 owns only
the dumpTypes writeSync expression, 63 owns dumpLegend, and
99-correctness-tracing-assert pins stopTracing. None pins either exact read or
all of dumpTypes. The new adaptation pins the complete composed dumpTypes
before and after and changes its two subtype reads together.

## Step 1: every later use

The objectFlags local has exactly five later reads. Every read applies a mask:

| Original adapted line | Mask | Effect on the written descriptor |
| --- | --- | --- |
| 237 | Anonymous | Enables display computation, together with TypeFlags.Literal |
| 256 | Reference | instantiatedType, typeArguments and referenceLocation |
| 286 | ReverseMapped | reverseMappedSourceType, reverseMappedMappedType and reverseMappedConstraintType |
| 296 | EvolvingArray | evolvingArrayElementType and evolvingArrayFinalType |
| 321 | Tuple | isTuple |

The raw objectFlags value is never serialized. descriptor.intrinsicName has no
later property read: JSON.stringify(descriptor) consumes it in the existing
writeSync call. Undefined is omitted; a string is retained.

## Step 1: whole lane

All 398 real gate TestShard_ units pass with child exit 0. Their manifests
validate 106,366 passing oracle cases, one sanctioned API acknowledgement
failure and zero skipped cases. All 398 units have counter-loaded process
coverage, across 797 completed Node processes. dumpTypes executes **0 times**
and visits **0 types** in the entire lane. Written-value differences are **0**.
This is a zero count, not direct coverage of the changed function. The three
real generateTrace projects below supply that execution coverage.

The counted product is a private copy. Its complete inventory contains 82,870
files, all 747 source files are byte-identical to baseline, and its 10 normal
JavaScript files differ only by the explicit counter hooks in _tsc.js, run.js
and typescript.js. The overlay selects that copy and passes the VM-aware preload
through the gate's own filtered environment. Production Go remains unchanged.

## Step 1: actual adapted compiler with generateTrace

Node runs the counted lane compiler against these real upstream oracle projects:

- jsFileCompilation/DifferentNamesNotSpecified
- projectOption/DefaultExcludeOnlyNodeModules
- decoratorMetadata/emitDecoratorMetadataCommonJSIsolatedModule

The first two exit 0. The third exits 2 for missing angular2/core and an implicit
any return, and still writes its types file. These diagnostics are retained and
are not described as a clean project run.

Across three dumpTypes calls, 84,642 visited types have 50,995 own objectFlags
properties and 88 own intrinsicName properties. getObjectFlags differs from the
raw read 33,693 times: 33,647 undefined reads become 0, and 46 types carry real
objectFlags outside TypeFlags.ObjectFlagsType. Of those 46, 33 hold 52953088 and
13 hold 14680064. None sets any of the five masks above. Differences for each
mask are 0, 0, 0, 0, 0.

The Intrinsic flag guard has 0 intrinsicName differences, both at loop entry and
at descriptor construction after display and recursion work. The in read also
has 0 differences. There are 0 own intrinsicName properties and 0 defined values
outside TypeFlags.Intrinsic. All 88 flagged visits have string names; flagged
objectFlags reads have 0 missing or nonnumeric values.

Stock 6.0.3 controls repeat the three projects. The original and two-replacement
runs write byte-identical types files of 7,153,496, 6,227,528 and 6,229,391 bytes.
The descriptor-time counter is separately checked against a pristine compiler
in the same directory, again with exact byte identity. Complete individual
records, original types files, hashes, commands and child exits are retained.

## Step 1: stage 0

Both intrinsicName forms pass ordinary lowering and complete refusal scans in
isolated executable probes, with 0 diagnostics and 0 findings. The selected
flags road also passes all 79 actual compiler roots. Its complete tracing census
visits all 23 units with 0 diagnostics and removes all three unchecked-cast events
at the two original sites: two refusal-scan events and one lowering event.
Other existing NotYet findings remain and deeper findings become visible. This
is not a claim that all compiler roots lower to native code.

## Step 2

The selected replacements are getObjectFlags(type) and the Intrinsic flag
conditional followed by a checked IntrinsicType read. No unknown substitution,
new helper or compiler change is introduced. Only tracing.ts changes among all
747 source files: two imports and the two owned expressions. All 747 source
files from the fresh full pipeline match the normally built candidate exactly.

| Check | Result |
| --- | --- |
| Whole-owner/import contract | 15 cases: 4 accepted with zero-write replay, 11 refusals before any write |
| LF/CRLF and unrelated text/imports | Preserved |
| Baseline and fresh complete apply | Both child exits 0 |
| Normal builds before and after | Both child exits 0 |
| All 79 roots, baseline and fresh candidate | Both child exits 0, zero diagnostics |
| Normal emitted JavaScript | 10 files checked; 3 changed files, 6 hunks, all inside dumpTypes |
| Emitted declarations | All 716 files byte-identical |
| Runtime export keys | All 2,256 own string keys identical; 0 symbol keys on each side |

The changed bundles are _tsc.js, run.js and typescript.js. Each complete file
matches baseline outside dumpTypes, and its entire changed owner equals the
baseline owner with only the two approved emitted replacements. The other seven
JavaScript files are byte-identical. Source maps and build metadata are outside
this JavaScript comparison.

The normal-build Node witness runs a small project with unions, tuples, a
conditional type and a generic mapping function. Both CLI images run in one
shared directory with the same library and project paths. Each --generateTrace
run exits 0 and writes **7,241,122 bytes**, exactly identical before and after.
SHA-256 on both written types files:

    bb0e074a3d6f5af54cbd064cf314e25917e15137bd99bb7dfacba7fe9cfd8bdb

Both restoring mutants use complete copies of all 747 source files, restoring
only one original cast expression in tracing.ts. All 23 tracing census units
run in each case with zero checker diagnostics. The two imports account for
the two-line coordinate shift:

| Restored original site | Mutant finding | Phases |
| --- | --- | --- |
| 232:34 objectFlags | 234:34 | refusal_scan |
| 318:33 intrinsicName | 320:33 | refusal_scan and lowering |

The candidate has zero unchecked-cast findings in the complete tracing census.
The baseline has three such events at these two sites; the object mutant
restores one and the intrinsic mutant restores two. Other existing NotYet and
Refused findings remain. Both no-output guards are enabled in the measurement
binary; production loader and lowering code are unchanged.

The adapter code and pins stay frozen throughout full apply, builds and final
source verification. Their retained hashes are in frozen-adapter.json. This
final prose was added after those checks and changes no generated source.


## Evidence and operational corrections

Raw scripts, logs, individual records and results belong to
[typescript/tracing-dumptypes-evidence](https://github.com/system-inc/adamic/tree/typescript/tracing-dumptypes-evidence/stage3/tracing-dumptypes-evidence),
not to stage3/adapt. Go evidence uses .go.txt names. Each driver records the
child's exit code and keeps complete logs in files.

The cancelled first lane attempt omitted 30 nested corpus files through an
overly broad node_modules ignore rule. It is excluded. The corrected copy
excludes only the root dependency directory, verifies the complete corpus, and
reruns all shards without splicing successful subsets.

Other retained initial attempts are not accepted evidence: setup begun on the
old checkout before submodule alignment; probe callers rejected by the
string-only console API; a basename census filter selecting no files; and a
trace comparison across different compiler-library paths. Corrected setup,
callers, absolute-path selection and same-directory comparisons all pass.

The first small-project witness invocation was also retained: TypeScript 6
rejected the repository tsconfig context. The accepted invocation uses the
project working directory and --ignoreConfig on both sides.
