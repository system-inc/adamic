# Acheck evidence, wave 3

Observed October 27, 2026. Parent: wave 2, b220aa3150086f3b56a845f3f340986a4aa1619d.
Read wave 1's stage3/census/repro/ACHECK.md and wave 2's
stage3/drivers/parser/ACHECK.md first. Only six headers and their six known.txt
entries change. Every program body and all four predicate programs are unchanged.

## Reader audit and scope

Before editing, ran git grep -n -F <file name> outside stage3/a-check-headers
and BLOCKERS.md for each of the twelve. Full matches are saved as
/tmp/wave3-readers-<file name>.log. Read each source, its git log, and directory
notes. Also inspected dynamic reader paths, not just literal basename matches.

The user's rule says to leave a program alone when a live script reads it.
Accordingly six programs and their known.txt entries remain byte-identical:

- 71-writable-views/witness.cjs:13 reads language-questions/${name}.a,
  selecting generic-storage, node-flags, or text-range at lines 10-11. It checks
  zero stock TypeScript diagnostics and Node output 1, with zero-write mutants.
  A local console definition would change that oracle's runtime input.
- 75-optional-widening/probe-read-write.cjs:12 reads probes/read-write.a,
  copies it into a scratch compiler tree, and checks eight alias/write sites.
- drivers/parser/evidence/front24/minimals.py:4-7 runs native-function-any-view.a.
- drivers/parser/evidence/front21/run-probe.py:8 and front22/run-probe.py:8
  run native-process-next-tick-read.a and compare native and Node results.

Other grep matches are retained reports, notes, or different fixtures.
The read-write.a Go-test matches name stage3/interface-downcasts/v2 fixtures,
not adaptation 75. Parser probe-stops.py reads front10's minimal_programs;
none of this wave's twelve is in that list. run.sh/recut.sh do not stage them.
No live reader was found for the six edited programs.

## Decisions

Paths below are relative to stage3/. Existing promises move verbatim to a
dated history comment on line two. Files without promises gain a standalone
failure explanation. No small body edit preserves the original host-binding,
callable-marker, unchecked-cast, or repeated-read evidence while restoring
its obsolete outcome.

range-cases.a needs ./src/compiler/parser.ts and types.ts from the adapted
tree; its first diagnostic is TS7006, before the TS2307 diagnostics.
Contrary to the request's shared-path description, range-check.a does not
import tsc source: it imports only adamic and fails on its non-null assertions.
Its new comment describes that actual standalone failure.

| Program | Original evidence / history | Decision | Today: exit and full stderr on one line |
| --- | --- | --- | --- |
| adapt/65-temporary-node-builtins/evidence/previous-process-host-types.a | Prior adaptation 65 official Node Process/browser contract (6c34bb145; previous-process-README.md) | header fixed: ambient refusal replaces TS2591 | exit 1; adamic: stage3/adapt/65-temporary-node-builtins/evidence/previous-process-host-types.a:6:1: Adamic 0.1 refuses an ambient value declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values |
| adapt/71-writable-views/language-questions/generic-storage.a | Captured DiagnosticWithLocation escapes through Diagnostic[]; wider alias writes undefined (f8dea48fd; DIAGNOSTICS.md) | left: witness.cjs reads it | exit 1; adamic: stage3/adapt/71-writable-views/language-questions/generic-storage.a:2:1: Adamic 0.1 refuses an ambient value declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values |
| adapt/71-writable-views/language-questions/node-flags.a | Mutable flags view changes a literal-zero holder (3bc499480; README.md) | left: witness.cjs reads it | exit 1; adamic: stage3/adapt/71-writable-views/language-questions/node-flags.a:2:1: Adamic 0.1 refuses an ambient value declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values |
| adapt/71-writable-views/language-questions/text-range.a | Generic position setter changes a literal-zero holder (3bc499480; README.md) | left: witness.cjs reads it | exit 1; adamic: stage3/adapt/71-writable-views/language-questions/text-range.a:2:1: Adamic 0.1 refuses an ambient value declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values |
| adapt/75-optional-widening/probes/read-write.a | Eight optional-view alias/write classifications (89119e9d7; READ_WRITE.md) | left: probe-read-write.cjs reads it | exit 1; stage3/adapt/75-optional-widening/probes/read-write.a:11:13: error TS2345: Argument of type 'number &#124; undefined' is not assignable to parameter of type 'string'. \n   Type 'undefined' is not assignable to type 'string'. |
| drivers/parser/native-function-any-view-called-only.a | Namespace-free Debug.shouldAssertFunction callable marker (e74c38e43) | header fixed: current return-domain writable-view refusal | exit 1; adamic: stage3/drivers/parser/native-function-any-view-called-only.a:7:56: Adamic 0.1 refuses a value of type (level: number) => boolean seen as AnyFunction, which can write void where boolean is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| drivers/parser/native-function-any-view.a | Debug.shouldAssertFunction callable marker with namespace (5d777de3b) | left: front24/minimals.py reads it | exit 1; adamic: stage3/drivers/parser/native-function-any-view.a:5:56: Adamic 0.1 refuses a value of type (level: number) => boolean seen as AnyFunction, which can write void where boolean is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| drivers/parser/native-node-binding.a | First standalone performanceCore require checker stop (7280bbde3) | header fixed: TS2591 plus standalone explanation | exit 1; stage3/drivers/parser/native-node-binding.a:5:20: error TS2591: Cannot find name 'require'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig. |
| drivers/parser/native-process-next-tick-read.a | Node-loader nextTick truthiness read in isNodeLikeSystem (05d1fe88c) | left: front21/front22 run-probe.py read it | exit 1; adamic: stage3/drivers/parser/native-process-next-tick-read.a:6:1: Adamic 0.1 refuses an ambient value declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values |
| drivers/parser/native-require-perf-hooks.a | Reduced performanceCore perf_hooks assertion (ff2d95cc9) | header fixed: unchecked cast replaces TS2591 | exit 1; adamic: stage3/drivers/parser/native-require-perf-hooks.a:4:50: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| drivers/parser/range-cases.a | Directed parser inputs for addRange repeated-read invariant (c6c109ac0) | header fixed: TS7006 plus adapted-tree import explanation | exit 1; stage3/drivers/parser/range-cases.a:15:24: error TS7006: Parameter 'child' implicitly has an 'any' type. \n stage3/drivers/parser/range-cases.a:4:48: error TS2307: Cannot find module './src/compiler/parser.ts' or its corresponding type declarations. \n stage3/drivers/parser/range-cases.a:5:42: error TS2307: Cannot find module './src/compiler/types.ts' or its corresponding type declarations. \n stage3/drivers/parser/range-cases.a:6:27: error TS2307: Cannot find module './src/compiler/types.ts' or its corresponding type declarations. |
| drivers/parser/range-check.a | Guarded repeated-read probe with invalid-getter mutation (c6c109ac0) | header fixed: non-null assertion refusal plus standalone explanation | exit 1; adamic: stage3/drivers/parser/range-check.a:10:13: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |

## Measurement

Requested setup exited 0 in 394.386s, log /tmp/setup.log.
Each shell sourced /workspace/adamic-tools/env.sh.

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /tmp/setup.log 2>&1
go build -o /tmp/acheck-wave3-adamic ./cmd/adamic > /tmp/acheck-wave3-build.log 2>&1
# For each of the twelve paths, from the repository root:
/tmp/acheck-wave3-adamic c <path>
```

Measurements use the acheck compiler invocation and working directory.
Stdout was discarded; stderr and exit codes are saved in
/tmp/acheck-wave3-results.json and /tmp/acheck-wave3-results/*.stderr.
All twelve exited 1. No emitted C is included.

## Mutants and full package proof

The requested old-body mutant could not be run: no witness body was changed,
because all three witness-fix candidates have the live witness.cjs reader.
The user's instruction, "If a live reader (a test, a script, a driver) depends
on the program or its first line, leave it alone and report it", controls this
stop. No witness-fixed program or old-body mutant result is claimed.

As additional header proof, temporarily restored only native-require-perf-hooks.a's
old TS2591 first line, preserving line count. Then restored its corrected header
and temporarily put only its original known.txt entry back. Each mutant ran:

```sh
go test ./internal/acheck/ -run '^TestEveryProgramOutsideAPackageKeepsItsPromise$' -count=1
```

Both exited 1. Full output went to /tmp/acheck-wave3-mutant-header.log and
/tmp/acheck-wave3-mutant-known.log. The two red lines were:

```text
    acheck_test.go:91: stage3/drivers/parser/native-require-perf-hooks.a breaks its first line's promise: refused, and its first line promises type error TS2591 (adamic: stage3/drivers/parser/native-require-perf-hooks.a:4:50: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast))
    acheck_test.go:91: stage3/drivers/parser/native-require-perf-hooks.a keeps its promise now: take it off internal/acheck/known.txt
```

Both mutants were restored. Full-package commands, with no output pipes:

```sh
go test ./internal/acheck/ > /tmp/acheck-wave3-green.log 2>&1
go test ./internal/acheck/ -count=1 > /tmp/acheck-wave3-final.log 2>&1
```

Both exited 0. Before mutants and after restoration, respectively:

```text
ok  	github.com/system-inc/adamic/internal/acheck	8.815s
ok  	github.com/system-inc/adamic/internal/acheck	6.016s
```

## Final wave: Lower ruling and reader comparisons

October 27, 2026. Branch typescript/acheck-known-last starts at
6118fbd3b84cca990a76894f0d5f59eb0575e9e3. This section supersedes wave 3's
live-reader stops under the user's explicit header-only authorization.
Lower's October 27 ruling admits the four predicate programs: a parameter's
own predicate type is admitted when every argument passed to it is proven;
a predicate never called narrows nothing; Array.isArray is the library exact
predicate. Their first lines are dated history comments without acheck promises.

Nine programs change only their existing first line, keeping line counts and
every body byte unchanged. No compiler, Go test, acheck logic, or reader changes.
Exactly nine known.txt entries come off. The single retained entry is
adapt/75-optional-widening/probes/read-write.a, whose required added header
would shift the reader's coordinate-bearing output.

Before editing, repeated git grep -n -F <file name> outside the dated header
records and BLOCKERS.md, saved as /tmp/acheck-wave4/readers-<file name>.log;
per-file git history is alongside it. The earlier source/notes audit remains
applicable. Old promises for the five header replacements live in the table,
not in additional source lines.

| Program relative to stage3/ | History: old first-line promise | Decision and purpose | Today's output on one line |
| --- | --- | --- | --- |
| adapt/71-writable-views/language-questions/generic-storage.a | refused the non-null assertion ! | header fixed: Ambient console stops before the diagnostic-storage question. | exit 1; adamic: stage3/adapt/71-writable-views/language-questions/generic-storage.a:2:1: Adamic 0.1 refuses an ambient value declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values |
| adapt/71-writable-views/language-questions/node-flags.a | refused a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | header fixed: Ambient console stops before the mutable Node flags question. | exit 1; adamic: stage3/adapt/71-writable-views/language-questions/node-flags.a:2:1: Adamic 0.1 refuses an ambient value declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values |
| adapt/71-writable-views/language-questions/text-range.a | refused a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | header fixed: Ambient console stops before the generic position-setter question. | exit 1; adamic: stage3/adapt/71-writable-views/language-questions/text-range.a:2:1: Adamic 0.1 refuses an ambient value declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values |
| adapt/75-optional-widening/probes/read-write.a | (none) | Left byte-identical: adding a first line moves every AST site line. | exit 1; stage3/adapt/75-optional-widening/probes/read-write.a:11:13: error TS2345: Argument of type 'number &#124; undefined' is not assignable to parameter of type 'string'. \n   Type 'undefined' is not assignable to type 'string'. |
| drivers/parser/native-array-is-array-predicate.a | refused a type predicate whose return is not proven (return expression is not a trusted check on value) | history fixed: Lower admits the library exact Array.isArray predicate. | exit 0; stderr empty; checked |
| drivers/parser/native-function-any-view.a | refused a namespace | header fixed: Current callable-view refusal replaces the namespace refusal. | exit 1; adamic: stage3/drivers/parser/native-function-any-view.a:5:56: Adamic 0.1 refuses a value of type (level: number) => boolean seen as AnyFunction, which can write void where boolean is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable) |
| drivers/parser/native-predicate-callback.a | refused a type predicate whose return is not proven (there is no body proving this parameter) | history fixed: Lower admits the parameter predicate when passed arguments are proven. | exit 0; stderr empty; checked |
| drivers/parser/native-predicate-function-view.a | refused a type predicate whose return is not proven (there is no body proving this parameter) | history fixed: The predicate is never called and narrows nothing. | exit 0; stderr empty; checked |
| drivers/parser/native-predicate-overload.a | refused a type predicate whose return is not proven (there is no body proving this parameter) | history fixed: Lower admits the proven callback arguments through the overload. | exit 0; stderr empty; checked |
| drivers/parser/native-process-next-tick-read.a | type error TS2591 | header fixed: Current ambient declaration refusal replaces TS2591. | exit 1; adamic: stage3/drivers/parser/native-process-next-tick-read.a:6:1: Adamic 0.1 refuses an ambient value declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values |

### Reader before/after evidence

Used Node v24.19.0 and stock TypeScript 6.0.3 from
/workspace/adamic-tools/typescript/lib/typescript.js (CENSUS_TYPESCRIPT and
TSC_ADAPT_TYPESCRIPT select that module; NODE_PATH names its package directory).
Every reader run captured stdout, stderr, and exit status separately in
/tmp/acheck-wave4/before/ and after/. All six witness runs were compared byte
for byte, including assertion stack traces; no normalization was needed.

| witness.cjs arguments | Before / after exit | Before / after output | Comparison |
| --- | --- | --- | --- |
| --generic-storage | 0 / 0 | {"diagnostics":0,"promised_type":"SourceFile","output":["1"],"source_write":"generic-storage.a:12"}; stderr empty | stdout/stderr byte-identical |
| --generic-storage --mutant | 1 / 1 | stdout empty; ERR_ASSERTION at witness.cjs:33:8, actual ['0'], expected ['1'] | stdout/stderr byte-identical |
| --node-flags | 0 / 0 | {"diagnostics":0,"promised_type":0,"output":["1"],"source_write":"binder.ts:1113:17"}; stderr empty | stdout/stderr byte-identical |
| --node-flags --mutant | 1 / 1 | stdout empty; ERR_ASSERTION at witness.cjs:33:8, actual ['0'], expected ['1'] | stdout/stderr byte-identical |
| (none) | 0 / 0 | {"diagnostics":0,"promised_type":0,"output":["1"],"source_write":"utilities.ts:10645:5"}; stderr empty | stdout/stderr byte-identical |
|  --mutant | 1 / 1 | stdout empty; ERR_ASSERTION at witness.cjs:33:8, actual ['0'], expected ['1'] | stdout/stderr byte-identical |

The no-argument row above is node witness.cjs. Each mutant preserves the
reader's existing zero-write mutation and fails the same output assertion.

probe-read-write.cjs ran before and after the trial added header; both exited
0 with byte-identical stdout and empty stderr. Its stdout on both runs was:

```json
{
  "status": "pass",
  "original": {
    "total": 8,
    "read_only": 1,
    "write": 7
  },
  "alias_edge_dropped": {
    "total": 8,
    "read_only": 2,
    "write": 6
  },
  "checks": [
    "direct multi-variable alias reaches c.p = 2",
    "callback alias reaches alias.p = 3",
    "unrelated p symbol leaves read-only site unchanged",
    "dropping alias changes the write count",
    "write to another missing member is detected",
    "computed keyof member write is detected",
    "Object.assign alias write is detected",
    "untyped alias does not hide a write",
    "library-returned receiver alias is detected"
  ]
}
```

The printed summary omits coordinates. A temporary Node preload copied the
reader's actual sites.json writes before its finally block deleted its scratch
directory; it did not change source or reader logic. The eight Line/Column
pairs differed as follows:

| Before | After trial added header |
| --- | --- |
| 5:19 | 6:19 |
| 10:18 | 11:18 |
| 19:22 | 20:22 |
| 21:27 | 22:27 |
| 24:22 | 25:22 |
| 27:20 | 28:20 |
| 30:21 | 31:21 |
| 33:22 | 34:22 |

All eight lines increase by one; columns are unchanged. The trial header was
removed and read-write.a verified byte-identical to parent 6118fbd3. Its known
entry stays. render-read-write.py does not read the probe's summary or its
eight-site fixture output: it takes a separate 391-row production report and
asserts len(rows) == 391. It was therefore not run on this fixture output.

front21/run-probe.py, front22/run-probe.py and front24/minimals.py are dated
evidence runners, not active tests, as clarified by the user. They were not
rerun against their historical scratch compilers. For both parser sources,
verified that only line one changes, the remaining source bytes and line count
are identical, and stock TypeScript comment-free transpilation is identical.
This is input/runtime-text parity, not a newly claimed historical native run.

### Final-wave commands and validation

Setup exited 0 in 12.738s, full log /tmp/acheck-wave4-setup.log. Each build/test
shell sourced /workspace/adamic-tools/env.sh.

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /tmp/acheck-wave4-setup.log 2>&1
go build -o /tmp/acheck-wave4-adamic ./cmd/adamic > /tmp/acheck-wave4/build.log 2>&1
# Each of the ten paths, from the repository root; stdout discarded:
/tmp/acheck-wave4-adamic c <path>
go test ./internal/acheck/ > /tmp/acheck-wave4/green.log 2>&1
```

Front-end exit codes and full stderr: /tmp/acheck-wave4/frontends.json.
No emitted C is included.

Temporarily restored only native-array-is-array-predicate.a's old first line,
with its known entry still removed. Ran:

```sh
go test ./internal/acheck/ -run '^TestEveryProgramOutsideAPackageKeepsItsPromise$' -count=1 > /tmp/acheck-wave4/mutant.log 2>&1
```

Exit 1, exact red line:

```text
    acheck_test.go:91: stage3/drivers/parser/native-array-is-array-predicate.a breaks its first line's promise: checked, and its first line promises refused a type predicate whose return is not proven (return expression is not a trusted check on value)
```

Restored the history comment, then ran the complete package again:

```sh
go test ./internal/acheck/ -count=1 > /tmp/acheck-wave4/final.log 2>&1
```

Both full-package runs exited 0, before and after the mutant respectively:

```text
ok  	github.com/system-inc/adamic/internal/acheck	6.357s
ok  	github.com/system-inc/adamic/internal/acheck	6.449s
```

## TypeScript fixture rename: final known entry

October 27, 2026. Branch typescript/acheck-known-readwrite starts at
cb9a9dd69544ec06d41a6321de2d4d8fdc74a3bb. This supersedes the preceding
decision to retain the read-write.a entry: the user identified it as stock
TypeScript input, not an Adamic program.

Renamed adapt/75-optional-widening/probes/read-write.a to read-write.ts with
git mv. Contents are byte-identical to the parent blob, with no added header.
Updated the sole live source-path reader, probe-read-write.cjs:12.
The reader still writes read-write.a in its temporary tree, uses that name in
its scratch tsconfig, and reports that name in its site records. Removed the
fixture's known.txt entry. No stage3 entries remain (known.txt has only its
existing comment preamble). No compiler, Go test, or acheck logic changed.

Ran git grep -n 'read-write.a' across the repository before editing;
full results: /tmp/acheck-wave5/readers.log. Other executable-script matches
refer to evidence/read-write/analysis.json.gz, a separate production analysis,
or to unrelated stage3/interface-downcasts/v2 fixtures. Historical acheck,
oracle count and review records are unchanged. render-read-write.py reads a
separate 391-row production report and does not consume probe-read-write.cjs's
eight-site fixture summary; no renderer run was required.

### Reader comparison

With /workspace/adamic-tools/env.sh sourced, ran the unchanged reader command
before and after the rename and path update:

```sh
TSC_ADAPT_TYPESCRIPT=/workspace/adamic-tools/typescript/lib/typescript.js node stage3/adapt/75-optional-widening/probe-read-write.cjs
```

Both runs exited 0. Full stdout and stderr are saved separately in
/tmp/acheck-wave5/before.stdout, before.stderr, after.stdout and after.stderr.
Compared both streams with cmp: stdout exit 0; stderr exit 0. Stderr is empty
on both runs. The stdout records have the following SHA256:

| Run | Exit | Stdout bytes | Stdout SHA256 | Stderr |
| --- | --- | --- | --- | --- |
| before | 0 | 629 | 2e22e1e7114d0aab6243da9fe96736ab2c744fd63dc039b9cba4bef038673e1f | empty |
| after | 0 | 629 | 2e22e1e7114d0aab6243da9fe96736ab2c744fd63dc039b9cba4bef038673e1f | empty |

Exact stdout, identical before and after:

```json
{
  "status": "pass",
  "original": {
    "total": 8,
    "read_only": 1,
    "write": 7
  },
  "alias_edge_dropped": {
    "total": 8,
    "read_only": 2,
    "write": 6
  },
  "checks": [
    "direct multi-variable alias reaches c.p = 2",
    "callback alias reaches alias.p = 3",
    "unrelated p symbol leaves read-only site unchanged",
    "dropping alias changes the write count",
    "write to another missing member is detected",
    "computed keyof member write is detected",
    "Object.assign alias write is detected",
    "untyped alias does not hide a write",
    "library-returned receiver alias is detected"
  ]
}
```

### Full-package and mutant proof

All test output went directly to log files, with no output pipes:

```sh
go test ./internal/acheck/ > /tmp/acheck-wave5/green.log 2>&1
# Temporarily rename only probes/read-write.ts back to read-write.a,
# leaving its known.txt entry removed and the live reader unchanged:
go test ./internal/acheck/ -run '^TestEveryProgramOutsideAPackageKeepsItsPromise$' -count=1 > /tmp/acheck-wave5/mutant.log 2>&1
# Restore read-write.ts:
go test ./internal/acheck/ -count=1 > /tmp/acheck-wave5/final.log 2>&1
```

The rename-back mutant exited 1; exact red line:

```text
    acheck_test.go:91: stage3/adapt/75-optional-widening/probes/read-write.a breaks its first line's promise: type error (stage3/adapt/75-optional-widening/probes/read-write.a:11:13: error TS2345: Argument of type 'number | undefined' is not assignable to parameter of type 'string'.)
```

Both full-package runs exited 0, before and after the mutant:

```text
ok  	github.com/system-inc/adamic/internal/acheck	6.300s
ok  	github.com/system-inc/adamic/internal/acheck	6.190s
```
