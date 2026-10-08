# require-await

Rechecked on lint-checker/facts `d845dccde413c89643293e808626344d12e3f023` for wave-19.

First stopping call: `checker.Checker_getResolvedSignature(ctx.TypeChecker, call, nil, checker.CheckModeNormal); resolved.Target(); declared.TypeParameters()`, at `cohere/internal/lint/rules/core/require_await.go:988, :992, :993`.
Needed question: declared-call-signature; type-projection and shared thenability remain later dependencies. It is absent from this facts slice; existing symbol declarations/provenance do not answer it. No descriptor, approximation, private checker or shared helper change is introduced.

Reproducer (Promise declarations):

```ts
declare function consume<T extends () => Promise<number>>(callback:T):void; consume(async () => 1);
```

New matched upstream cases: 0. New mutant runs: 0. This note is not a port certification. Rule-specific program reads will match upstream when the missing contract arrives.
