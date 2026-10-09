Historical initial investigation. For the final implementation and disposition, see [FINAL.md](FINAL.md).

Investigated all 22 assigned new-expression roots; no lowering change is claimed.
Base b410340dc8f889b5799c3bc519117c63def3aa24; replay 9a1f14c5d994aa855625e7cfa295677060348fec; merge 915bfb9ef2063f3e9300c4469919c33dcbdfdbd0.
Five census replays: two exact reproductions, three earlier stops; setup succeeded in 231.369s, nproc 5.
No implementation rules or checks were added, and no mutants were run.
All 22 roots remain uncovered by the required native, JavaScript, sanitizer, and mutant validation.

Assumption: the unit-specific origin/area/compiler base overrides the generic
origin/main starting instruction. The main-only fetch refspec required explicit
fetching of the three named refs. Current fetched origin/main
(d65e2d5e65197969480c2417832ca58fa007c933) is already an ancestor of this branch.
Only this report and evidence were added after the requested replay-tool merge.
No production compiler function, shared hook, or cohere code was changed.

| Kind | Assigned roots | Disposition | Validated lowered roots |
| --- | ---: | --- | ---: |
| new an Identifier | 13 | skipped: several independent facilities outside newExpression | 0 |
| new a ParenthesizedExpression | 8 | skipped: runtime constructor values need representation and dispatch work | 0 |
| new a PropertyAccessExpression | 1 | skipped: external Node inspector constructor; replay encounters earlier closure stop | 0 |

These are blocked scope findings, not new language refusals. No root is cancelled
as an echo, and no NotYet is relabeled Refused. Weak collection lifetime semantics
need a memory-model decision before implementation; this report does not assert
that Adamic forbids them by design.

## Observations

The fresh adapted corpus matches all 81 entries of the table branch's
source-manifest.json, byte-for-byte by SHA-256. Source paths differ only in the
scratch directory prefix. The table's compiler base is 44583d32; this worker uses
the requested newer compiler-area base. Replay does not establish that an earlier
stop is an echo or that the original census finding is gone.

| Site | Exit | Observed selected-unit result | Replay total |
| --- | ---: | --- | ---: |
| binder.ts:632:16 | 1 | createSymbol signature stops at 630:47, a value of type __String | 117.877s |
| checker.ts:2635:24 | 1 | createSymbol signature stops at 2633:47, a value of type __String | 2.775s |
| checker.ts:2940:58 | 0 | reproduced new a ParenthesizedExpression in getNodeLinks | 2.586s |
| factory/baseNodeFactory.ts:42:16 | 0 | reproduced new a ParenthesizedExpression in createBaseSourceFileNode | 2.275s |
| sys.ts:1656:29 | 1 | selected enclosing function-expression statement at 1463:1; closure stop at 1469:5 and NonNullExpression at 1964:12 | 2.254s |

The first replay includes cold overlay compilation while setup was warming the
build cache. Timings are single observations, not comparative benchmarks.
[evidence/replays.json](evidence/replays.json) preserves the selected units,
complete findings, and command messages. Compressed original logs preserve the
complete JSON records, including declaration ledgers.

The identifier bucket's 13 distinct sites split as follows:

- Six weak collections: WeakMap at debug.ts:617 and :618,
  utilities.ts:10222, factory/nodeChildren.ts:11; WeakSet at parser.ts:9927
  and :9936.
- One Uint16Array at utilities.ts:10491.
- Five allocator-provided constructors: Symbol at binder.ts:632 and
  checker.ts:2635, Type at checker.ts:5518 and :5532, Signature at
  checker.ts:14134.
- One class-expression binding, SymbolLinks at checker.ts:2935.

The parenthesized bucket is seven cached allocator-constructor selections
(baseNodeFactory.ts:42, :46, :50, :54, :58; nodeFactory.ts:7415;
parser.ts:433), plus checker.ts:2940's NodeLinks function cast through any.
The property site is new inspector.Session() after require("inspector").
[evidence/sites.json](evidence/sites.json) records all 22 original expressions.

## Why this exceeds the owned function

Observed implementation: newExpression recognizes registered class declaration
symbols and selected library globals. construct instantiates a source class and
calls its fixed allocator. staticConstruct checks source-class readiness but does
not dispatch through the runtime value written after new. The IR records class
constructors as fixed function indices. representation treats a construct-only
structural type as Object; only call signatures select Closure.

Inference: dispatching through constructor parameters, mutable cached bindings,
returned constructor values, or object properties requires a constructor value
representation, runtime target dispatch, captured state and ownership support.
Choosing the class from its result type would discard the actual constructor
value and its evaluation effects. Those changes involve class/function/value
lowering, IR and both backends, rather than a minimal shared hook. Function-based
constructors additionally require JavaScript allocation, this binding, and return
replacement semantics. They cannot soundly become ordinary function calls.
Class-expression registration is also outside this unit's owned function.

Uint16Array is explicitly outside the three supported typed-array element kinds
in typed_arrays.go; adding it requires representation, conversion and backend
runtime support outside this territory. WeakMap/WeakSet cannot become ordinary
strong maps or sets without changing lifetime semantics; weak-key lifetime and
WeakMap value/key cycles need a separately owned memory implementation.
inspector.Session is a Node host API, requiring host bindings and behavior beyond
an ordinary class allocator. The NodeLinks cast through any also needs a ruling
on the source contract; this worker did not accept the unchecked cast.

## Commands and validation limits

After exporting GOPROXY='https://proxy.golang.org|direct', ran bash cloud/setup.sh
with output in /tmp/new-expression-setup.log, then sourced
/workspace/adamic-tools/env.sh for commands. Setup timing lines: Node 0.026s,
Go 0.037s, clang 0.272s, markdown install 0.940s, markdown ready 1.037s,
submodules 17.278s, Go build 231.234s, test binaries deferred 231.336s,
build cache warm 231.338s, done 231.369s. nproc was 5, cpu.max was
400000 100000. [Raw setup output](evidence/setup.log.txt).

Ran bash stage3/apply.sh /tmp/new-expression-adapted with output in
/tmp/new-expression-adapt.log, exit 0. Each replay used:

```sh
go run ./stage3/census/latent/replay \
  -project /tmp/new-expression-adapted/src/tsc/tsc.ts \
  -where /tmp/new-expression-adapted/src/compiler/FILE:LINE:COLUMN \
  -kind NotYet -reason 'EXACT REASON' > /tmp/new-expression-replay-NAME.log 2>&1
```

FILE:LINE:COLUMN and EXACT REASON are the five rows above and their assigned
kind strings. The replay worker checks its no-output guards; it does not invoke
a backend. No reduced fixture was added because neither backend can execute the
required dynamic constructor lesson within the authorized territory. No oracle
test, sanitizer run, counts refresh, mutant, whole-package test or full gate was
run. counts.md has no new fixture to record. git diff --check passed.

This is an evidence-only branch. It must not be reported as lowering any of the
22 sites or as completing the unit's semantic acceptance requirements.
