# react/jsx-no-undef: certified on shared checker facts

Source: ba85c35ba; shared facts: d845dccde (already an ancestor, merge reports up to date).
All 45 captured upstream cases and the firing witness match unchanged Go byte for byte on source Node, emitted JavaScript and ASan/UBSan native, including findings, fixes and suggestions.
The owned message mutant compiles, runs and is caught by the differential comparison on all three runtimes.
SymbolDetails from checker_declarations.a replaces the private symbol decoder; queries use the run's RuleContext.checker, with no private program or recorded-answer stand-in.
The unused private bindings module and private symbol decoder are deleted. Explicit allowGlobals options are included in the upstream comparisons.
programReads is explicitly empty, exactly as cohere/internal/lint/rules/react/jsx_no_undef.go:83 declares at the pinned cohere commit.
Matrix evidence: /workspace/wave-23/facts-wave07-evidence/result.json and events.jsonl.
Complete lint log: /workspace/wave-23/facts-wave07-lint-complete.jsonl; mutant lines: /workspace/wave-23/facts-wave07-mutants.log.
Registry, gofmt and vet pass. The aggregate gate retains the shared inventory, graph and recovery failures described in /workspace/wave-23/facts-wave07-final-report.md.

Full package: 26 top-level passes, 8 shared failures, one TestCheckerBridgeRefusalPending skip; 84 registered mutants caught and the indexed mutant blocked. Wall 2400.15s, nproc 5. No input skips or relaxed checks.
