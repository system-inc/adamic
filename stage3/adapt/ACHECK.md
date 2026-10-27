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
