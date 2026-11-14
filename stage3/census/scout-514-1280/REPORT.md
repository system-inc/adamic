# Scout: sys.ts:514 and utilitiesPublic.ts:1280

Status: stock diagnostics and isolated emitted-helper checkpoint; full Adamic census pending. Compiler and adaptations unchanged.

Measured-head target: `4f460a611b61a5b11c12ce61fde8ad5d0f11a912` (fresh detached worktree, then report branch). Prior evidence: `716d205b`, not counted as a new measurement.

## Recommended roads (provisional)

- Wall A: provisional road A, two concretely typed helpers. Stock 6.0.3: baseline 0 diagnostics / 79 roots; A 0 / 79, C 0 / 79; B adds 1 TS2345. A and C match baseline isolated emitted-helper traces for both callback types. Full Adamic sys.ts census is still pending, so neither is recommended for implementation yet.
- Wall B: investigate the language/compiler proof transport boundary; do not rewrite getFirstJSDocTag merely to bypass it. New measured numbers: none. At this head, predicateArgumentProof explicitly handles bodies and producer calls but not a parameter declaration.

## Setup

The shared checkout was not edited. The requested commit was initially absent (`fatal: invalid reference: 4f460a611b61a5b11c12ce61fde8ad5d0f11a912`); fetching the two named branches succeeded. Scratch worktree: `/workspace/scratch/scout-514-1280`. `npm ci` in its stage3/api succeeded (3 packages). Upstream shallow clone confirmed `v6.0.3`, `050880ce59e30b356b686bd3144efe24f875ebc8`; package.json version 6.0.3. ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-603-scout. Apply completed successfully with this head. Recursive submodule initialization is still downloading the nested compiler; no credential error observed.

## What this checkpoint cannot see

No baseline or candidate Adamic census has run yet because recursive initialization is still pending. No claim that 514:17 clears is supported. Runtime traces are isolated helper experiments, not the compiler test suite or external host implementations. Stock diagnostics use the same strict gate options plus the pinned @types/node 25.3.3 (needed by stock tsc instead of Adamic's Node library shim); the initial no-Node-types setup was corrected and is not the reported measurement. All 79 compiler roots are loaded together; this is not 79 independent ordinary first-stop compilations. sys.js is emitted by stock 6.0.3 transpileModule at ES2024/ESNext; see the exact script, source patches, emitter diffs and traces.

## Candidate checkpoint

A duplicates the helper specialized for FileWatcherCallback and FsWatchCallback, selects the appropriate helper at each call, and leaves each forwarder as a three-parameter arrow. C accepts a caller-supplied factory `(getCallbacks: () => T[] | undefined) => T`; each caller supplies a concrete three-parameter arrow. Neither uses a cast. B uses the union common call domain `(param1: "rename" | "change", param2: never, param3?: Date)`; it cannot be related to arbitrary T. B is rejected before it can be a zero-new-diagnostic candidate.

The isolated emitted-helper experiment preserves cache lookup, snapshot dispatch, insertion before/after synchronous invocation, callback addition/removal during dispatch, argument count (three even for omitted timestamp), argument values/order, forwarder length (3), per-watcher identity and closing. Both A and C match the baseline traces for both callback types. Their sys.js is not byte-identical; value identity is supported for these helper observations, not arbitrary reflective access to source text or external injected hosts. B's erased emit has the same forwarder body but is ill-typed.

## Wall B: static evidence at the exact head

`internal/lower/predicates.go:760` resolves an identifier to a function declaration or variable initializer, then requires a function body (or handles a proven producer call). A parameter declaration supplies neither. Its explicit annotation branch also requires exact target identity and a proven body or a checked declaration/closure. The forwarding argument `predicate` in getFirstJSDocTag therefore fails the body requirement.

`find` is core.ts's overload at lines 162–166: its narrowing overload takes a predicate, but its implementation takes a plain boolean callback. `predicateParameter` maps that overload to its boolean implementation, so checkedFunctionTypeClaim does not make the overload parameter checked. It consequently retains argument proof. This is not simply a lib.d.ts find call. The getFirstJSDocTag parameter itself is a checked bodiless function-type claim: predicateCallbackCallers cannot vouch it because its use as find's argument escapes the direct-call-only rule. A check at calls directly through that signature does not transport a check into find's boolean callback invocation.

All 21 getFirstJSDocTag callers are in utilitiesPublic.ts, passing 15 named isJSDoc*Tag functions. The all-src search includes harness and testRunner and finds no other calls. These resolve to explicit type guards in factory/nodeTests.ts, each returning the corresponding direct `node.kind === SyntaxKind.JSDoc…Tag` comparison. They are subject to Adamic's declaration proof or checked-value mechanism; passing a guard by name is not an unchecked annotation escape. The 21 call occurrences and locations are in callers.json.

Owner: Language/Lower proof transport, rather than a required TypeScript adaptation. docs/rulings.md:134–149 preserves the proven-body rule, .ts checked predicates (#tzd3gjg), checked predicate values (#vdt9f1q), and checked bodiless signatures (#1w5bve2). A bodiless signature carries no proof of its own; even a proven function called through an unproven signature is checked. Any forwarding solution must preserve the proof or checked closure, not merely trust the parameter annotation or collect today's callers. Source comments in predicates_function_type.go explicitly keep proof for a forwarded checked parameter at an unchecked callback consumer. The precise general forwarding rule needs Language/Lower ownership. No such rule is implemented or proposed here. The tasks themselves are not available through a task connector; the repository's attributed rulings and implementation are the evidence available here.
