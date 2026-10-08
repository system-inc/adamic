# nexus/correctness-no-uncleared-race-timeout

Blocked after the RuleContext checker integration at area 9b7547976. No descriptor is installed and no port/parity or mutant pass is claimed.

Upstream call: `ast.IsGlobalScopeAugmentation(container.Parent)` at `cohere/internal/lint/rules/nexus/correctness_no_uncleared_race_timeout.go:263`.

Current declaration-details exposes only one parent kind/name/span, without ancestor flags or global augmentation. declaration-ancestry is not registered in the shared Inspect switch.

Minimal source:

```ts
Promise.race([work(), new Promise(resolve => setTimeout(resolve, 10))]);
```

The old isolated implementation stays on codex/typeaware-wave-19 for reference only. No private checker, private shared helper, shared dispatch edit, or syntax approximation was added.
