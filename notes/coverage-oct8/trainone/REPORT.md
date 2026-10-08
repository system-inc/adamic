Built 13 .a coverage programs and reproducible four-way and mutation runners; 6 agree and 7 stop at compile time.
Commits: base 031a1259bc7973934792dc6cb1bd4074fc2204b9; reviewed range 54cbc125..9f16421c; programs/evidence commit 5ea8459c131a463688c16f14f83b15f6492dff96; report/cleanup follow-up is on codex/coverage-oct8-trainone.
Commands and outputs: run.py records stdout, stderr and exit codes; setup took 210.148s on nproc=5; package logs are below.
Mutants: numeric-index and structural-map guards caught by lowering assertions; missing substr operand caught by compiler panic; deleted-key filter caught by ASan; two-index sort and allocation-size guards survived focused native tests.
Not covered: admitted source deletion/presence-slot enumeration or optional Record access, combined null/undefined receivers, allocation failure injection, 32-bit targets, and the full repository gate.

## Observations

This is coverage-only work based on current origin/main. No compiler or runtime change is committed, no main merge and no PR. The full named files and slice history were read, plus CLAUDE.md, README.md, docs/0.1.md and docs/memory.md. The emit_statements.go and emit_slots.go slice hunks were read in history and current context. [slice-history.log](evidence/slice-history.log) preserves the requested git log -p output.

The landing commits are cfa29460 (optional indexing), c4b29169 (substr), 8c962b67 (record aliases) and be356537 (checked enumeration). The range also contains the optional ArrayIndex slot hunk in 6eef75d9, debugger in 6c38c305, namespace readiness in 70ae18c3, and the Uint16Array addition/withdrawal pair 277e3937/f381d094. There is no new union.c comment in this slice.

Source truth uses Node 24.19.0 via oracle/node.mjs, which strips types with Node itself. Native uses the compiler CLI with --sanitize (the oracle flags: -O1, ASan, UBSan, no recovery) and without it (-O2). Emitted JavaScript runs through the same Node runner. Comparison checks the raw stdout/stderr bytes and exit status. Each successful sanitized program runs again with detect_leaks=1. Every supported case exits 0 with empty stderr, including the separate leak runs. No baseline runtime abort, invalid C, wrong runtime output or leak was observed.

| Program | What was tried | Four-way result |
|---|---|---|
| optional-nested.a | a?.[i]?.[j], absent base/inner row, empty rows, runtime strings, index replacing base binding | Agree |
| optional-map.a | optional Map get followed by optional indexing; missing map/key/value; effects k then i | Agree |
| optional-null.a | separate null array and Map receivers, skipped index/key, runtime values | Agree |
| substr.a | negative/fractional/NaN/finite huge arguments, optional length, both surrogate halves, receiver changed by start operand | Agree |
| entries-order.a | checked record entries, saved snapshot before overwrite, integer boundaries, empty/huge/ordinary keys | Agree |
| entries-two-indices.a | exactly two numeric keys inserted out of numeric order | Agree |
| entries-delete.a | delete then reinsert numeric key and delete ordinary key | Compiler refusal |
| entries-absent.a | omitted optional number field with contextual homogeneous entries type | Compiler refusal |
| entries-presence.a | omitted optional field versus explicitly present undefined | Compiler refusal |
| optional-record.a | optional string-key Record lookup with side effect | NotYet |
| optional-map-nullish.a / optional-nested-nullish.a | receiver includes null and undefined together | NotYet at representation |
| substr-omitted-start.a | omitted/undefined start accepted by Node but outside Adamic library declaration | Type-check errors |

The existing syntax_substr.a already covers a numeric cross product including infinities and split UTF-16 boundaries. These additions exercise finite magnitude 1e300, negative zero, receiver mutation during argument evaluation and additional half-surrogate boundaries. Existing optional_indexing_map.a already replaces its map binding in the key; the added map program composes the optional get with a separately effectful optional index. Existing record runtime harnesses cover deletion below the source-language refusal.

## Compiler-stop disagreements

For the three compiler columns below, these are compile-command outputs, not outputs from nonexistent binaries. That phase is recorded in results.json. An empty output block means zero bytes. These stops are observations, not claims of silent miscompilation.

### entries-absent

**Kind: compiler error.** internal/lower/library_object.go:140-141 rejects optional field layouts even with a homogeneous result contract.

Program:
```typescript
interface Fields { present: number; absent?: number; }
const source: Fields = { present: 1 };
const entries: [string, number][] = Object.entries(source);
entries.forEach(entry => console.log(`${entry[0]}:${entry[1]}`));
```

node: exit 0, phase run; stdout verbatim:
```text
present:1
```
stderr verbatim:
```text

```

sanitize: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/entries-absent.a:3:37: Adamic 0.1 refuses Object.entries; every present field must have tsc's result element representation; optional or heterogeneous fields cannot be read soundly
```

O2: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/entries-absent.a:3:37: Adamic 0.1 refuses Object.entries; every present field must have tsc's result element representation; optional or heterogeneous fields cannot be read soundly
```

javascript: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/entries-absent.a:3:37: Adamic 0.1 refuses Object.entries; every present field must have tsc's result element representation; optional or heterogeneous fields cannot be read soundly
```

### entries-delete

**Kind: compiler error.** internal/native/runtime/record.c:108 implements runtime deletion, but internal/lower/refusals.go:27 rejects source delete before that path.

Program:
```typescript
const source: { visible: number; [key: string]: number | string } = { visible: 1 };
source['10'] = 10;
source['2'] = 2;
source['4294967294'] = 94;
source['4294967295'] = 95;
source['01'] = 1;
source['-0'] = 0;
source['1e0'] = 100;
source['gone'] = 8;
delete source['gone'];
delete source['2'];
source['2'] = 22;
const view: { readonly visible: number } = source;
const entries: [string, number][] = Object.entries(view);
entries.forEach(entry => console.log(`${entry[0]}:${entry[1]}`));
```

node: exit 0, phase run; stdout verbatim:
```text
2:22
10:10
4294967294:94
visible:1
4294967295:95
01:1
-0:0
1e0:100
```
stderr verbatim:
```text

```

sanitize: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/entries-delete.a:10:1: Adamic 0.1 refuses delete; an object's shape is fixed; use a Map for keys that come and go
```

O2: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/entries-delete.a:10:1: Adamic 0.1 refuses delete; an object's shape is fixed; use a Map for keys that come and go
```

javascript: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/entries-delete.a:10:1: Adamic 0.1 refuses delete; an object's shape is fixed; use a Map for keys that come and go
```

### entries-presence

**Kind: compiler error.** internal/lower/library_object.go:124-125 rejects a result representation containing undefined.

Program:
```typescript
interface Fields { present: number; absent?: number; explicit?: number | undefined; }
const source: Fields = { present: 1, explicit: undefined };
const entries = Object.entries(source);
entries.forEach(entry => console.log(`${entry[0]}:${entry[1]}`));
```

node: exit 0, phase run; stdout verbatim:
```text
present:1
explicit:undefined
```
stderr verbatim:
```text

```

sanitize: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/entries-presence.a:3:17: Adamic 0.1 refuses Object.entries; tsc's result must have one homogeneous number, string or boolean value type; any and widened field views are unsound
```

O2: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/entries-presence.a:3:17: Adamic 0.1 refuses Object.entries; tsc's result must have one homogeneous number, string or boolean value type; any and widened field views are unsound
```

javascript: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/entries-presence.a:3:17: Adamic 0.1 refuses Object.entries; tsc's result must have one homogeneous number, string or boolean value type; any and widened field views are unsound
```

### optional-map-nullish

**Kind: compiler error.** internal/lower/expression.go:40 reports the unrepresentable receiver type before optional_indexing_map.go:44 can lower it.

Program:
```typescript
let effects = '';
function key(): string { effects += 'k'; return ['ro', 'w'].join(''); }
function index(): number { effects += 'i'; return 1; }
function pick(map: ReadonlyMap<string, readonly string[]> | null | undefined): void {
    effects = '';
    const result = map?.get(key())?.[index()];
    console.log(`${result ?? 'missing'}:${effects}`);
}
pick(undefined);
pick(null);
pick(new Map<string, readonly string[]>());
pick(new Map<string, readonly string[]>([['row', []]]));
pick(new Map<string, readonly string[]>([['row', ['a', 'b'.repeat(2)]]]));
```

node: exit 0, phase run; stdout verbatim:
```text
missing:
missing:
missing:k
missing:ki
bb:ki
```
stderr verbatim:
```text

```

sanitize: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/optional-map-nullish.a:4:15: stage 0 can't lower a value of type ReadonlyMap<string, readonly string[]> | null | undefined yet
```

O2: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/optional-map-nullish.a:4:15: stage 0 can't lower a value of type ReadonlyMap<string, readonly string[]> | null | undefined yet
```

javascript: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/optional-map-nullish.a:4:15: stage 0 can't lower a value of type ReadonlyMap<string, readonly string[]> | null | undefined yet
```

### optional-nested-nullish

**Kind: compiler error.** internal/lower/expression.go:40 reports the unrepresentable receiver type before optional_indexing.go:10 can lower it.

Program:
```typescript
let effects = '';
function index(label: string, value: number): number { effects += label; return value; }
function pick(values: readonly (readonly string[] | undefined)[] | null | undefined): void {
    effects = '';
    const result = values?.[index('i', 0)]?.[index('j', 1)];
    console.log(`${result ?? 'missing'}:${effects}`);
}
pick(undefined);
pick(null);
pick([]);
pick([undefined]);
pick([[]]);
pick([['a', ['b', 'c'].join('')]]);
let selected: readonly (readonly string[])[] | undefined = [['old'.repeat(2)]];
function replace(): number { selected = undefined; effects += 'x'; return 0; }
effects = '';
console.log(`${selected?.[replace()]?.[index('y', 0)] ?? 'missing'}:${effects}`);
```

node: exit 0, phase run; stdout verbatim:
```text
missing:
missing:
missing:i
missing:i
missing:ij
bc:ij
oldold:xy
```
stderr verbatim:
```text

```

sanitize: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/optional-nested-nullish.a:3:15: stage 0 can't lower a value of type readonly (readonly string[] | undefined)[] | null | undefined yet
```

O2: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/optional-nested-nullish.a:3:15: stage 0 can't lower a value of type readonly (readonly string[] | undefined)[] | null | undefined yet
```

javascript: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/optional-nested-nullish.a:3:15: stage 0 can't lower a value of type readonly (readonly string[] | undefined)[] | null | undefined yet
```

### optional-record

**Kind: compiler error.** internal/lower/collections.go:367 stops optional non-tuple indexing before optional_indexing.go can run.

Program:
```typescript
let calls = 0;
function key(): string { calls++; return 'name'; }
function pick(record: Record<string, string> | undefined): void {
    console.log(`${record?.[key()] ?? 'missing'}:${calls}`);
}
pick(undefined);
pick({ name: 'present' });
```

node: exit 0, phase run; stdout verbatim:
```text
missing:0
present:1
```
stderr verbatim:
```text

```

sanitize: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/optional-record.a:4:20: stage 0 can't lower ?.[] on anything but a tuple, at a position written out yet
```

O2: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/optional-record.a:4:20: stage 0 can't lower ?.[] on anything but a tuple, at a position written out yet
```

javascript: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
adamic: /workspace/adamic/notes/coverage-oct8/trainone/optional-record.a:4:20: stage 0 can't lower ?.[] on anything but a tuple, at a position written out yet
```

### substr-omitted-start

**Kind: compiler error.** internal/lower/string_substr.go:33-34 has an absent-start fallback, but the embedded substr declaration requires a number start; checker errors happen first.

Program:
```typescript
console.log('abc'.substr());
console.log('abc'.substr(undefined, undefined));
```

node: exit 0, phase run; stdout verbatim:
```text
abc
abc
```
stderr verbatim:
```text

```

sanitize: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
notes/coverage-oct8/trainone/substr-omitted-start.a:1:19: error TS2554: Expected 1-2 arguments, but got 0.
notes/coverage-oct8/trainone/substr-omitted-start.a:2:26: error TS2345: Argument of type 'undefined' is not assignable to parameter of type 'number'.
```

O2: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
notes/coverage-oct8/trainone/substr-omitted-start.a:1:19: error TS2554: Expected 1-2 arguments, but got 0.
notes/coverage-oct8/trainone/substr-omitted-start.a:2:26: error TS2345: Argument of type 'undefined' is not assignable to parameter of type 'number'.
```

javascript: exit 1, phase compile; stdout verbatim:
```text

```
stderr verbatim:
```text
notes/coverage-oct8/trainone/substr-omitted-start.a:1:19: error TS2554: Expected 1-2 arguments, but got 0.
notes/coverage-oct8/trainone/substr-omitted-start.a:2:26: error TS2345: Argument of type 'undefined' is not assignable to parameter of type 'number'.
```

## Comment and documentation audit

Every comment block in the named whole files was inspected. Matching here means consistent with the implementation, not an independently proven specification claim. There are no comments or doc blocks in entries_records.go.

| File and comment | Observation |
|---|---|
| optional_indexing.go:8-9 | Effects saves the receiver, Conditional guards null/undefined, and only the true arm contains ArrayIndex/StringIndex. The new nested probe observes the claimed effects. |
| optional_indexing_chain.go:8-14 | Only a direct optional property receiver with required array/string/typed-array field is admitted; parentheses are not skipped. Missing element representation can flow into a following optional read. |
| optional_indexing_map.go:8-9,67-69 | Conditional guards receiver before key evaluation; structural object/new hazards are scanned even behind a library annotation. Structural-guard removal is caught. |
| string_substr.go:8-9 | Helper call receives operands once; stringInteger normalizes, negative start is relative, count determines slice end. Receiver mutation probe prints the original string and effects sl. |
| library_object.c:1 | Comment says static Object methods over proven, fixed shapes. Actual keys and checked values dispatch to mutable record storage. **Kind: comment that does not match the code**, if fixed shapes is intended as the scope of this translation unit. |
| library_object.c:32,52-53,78,98,157 | Private # fields bypass freeze check; canonical uint32 index recognition excludes the listed strings; insertion sort is stable for ordinary names; class keys delegate to class descriptors; checked reflection uses ordered keys and one actual-value read per key. No mismatch found in these bodies. |
| record.c:1,8-11 | Record stores keys in a counted Map behind wrapper slots; observable keys come from table, not wrapper shape. Iterator wraps record and snapshot with scalar cursor. Offsets are 0/1/key as documented. No layout proof changes were made. |
| record.c:42-43,75 | Prototype member guard runs only after an own miss; byte comparisons are dispatched by lengths; diagnostic fits the stated maximum without allocation. Existing record tests compare the member list with Node. |
| record.c:109,118-119 | delete always returns true; index recognizer bounds length to ten bytes before uint64 arithmetic, rejects uint32 max, -0 and 01. Source deletion is nevertheless refused. |
| union.c:1,35,40,57-58,88,164 | Union box equality follows numeric/string value equality and identity otherwise; conversion defaults panic; shape metadata is a static linked registry; scalar tags match ir.Type values. Comments describe these code paths. |
| emit_slots.go:9-10,14-15,51-52 | Optional slot path snapshots before aside(index), emits key and cleanup inside present branch; TypeOf retains lookup presence to distinguish missing from null. |
| emit_statements.go slice comments at 67,205-206,261 | Debugger emits no native instruction; ordinary property write evaluates value before undefined check; kept reference is taken before releasing old slot. No mismatch found. Namespace readiness and withdrawn Uint16 hunk add no conflicting comment. |

The library_object.c header mismatch is a description-scope inference grounded in the record branches at lines 45,99,176; changing the header is outside this coverage-only delivery.

## Mutations and package checks

Each production mutation was independent and restored with finally. The mutant runners record exact replacement text and commands. No compiler-warning failure is counted as a successful detector. Focused checks:

| Mutant | Check and observation |
|---|---|
| numeric-index-guard, optional_indexing.go:35 | TestOptionalIndexingKeepsUnsupportedStorageNotYet/string_numeric_key fails its diagnostic assertion, exit 1. |
| structural-map-guard, optional_indexing_map.go:27 | TestOptionalIndexingMapShapeRefused fails its structural-receiver diagnostic assertion, exit 1. |
| substr-missing-argument-guard, string_substr.go:33 | Existing syntax_substr.a reaches missing length; compiler panics index out of range [2] with length 2, exit 1. This proves the guard is exercised, not a runtime-output detector. |
| record-deleted-key-filter, record.c:188 | TestRecordsAgainstNode/semantics fails with ASan heap-use-after-free in array_index through record_keys, exit 1. |
| record-two-index-sort, record.c:178 | count > 1 changed to count > 2; TestRecordsAgainstNode exits 0. See full-package and new-witness results below. |
| record-allocation-size-guard, record.c:162 | Removed count > SIZE_MAX / sizeof *indices panic branch; TestRecordsAgainstNode exits 0. See full-package result below. |

Full native package under record-two-index-sort: `go test ./internal/native -count=1 -timeout 30m`, exit 0.
```text
ok  	github.com/system-inc/adamic/internal/native	276.568s
```

Full native package under record-allocation-size-guard: `go test ./internal/native -count=1 -timeout 30m`, exit 0.
```text
ok  	github.com/system-inc/adamic/internal/native	273.529s
```

**Kind: branch no native-package test covers for exactly two indices.** internal/native/runtime/record.c:178-179 sorts two keys in main; threshold mutant skips it. New program is entries-two-indices.a. Four outputs under that mutant follow. This is a deliberate mutant, not a main defect.

node: exit 0; stdout verbatim:
```text
2:twotwo
10:tenten
label:seedseed
```
stderr verbatim:
```text

```

sanitize: exit 0; stdout verbatim:
```text
10:tenten
2:twotwo
label:seedseed
```
stderr verbatim:
```text

```

O2: exit 0; stdout verbatim:
```text
10:tenten
2:twotwo
label:seedseed
```
stderr verbatim:
```text

```

javascript: exit 0; stdout verbatim:
```text
2:twotwo
10:tenten
label:seedseed
```
stderr verbatim:
```text

```

The existing entries_record_alias source oracle under the same sort mutant:
```text
=== RUN   TestNativeAgreesWithNode
=== PAUSE TestNativeAgreesWithNode
=== CONT  TestNativeAgreesWithNode
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/entries_record_alias.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/entries_record_alias.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/entries_record_alias.a
    oracle_test.go:774: stdout differs
        node:   exit 0, stdout "visible\nvisible:4\nown:true\n2,10,visible,late,01\n2:2\n10:10\nvisible:4\nlate:2\n01:1\n2,10,4,2,1\nlabel:changedchanged\nlast:second\nenabled:true\ndisabled:false\n", stderr ""
        native: exit 0, stdout "visible\nvisible:4\nown:true\n10,2,visible,late,01\n10:10\n2:2\nvisible:4\nlate:2\n01:1\n10,2,4,2,1\nlabel:changedchanged\nlast:second\nenabled:true\ndisabled:false\n", stderr ""
--- FAIL: TestNativeAgreesWithNode (0.01s)
    --- FAIL: TestNativeAgreesWithNode/internal/oracle/testdata/entries_record_alias.a (0.53s)
FAIL
gate cache: native hits=0 misses=3
gate cache: node hits=0 misses=2
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	0.554s
FAIL
```
Observed: the existing source oracle catches this mutant with stdout differs, exit 1. Thus exactly-two-key ordering is already covered at repository level; only the native-package test set missed it.

**Kind: branch no test covers in the runs performed.** record.c:162 overflow panic branch survives removal. On this 64-bit target, at most 2^32-1 canonical numeric keys can exist, below SIZE_MAX / sizeof(indexed_key). Inference: this guard cannot fire for a valid record here; the result does not prove 32-bit safety. No enormous allocation or fabricated runtime state was attempted.

The candidate guard audit is not an exhaustive branch census. Error propagation, malloc failure paths, synthetic IR-only storage states, all optional-chain rejection variants and every scalar box branch were not independently mutated. Existing substr IR mutants and record runtime mutants are additionally exercised by the validation commands below.

Existing mutant checks run as part of the package/filtered validation:

| Existing mutants | What catches them |
|---|---|
| substr relative-start, truncate, NaN, undefined-length, length-not-end, length-clamp, evaluation-order | TestSyntaxSubstrMutants requires native and emitted-JavaScript stdout to differ from source Node while finishing and passing leak checks. |
| record indices-in-insertion-order, uint32-max-as-index, deleted-key-iterated | TestRecordMutants requires differing Node stdout with balanced counts and clean sanitizer exits. |
| record overwrite-key-leaked | TestRecordMutants requires LeakSanitizer. |
| record stored-key-freed / own-slot-null-read | TestRecordMutants requires ASan use-after-free / UBSan null-member access respectively. |
| record prototype-membership-restored, missing-read-silent, own-read-checked-as-missing | TestRecordReadMutants requires failure of the exact own-only stop contract, or erroneous stop for an own hit. |
| entries unproven-as-proven on checked_misfit, checked_literal and record_misfit | TestEntriesProvenance pins exit 70 and diagnostic; mutants instead match source Node completion in native and JavaScript. |
| entries drop literal membership | TestEntriesProvenance/entries_checked_literal pins the literal-contract exit/message. |
| entries unchecked readiness primitive | TestEntriesRuntimeReadiness pins undefined-field exit/message in both backends; altered IR instead matches Node completion. |

## Commands, setup and limits

All test output went directly to log files. Setup succeeded; its timing lines and environment are verbatim:
```text
go version go1.27.1 linux/amd64
setup: go ready (0.071s)
v24.19.0
setup: node ready (0.088s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.452s)

added 3 packages in 709ms
setup: markdown dependencies installed (npm ci, integrity verified); step-duration=1.305s
setup: markdown dependencies ready (1.512s)
submodule cache: restored cohere 7945d102a6c18dd36adf9114a758ce646e8b2359
setup: submodules ready (18.798s)
setup: go build ready (210.028s)
setup: test binaries deferred (use --warm-tests) (210.121s)
setup: build cache warm (210.123s)
setup: build-flags commit=031a1259bc7973934792dc6cb1bd4074fc2204b9 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false load-before=0.04 0.01 0.00 1/133 709 load-after=6.64 3.62 1.44 1/142 3502
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (210.148s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.gw8d4F
```
Initial full lowering test exited 1 because the Node type dependency was absent; see lower-initial.log. Workaround: `npm ci --prefix stage3/api --ignore-scripts` installed the lockfile-pinned packages, with no tracked lockfile change. The rerun exited 0:
```text
ok  	github.com/system-inc/adamic/internal/lower	35.151s
```
Reproduction from repository root:
```sh
source /workspace/adamic-tools/env.sh
go build -o /tmp/trainone-adamic ./cmd/adamic > /tmp/trainone-build.log 2>&1
python3 notes/coverage-oct8/trainone/run.py > /tmp/trainone-fourway.log 2>&1
python3 notes/coverage-oct8/trainone/mutants.py > /tmp/trainone-mutants.log 2>&1
python3 notes/coverage-oct8/trainone/survivors.py > /tmp/trainone-survivors.log 2>&1
python3 notes/coverage-oct8/trainone/oracle-sort-mutant.py > /tmp/trainone-sort-oracle.log 2>&1
```
Final validation uses `bash notes/coverage-oct8/trainone/validate.sh`; its exact commands are saved in that script and exit statuses in evidence/validation-status.log. The unmutated lowering command is `go test ./internal/lower -count=1 -timeout 30m`. Full native runs under each surviving mutant used `go test ./internal/native -count=1 -timeout 30m`.

Use the path printed by setup on a different machine. Mutant runners edit the shared tree temporarily; run them sequentially and do not overlap a baseline compiler rebuild with them. run.py uses a prebuilt compiler and stores compile output separately from execution output.

validation-status.log:
```text
native exit=0
oracle exit=0
vet exit=0
format exit=0
diff-check exit=0
```

validation-native.log:
```text
ok  	github.com/system-inc/adamic/internal/native	12.429s
```

validation-oracle.log:
```text
=== RUN   TestEntriesProvenance
=== PAUSE TestEntriesProvenance
=== RUN   TestEntriesRuntimeReadiness
=== PAUSE TestEntriesRuntimeReadiness
=== RUN   TestNativeAgreesWithNode
=== PAUSE TestNativeAgreesWithNode
=== RUN   TestSyntaxSubstrMutants
=== PAUSE TestSyntaxSubstrMutants
=== CONT  TestEntriesProvenance
=== RUN   TestEntriesProvenance/entries_scanner
=== CONT  TestNativeAgreesWithNode
=== PAUSE TestEntriesProvenance/entries_scanner
=== CONT  TestEntriesRuntimeReadiness
=== CONT  TestSyntaxSubstrMutants
=== RUN   TestEntriesProvenance/entries_proven
=== PAUSE TestEntriesProvenance/entries_proven
=== RUN   TestEntriesProvenance/entries_checked_fit
=== PAUSE TestEntriesProvenance/entries_checked_fit
=== RUN   TestEntriesProvenance/entries_checked_misfit
=== PAUSE TestEntriesProvenance/entries_checked_misfit
=== RUN   TestSyntaxSubstrMutants/relative-start
=== PAUSE TestSyntaxSubstrMutants/relative-start
=== RUN   TestSyntaxSubstrMutants/truncate
=== RUN   TestEntriesProvenance/entries_proven_import
=== PAUSE TestSyntaxSubstrMutants/truncate
=== PAUSE TestEntriesProvenance/entries_proven_import
=== RUN   TestEntriesProvenance/entries_checked_boolean
=== RUN   TestSyntaxSubstrMutants/NaN
=== PAUSE TestSyntaxSubstrMutants/NaN
=== PAUSE TestEntriesProvenance/entries_checked_boolean
=== RUN   TestEntriesProvenance/entries_checked_literal
=== PAUSE TestEntriesProvenance/entries_checked_literal
=== RUN   TestSyntaxSubstrMutants/undefined-length
=== PAUSE TestSyntaxSubstrMutants/undefined-length
=== RUN   TestSyntaxSubstrMutants/length-not-end
=== RUN   TestEntriesProvenance/entries_checked_boxed
=== PAUSE TestSyntaxSubstrMutants/length-not-end
=== PAUSE TestEntriesProvenance/entries_checked_boxed
=== RUN   TestSyntaxSubstrMutants/length-clamp
=== PAUSE TestSyntaxSubstrMutants/length-clamp
=== RUN   TestSyntaxSubstrMutants/evaluation-order
=== PAUSE TestSyntaxSubstrMutants/evaluation-order
=== CONT  TestSyntaxSubstrMutants/relative-start
=== CONT  TestSyntaxSubstrMutants/length-clamp
=== RUN   TestEntriesProvenance/entries_record_alias
=== PAUSE TestEntriesProvenance/entries_record_alias
=== RUN   TestEntriesProvenance/entries_record_misfit
=== PAUSE TestEntriesProvenance/entries_record_misfit
=== CONT  TestSyntaxSubstrMutants/length-not-end
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/entries_scanner.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/entries_scanner.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/entries_proven.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/entries_proven.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_fit.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_fit.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_misfit.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_misfit.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/entries_proven_import.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/entries_proven_import.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_boolean.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_boolean.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_literal.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_literal.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_boxed.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_boxed.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/entries_record_alias.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/entries_record_alias.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/entries_record_misfit.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/entries_record_misfit.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/notyet_library_object_entries_const.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/notyet_library_object_entries_const.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_array.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_array.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_string.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_string.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_string_gap.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_string_gap.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_typed_array.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_typed_array.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_map.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_map.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_chain.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_chain.a
=== RUN   TestNativeAgreesWithNode/internal/oracle/testdata/syntax_substr.a
=== PAUSE TestNativeAgreesWithNode/internal/oracle/testdata/syntax_substr.a
=== CONT  TestSyntaxSubstrMutants/undefined-length
=== NAME  TestSyntaxSubstrMutants/relative-start
    syntax_substr_test.go:114: caught in both backends by stdout comparison with source Node
=== CONT  TestSyntaxSubstrMutants/NaN
=== NAME  TestSyntaxSubstrMutants/undefined-length
    syntax_substr_test.go:114: caught in both backends by stdout comparison with source Node
=== CONT  TestSyntaxSubstrMutants/truncate
=== NAME  TestSyntaxSubstrMutants/length-clamp
    syntax_substr_test.go:114: caught in both backends by stdout comparison with source Node
=== CONT  TestSyntaxSubstrMutants/evaluation-order
=== NAME  TestSyntaxSubstrMutants/length-not-end
    syntax_substr_test.go:114: caught in both backends by stdout comparison with source Node
=== CONT  TestEntriesProvenance/entries_scanner
=== NAME  TestEntriesRuntimeReadiness
    entries_provenance_test.go:145: unchecked readiness primitive mutant caught by pinned exit/message in both backends
--- PASS: TestEntriesRuntimeReadiness (3.76s)
=== CONT  TestEntriesProvenance/entries_record_misfit
=== NAME  TestSyntaxSubstrMutants/NaN
    syntax_substr_test.go:114: caught in both backends by stdout comparison with source Node
=== CONT  TestEntriesProvenance/entries_record_alias
=== NAME  TestSyntaxSubstrMutants/truncate
    syntax_substr_test.go:114: caught in both backends by stdout comparison with source Node
=== CONT  TestEntriesProvenance/entries_checked_boxed
=== CONT  TestEntriesProvenance/entries_checked_literal
=== NAME  TestSyntaxSubstrMutants/evaluation-order
    syntax_substr_test.go:114: caught in both backends by stdout comparison with source Node
--- PASS: TestSyntaxSubstrMutants (0.01s)
    --- PASS: TestSyntaxSubstrMutants/relative-start (3.47s)
    --- PASS: TestSyntaxSubstrMutants/undefined-length (3.42s)
    --- PASS: TestSyntaxSubstrMutants/length-clamp (3.52s)
    --- PASS: TestSyntaxSubstrMutants/length-not-end (3.54s)
    --- PASS: TestSyntaxSubstrMutants/NaN (0.55s)
    --- PASS: TestSyntaxSubstrMutants/truncate (0.56s)
    --- PASS: TestSyntaxSubstrMutants/evaluation-order (0.55s)
=== CONT  TestEntriesProvenance/entries_checked_boolean
=== NAME  TestEntriesProvenance/entries_record_misfit
    entries_provenance_test.go:80: unproven-as-proven mutant prints Node's completed and exits 0; pinned exit 70 catches both backends
=== CONT  TestEntriesProvenance/entries_proven_import
=== CONT  TestEntriesProvenance/entries_checked_misfit
=== CONT  TestEntriesProvenance/entries_checked_fit
=== NAME  TestEntriesProvenance/entries_checked_literal
    entries_provenance_test.go:80: unproven-as-proven mutant prints Node's completed and exits 0; pinned exit 70 catches both backends
=== CONT  TestEntriesProvenance/entries_proven
=== NAME  TestEntriesProvenance/entries_checked_literal
    entries_provenance_test.go:102: drop literal membership mutant caught by pinned literal-contract exit/message in both backends
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/entries_scanner.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/entries_record_misfit.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/syntax_substr.a
=== NAME  TestEntriesProvenance/entries_checked_misfit
    entries_provenance_test.go:80: unproven-as-proven mutant prints Node's completed and exits 0; pinned exit 70 catches both backends
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_chain.a
--- PASS: TestEntriesProvenance (0.01s)
    --- PASS: TestEntriesProvenance/entries_scanner (0.51s)
    --- PASS: TestEntriesProvenance/entries_record_misfit (0.68s)
    --- PASS: TestEntriesProvenance/entries_checked_boolean (0.58s)
    --- PASS: TestEntriesProvenance/entries_checked_boxed (0.64s)
    --- PASS: TestEntriesProvenance/entries_record_alias (0.76s)
    --- PASS: TestEntriesProvenance/entries_checked_literal (0.89s)
    --- PASS: TestEntriesProvenance/entries_proven_import (0.65s)
    --- PASS: TestEntriesProvenance/entries_checked_fit (0.60s)
    --- PASS: TestEntriesProvenance/entries_checked_misfit (0.68s)
    --- PASS: TestEntriesProvenance/entries_proven (0.62s)
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_map.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_typed_array.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_string_gap.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_string.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_array.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/notyet_library_object_entries_const.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_boolean.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/entries_record_alias.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_boxed.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_literal.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_misfit.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/entries_proven_import.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_fit.a
=== CONT  TestNativeAgreesWithNode/internal/oracle/testdata/entries_proven.a
--- PASS: TestNativeAgreesWithNode (0.08s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/entries_scanner.a (0.53s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/entries_record_misfit.a (0.52s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/syntax_substr.a (0.68s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_map.a (0.57s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_chain.a (0.71s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_typed_array.a (0.57s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_string_gap.a (0.58s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_string.a (0.56s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/notyet_library_object_entries_const.a (0.61s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_boolean.a (0.63s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/optional_indexing_array.a (0.78s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/entries_record_alias.a (0.61s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_boxed.a (0.65s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_misfit.a (0.57s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_literal.a (0.62s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/entries_proven_import.a (0.58s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/entries_checked_fit.a (0.58s)
    --- PASS: TestNativeAgreesWithNode/internal/oracle/testdata/entries_proven.a (0.41s)
PASS
gate cache: native hits=0 misses=65
gate cache: node hits=0 misses=77
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	7.609s
```

vet.log:
```text

```

format.log:
```text

```

diff-check.log:
```text

```

fourway.log:
```text
entries-absent agree=False {'node': 0, 'sanitize': 1, 'O2': 1, 'javascript': 1}
entries-delete agree=False {'node': 0, 'sanitize': 1, 'O2': 1, 'javascript': 1}
entries-order agree=True {'node': 0, 'sanitize': 0, 'O2': 0, 'javascript': 0}
entries-presence agree=False {'node': 0, 'sanitize': 1, 'O2': 1, 'javascript': 1}
entries-two-indices agree=True {'node': 0, 'sanitize': 0, 'O2': 0, 'javascript': 0}
optional-map-nullish agree=False {'node': 0, 'sanitize': 1, 'O2': 1, 'javascript': 1}
optional-map agree=True {'node': 0, 'sanitize': 0, 'O2': 0, 'javascript': 0}
optional-nested-nullish agree=False {'node': 0, 'sanitize': 1, 'O2': 1, 'javascript': 1}
optional-nested agree=True {'node': 0, 'sanitize': 0, 'O2': 0, 'javascript': 0}
optional-null agree=True {'node': 0, 'sanitize': 0, 'O2': 0, 'javascript': 0}
optional-record agree=False {'node': 0, 'sanitize': 1, 'O2': 1, 'javascript': 1}
substr-omitted-start agree=False {'node': 0, 'sanitize': 1, 'O2': 1, 'javascript': 1}
substr agree=True {'node': 0, 'sanitize': 0, 'O2': 0, 'javascript': 0}
All supported comparisons and explicit compiler stops matched expectations.
```

Successful raw outputs, every failed compile output, commands and exit codes are preserved in evidence/results.json and the individual .stdout/.stderr/.json files. Generated JavaScript is preserved as each successful JavaScript compile stdout. Logs are committed so temporary-directory cleanup does not remove the observations.

## Inferences and remaining scope

All four agreed on everything the current compiler admitted in these probes. Optional-record access, delete, optional-presence enumeration and the combined nullish type stop before runtime, so this unit cannot claim four-way execution coverage for them. The source Node outputs establish what those programs do, while existing C harness checks establish only runtime primitive behavior below the admission boundary. No absence of defects beyond these inputs is inferred.

The full repository gate was not run. The selected package/oracle checks and their limits are recorded above. No counts table was changed because these programs live under notes and are not registered as ordinary oracle fixtures. No main branch update, PR, compiler fix or changed refusal contract is part of this work.
