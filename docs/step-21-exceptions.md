# Step 21: native exceptions

## Census: observed evidence

Delivery base: `8cb5e7c189fc431ab3cab7e0be2c0530e722a598`, the specified
`origin/compiler/area-next-fixtures` candidate. The delivery branch is
`codex/scout-exceptions`. No other worker branch is merged.

The full compiler-scope census committed on this base is
`stage3/meter/runs/20261008T035244Z.latent-full/compiler/full.jsonl.gz`.
It is measured on a checker-rejected, adapted program. A root here means a
unique attempted unit (`unit`), not a repeated finding or an entire imported file.
The hidden ranking at `6c4fc1af` instead measures compiler `ed6e2975`:
its bytes are outermost-cause credit, an estimate of bytes exposed by removing
one blocker, not successful compilation or dynamic heat. Those snapshots cannot
be added or presented as one current-base measurement.

**Observation:** neither snapshot records an exception-specific Refused or NotYet
reason. The exact-reason census has zero roots for the production exception
reasons below. The historical ranking has no matching reason entry, so its hidden
credit is zero, not a measurement that these constructs have no hidden source.
There cannot be three diagnostic witnesses for a reason with no observations.

| Production diagnostic shape | Base roots | Historical hidden credit | Witnesses |
| --- | ---: | ---: | --- |
| Refused: `throwing a <type>` | 0 | 0 | none recorded |
| NotYet: `throwing an Error that isn't made where it's thrown or caught by the catch around it` | 0 | 0 | none recorded |
| NotYet: `a catch that destructures what it caught` | 0 | 0 | none recorded |
| NotYet: `new Error with options` | 0 | 0 | none recorded |
| NotYet: `new Error with a message that isn't a string` | 0 | 0 | none recorded |
| NotYet: `a try around <operation>, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md)` | 0 | 0 | none recorded |

The last template covers repeat, normalization, number formatting, array and
 typed-array ranges, RegExp flag checks, frozen-object writes and Hash
finalization. It is a representation/behavior barrier, not a request to silently
turn a catchable failure into a terminal panic.

The refusal table supplied inside the ranking has no exception-specific row.
`reading exception` is **not** catch narrowing: it names the generator transform's
`ExceptionBlock` local. Its 12 historical boundaries receive zero hidden credit;
three witnesses are `transformers/generators.ts:2250`, `:2252`, `:2257`.
`reading error` names diagnostic-message locals in `checker.ts:36462`, `:36467`,
`:36489`, also zero credit. CatchClause/ThrowStatement type names denote compiler
AST data; their checked-view and writable-variance barriers belong to those steps.
Detached `throwIfCancellationRequested` belongs to method binding. Counting these
as retired exception roots would be incorrect.

The original stock census does expose the catch boundary through **checker**
TS18046: eight diagnostics, seven distinct file:lines, five file roots. Witnesses:
`commandLineParser.ts:2301`, `program.ts:406`, `sys.ts:1553`.
The other lines are `program.ts:441`, `sys.ts:1280`, `tracing.ts:66`, and
`transformers/jsx.ts:187` (two accesses). Hidden bytes are unavailable for this
exact checker code: the ranking attributes checker bodies as one cause, not by
TS diagnostic. Do not substitute that bucket's 1,730,511 bytes for TS18046.
Disabling unknown catch variables would change the proof contract and is not built.

## Source exposure, distinct from refusal roots

A pinned TypeScript 6.0.3 AST walk over `src/compiler/**/*.ts` finds 34 executable
try statements: 26 try/catch and eight try/finally. There are nine `throw new Error`
and three identifier throws. There are no executable try/catch/finally combinations,
other throw expressions or Error subclasses in this corpus. Helper template strings
and comments are not executable compiler-source AST and are excluded.

| Shape | Source sites | Three original file:line witnesses |
| --- | ---: | --- |
| try/catch | 26 | `commandLineParser.ts:2297`, `moduleSpecifiers.ts:134`, `sys.ts:1549` |
| try/finally | 8 | `checker.ts:1937`, `symbolWalker.ts:50`, `utilities.ts:787` |
| throw new Error | 9 | `core.ts:1583`, `moduleNameResolver.ts:1681`, `utilitiesPublic.ts:448` |
| throw identifier | 3 | `debug.ts:203`, `program.ts:2854`, `sys.ts:1555` |

Two identifier throws rethrow catch bindings; debug.ts constructs an Error in a
local before throwing it, currently NotYet when reduced. None of these counts is
a count of independently compilable roots. Earlier checker, host, signature and
body barriers can prevent exception lowering from being attempted.

## Reproduction and validation

`python3 docs/step-21-exceptions/census.py` reads the two pinned artifacts and
checks the distinction between actual exceptions and compiler AST names.
`census.json` retains its output. Its root count deduplicates attempted units.
A selector mutant that admits `reading exception` must fail the exclusion
assertion. An accounting mutant that drops one unknown-catch diagnostic must fail
the independent eight-diagnostic assertion. No native fixture is added in this unit.

Setup succeeded with `GOPROXY='https://proxy.golang.org|direct'` and selected
`/workspace/adamic-tools/env.sh`. Cumulative timing lines: Node 0.229s,
Go 0.235s, clang 0.950s, markdown ready 1.485s, submodules 23.056s,
Go build 211.122s, test binaries deferred 211.235s, cache warm 211.237s,
done 211.267s. `nproc` printed 5; cgroup quota is four CPUs.
