# @typescript-eslint/dot-notation

Rechecked on lint-checker/facts `d845dccde413c89643293e808626344d12e3f023` for wave-19.

First stopping call: `ctx.Program.Options().NoPropertyAccessFromIndexSignature.IsTrue()`, at `cohere/internal/lint/rules/typescript/dot_notation.go:88`.
Needed question: index-signature-access (compiler option), ordered property modifiers and index infos. It is absent from this facts slice; existing symbol declarations/provenance do not answer it. No descriptor, approximation, private checker or shared helper change is introduced.

Reproducer (noPropertyAccessFromIndexSignature: true):

```ts
declare const x: Record<string, number>; x["p"];
```

New matched upstream cases: 0. New mutant runs: 0. This note is not a port certification. Rule-specific program reads will match upstream when the missing contract arrives.
