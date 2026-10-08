# Async and suspension gaps for stage 3

Measured on 2026-10-08. The highest-count suspension gap is **generators: one
observed tsc NotYet root**. Promise combinators and async iteration remain native
compatibility gaps, but this census gives them no measured root-count priority.
The generator work is deferred here: it needs a separate ownership and resume
protocol, beyond a small extension of the approved async-function machinery.

## Inputs and counting method

- Runtime baseline: `origin/area/runtime` at
  `c1c6073021e8e5cd6005fcdee59eb3b377395fd8`.
- Evidence branch: `origin/codex/stage3-notyet-table` at
  `e8c283b5ed32477805357b652a170b85a04b2469`.
- Primary ranking input: [morning after roots.csv](https://github.com/system-inc/adamic/blob/e8c283b5ed32477805357b652a170b85a04b2469/stage3/notyet-table/rerun-0730/after/roots.csv).
  Its measured compiler is `0e5661e4243af95ae3247d066de14ccb2b582a94`,
  not the runtime baseline above. The table is evidence of demand, not a fresh
  census of this branch.
- Cross-check: [before roots.csv](https://github.com/system-inc/adamic/blob/e8c283b5ed32477805357b652a170b85a04b2469/stage3/notyet-table/rerun-0730/before/roots.csv),
  the original root summary, and the after referenced-other-roots ledger.
- Native capability evidence: [async lowering probes](../internal/lower/async_test.go),
  [oracle reader probes](../internal/oracle/async_test.go),
  [oracle async refusal sources](../internal/oracle/testdata/async_refused),
  [async normalization](../internal/flow/async_normalize.go), and the generator
  refusal in [refusals.go](../internal/lower/refusals.go).

Count distinct `(kind, where, reason, text)` root signatures from the root-filtered
CSV. Do not count repeated attempts, symbol-linked echoes, or words in compiler
variable names. The after census has 6,703 NotYet roots, 472 root reasons and
4,047 removed echo-only sites. It ran on a checker-rejected entry program with
324 checker diagnostic sites and 147 checker-split units; failed compound
boundaries also limit coverage. **Zero means no root observed, not proof that a
feature is implemented or absent from all code paths.** Stage 1 has no equivalent
root census in this input; its source scan below is not added to tsc root counts.

## Ranked native gaps

All zero-count rows tie. Their order below groups related work and does not imply
an extra measured priority. Counts describe the after census before any work on
this branch.

| Rank | Feature | Observed tsc roots | Evidence and native status on the runtime baseline |
| --- | --- | ---: | --- |
| 1 | Generators, `yield`, and generator delegation | 1 | `src/compiler/core.ts:1045:9`: `a YieldExpression as a statement`. This is also present in the before census. Production lowering refuses every generator function or method, even without a yield. This observed site is a synchronous generator demand, not evidence of an async generator. |
| =2 | `Promise.all` | 0 | `TestAsyncGapsNameTheMissingPiece` refuses `await Promise.all([Promise.resolve(1)])`. The design requires input-position results, first-observed rejection, live siblings and queued observation of an empty input. |
| =2 | `Promise.race`, `any`, `allSettled`, and Promise reaction APIs | 0 | They are outside the trusted Promise surface in `lower/async.go`. Their settlement, rejection and lifetime rules require separate witnesses; the generic diagnostic naming Promise.all is not evidence that only all is missing. |
| =2 | Async iterators, `for await`, and async generators | 0 | `lower/object.go`, `library_array.go` and `typed_arrays.go` explicitly reject for-await. Generator refusal applies to async generators too. Iterator closing and pending next/return/throw ownership are additional work. |
| =2 | Await in catch or finally; async finally completion routing | 0 | Named probes refuse await in catch/finally. Even a non-suspending finally in an async function remains NotYet. Saved return/throw/break completions must survive suspension and can be replaced by finally. |
| =2 | For-of and switch inside async functions | 0 | This baseline's normalizer refuses `async for-of and switch regions`. Arrays, strings, Maps/Sets and switch routing need their own continuation evidence. Support on another feature branch does not change this pinned baseline. |
| =2 | Captured per-iteration bindings | 0 | `async_refuse_loop_body_capture.a` and repeated-binding probes require a fresh environment for each iteration. Reusing one frame cell changes escaped closures' values. |
| =2 | Promise return/adoption and thenable assimilation | 0 | Named Promise-adoption probes and `async_refuse_return_thenable.a` / `async_refuse_arrow_thenable.a` refuse these shapes. A synchronous `then` member is not a substitute for queued, once-only Promise assimilation. |
| =2 | User Promise executors, escaping resolvers and pending host I/O | 0 | `new Promise` is refused as `Promise executors and pending-I/O cancellation`. Producer liveness, shutdown and request acknowledgement are not represented by a scalar fulfilled Promise. |
| =2 | Await among structural method operands; boolean-or-undefined frame slots | 0 | Both have explicit normalization refusals. Bound receiver/callee snapshots and slot representation need separate proofs. |
| =2 | Cyclic captured async frames and unowned tasks | 0 | `async_refuse_frame_capture_cycle.a` is a deliberate ownership refusal. Discarded tasks and `void task()` are also refused. These are safety constraints, not permissions to drop ownership checks for compatibility. |

A keyword search of the original unfiltered table also finds `reading isAsync`,
`reading isGenerator`, `reading awaitedType`, and similar compiler locals. Those
are ordinary binding/lowering findings, frequently tagged echoes, not evidence
that executing await or a Promise combinator needs that many roots. None occurs
as an async-semantic reason in the after root ledger. The referenced-other-root
ledger contributes no async/generator reason either.

## What the stage 1 ports actually execute

The stage 1 parser, ESTree converter, printer and lint rules recognize or emit
async/generator syntax. Examples include `typescript/parser/parser.ts`,
`cohere/estree/sourceParser.ts`, `cohere/tsprinter/expressions.ts`, and
`cohere/lint/rules/no-await-in-loop/rule.ts`. An `async: boolean` parameter, an AST
AwaitExpression case, or emitted text `await ` does not suspend the port itself.

A syntax-tree scan of every tracked stage1 `.ts` and `.a` input uses the pinned
TypeScript parser, traverses function modifiers, generator asterisks, Await/Yield
nodes, for-await modifiers and direct `Promise.all/race/any/allSettled` calls.
Comments and string contents are not counted. The scan found 790 inputs and zero
instances of those executing constructs; the input count also matches
`git ls-files stage1`. A source-syntax occurrence would still need
lowering provenance before being called a measured root.

The older syntax census in `stage3/census/data/sites.json` contains 16
`yield (generators)` sites, including checker generators and `yield array[i]` at
the core location. That is corroborating source demand, **not 16 roots in the
newer root-filtered census**. It has a different source manifest and counting
method.

## Class fields and default parameters

Direct await in a class field initializer or a function parameter initializer is
invalid JavaScript, including in modules and async functions. Node v24.19.0 rejects:

```js
class C { value = await Promise.resolve(1); }
async function f(value = await Promise.resolve(1)) {}
```

The errors are `Unexpected reserved word` and
`Illegal await-expression in formal parameters of async function`, respectively.
Those invalid forms have no ranked runtime gap. An awaited computed field name
is different valid syntax: Node v24.19.0 prints `1` for
`class C { [await Promise.resolve("value")] = 1; } console.log(new C().value);`
in a module. The baseline rejects computed field names in
`lower/class_inheritance.go`; the census does not identify an async-specific root
for that general class-lowering boundary. Calling an async function from an
initializer,
or placing an async arrow there, is different valid syntax; any failure must be
attributed to its actual Promise ownership, result representation or enclosing
lowering boundary. A compiler that parses or diagnoses the invalid snippets also
does not need native suspension in those positions to do its job.

## Why the top feature is not implemented here

[The async design](concurrency-async.md#throws-cleanup-and-cancellation) requires
owned live slots, explicit cleanup and pending completions. A generator can use
that general approach, but an async function's await/resume path alone is not its
protocol. [The generator design boundary](user-iterators.md#generators-need-owned-suspended-frames)
requires heap frames, explicit `next`, `return` and `throw` resume modes,
yielding from finally, delegation completion precedence, catchable reentrancy
failure and cycle analysis through erased iterator views. Dropping an abandoned
generator must release slots without executing user finally code.

A full implementation needs Node fixtures for lazy eager-prefix behavior,
mutation between next calls, throw into a suspended catch, return through a
finally that itself yields, reentrancy and abandonment. Mutants must resume the
wrong state, lose a live reference, route return/throw incorrectly and run finally
on abandonment. Materializing yields or implementing just the observed core loop
would evade those requirements and change JavaScript behavior.

This is a distinct generator landing, larger than the bounded implementation
requested here. The report preserves its one-root priority and stops the runtime
work; it does not replace the top item with a zero-root Promise feature. No
production semantics or registered oracle fixtures are changed, so fixture counts
need no regeneration.

## Validation

All checks passed before the report-only push:

- Pinned root audit: 10,083 before roots and 6,703 after roots; the same single
  Yield root in both; 409 referenced-other roots with no async/generator reason.
  All 11 ranked rows match the audit. Injecting either an extra Yield root or a
  falsely credited Promise.all root fails the expected-signature assertion.
- Stage 1 AST scan: 790 `.ts`/`.a` files, matching the tracked input count; zero
  executing constructs listed above.
- `go test ./internal/lower -run
  '^(TestAsyncGapsNameTheMissingPiece|TestAsyncReaderRefusals|TestAsyncRepeatedBindingRefusals|TestAsyncReturnThenableShapes|TestGeneratorsAreRefusedEvenWithoutYield)$'
  -count=1`: passed, package test time 3.339 s.
- `go test ./internal/oracle -run '^TestAsyncReaderProbes$' -count=1`: passed,
  package test time 96.367 s. Its existing sources run on Node and retain their
  named native lowering refusals.
- Node syntax checks: both invalid initializer examples produce the expected
  SyntaxErrors; the valid awaited computed field name prints `1`.
- `git diff --check`: passed. No new runtime fixture or production mutant was
  introduced; the two mutants above validate this report's measurement audit.

Setup used `GOPROXY=https://proxy.golang.org|direct`,
`bash cloud/setup.sh --wasi-sdk`, and its printed environment file
`/workspace/adamic-tools/env.sh`. Versions: Go 1.27.1, clang 20.1.8, WASI SDK 27,
Node v24.19.0. Machine `11f79d9f08c2`: Xeon Platinum 8573C, five visible CPUs,
four-CPU cgroup quota, 17.6 GB RAM. Shell `time` measured 52.336 s for the lower
command and 141.215 s for the oracle command, including rebuilding/linking test
binaries; these are validation durations, not runtime benchmarks. Setup took
635.586 s; its recorded load went from 0.47 to 7.04, and the observed concurrent
build load reached 14.90.

The root audit can be reproduced after fetching the pinned evidence commit:

```python
import csv, io, re, subprocess
pin = "e8c283b5ed32477805357b652a170b85a04b2469"
path = "stage3/notyet-table/rerun-0730/after/roots.csv"
text = subprocess.check_output(["git", "show", pin + ":" + path], text=True)
rows = list(csv.DictReader(io.StringIO(text)))
def semantic(rows):
    return {(r["kind"], r["where"], r["reason"], r["text"]) for r in rows
            if re.search(r"async|await|promise|generator|yield", r["reason"], re.I)
            and not r["reason"].startswith("reading ")}
roots = semantic(rows)
assert len(rows) == 6703 and len(roots) == 1
assert next(iter(roots))[1:3] == (
    "src/compiler/core.ts:1045:9", "a YieldExpression as a statement")
for reason in ("a YieldExpression as a statement", "Promise.all"):
    mutant = {**rows[0], "where": "synthetic.ts:1:1", "reason": reason,
              "text": "synthetic extra root"}
    assert semantic(rows + [mutant]) != roots  # both incorrect credits caught
```
