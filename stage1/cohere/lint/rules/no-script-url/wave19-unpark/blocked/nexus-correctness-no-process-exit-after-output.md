# nexus/correctness-no-process-exit-after-output

Blocked after the RuleContext checker integration at area 9b7547976. No descriptor is installed and no port/parity or mutant pass is claimed.

Upstream call: `writers.ctx.TypeChecker.GetResolvedSignature(call), then signature.Declaration() at line 318` at `cohere/internal/lint/rules/nexus/correctness_no_process_exit_after_output.go:314`.

signature-shape supplies types, not the resolved signature declaration and its flags, body/source ownership. wave19-resolved-callee is not registered. Global declaration ancestry is missing too.

Minimal source:

```ts
function write(){process.stdout.write("x");} write(); process.exit(0);
```

The old isolated implementation stays on codex/typeaware-wave-19 for reference only. No private checker, private shared helper, shared dispatch edit, or syntax approximation was added.
