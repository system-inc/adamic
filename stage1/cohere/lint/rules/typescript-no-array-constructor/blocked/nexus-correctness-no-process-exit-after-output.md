# nexus/correctness-no-process-exit-after-output

Rechecked on lint-checker/facts `d845dccde413c89643293e808626344d12e3f023` for wave-19.

First stopping call: `writers.ctx.TypeChecker.GetResolvedSignature(call); signature.Declaration()`, at `cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output.go:314 and :318`.
Needed question: resolved-callee declaration/body ownership and function flags. It is absent from this facts slice; existing symbol declarations/provenance do not answer it. No descriptor, approximation, private checker or shared helper change is introduced.

Reproducer (Node declarations):

```ts
function write(){process.stdout.write("x");} write(); process.exit(0);
```

New matched upstream cases: 0. New mutant runs: 0. This note is not a port certification. Rule-specific program reads will match upstream when the missing contract arrives.
