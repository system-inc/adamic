# @typescript-eslint/consistent-generic-constructors

Rechecked on lint-checker/facts `d845dccde413c89643293e808626344d12e3f023` for wave-19.

First stopping call: `ctx.Program.Options().IsolatedDeclarations.IsTrue()`, at `cohere/internal/lint/rules/typescript/consistent_generic_constructors.go:240`.
Needed question: isolated-declarations. It is absent from this facts slice; existing symbol declarations/provenance do not answer it. No descriptor, approximation, private checker or shared helper change is introduced.

Reproducer (isolatedDeclarations: true):

```ts
class Box<T>{} const x: Box<string> = new Box();
```

New matched upstream cases: 0. New mutant runs: 0. This note is not a port certification. Rule-specific program reads will match upstream when the missing contract arrives.
