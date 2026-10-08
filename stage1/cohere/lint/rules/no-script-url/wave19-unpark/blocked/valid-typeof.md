# valid-typeof

Blocked after the RuleContext checker integration at area 9b7547976. No descriptor is installed and no port/parity or mutant pass is claimed.

Upstream call: `identifierIsShadowed(ctx, operand), defined at core/no_undef_init.go:159` at `cohere/internal/lint/rules/core/valid_typeof.go:147`.

The shared identifierIsShadowed helper has no RuleContext implementation. Global undefined can have zero declarations; the global helper is not its negation.

Minimal source:

```ts
typeof value === undefined; function f(undefined:number){return typeof value === undefined;}
```

The old isolated implementation stays on codex/typeaware-wave-19 for reference only. No private checker, private shared helper, shared dispatch edit, or syntax approximation was added.
