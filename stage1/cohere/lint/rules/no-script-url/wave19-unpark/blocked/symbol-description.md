# symbol-description

Blocked after the RuleContext checker integration at area 9b7547976. No descriptor is installed and no port/parity or mutant pass is claimed.

Upstream call: `resolvesToAGlobal(ctx, callee), defined at core/no_new_native_nonconstructor.go:99` at `cohere/internal/lint/rules/core/symbol_description.go:95`.

The shared resolvesToAGlobal helper has no RuleContext implementation.

Minimal source:

```ts
Symbol(); function f(Symbol:any){return Symbol();}
```

The old isolated implementation stays on codex/typeaware-wave-19 for reference only. No private checker, private shared helper, shared dispatch edit, or syntax approximation was added.
