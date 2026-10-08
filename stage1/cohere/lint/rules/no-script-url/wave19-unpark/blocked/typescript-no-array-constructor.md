# @typescript-eslint/no-array-constructor

Blocked after the RuleContext checker integration at area 9b7547976. No descriptor is installed and no port/parity or mutant pass is claimed.

Upstream call: `resolvesToAGlobal(ctx, callee), defined at core/no_new_native_nonconstructor.go:99` at `cohere/internal/lint/rules/core/no_array_constructor.go:205`.

The shared resolvesToAGlobal helper has no RuleContext implementation. Raw node-symbol-origin is also unregistered on area; a private copy of this shared judgment is forbidden.

Minimal source:

```ts
Array(1,2); function f(Array:any){return Array(1,2);}
```

The old isolated implementation stays on codex/typeaware-wave-19 for reference only. No private checker, private shared helper, shared dispatch edit, or syntax approximation was added.
