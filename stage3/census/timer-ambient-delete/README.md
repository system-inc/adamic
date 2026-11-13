# Ambient timer deletion: required gate stopped

Task 995spev, Nov 13 2026 07:35Z; Language README addition at 07:38Z.
Base: `f37253b347c4f7e7bc223ed05fe04edb8cef0cd4`.
Adaptation branch: `typescript/timer-ambient-delete`,
`bd54305c48a1a4c3285083276d69ee8d6ee7a5b7`.
That single commit adds only the requested “Premises a reader can re-check”
section. The ambient deletion is **not implemented**: stock tsc refuses the
unchanged cancellation forwarder. No cast, new widening, or adapter change.

Read first: the base timer-closures README and PROOF.md, the existing evidence
branch `37f27baefd4166ae9509169f23bb59ed8de6e3f9`, and the older deletion
`91af96c4157e4437d5c327d15fa20244d3ed3f26` with documentation follow-up
`5ba70a9486a9d297f42a6a29fd9b2388caaf093d`. No older types are reused.

## Focused executable gate

`forwarders.ts` contains exactly 93's two forwarding expressions, in an
external module that requests Node globals through a type-only Node import,
plus two global-value aliases for inspection. There are no local timer heads.
`stock-forwarders-final` is the definitive focused stock receipt; the earlier
focused receipt predates the type-only import and global inspection aliases.
`stock-probe.cjs` runs pinned stock TypeScript 6.0.3 against the unchanged
published @types/node 25.3.3 pin. `adamic-forwarders` uses the actual Adamic
executable built from the base code (the sole commit is documentation).

| Checker | Exit | Diagnostics |
|---|---:|---|
| Stock TypeScript 6.0.3 | 1 | TS2345 at forwarders.ts:4:62 |
| Adamic types | 0 | None |

Exact stock diagnostic:

```
Argument of type 'unknown' is not assignable to parameter of type 'string | number | Timeout | undefined'.
```

The setTimeout forwarder is accepted by both. Both declare its global as:

```ts
function setTimeout<TArgs extends any[]>(
    callback: (...args: TArgs) => void,
    delay?: number,
    ...args: MakeVoidParameterOptional<TArgs>
): NodeJS.Timeout;
```

`MakeVoidParameterOptional<TArgs extends any[]>` is
`[void] extends TArgs ? Partial<TArgs> : TArgs`.
Stock declares the cancellation global as:

```ts
function clearTimeout(timeout: NodeJS.Timeout | string | number | undefined): void;
```

Adamic declares it as:

```ts
function clearTimeout(timeout: unknown): void;
```

This difference already exists at the base: `internal/load/host_axioms.json`
has the line-41 correction ruled on #995spev, Nov 7. `node_library_fs.go`
serves it over the unchanged bundled publication. The prelude adds no timer
declaration. `published-timers.d.ts.txt`, `adamic-served-timers.d.ts.txt`, and
`timer-host-axioms.json` preserve the published text and the existing served
view. The stock probe records compiler-resolved signatures and declarations.

## Scope and receipts

Each status JSON reports the actual child exit; run-command.py propagates it
and bounds child commands to 1700 seconds. The first base apply setup failure
(missing ADAMIC_TYPESCRIPT_SOURCE) is retained separately from the corrected
run. Evidence stays outside stage3/adapt.

The failed stock gate requires an Adamic ruling. Therefore there is no deletion
adapter, deletion replay, 398-shard run, restoring-ambient mutant, ledger
amendment, full-mode census pair, or binder first-stop claim.

## Actual adapted-tree counterfactual

The full base apply succeeded (exit 0, 396.909 seconds). After it finished,
only the two ambient declaration lines were removed from its sys.ts, using
the current 41 rows' concrete `after` text. No older types were reused.
`sys.before.ts.txt`, `sys.after.ts.txt`, `counterfactual.diff`, and
`counterfactual.json` preserve the exact change and SHA-256 pair. This is a
measurement counterfactual, not a deletion adaptation commit.

| Check on all 79 sorted compiler roots | Exit | Result |
|---|---:|---|
| Actual adamic types | 0 | No diagnostics |
| Stock 6.0.3 under the established 79-root proof options | 1 | Exactly two TS2345 diagnostics |

Exact stock diagnostics, in the declaration-deleted tree:

```
src/compiler/sys.ts:1520:70 TS2345: Argument of type 'unknown' is not assignable to parameter of type 'string | number | Timeout | undefined'.
src/compiler/sys.ts:1618:70 TS2345: Argument of type 'unknown' is not assignable to parameter of type 'string | number | Timeout | undefined'.
```

Both point to the argument in `clearTimeout(timeoutId)` inside 93's unchanged
forwarders. No setTimeout diagnostic or other diagnostic occurs. The JSON
receipt records both compiler-resolved global signatures and original
declaration text. The stock roots harness retains the prior evidence harness's
prelude and RegExp library compatibility view; it applies no Node host-axiom
correction. The focused stock check also fails with the untouched stock libs.

The requested conditional deletion proof suite was not started. The zero
diagnostics result applies to Adamic only; stock does not pass this gate.
