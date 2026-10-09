# Concurrency, part 2: moves

Status: **inferred moves approved by @system_adamic, Kirk briefed** (#p286ycm). Moves have no source annotation. The gospel's no borrow checker rule holds: users write neither ownership nor lifetimes, and move proofs can refuse only at a task boundary. Part 1's Shareable contract remains.

A task may take a mutable value when its variable uniquely owns the whole reachable graph. No other variable, field, closure, global or Weak may reach any mutable part of it. A count of one at the root is insufficient: a child can have another owner, and a Weak counts nothing. Each parallelMap item must also own a disjoint graph. `[value, value]` cannot give two tasks ownership of one object. Immortal scalar constants are harmless; other shared descendants need part 1's publication protocol.

## Source and proof

Infer the move at the existing call, with no annotation:

```a
import { parallelMap } from 'adamic';

function run(): void {
    const items = [{ value: 1 }, { value: 2 }];
    const results = parallelMap(items, (item) => {
        item.value += 10;
        return item;
    });
    console.log(results.map((item) => `${item.value}`).join(','));
}
run();
```

The call consumes `items`. Every later use through that binding is a compile error, including a read after the join. Results are the new owners. A closure may not retain an element merely because the parent never calls it again. Inferred last use is a necessary condition, not the ownership proof. There is no `move(items)` marker or ownership annotation. Diagnostics at this boundary include the path of a later use.

The general proof belongs on the existing IR graph. `internal/flow` provides reaching definitions, exception-aware liveness and Assign/CreateFrom/Capture/MaybeAlias edges. `internal/fresh` already interprets reachable abstract heaps for the cycle finder's relaxation: allocation recency, escaped objects, strong and weak reachability, joins and direct-call summaries. Extend that interpretation with a query at the transfer: every reachable object is confined, no other root whose ownership has not ended reaches it, no weak observer exists, and item graphs are disjoint. Unknown objects, summary fallback and merged allocation instances must refuse. Confinement alone is insufficient: two locals can hold a confined object. Mutable ranges alone do not prove uniqueness either.

Borrow and reuse plans must agree on the transfer. A borrowed variable owns no count to hand over. Reuse's dead-source proof and runtime root count check cannot establish graph exclusivity. Record the ownership decision explicitly in IR; native consumes the binding only after operand evaluation succeeds, with cleanup on exceptions. Node keeps the ordinary sequential map as the outside witness.

## Narrow prototype

The prototype proves a smaller shape directly at the source boundary, where part 1 currently refuses transfers: a function-local, immutable binding initialized by a mutable array literal of distinct flat object literals, with numeric or boolean literal fields. No spread, computed field, helper call, class, nested reference, parameter, global, reassignment, capture, loop-carried transfer or other use of the array qualifies. The only value reference to the array is the parallelMap argument. The callback is an inline arrow, writes only numeric/boolean fields of its first parameter and its own locals, and returns that same parameter. No reference-field replacement, escaping callback or effectful helper qualifies. This syntax is a sufficient freshness and alias proof, not a general implementation of flow/fresh ownership inference.

The runtime borrows the private array shell through the join, assigns each disjoint item to one task, and never touches item counts in another executor concurrently. Callback entry/exit retains remain local. Returned references wait for the join before the parent uses them. No sharing mark is needed for those item/result graphs; their counts remain plain. The private shell may retain administrative ownership until join cleanup, but source access is consumed. The move callback admits no captures; other parallelMap calls still follow part 1. Exceptions stay on part 1's shared error path; inputs and completed outputs are released after join. This prototype does not remove every retain or take array storage over in place.

## Refusals

Diagnostics include source position, the failing ownership or task-result proof path and exactly one of the two approved fixes. Unknown ownership names the path whose graph the prototype cannot prove, rather than inventing an alias. Current messages:

| Rule | Message | Fix |
| --- | --- | --- |
| use-after-move | `use after move: items.length; items was moved into parallelMap` | `don't use it after the parallelMap` |
| aliased array | `cannot move items[0]: another variable, field or closure may reach the graph` | `return it through the results` |
| unknown element, including global/closure aliases | `cannot move items[0]: element is not a fresh object literal with only scalar literal fields` | `return it through the results` |
| nested object | `cannot move items[0].child: nested reachable ownership is not proven` | `return it through the results` |
| nested array | `cannot move items[0][]: nested array ownership is not proven` | `return it through the results` |
| nested Map | `cannot move items[0].values[]: nested Map ownership is not proven` | `return it through the results` |
| unknown owner | `cannot move items: whole reachable ownership is not proven` | `return it through the results` |
| captured work value | `cannot move amount: captures are outside the move prototype` | `return it through the results` |
| unproven result | `cannot move work.result: return the owned item` | `return it through the results` |

The narrow slice may conservatively report unknown ownership for graphs requiring the general query; it must never claim a precise alias path it has not established. Every admitted shape and every refusal needs a fixture. Letting an alias through must fail Linux TSan or ASan; allowing a later source use must fail the exact refusal witness; gratuitous sharing must fail a measured shared-count assertion. A build failure or timeout proves none of these.

## Prior art read

- Rust's [ownership chapter](https://github.com/rust-lang/book/blob/main/src/ch04-01-what-is-ownership.md) makes the old binding invalid after a move. [Send and Sync](https://github.com/rust-lang/book/blob/main/src/ch16-04-extensible-concurrency-sync-and-send.md) separate transfer from sharing. Rust refuses `Rc<T>` across threads because its count can be aliased. Adamic must prove the entire graph exclusive before keeping plain counts.
- Pony's [reference capabilities](https://github.com/ponylang/pony-tutorial/blob/main/docs/reference-capabilities/reference-capabilities.md) give `iso` an isolation boundary, permitting internal references. [recover](https://github.com/ponylang/pony-tutorial/blob/main/docs/reference-capabilities/recovering-capabilities.md) restricts outside access so mutable aliases cannot leak into the recovered value. Adamic proposes inferring this boundary rather than spelling capabilities.
- Swift [SE-0414](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0414-region-based-isolation.md) merges possibly connected values into isolation regions and diagnoses access after transfer. [SE-0430](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0430-transferring-parameters-and-results.md) makes `sending` parameters and results disconnected at the boundary. Adamic's proposed rule is stricter: another surviving mutable owner refuses even when unused.
- Verona's [region tracking](https://github.com/microsoft/verona/blob/old_version/docs/internal/region-tracking.md) associates mutable aliases with their owning region and distinguishes region entry points from internal aliases. Its [FAQ](https://github.com/microsoft/verona/blob/old_version/docs/faq.md) states the aim of linear regions without per-object linearity. Internal sharing need not imply external sharing; splitting one region between tasks still requires disjointness.
- Lean's [reset/reuse pass](https://github.com/leanprover/lean4/blob/master/src/Lean/Compiler/LCNF/ResetReuse.lean) searches for dead owned values and separates borrowed from owned arguments. Koka's [reuse transformation](https://github.com/koka-lang/koka/blob/master/src/Backend/C/ParcReuse.hs) branches on `genIsUnique` before destructive reuse. These are memory reuse proofs, not permission to transfer an aliased descendant between threads.

## Remaining design

Inferred moves, the absence of annotations and lifetime syntax, task-boundary-only refusal and the two diagnostic fixes are approved. Whether dead aliases may be consumed as a group, splitting regions with internal aliases, mixing moved and Shareable descendants and general helper/result summaries remain for Kirk and Ahra to decide. Async and await remain part 3. The flat-literal experiment cannot establish how often real programs will satisfy the general proof.

Nested object graphs, arrays of arrays and Maps of objects are all explicitly refused in this slice, even when a human can see they are unique. Their fixtures name the child, array-element or Map-value path. No nested acceptance is claimed. Callback property traversal itself does not create an alias; the input graph proof remains mandatory before any scalar leaf write. The nested-flat mutant breaks only that graph proof, keeping the callback and source-use checks intact.

## Prototype evidence, October 7

Built in `internal/lower/moves.go`, with `ir.ParallelMap.Moved` set only by lowering and a small native/runtime ABI hook. The emitter clears the source binding after evaluating work, gives its count to a statement temporary and releases the private shell after join. The accepted `.a` fixture mutates and returns 2,048 independent objects; 27 exact refusal fixtures include later array reads, element aliases, globals, closures, repeated elements and unsupported proof shapes. This is construction-based exclusivity, not the general flow/fresh query above.

Source Node, emitted JavaScript and native agree: `2048 moved objects, ordered digest 661084526`. One thread and the four-thread default passed ASan/UBSan with leaks, sanitized size classes, release and separate Linux TSan. All lower/native package tests passed (18.351s and 110.906s), including part 1's runtime mutants. Final lower, part 1 source fixtures, all 27 move refusals, counts and vet passed. A filtered uncached oracle run initially failed only because the new closure fixture's expected path omitted `.value`; the corrected refusal suite passed. That initial prototype run did not include the whole-repository gate or Darwin. Commands and actual outputs are in [evidence.json](../internal/oracle/testdata/moves/evidence.json).

All three requested mutants compile in isolation and are caught: accepting identifier elements gives aliased tasks a TSan race in `adamic_retain` (exit 66); allowing a later array read fails its exact refusal (`got <nil>`, test exit 1); using shared publication for moved values changes graph-header counts from **2,050 plain / 0 shared** to **1 plain / 2,049 shared**. Ordinary retain totals are identical, so they alone cannot catch the third mutant. Reproduce with `python3 internal/oracle/testdata/moves/prove.py`. Conservative syntax restrictions were not each given an independent mutant.

Best whole-process seconds, five interleaved rounds after tests finished, clang 20.1.8 and Node 24.19.0:

| Native threads | 1 | 2 | 3 | 4 | 5 | Node, sequential |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Seconds | 0.3293 | 0.1657 | 0.1148 | 0.0888 | 0.0801 | 0.1728 |

`nproc` and affinity are 5, with a four-CPU quota (`400000 100000`). Load before/after was 1.06/2.13/1.73 and 1.05/2.11/1.72. Every timed and counted run matched Node. All thread settings counted allocations = frees = 2,056, retains 8,194, releases 6,153, peak 2,053 and regions 0. One native thread is 1.91x Node's time here; four threads take 0.51x Node's time. These are shared-machine observations of this fixture, not a general speed claim. Every sample is in [measurements.json](../internal/oracle/testdata/moves/measurements.json); reproduce with `python3 internal/oracle/testdata/moves/measure.py --threads 1,2,3,4,5 --rounds 5`.

## Approval follow-up

All 27 original refusals now check both their path-bearing message and their exact approved fix. With the nested-array, Map and shared-child witnesses there are 30 refusal fixtures. Nested objects, arrays of arrays and Maps of objects are explicitly unsupported; none is silently handled as a flat record.

The nested-flat mutant bypasses only the field graph guard. Distinct outer objects then share one mutable child; Linux TSan reports a race in that child's plain count (`adamic_retain`, exit 66). The same binary at one thread agrees with source Node: `1024 37869`. The aliased-element, later-use and shared-marking mutants are caught again. Two diagnostic mutants, substituting an unapproved fix and dropping the nested child path, each fail an exact refusal assertion (test exit 1). Reproduce all six with `prove.py`.

The accepted fixture's emitted C is byte-identical to the measured prototype (SHA-256 `44fc1105cec03d2b48be19349cf7e3bfcdf64edb927dd9daf04acdb484334c27`). The timings above remain the original observations; no new timing claim is made. Follow-up commands and outputs are in [approval-evidence.json](../internal/oracle/testdata/moves/approval-evidence.json).

The full Linux gate passed uncached: `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...`, exit 0. Lowering passed in 26.117s, native in 269.866s, the oracle in 765.961s, Unicode properties in 957.988s and every stage-1 package passed. Formatting and `go vet ./...` were clean. Full output is recorded in the follow-up evidence; the original log is `/tmp/concurrency-moves-full-gate.log`. Darwin is left to Kirk's run.
