# Concurrency, part 2: moves

Status: **proposal and prototype plan, not approved** (#p286ycm). The design is Kirk and Ahra's to approve. Part 1's Shareable contract remains.

A task may take a mutable value when its variable uniquely owns the whole reachable graph. No other variable, field, closure, global or Weak may reach any mutable part of it. A count of one at the root is insufficient: a child can have another owner, and a Weak counts nothing. Each parallelMap item must also own a disjoint graph. `[value, value]` cannot give two tasks ownership of one object. Immortal scalar constants are harmless; other shared descendants need part 1's publication protocol.

## Source and proof

Recommend inferring the move at the existing call, with no annotation:

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

The call consumes `items`. Every later use through that binding is a compile error, including a read after the join. Results are the new owners. A closure may not retain an element merely because the parent never calls it again. Inferred last use is a necessary condition, not the ownership proof. A `move(items)` marker would express intent but could not replace that proof; do not add it for this prototype.

The general proof belongs on the existing IR graph. `internal/flow` provides reaching definitions, exception-aware liveness and Assign/CreateFrom/Capture/MaybeAlias edges. `internal/fresh` already interprets reachable abstract heaps for the cycle finder's relaxation: allocation recency, escaped objects, strong and weak reachability, joins and direct-call summaries. Extend that interpretation with a query at the transfer: every reachable object is confined, no other live root reaches it, no weak observer exists, and item graphs are disjoint. Unknown objects, summary fallback and merged allocation instances must refuse. Confinement alone is insufficient: two locals can hold a confined object. Mutable ranges alone do not prove uniqueness either.

Borrow and reuse plans must agree on the transfer. A borrowed variable owns no count to hand over. Reuse's dead-source proof and runtime root count check cannot establish graph exclusivity. Record the ownership decision explicitly in IR; native consumes the binding only after operand evaluation succeeds, with cleanup on exceptions. Node keeps the ordinary sequential map as the outside witness.

## Narrow prototype

Prove a smaller shape directly at the source boundary, where part 1 currently refuses transfers: a function-local, immutable binding initialized by an array literal of distinct flat object literals, with numeric or boolean literal fields. No spread, computed field, helper call, class, nested reference, parameter, global, reassignment, capture, loop-carried transfer or other use of the array qualifies. The only value reference to the array is the parallelMap argument. The callback is an inline arrow, writes only numeric/boolean fields of its first parameter and its own locals, and returns that same parameter. No reference-field replacement, escaping callback or effectful helper qualifies. This syntax is a sufficient freshness and alias proof, not a general implementation of flow/fresh ownership inference.

The runtime borrows the private array shell through the join, assigns each disjoint item to one task, and never touches item counts in another executor concurrently. Callback entry/exit retains remain local. Returned references wait for the join before the parent uses them. No sharing mark is needed for those item/result graphs; their counts remain plain. The private shell may retain administrative ownership until join cleanup, but source access is consumed. Closure and immutable global captures still follow part 1. Exceptions stay on part 1's shared error path; inputs and completed outputs are released after join. This prototype does not remove every retain or take array storage over in place.

## Refusals

Diagnostics include source position, the reachable path and a fix. Proposed names and messages:

| Rule | Message | Fix |
| --- | --- | --- |
| use-after-move | `use after move: items.length; items was moved into parallelMap` | use the returned results |
| aliased graph | `cannot move items[0]: another variable or field may reach this element` | construct independent elements under the array's sole owner |
| global reach | `cannot move items[0]: reachable from global saved` | remove the global owner before transfer |
| closure reach | `cannot move items[0]: reachable from closure observer` | end the capture's ownership before transfer |
| overlapping items | `cannot move items[1]: also reachable from items[0]` | give each task an independent graph |
| unknown ownership | `cannot move items: whole reachable ownership is not proven` | use the supported fresh construction or make it Shareable |
| borrowed owner | `cannot move items: the binding borrows its value` | transfer from its owning binding |
| unproven work/result | `cannot move task result: ownership or effects are not proven` | return the owned item or a proven fresh result |

The narrow slice may conservatively report unknown ownership for graphs requiring the general query; it must never claim a precise alias path it has not established. Every admitted shape and every refusal needs a fixture. Letting an alias through must fail Linux TSan or ASan; allowing a later source use must fail the exact refusal witness; gratuitous sharing must fail a measured shared-count assertion. A build failure or timeout proves none of these.

## Prior art read

- Rust's [ownership chapter](https://github.com/rust-lang/book/blob/main/src/ch04-01-what-is-ownership.md) makes the old binding invalid after a move. [Send and Sync](https://github.com/rust-lang/book/blob/main/src/ch16-04-extensible-concurrency-sync-and-send.md) separate transfer from sharing. Rust refuses `Rc<T>` across threads because its count can be aliased. Adamic must prove the entire graph exclusive before keeping plain counts.
- Pony's [reference capabilities](https://github.com/ponylang/pony-tutorial/blob/main/docs/reference-capabilities/reference-capabilities.md) give `iso` an isolation boundary, permitting internal references. [recover](https://github.com/ponylang/pony-tutorial/blob/main/docs/reference-capabilities/recovering-capabilities.md) restricts outside access so mutable aliases cannot leak into the recovered value. Adamic proposes inferring this boundary rather than spelling capabilities.
- Swift [SE-0414](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0414-region-based-isolation.md) merges possibly connected values into isolation regions and diagnoses access after transfer. [SE-0430](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0430-transferring-parameters-and-results.md) makes `sending` parameters and results disconnected at the boundary. Adamic's proposed rule is stricter: another surviving mutable owner refuses even when unused.
- Verona's [region tracking](https://github.com/microsoft/verona/blob/old_version/docs/internal/region-tracking.md) associates mutable aliases with their owning region and distinguishes region entry points from internal aliases. Its [FAQ](https://github.com/microsoft/verona/blob/old_version/docs/faq.md) states the aim of linear regions without per-object linearity. Internal sharing need not imply external sharing; splitting one region between tasks still requires disjointness.
- Lean's [reset/reuse pass](https://github.com/leanprover/lean4/blob/master/src/Lean/Compiler/LCNF/ResetReuse.lean) searches for dead owned values and separates borrowed from owned arguments. Koka's [reuse transformation](https://github.com/koka-lang/koka/blob/master/src/Backend/C/ParcReuse.hs) branches on `genIsUnique` before destructive reuse. These are memory reuse proofs, not permission to transfer an aliased descendant between threads.

## Left for Kirk and Ahra

Inferred moves, permanent invalidation after a structured join, whether dead aliases may be consumed as a group, splitting regions with internal aliases, mixing moved and Shareable descendants, general helper/result summaries and ownership diagnostics are recommendations here. No syntax or language decision is approved by this unit. Async and await remain part 3. The flat-literal experiment cannot establish how often real programs will satisfy the general proof.
