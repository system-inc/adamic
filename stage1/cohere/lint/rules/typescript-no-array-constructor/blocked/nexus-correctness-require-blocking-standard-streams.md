# nexus/correctness-require-blocking-standard-streams

Rechecked on lint-checker/facts `d845dccde413c89643293e808626344d12e3f023` for wave-19.

First stopping call: `program.ResolveModule(importer, specifier); program.GetSourceFileForResolvedModule(resolved.ResolvedFileName)`, at `cohere/internal/lint/rules/nexus/correctness_require_blocking_standard_streams.go:279 and :283`.
Needed question: runtime-modules: program file graph and resolved runtime edges. It is absent from this facts slice; existing symbol declarations/provenance do not answer it. No descriptor, approximation, private checker or shared helper change is introduced.

Reproducer (Node declarations and a sibling blocking.ts):

```ts
import "./blocking"; process.stdout.write("x");
```

New matched upstream cases: 0. New mutant runs: 0. This note is not a port certification. Rule-specific program reads will match upstream when the missing contract arrives.
