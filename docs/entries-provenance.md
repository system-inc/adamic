# Object enumeration provenance, roadmap step 12

Delivery branch: `codex/entries-provenance`, based only on
`origin/compiler/area-next-fixtures` at `4885cec50290686df487b62aac47c85d871ed40c`.
Research: `codex/step12-scout` at `ebaf1bc0`, read without merging.
Task: `#d9eemrs`.

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
