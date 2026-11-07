# Core root first stop after b4b8c4d7

Measured origin/main tip e8f3723bd41abcebedabfcc7546fdf104a9bbf08. b4b8c4d7 is an ancestor (git merge-base --is-ancestor exit 0). TypeScript 6.0.3 source: 050880ce59e30b356b686bd3144efe24f875ebc8.

One full stage3/apply.sh --build-product tree, all 95 adaptations completed, and the ordinary stage3/census/tool built from that same main SHA. No overlays or partial-root fallback. The reference is stage3/census/next-stops-af42d3dc/REPORT.md at ba90529b3d16768a886a7570c47a0d1c07b06a32 on typescript/next-stops-after-wave, specifically run-groups.py's single-file invocations. This measures only src/compiler/core.ts and src/compiler/binder.ts.

| Root | First stop | Kind | Exact reason | Lowering decision |
| --- | --- | --- | --- | --- |
| core.ts | debug.ts:267:72 | Refused | a value of type (node: Node) => boolean seen as AnyFunction, which can write void where boolean is read | internal/lower/invariance.go:1039, (*lowering).wideningRefusal (starts 1023, emits 1053); refuseWidening at 459 calls it at 610 |
| binder.ts | core.ts:2105:26 | Refused | a method read as a value (compare would lose its object, and this with it) | internal/lower/refusals.go:282, (*lowering).methodValueRefusal (starts 260) |

Core does not match Optimizer's prediction. It stops in assertEachNode, at test! in getFunctionName(test!), rather than the checked predicate overload target at assertEachIsDefined. debug.ts:243:59 appears zero times anywhere in core.jsonl; it is also absent from core.log. The ordinary stream contains two observations, each independently lowering the same one-root program. Their non-timing fields agree for each root.

The whole enclosing core-stop function and its overloads (15 lines) are retained in debug-stop.ts.txt:

```ts
    export function assertEachNode<T extends Node, U extends T>(nodes: NodeArray<T>, test: (node: T) => node is U, message?: string, stackCrawlMark?: AnyFunction): asserts nodes is NodeArray<U>;
    export function assertEachNode<T extends Node, U extends T>(nodes: readonly T[], test: (node: T) => node is U, message?: string, stackCrawlMark?: AnyFunction): asserts nodes is readonly U[];
    export function assertEachNode<T extends Node, U extends T>(nodes: NodeArray<T> | undefined, test: (node: T) => node is U, message?: string, stackCrawlMark?: AnyFunction): asserts nodes is NodeArray<U> | undefined;
    export function assertEachNode<T extends Node, U extends T>(nodes: readonly T[] | undefined, test: (node: T) => node is U, message?: string, stackCrawlMark?: AnyFunction): asserts nodes is readonly U[] | undefined;
    export function assertEachNode(nodes: readonly Node[], test: ((node: Node) => boolean) | undefined, message?: string, stackCrawlMark?: AnyFunction): void;
    export function assertEachNode(nodes: readonly Node[] | undefined, test: ((node: Node) => boolean) | undefined, message?: string, stackCrawlMark?: AnyFunction) {
        if (shouldAssertFunction(AssertionLevel.Normal, "assertEachNode")) {
            assert(
                test === undefined || every(nodes, test),
                message || "Unexpected node.",
                () => `Node array did not pass test '${getFunctionName(test!)}'.`,
                stackCrawlMark || assertEachNode,
            );
        }
    }
```

Setup, npm ci, apply, census build retry and both root invocations exited 0. An earlier census build exited 1 while setup was still populating cohere/TypeScript/tsc/go.mod; it was retried after the submodule became ready. Census process groups were killed after each child was reaped. No task child remains running.

Reproduction, from the recorded main SHA:

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > setup.log 2>&1
npm ci --prefix stage3/api > npm-ci.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
bash stage3/apply.sh --build-product /workspace/core-root-e8f3723b-adapted > apply.log 2>&1
go build -o /tmp/core-root-avqasqt/census ./stage3/census/tool > census-build.log 2>&1
/tmp/core-root-avqasqt/census /workspace/core-root-e8f3723b-adapted/src/compiler/core.ts core.jsonl > core.log 2>&1
/tmp/core-root-avqasqt/census /workspace/core-root-e8f3723b-adapted/src/compiler/binder.ts binder.jsonl > binder.log 2>&1
```

The retained supervise.py and run-roots.py wrap these commands with child exit reporting and process group cleanup. Logs were written to files throughout. No scratch Go source was created.
