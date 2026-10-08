# Shared checker migration

prefer-rest-params uses only the area's RuleContext.checker. It declares typed and node
listeners in rule.json, receives its subscribed ParseNode, and keeps its helpers
inside this directory. The upstreamTest prefix captures every corresponding real
cohere test. The Go adapter uses the unchanged upstream rule; configured rules
decode their own options.

Independent captured-case comparison: 23/23 byte-identical on unchanged
Go, source Node, emitted JavaScript and sanitized native, including complete
findings, individual fixes, suggestions and resulting source. All captured rows,
including failures, are preserved in typeaware/wave07_jsx/checker-landing-evidence.
No recovery source was silently renamed or discarded.

Go process/checker/lint elapsed sum: 1.351764 s. Native sum:
1.711808 s. These are per-case measurements under concurrent
package checks, not isolated throughput benchmarks. Each runtime's raw output
and stderr and each manifest/project are in owned-cases.tar.gz.

This rule's valid message mutant is named in mutant.json. The exact six-mutant
rerun after whitespace cleanup passed; independent Go bytes catch every mutation
on all three port runtimes. Full package witnesses and any aggregate blockers are
reported in typeaware/wave07_jsx/CHECKER_LANDING_REPORT.md and CHECKER_BLOCKERS.md.
