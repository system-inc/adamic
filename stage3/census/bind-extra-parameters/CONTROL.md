# Positive-control failure and omitted whole-program check

The exact head fixture `internal/oracle/testdata/bind_extra_parameters/narrow_type.a` was copied unchanged to scratch `control/main.ts`: a two-parameter `seen(shape, index?)` held as `(shape: Shape) => boolean`, bound as that narrower type, then called by `find` with `(value, i)`.

Both unchanged latent binaries ran it with CommonJSBundle and `LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_ONLY_FILE=<absolute control/main.ts>`. Both checker headers have zero diagnostics. Both exited 0 and emitted **zero findings**, including **zero occurrences of the new bind reason**. Thus the required positive control fails at head. The base negative control passes, but this does not rescue the head observation. [CONTROL.json](CONTROL.json) contains commands and summary receipts; raw streams stay in scratch.

At both requested commits, `stage3/census/latent/pass_ledger.go.txt:28` lists `{Name: "checkBoundArguments"}` with no `Scoped: true`. `scoped_tail.go.txt:19` disables the tail unless `LATENT_TAIL=1`; even then its loop at line 73 skips every non-scoped pass, and its dispatcher has no checkBoundArguments case. The full-unit and legacy first-error drivers do not call the production finalization tail. The generated lower.go replaces production's public lowering body with the no-output guard. Therefore this check does not run at all in these census measurements.

Production `internal/lower/lower.go:218` calls `checkBoundArguments` after lowering and `findCycles`. At head, `function_bind.go:450` iterates bound closure records, asks `boundValueReach` when a typed target declares excess parameters, uses only reached functions for a known reach, and falls back to all type-compatible targets for an unknown reach. `boundValueReach` at line 405 assigns allocation sites and constructs the ordinary flow graph from completed IR.

The projection index in `shape_flow.go:250` walks **all** main IR and function bodies; `CallClosure`, `ObjectCall`, `GeneratorNext`, `IteratorNext` and `Yield` set its `opaqueCalls` at line 272. `projectedSources` at line 325 adds "opaque call may mutate projected fields" when this flag is set, preventing a known projection certificate and causing the bind check's type-wide fallback. The overlay leaves this helper's source unchanged, but does not invoke the bind check, and its isolated unit IR cannot stand in for those whole-program facts. There is no whole-program `opaqueCalls` computation in the completed sys latent measurement that validates this new stop.

Consequently every latent zero in this task is unproven for the new reason. Running production LowerWithOptions whole for each source root is the real measure of its *first reachable stop*, subject to earlier production stops. No compiler or census behavior is being changed to make the control pass.

Core's first base process had just started when the control request arrived; it was terminated before completion and before its head pair. Its partial raw stream is retained under scratch `aborted-before-control/full-core.ts`, excluded from all counts. Core will restart after controls.

Control source:

```ts
// seen declares an index, held as (shape) => boolean and bound: Node hands it find's index (#04f5s0s).
interface Circle {
    readonly kind: "circle";
    readonly radius: number;
}
interface Square {
    readonly kind: "square";
    readonly side: number;
}
type Shape = Circle | Square;
// core.ts's find, which calls its predicate with (value, i).
function find<T, U extends T>(array: readonly T[] | undefined, predicate: (element: T, index: number) => element is U, startIndex?: number): U | undefined;
function find<T>(array: readonly T[] | undefined, predicate: (element: T, index: number) => boolean, startIndex?: number): T | undefined;
function find<T>(array: readonly T[] | undefined, predicate: (element: T, index: number) => boolean, startIndex?: number): T | undefined {
    if (array === undefined) return undefined;
    for (let i = startIndex ?? 0; i < array.length; i++) {
        const value = array[i]!;
        if (predicate(value, i)) {
            return value;
        }
    }
    return undefined;
}
const shapes: Shape[] = [{ kind: "square", side: 2 }, { kind: "circle", radius: 3 }];
function seen(shape: Shape, index?: number): boolean {
    console.log(`seen at ${String(index)}`);
    return shape.kind === "circle";
}
const narrow: (shape: Shape) => boolean = seen;
const bound = narrow.bind(undefined);
const found = find(shapes, bound);
console.log(found === undefined ? "none" : found.kind);

```
