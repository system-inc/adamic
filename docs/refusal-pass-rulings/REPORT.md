Built up-front class-merge, Function annotation, permanent-construct and temporary Record refusals, with 28 pinned fixtures.
Code commit: eacadddaec5cd19d9bd4d0a2f2d01f32f2cadc59; base and landing main: 39638d9e278d38bb5aeae887f46d55a70e47aaad.
Commands: affected package tests, filtered uncached oracle, vet and formatting passed; all 282 relevant probe programs now refused.
Mutants: class merge, Function annotation, lowering first, Record, any, eval, expando and new Function each failed TestRefusalPassRulings.
Not covered: full repository gate, optional widening, exhaustive module augmentation cases, or implementing the records design.

Follow-up: [developer tools runtime effects and record operation diagnostics](FOLLOWUP.md).

## Scope and source

Branch `codex/refusal-pass-rulings`, from current origin/main. The final fetch and
`git merge origin/main` reported `Already up to date.` No PR was opened.
The probe report was read from fetched `devtools/refusal-probes` (the user's
`5e3c7b7`), `internal/refusalprobe/REPORT.md`, its catalog, and
`results/main-2000.jsonl`. Its target is exactly this unit's base, `39638d9`.
No cohere source was copied. The expando check calls the existing AST shim.
No protected compiler file has a committed change. `lower.go` was changed only
transiently for the explicitly requested ordering mutant, then restored.

## The rulings

1. Class/interface merges that add properties, inherited interface members, call
   signatures or construct signatures are refused. Class/namespace merges name
   the class and namespace rather than returning the generic namespace message.
   Declaration order does not change the result. Real class members described
   again by an interface remain allowed. Interface/interface merging remains
   allowed, with both members required by the checker at construction.
2. The library Function type is refused at every TypeReference, including an
   unused parameter, field, variable, type argument, union member and alias.
   The fix spells out a function type with its parameters and result. A user's
   own interface or class named Function remains allowed.
3. Any keywords, the library eval and Function constructor, and properties whose
   checker declarations are expando assignments are caught by the up-front pass.
   Dot writes and literal-key computed writes are covered. Each construct has a
   pinned complete diagnostic, plus a fixture with an earlier lowering limitation.
4. Main does not implement #p9v82wa. All 40 exact Record finding programs from the
   report compile on main (replayed with `adamic js`); all 40 are now refused.
   The temporary refusal checks the resolved string index type, so a generic alias
   such as `Bag<string, number>` cannot bypass it. Finite literal-key Records and
   a user's own unrelated Record alias remain allowed.

## What reaches NotYet first

Observed in main: `typeOf` returns NotYet for an any-valued declaration;
identifier lowering returns NotYet for reading eval; assignment lowering returns
NotYet for a function's expando property; `newExpression` returns NotYet for new
Function. These constructs had no corresponding check in `refusals.go`.

History inspection: `7c9f7d2` already runs refusal before declaration and statement
lowering, just as current main does. Its refusal pass also lacks these four checks;
its `typeOf` already has the any NotYet fallback. Therefore no reversal of pass
order was found. The repair adds the missing checks to the existing first pass.
The ordering mutant proves the fixtures detect an actual reversal. No bisect was
needed, and no historical runtime result is inferred from inspecting source.

## Main against Node, before the fix

The observation programs are now pinned refusal fixtures:

```a
class Box { n = 1; } interface Box { extra: number; }
console.log(typeof new Box().extra);
```

Node source: stdout `undefined\n`, exit 0. JavaScript backend: the same stdout
and exit 0. Native: empty stdout, stderr
`adamic: panic: compiler bug: a field the checker proved is there is missing`,
exit 70. Both backend compilations exit 0. **No silent wrong output** was observed:
the JavaScript backend agrees with Node; native stops loudly. The class's claimed
number member still does not exist, which the Node observation demonstrates.

```a
function take(value: Function): void { console.log(`${value(42)}`); }
take((value: string): string => value.toUpperCase());
```

Node source reaches a TypeError; through the repository runner it has empty stdout,
stderr `adamic: panic: TypeError: value.toUpperCase is not a function`, exit 70.
Both backend compilation attempts exit 1 with
`stage 0 can't lower a call to an Identifier yet`, at line 1, column 55.
**No silent wrong output** was observed: main cannot lower this call, even though
it accepts the unused Function annotation in the original probe. A backend run
of this observation program is impossible on main because no output is generated.

The exact Record program demonstrating acceptance without the design is:

```a
const value: Record<string, number> = {};
console.log(typeof value.missing);
```

Node source and the JavaScript backend print `undefined\n`, exit 0. Native
compiles successfully and then panics with the same missing-field compiler-bug
message, exit 70. A genuinely dynamic key also shows the missing implementation:

```a
const value: Record<string, number> = {};
const key: string = 'missing';
console.log(typeof value[key]);
```

Both backend compilation attempts return NotYet for an ElementAccessExpression
at line 3, column 20. Node prints `undefined\n`, exit 0. Reading
`value['toString']` is already refused as an inherited library member, and
`'toString' in value` is already refused by the generic in rule; Node prints
`function\n` and `true\n` respectively. Those two generic refusals alone do not
establish the own-key record design. The named missing read and the dynamic read
show why accepting the annotation is premature.

The first direct JavaScript execution lacked the adamic runtime package and failed
module resolution. That is a discarded invocation, not a compiler observation.
The reported JavaScript results were rerun with `oracle/node.mjs`, which supplies
the repository runtime. Only those successful runner results are retained.

## Developer tools note

For now the Record refusal is temporary, with #p9v82wa named in its fix. When the
records design lands, its own-key operations and T | undefined dynamic reads
should be executed against Node, with prototype-name reads and in required to
stop loudly. A blanket expectation that Record must forever be refused would be
wrong. This note is recorded here and in docs/0.1.md; no external message was sent.

## Verification and commands

Every test command wrote to a log, never a pipe. The printed environment file was
`/workspace/adamic-tools/env.sh`, sourced in each toolchain shell.
`bash cloud/setup.sh` printed Go ready 0s, clang ready 1s, Node ready 1s,
submodules ready 1s, build cache warm 120s, done 120s. `nproc` printed 5;
cgroup cpu.max was `400000 100000`. Versions: Go 1.27.1, clang 20.1.8,
Node 24.19.0. See [setup.log](setup.log).

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/lower ./internal/load > /tmp/refusal-packages.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^load$/^testdata$/^0.1$/^compile$/^01_hello[.]ts$' > /tmp/refusal-oracle.log 2>&1
go vet ./... > /tmp/refusal-vet.log 2>&1
gofmt -l cmd internal > /tmp/refusal-format.log
git diff --check
ADAMIC_GATE_UNCACHED=1 python3 docs/refusal-pass-rulings/mutants.py > /tmp/refusal-mutants.log 2>&1
```

Final package results: lower passed in 10.773s, load in 0.747s. The filtered oracle
passed in 11.736s and actually ran 01_hello.ts: native misses 3, Node misses 2,
all cache hits 0. Vet exited 0; formatting and diff checks printed nothing.
No runtime fixture was added to the oracle, so its counts table needs no new row.
The full repository gate was not run; this is a focused worker gate.

Replay of the exact finding Sources with the fixed compiler returned Refused with
its expected construct text for 41 any, 41 expando, 40 eval, 40 Function type,
40 new Function, 40 Record and 40 merging programs: **282 of 282**. Optional widening
was excluded because it is outside these four rulings. See [probe-replay.log](probe-replay.log).
Main accepted **40 of 40** Record programs; see [record-main.log](record-main.log).

For each observation program, the commands were:

```sh
go build -o /tmp/refusal-main ./cmd/adamic
# This binary was built on unchanged 39638d9 before editing compiler source.
node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/refusal-observations/NAME.a
/tmp/refusal-main js /tmp/refusal-observations/NAME.a > /tmp/refusal-observations/NAME.js
node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/refusal-observations/NAME.js
/tmp/refusal-main build /tmp/refusal-observations/NAME.a -o /tmp/refusal-observations/NAME.native
/tmp/refusal-observations/NAME.native
```

Each stdout/stderr was redirected to its own observation log. NAME was merge,
function, record-lie, record-missing, record-prototype or record-in. Backend runs
were skipped when compilation failed. The retained logs are in [observations](observations).

## Mutants actually run

Each mutant began from the final source independently, and both compiler files
were restored in a finally block after every run. Each ran
`go test ./internal/lower -run '^TestRefusalPassRulings$' -count=1`, with
ADAMIC_GATE_UNCACHED=1. All eight exited 1 through the intended assertion, with
no build failure or clang warning. The affected packages were re-greened afterward.

| Mutant | What caught it |
|---|---|
| Drop class merge check | class-interface.a accepted; namespace fixtures also caught the wrong generic message |
| Drop Function annotation check | function-parameter.a, field, variable, type argument, union and alias fixtures accepted |
| Put refusal after lowering | any.a, eval.a, expando.a and new-function.a returned NotYet; order fixtures also failed |
| Drop Record check | record.a, record-alias.a, record-generic-alias.a and observe-record.a accepted |
| Drop any check | any.a returned NotYet |
| Drop eval check | eval.a returned NotYet |
| Drop expando check | expando.a and expando-index.a returned NotYet |
| Drop new Function check | new-function.a returned NotYet |

See [mutants.log](mutants.log), the individual mutant logs, and [mutants.py](mutants.py).

## Limits

No exhaustive cross-module augmentation suite, generic key-space census, or
full source corpus gate was run. The Function runtime call was blocked on main,
so no backend runtime result is claimed for it. Records were refused rather than
implemented. No change addresses the optional-widening finding or any other
construct from the original probe report. Nothing was merged or pushed to main
or an area branch.
