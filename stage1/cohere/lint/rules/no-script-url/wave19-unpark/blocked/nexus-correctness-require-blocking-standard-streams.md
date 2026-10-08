# nexus/correctness-require-blocking-standard-streams

Blocked after the RuleContext checker integration at area 9b7547976. No descriptor is installed and no port/parity or mutant pass is claimed.

Upstream call: `program.ResolveModule(importer, specifier), then GetSourceFileForResolvedModule at line 283` at `cohere/internal/lint/rules/nexus/correctness_require_blocking_standard_streams.go:279`.

The checker has no program source/module graph question. wave19-program-modules is not registered.

Minimal source:

```ts
import "./blocking"; process.stdout.write("x");
```

The old isolated implementation stays on codex/typeaware-wave-19 for reference only. No private checker, private shared helper, shared dispatch edit, or syntax approximation was added.
