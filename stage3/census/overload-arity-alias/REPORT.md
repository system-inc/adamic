# Overload arity at value uses (#5fws6q8, #ev14ch9)

**No live miscompile found at `49f1d643064a9231f7f343d8bbd02ba718866e6a`.** All attempted unsafe entry roads are refused. Of 24 programs, 20 refuse before producing a native binary; four controls compile natively, exit 0, and match Node byte for byte on both stdout and stderr. Refusal is not counted as runtime agreement.

The body proof is declaration-based, but an accepted overload set is **not trusted as a function value**: Adamic explicitly refuses those reads. That refusal is essential to the safety argument; it is not a proof that the wider callable assignment is sound.

## The two questions

1. **The proof premise is per overload declaration, not the actual arguments at each call.** `censusOverloads` visits each bodyless declaration ([`internal/lower/census_small.go:115`](https://github.com/system-inc/adamic/blob/49f1d643064a9231f7f343d8bbd02ba718866e6a/internal/lower/census_small.go#L115)). Its parameter comparison treats an omitted implementation parameter as `undefined` ([`internal/lower/census_small.go:202`](https://github.com/system-inc/adamic/blob/49f1d643064a9231f7f343d8bbd02ba718866e6a/internal/lower/census_small.go#L202)). `censusProveOverloadResult` calls `censusReturnProof` with no actual-argument admission list ([`internal/lower/census_overload_proof.go:42`](https://github.com/system-inc/adamic/blob/49f1d643064a9231f7f343d8bbd02ba718866e6a/internal/lower/census_overload_proof.go#L42)); that proof seeds each implementation parameter from the overload's parameter list and seeds missing entries as undefined (lines 64–83). Reachability then prunes branches using those seeded types (lines 110–125 and 175–204). `overloadResultCheck` runs at individual calls, but omits their result check when that same declaration-level proof succeeds ([`internal/lower/overload_result_check.go:146`](https://github.com/system-inc/adamic/blob/49f1d643064a9231f7f343d8bbd02ba718866e6a/internal/lower/overload_result_check.go#L146)). The proof can be recomputed; it is not a separate actual-arity proof. `narrowOverload` itself is explicitly memoized by overload AST declaration ([`internal/lower/overload_result_check.go:28`](https://github.com/system-inc/adamic/blob/49f1d643064a9231f7f343d8bbd02ba718866e6a/internal/lower/overload_result_check.go#L28) and [`internal/lower/lower.go:426`](https://github.com/system-inc/adamic/blob/49f1d643064a9231f7f343d8bbd02ba718866e6a/internal/lower/lower.go#L426)).

2. **Yes, overload sets used as values are already refused.** [`internal/lower/expression.go:2135`](https://github.com/system-inc/adamic/blob/49f1d643064a9231f7f343d8bbd02ba718866e6a/internal/lower/expression.go#L2135) defines `overloadSetValueStop`: any bodyless function declaration with an implementation causes “an overloaded function as a value”. Module-level `functionValue` calls it before making/reusing its adapter (line 2159). It also resolves shorthand assignment value symbols (lines 2152–2157). Nested functions use the same guard in [`internal/lower/nested_functions.go:239`](https://github.com/system-inc/adamic/blob/49f1d643064a9231f7f343d8bbd02ba718866e6a/internal/lower/nested_functions.go#L239) except at the direct callee node (`node == l.nestedCallee`). These are unconditional overload-set guards, not restricted to `narrowOverload` or to unproven bodies. Named/imported symbol aliases are resolved by [`internal/lower/locals.go:195`](https://github.com/system-inc/adamic/blob/49f1d643064a9231f7f343d8bbd02ba718866e6a/internal/lower/locals.go#L195). Namespace value paths also reach `functionValue` (`namespaces.go:119`, `namespace_value.go:84`); object shorthand does at `object.go:2042`. The tested alias/callback/object/array/nested/shorthand roads all hit this guard.

The source already explains the reason: a call through a value resolves to the slot's signature, not the overload declaration, so an overload result check would not run (`expression.go:2131–2134`). Existing regression coverage makes the same point in [`internal/lower/overload_result_check_test.go:261`](https://github.com/system-inc/adamic/blob/49f1d643064a9231f7f343d8bbd02ba718866e6a/internal/lower/overload_result_check_test.go#L261).

## Programs and observed results

Every row links its complete standalone source and exact build stderr. All builds used the unchanged baseline compiler's native `build` command. `—` means no native binary was produced. Node executes the original `.ts` with Node 24 type stripping, not Adamic-emitted JS. Numeric observations are `String(r + 1)` and `r.toFixed(0)`; `String` is needed because Adamic's console prelude accepts strings. The array holder is guarded before invocation; the body-only array result uses `[0] ?? 0` to satisfy checked indexing.

For Node failures, `TypeError` below means exactly `TypeError: r.toFixed is not a function`, after the shown stdout. Raw stack traces are preserved separately. Build diagnostics appear verbatim below the table and in each linked file.

| Program / road | Adamic native build result | Node stdout (escaped), exit | Native result |
| --- | --- | --- | --- |
| [01-alias](programs/01-alias.ts): typed wider alias | [refuses: overload value](outputs/01-alias.build.stderr) | `"bad1\n"`, 1 (TypeError) | — |
| [02-callback](programs/02-callback.ts): callback argument | [refuses: overload value](outputs/02-callback.build.stderr) | `"bad1\n"`, 1 (TypeError) | — |
| [03-object](programs/03-object.ts): object field | [refuses: overload value](outputs/03-object.build.stderr) | `"bad1\n"`, 1 (TypeError) | — |
| [04-array](programs/04-array.ts): array element | [refuses: overload value](outputs/04-array.build.stderr) | `"bad1\n"`, 1 (TypeError) | — |
| [05-call](programs/05-call.ts): .call, number slot | [refuses: TS2322](outputs/05-call.build.stderr) | `"bad1\n"`, 1 (TypeError) | — |
| [06-apply](programs/06-apply.ts): .apply, number slot | [refuses: TS2322](outputs/06-apply.build.stderr) | `"bad1\n"`, 1 (TypeError) | — |
| [07-bind](programs/07-bind.ts): .bind then f, number slot | [refuses: TS2322](outputs/07-bind.build.stderr) | `"bad1\n"`, 1 (TypeError) | — |
| [08-spread-alias](programs/08-spread-alias.ts): tuple spread through wider alias | [refuses: overload value](outputs/08-spread-alias.build.stderr) | `"bad1\n"`, 1 (TypeError) | — |
| [09-spread-direct-number](programs/09-spread-direct-number.ts): direct tuple spread, number slot | [refuses: TS2322](outputs/09-spread-direct-number.build.stderr) | `"bad1\n"`, 1 (TypeError) | — |
| [10-direct-control](programs/10-direct-control.ts): original direct numeric control | [refuses: representation](outputs/10-direct-control.build.stderr) | `"2\n1\n"`, 0 | — |
| [11-no-if-alias](programs/11-no-if-alias.ts): no-if optional implementation, alias | [refuses: overload value](outputs/11-no-if-alias.build.stderr) | `"2\n1\n"`, 0 | — |
| [12-no-if-direct](programs/12-no-if-direct.ts): no-if optional implementation, direct | [refuses: representation](outputs/12-no-if-direct.build.stderr) | `"2\n1\n"`, 0 | — |
| [13-body-only-alias](programs/13-body-only-alias.ts): body-proof-only, wider alias | [refuses: overload value](outputs/13-body-only-alias.build.stderr) | `"b1\n"`, 1 (TypeError) | — |
| [14-body-only-direct](programs/14-body-only-direct.ts): body-proof-only, direct numeric read | [refuses: representation](outputs/14-body-only-direct.build.stderr) | `"2\n1\n"`, 0 | — |
| [15-spread-direct-honest](programs/15-spread-direct-honest.ts): direct tuple spread, honest string result | [refuses: tuple spread](outputs/15-spread-direct-honest.build.stderr) | `"bad1\n3\n"`, 0 | — |
| [16-call-inferred](programs/16-call-inferred.ts): .call, inferred string | [refuses: prototype member](outputs/16-call-inferred.build.stderr) | `"bad1\n3\n"`, 0 | — |
| [17-apply-inferred](programs/17-apply-inferred.ts): .apply, inferred string | [refuses: prototype member](outputs/17-apply-inferred.build.stderr) | `"bad1\n3\n"`, 0 | — |
| [18-bind-inferred](programs/18-bind-inferred.ts): .bind, inferred string | [refuses: bind signature](outputs/18-bind-inferred.build.stderr) | `"bad1\n3\n"`, 0 | — |
| [19-nested-alias](programs/19-nested-alias.ts): nested wider alias | [refuses: overload value](outputs/19-nested-alias.build.stderr) | `"bad1\n"`, 1 (TypeError) | — |
| [20-shorthand](programs/20-shorthand.ts): shorthand object then wider alias | [refuses: overload value](outputs/20-shorthand.build.stderr) | `"bad1\n"`, 1 (TypeError) | — |
| [21-original-declaration](programs/21-original-declaration.ts): original set, declaration only | [lowers and agrees with Node](outputs/21-original-declaration.build.stderr) | `"original set accepted\n"`, 0 | stdout and stderr byte-equal; exit 0 |
| [22-body-only-declaration](programs/22-body-only-declaration.ts): body-proof-only set, declaration only | [lowers and agrees with Node](outputs/22-body-only-declaration.build.stderr) | `"body-only set accepted\n"`, 0 | stdout and stderr byte-equal; exit 0 |
| [23-no-if-native](programs/23-no-if-native.ts): no-if optional implementation, native control | [lowers and agrees with Node](outputs/23-no-if-native.build.stderr) | `"2\n1\n"`, 0 | stdout and stderr byte-equal; exit 0 |
| [24-same-result-native](programs/24-same-result-native.ts): same-result overloads, native control | [lowers and agrees with Node](outputs/24-same-result-native.build.stderr) | `"2\n1\n"`, 0 | stdout and stderr byte-equal; exit 0 |

There is no runnable string-in-number-slot witness in this matrix: each candidate stops before code generation. In particular, direct `.call`/`.apply`/`.bind` with `r: number` and direct tuple spread do not select the one-argument overload: the checker sees a string result and reports TS2322. The inferred-result variants remove that annotation to expose the lowerer's independent stops. Direct tuple spread honestly selects the two-argument string overload but is unsupported by this lowerer.

The original direct one-argument control is also refused after census acceptance: fitting the implementation's number|string representation to the resolved number result is unsupported. This is an additional barrier, not an output mismatch.

### Exact refusal words

The full diagnostics include their source paths and positions in the linked stderr files. The complete distinct messages are:

```text
stage 0 can't lower an overloaded function as a value yet
error TS2322: Type 'string' is not assignable to type 'number'.
stage 0 can't lower an overload result requiring another representation yet
stage 0 can't lower a call spreading something other than an array yet
Adamic 0.1 refuses inherited library member call read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method)
Adamic 0.1 refuses inherited library member apply read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method)
stage 0 can't lower bind without one declared function signature yet
```

`.call` and `.apply` reach the inherited-library-member refusal defined in [`internal/lower/prototype.go:120`](https://github.com/system-inc/adamic/blob/49f1d643064a9231f7f343d8bbd02ba718866e6a/internal/lower/prototype.go#L120). `.bind` reads the target's call signatures and refuses unless there is exactly one ([`internal/lower/function_bind.go:37`](https://github.com/system-inc/adamic/blob/49f1d643064a9231f7f343d8bbd02ba718866e6a/internal/lower/function_bind.go#L37)). Tuple spread fails at [`internal/lower/arguments_length.go:115`](https://github.com/system-inc/adamic/blob/49f1d643064a9231f7f343d8bbd02ba718866e6a/internal/lower/arguments_length.go#L115).

## Which road accepts the set?

The original first overload passes the parameter comparison: missing `f` is undefined, which the optional implementation accepts. Its result number|string does not relate to number. No phantom brand, presence-only conversion, predicate callback, or void promise applies. **Body proof accepts overload 1**: under `f = undefined`, the `return f(a)` branch is unreachable, leaving the number parameter return. Its number promise is also checkable by `narrowOverload`, but that fallback is later in census order.

The original second overload instead reaches **`narrowOverload`'s call-check fallback**: its reachable `return f(a)` has a call effect, so body proof fails (`predicates.go:358–365`, `census_overload_proof.go:146`). Its promised string is a checkable member of number|string. The two-overload no-if variant is accepted similarly: first overload by body proof, second by narrow fallback, since `return a` cannot prove string. Probes 21 and 22 demonstrate declaration acceptance separately from value or result-representation refusals.

The body-only variant (13, 14, 22) is:

```ts
function pick(a: number): number[];
function pick(a: number, f?: (x: number) => string): number[] | string {
    if (f) return "bad";
    return [a];
}
```

Its first and only exposed overload requires **body proof as the sole accepting road**. Produced number[]|string is not covariant with number[], has no phantom difference, no undefined presence conversion, no predicate callback and no void promise. `narrowOverload` cannot accept number[]: arrays are explicitly uncheckable (`overload_result_check.go:105`). The proof seeds omitted `f` as undefined and proves the reachable array-literal return. The wider alias is refused (13); a direct numeric read hits representation fitting (14); merely declaring the set builds natively and matches Node (22). Declaration census runs even for unused module functions at `modules.go:195`. This confirmation follows the census code; `--explain-checks` does not print a census-road trace.

The no-if native control (23) has one exposed number overload and an optional implementation parameter, returns `a` without a test, and declares the implementation result as number. It uses the ordinary result-relation road and natively prints `2\n1\n`. Probe 24 has two number-result overloads and the same output. Both stdout and stderr compare byte for byte equal to Node.

## Arity assumptions by accepted road

All result roads run **after the common overload-vs-implementation parameter comparison** at `census_small.go:186–220`, which uses the overload's own list, absent = undefined (and special handling for rest). None of those declaration acceptances establishes that a value assigned to a wider callable receives only the overload's arguments.

| Accepted result road | Code at baseline | Does its result justification depend on the overload's own arity? | Value use / call enforcement |
| --- | --- | --- | --- |
| Phantom, including nested array brands | `census_small.go:224–229`; `overload_phantom_result.go:26–53` | The common input check does. The result test itself compares whole implementation/result types after permitted brand stripping and, for array differences, freshness; it does not prune an extra-argument branch. | Overload function values are refused regardless. Brand-presence refusals are separate. |
| Presence | `census_small.go:231`; `census_overload_proof.go:25–35` | The common input check does. The result test requires every present implementation result member to serve the promise, independent of arity. **Omitting the presence check by body proof does use declared arity.** | Direct resolved calls run `censusOverloadResult`; lines 353–362 insert the undefined check unless body proof succeeds. Values are refused. |
| Predicate result | `census_small.go:234–240`; `overload_predicate_result.go:29–56,84–124` | Positional overload callback contracts matter. This result proof checks all returns, not reachability under absent-as-undefined seeds; the accepted callback claim must justify the returned value. It is not a general omission-of-extra-arguments proof. | Callback proofs are demanded where arguments are passed (`overload_predicate_result.go:8–15`); overload values still refuse. Return-type predicate overloads also use `census_small.go:326–327` / `checkedPredicateResult`. |
| Void | `census_small.go:241–242,290–310,329–332` | No extra arity dependency in the result rule: only exactly void over unowned scalar/undefined results is accepted, and no read of void lowers. Common input check still uses declared arity. | Values refuse; direct calls keep implementation representation so discarded values release correctly. |
| Body proof | `census_small.go:244–247`; `census_overload_proof.go:42–83` | **Yes, directly.** Omitted parameters seed undefined; reachable returns are proven under those types. It would be unsound to reuse that premise on an unguarded wider callable receiving an extra observed argument. | The same declaration proof suppresses checks at `overload_result_check.go:146–148` and `census_small.go:353`. Values refuse first. |
| `narrowOverload` fallback / call check | `census_small.go:249–250`; `overload_result_check.go:28–68,135–190` | Classification is declaration-based and the common input check assumes own arity. An **executed** runtime membership check tests the actual result, so it does not require an arity premise. **Its body-proven exemption does**, as above. | Only `.ts` checkable promises qualify. Checks are per call resolving to that overload; discarded results can erase them. `checkOverloadResults` (`:288–298`) guards that every lowered narrow-overload call was settled. Calls through wider value slots would not resolve to that declaration; this is why values are refused. |

The parameter-side `overloadCallbackUse` exception also reasons from the overload's positional types (`census_small.go:209–212`, `overload_callback_use.go:8–47`); it is not evidence about arbitrary extra arguments at a wider value call.

## Reproduction and artifacts

Fresh detached worktree was created with:

```sh
git fetch origin 49f1d643064a9231f7f343d8bbd02ba718866e6a
git worktree add --detach /workspace/scratch/overload-arity-alias 49f1d643064a9231f7f343d8bbd02ba718866e6a
cd /workspace/scratch/overload-arity-alias
git submodule update --init --recursive --depth 1
mkdir -p /tmp/adamic-gate
source /workspace/adamic-tools/env.sh
go build -o /workspace/scratch/overload-arity-alias-adamic ./cmd/adamic
python3 stage3/census/overload-arity-alias/run.py /workspace/scratch/overload-arity-alias-adamic
```

The root README describes the Go → C → clang compiler but only gives a `dedication/` build command. Stage-0 CLI/build instructions are in `docs/cli.md:11–46` and native release settings in `docs/native-builds.md:3–11`; these native runs use `adamic build <file.ts> -o <binary> --explain-checks` followed by executing the binary. Go 1.27.1, clang 20.1.8, Node v24.19.0, linux/amd64. Submodules: cohere `5360753e19e4417675a90e92a1823bea66b13918`, TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`.

- [Runner](run.py) captures each Node execution, native build, and successful native execution separately. It compares stdout and stderr as raw bytes, with no normalization. Node warnings are suppressed with `--no-warnings`; exceptions are retained.
- [Summary JSON](outputs/summary.json) records each status and both byte-equality flags. Null means no native execution, not a passing comparison.
- [Environment](outputs/environment.txt) records tool versions and pins. Report-only commits were made while the baseline binary built; no compiler source changed.
- `outputs/<program>.{node,build,native}.{stdout,stderr,exit}` stores exact streams and exit status. Native files exist only for successful builds. Compiler diagnostic output is never compared to Node runtime output.
- Initial probes needed string conversion for console logging and indexed-read guards; the saved final programs and outputs include those corrections. No compiler patch, guard bypass, or mutation was used.

Partial findings were pushed at `5b78c07dd` and `3a15ea061`; this final report and all evidence are on `typescript/overload-arity-alias`, pushed without force. The conclusion is bounded to this baseline and these roads: the proposed wider-value premise is unsound in isolation, but the tested compiler does not permit the value conversion needed to exploit it.
