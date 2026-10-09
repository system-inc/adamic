# Object enumeration provenance, roadmap step 12

Delivery branch: `codex/entries-provenance`, based only on
`origin/compiler/area-next-fixtures` at `4885cec50290686df487b62aac47c85d871ed40c`.
Research: `codex/step12-scout` at `ebaf1bc0`, read without merging.
Task: `#d9eemrs`.

## Acceptance follow-up

Dependency: `codex/step12-entries-fixtures` at `8d864f9c`, read and run without
merging its branch or committing its fixtures. Its six source programs, Node
observations, and six source mutations remain unchanged. Supply the extracted
bucket with `ADAMIC_ENTRIES_ACCEPTANCE` when it is not present in Stage 3.
After the unchanged acceptance runner passed 12/12, the bucket was moved out of
the delivery checkout and independently retested:

```sh
ADAMIC_ENTRIES_ACCEPTANCE=/tmp/entries-acceptance-dependency-8d864f9c ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestEntriesAcceptance$' -count=1 -v > /tmp/entries-acceptance-external.log 2>&1
```

All 12 contracts pass, with reflection modes inspected and compiled cases run
in both backends, native release, and native sanitizers. The ten own source
fixtures, readiness witness, allocation proof, refusal boundaries, related
Object/class/regex checks, vet and Linux counts regeneration pass. The counts
diff adds only the two new own fixture rows.

| Acceptance program | Result |
|---|---|
| 01 scanner keywords | All 84 original keyword entries match Node, including the exact computed constructor key. Allocation proof removes the check. |
| 02 extra numeric field | Both visible and hidden entries match Node. Allocation proof removes the check. |
| 03 extra string field | Checked enumeration exits 70 at hidden, naming actual string and declared number. |
| 04 alias adds a field | Checked enumeration includes late in Node order. Its string mutation exits 70 at late. |
| 05 getter | Explicitly refused for the getter/accessor. Replacing it with a data field matches Node. |
| 06 symbol key | Explicitly refused for the symbol key. Replacing it with a string data field matches Node. |

The initial compiler refused the scanner's index annotation, the interface
results' permissive any overload, and the alias's index annotation. The getter
was refused for an unrelated reason and the symbol was NotYet. The acceptance
adapter now checks refusal text without the filename, so a filename containing
getter or symbol cannot satisfy the diagnostic assertion accidentally.

The result destination supplies the declared enumeration value contract when
TypeScript's interface overload returns any. Effect-free constant string keys
preserve the scanner's computed constructor spelling. String display of number
and boolean values uses their existing conversions.

String index annotations are admitted in modules using Object enumeration.
Const literal origins that receive indexed writes use the existing counted,
ordered record table. Alias writes invalidate unchecked enumeration. Actual own
keys and values are used through narrower views, including hasOwnProperty and
all Object.keys paths. Numeric keys sort numerically; other keys keep creation
order. Records retain representation checks through narrowed parameters.

The new own witness `internal/oracle/testdata/entries_record_alias.a` caught a
real exit-0 error: native Object.keys printed the table wrapper's key 0, and a
narrow numeric field read printed pointer bits instead of 4. Record-aware key
routes and representation checks repair both. This witness also checks scalar
number, string and boolean records, repeated strings, reads and writes through
parameters, overwriting a field, and keys before and after mutation. It matches
Node in release native, ASan/UBSan native and JavaScript. The second new witness,
entries_record_misfit.a, pins exit 70 and the complete late/string/number
message. Counts are regenerated on Linux for both new registered witnesses.

Conservative boundary: indexed mutations with unknown allocation origins,
nonconstant write keys, record allocation with spreads, and union or optional
field reads in record programs remain NotYet. Record spread copying stops
explicitly at runtime. Getter and symbol layouts remain outside this reflection
implementation. These restrictions prevent unsupported storage from being
silently treated as inline object fields. This is not general dictionary lowering.

Focused checks (complete output in logs):

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/oracle -run '^TestEntries|^TestObjectRefusalsExplainSoundness$|^TestObjectUnprovenShapesStayNotYet$|^TestUniformFieldsMatchNode$|^TestRuntimeFieldLayoutsAreIncluded$|^TestRegexProgramsKeepCheckedFieldReads$|TestNativeAgreesWithNode/internal/oracle/testdata/(library_object_|class_features_|regexp_match)' -count=1 -v > /tmp/entries-acceptance-related-final.log 2>&1
python3 stage3/fixtures/entries/check.py --acceptance --scratch /tmp/entries-acceptance-delivery > /tmp/entries-acceptance-delivery.log 2>&1
python3 internal/oracle/testdata/run-entries-provenance-mutants.py > /tmp/entries-acceptance-mutants.log 2>&1
python3 internal/oracle/testdata/run-entries-provenance-mutants.py lowering-drop-indexed-origin-boundary lowering-drop-indexed-read-boundary lowering-drop-record-spread-boundary > /tmp/entries-boundary-mutants.log 2>&1
go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle > /tmp/entries-acceptance-vet.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/entries-acceptance-counts.log 2>&1
```

Eleven implementation mutants are run and restored by the mutation runner:
removing native or JavaScript readiness checks; treating the hidden-value or
record-value misfit as proven; dropping a native or JavaScript enumeration key;
removing native record key dispatch; removing record representation marks; and
removing the indexed-origin, indexed-read and record-spread boundaries.
Pinned exit 70 catches removal of readiness and enumeration checks. Exact Node
stdout catches all three dropped-key mutants. Removing record representation
marks is caught because a successful Node program instead stops with exit 70. Existing fixture tests also remove the IR check and finite-literal
membership checks, proving both backend results revert to Node's exit-0 output
rather than the specified checked stop. Removing each refusal boundary is caught by its explicit NotYet assertion:
unknown indexed writers, undeclared indexed field reads, and spreads into record
storage. Their separate log is /tmp/entries-boundary-mutants.log. Counts initially
found a regression in existing regexp_match named-group reads; the indexed-read
boundary is now scoped to record storage, and that fixture and the regex field
layout check pass.

No full suite or shared gate is run locally. No Stage 3 status or recorded Node
observation is regenerated. The full scanner remains subject to its earlier
closed-program blockers; the exact scanner allocation in acceptance program 01
passes independently. The final native scanner driver was also rerun:

```sh
bash stage3/drivers/scanner/run.sh /tmp/entries-scanner-acceptance-after --tree /tmp/entries-scanner-slice --inputs /tmp/entries-scanner-before/adapted --oracle /tmp/entries-scanner-slice-baseline/node.stdout > /tmp/entries-scanner-acceptance-after.log 2>&1
```

Its first stop remains `corePublic.ts:9:5`, Refused index signature, build exit 1.
The module containing that earlier declaration has no Object enumeration call,
so the deliberately narrow declaration exception does not admit it. Node output
is byte-identical to the initial baseline: 108,019,868 bytes, 1,369,441 tokens,
466 diagnostics, 81 files, SHA-256
`ef99bf424a5b54ccdcbdf6eee2e4267aa59856887239008bca1d68582b22fc3b`.
Comparison control exits 0 and the end-token mutant is caught with diff exit 1.
The initial driver comparison and setup evidence below are retained as the
initial delivery's observations.

## Initial delivery evidence

Object.entries and Object.values now follow stable literal origins through annotated
aliases and imported symbols. A const or never-reassigned binding is eligible;
assignment and destructuring writes invalidate the proof. Every actual literal
field, including hidden fields, must fit the inferred result contract. Eligible
calls use the existing unchecked runtime enumeration. Remaining supported scalar
contracts enumerate the actual shape in ECMA key order and check each value before
returning the result. A mismatch exits 70 and names the key, actual type, and declared
type. Finite literal contracts are checked separately; numeric enums remain open.
Object.keys enumerates the actual shape: its declared string key contract needs no
additional value check. Enumeration does not substitute the structural view's fields
for the allocation's fields.

Conservative assumption: only compatible known data-property literals prove the
unchecked path. Parameters and calls use checks rather than inferred return or
parameter origins. The whole-program write scan conservatively treats property
writes and unary operands as possible binding writes. Whole-program descriptor
checks reject accessor layouts on checked calls, reserved or symbol-key storage,
and nullable spreads with synthetic absent slots. Existing homogeneous scalar
result restrictions remain: this unit does not admit arbitrary object-valued or
mixed-value contracts, any, or standalone explicit index signatures.

Eight registered source fixtures cover the scanner reduction, stable aliases,
imports, unknown origins that fit, hidden values that fail, booleans, finite string
literals, boxed scalar slots, strings, and open numeric enums (including 37 and
12.5). Successful observations match source Node byte for byte in the JavaScript,
release native, and ASan/UBSan native backends. Misfits pin exit 70 and the complete
diagnostic in all three. A hand-built IR witness separately checks uninitialized
slots; the source refusal pass still rejects its staged initializer, so that test
makes no source-admission claim.

## Verification

All test output is retained in log files. No whole package suite or full gate was run.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/entries-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run '^TestEntries|^TestObjectRefusalsExplainSoundness$|^TestObjectUnprovenShapesStayNotYet$|TestNativeAgreesWithNode/internal/oracle/testdata/library_object_' -count=1 -v > /tmp/entries-verified.log 2>&1
python3 internal/oracle/testdata/run-entries-provenance-mutants.py > /tmp/entries-mutants.log 2>&1
go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle > /tmp/entries-vet.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/entries-counts.log 2>&1
```

Focused proof, fixture, readiness, and related Object tests passed. Vet passed.
Counts regeneration passed. The diff adds eight rows, moves the existing
logical_and_reference_maybe row without changing its numbers, and removes the stale
binder_flow row: that fixture was already registered with lowers=false in the base.
Counts are generated on this Linux machine, never edited by hand.

| Mutant actually run | Catcher |
|---|---|
| Lowering treats an unproven call as proven | Pinned misfit exit 70 fails: mutant exits 0 |
| Native enumeration drops the first shape key | Node stdout comparison fails |
| JavaScript checked enumeration drops its first key | Node stdout comparison fails |
| JavaScript checked enumeration drops the readiness test | Pinned undefined-slot exit 70 fails |
| Native checked enumeration drops the readiness test | Pinned undefined-slot exit 70 fails |
| IR removes checks, number and literal misfits, both backends | Mutant prints exactly Node's `completed` and exits 0; pinned exit/message rejects it |
| IR removes finite literal membership, both backends | Mutant prints exactly Node's `completed` and exits 0; literal contract assertion rejects it |
| IR removes the readiness check path, both backends | Mutant matches Node exit 0; pinned undefined-slot exit/message rejects it |

The real implementation mutants are restored after each run. Their individual logs
are under `/tmp/entries-provenance-mutants`. They fail on program behavior, not
compilation. The protected emitter, main lowering/native files, and oracle fixture
registry were not edited. New fixtures are registered from the unit's test file.

## Scanner before and after

The reduced table retains scanner.ts:222's abstract, constructor, and of keyword
shape. Its computed constructor spelling is normalized to the equivalent literal
key, and its open view uses the supported builtin Readonly<Record<string, SyntaxKind>>.
No cohere code is copied.

Before, the reduced program's first stop was:

```text
entries_scanner.a:8:46: stage 0 can't lower Object.entries on a shape not proven by a plain literal or its const binding yet
```

After, it compiles and matches Node:

```text
3 137
abstract,constructor,of
```

The actual TypeScript native scanner slice was run with both the unchanged base
compiler and the changed compiler:

```sh
bash stage3/drivers/scanner/run.sh /tmp/entries-scanner-slice-baseline --tree /tmp/entries-scanner-slice --inputs /tmp/entries-scanner-before/adapted --compiler /tmp/entries-before-compiler > /tmp/entries-scanner-slice-baseline.log 2>&1
bash stage3/drivers/scanner/run.sh /tmp/entries-scanner-slice-after --tree /tmp/entries-scanner-slice --inputs /tmp/entries-scanner-before/adapted --compiler /tmp/entries-after-compiler > /tmp/entries-scanner-slice-after.log 2>&1
cmp /tmp/entries-scanner-slice-baseline/node.stdout /tmp/entries-scanner-slice-after/node.stdout > /tmp/entries-scanner-node-cmp.log 2>&1
```

Both driver runs stop before Object.entries at
`adapted/src/compiler/corePublic.ts:9:5`: Adamic refuses an index signature and
suggests Map. Both build exits are 1. This earlier MapLike admission dependency is
outside the unit and was not bypassed. The Node observation files are byte-identical:
1,369,441 tokens, 466 scanner diagnostics, 81 input files, 108,019,868 output bytes,
SHA-256 `ef99bf424a5b54ccdcbdf6eee2e4267aa59856887239008bca1d68582b22fc3b`.
Both comparison controls exit 0 and both end-token mutants are caught (diff exit 1).
No Stage 3 status record or recorded Node observation was rewritten. The observation
is an unchanged earlier native stop; the reduction proves this Object.entries
blocker is removed, not that the complete scanner now runs natively.

## Setup

`nproc` reported 5. Setup succeeded with the requested GOPROXY fallback.

```text
setup: submodules ready (0.079s)
setup: node ready (0.086s)
setup: go ready (0.086s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.486s)
setup: markdown dependencies installed (npm ci, integrity verified); step-duration=0.713s
setup: markdown dependencies ready (0.929s)
setup: go build ready (46.226s)
setup: test binaries deferred (use --warm-tests) (46.438s)
setup: build cache warm (46.439s)
setup: build-flags commit=4885cec50290686df487b62aac47c85d871ed40c nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false load-before=0.04 0.01 0.00 1/137 697 load-after=5.91 1.63 0.55 1/140 1700
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (46.466s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.EURwnu
```
