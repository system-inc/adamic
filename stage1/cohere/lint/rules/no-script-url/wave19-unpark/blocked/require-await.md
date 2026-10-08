# require-await

Blocked after the RuleContext checker integration at area 9b7547976. No descriptor is installed and no port/parity or mutant pass is claimed.

Upstream call: `checker.Checker_getResolvedSignature(...), signature target and declared generic parameters` at `cohere/internal/lint/rules/core/require_await.go:988`.

signature-shape is instantiated parameter types only; no declared target signature, type-parameter substitutions or complete thenable member/index metadata. wave19-generic-call / wave19-type-signatures / wave19-type-members / wave19-heritage-members are unregistered.

Minimal source:

```ts
declare function consume<T extends () => Promise<number>>(callback:T):void; consume(async () => 1);
```

The old isolated implementation stays on codex/typeaware-wave-19 for reference only. No private checker, private shared helper, shared dispatch edit, or syntax approximation was added.
