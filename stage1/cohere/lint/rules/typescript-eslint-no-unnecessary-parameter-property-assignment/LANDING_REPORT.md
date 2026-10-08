# Recovered-parser landing

The wave 2 base restores this rule after the historical parser refusal.
All 92 upstream cases are included, including modified binding-pattern
parameters. The original gap input remains under testdata and the additional
modified-object-binding.ts.txt witness also fires a real finding.

The trailing-semicolon mutant is caught by Go byte comparison on Node, emitted
JavaScript and sanitized native. Restored sources match Go on all three over
94 upstream/witness inputs and 1181 total inputs, including all stage1 sources
and TypeScript 6.0.3 src/compiler. Runtime exits are zero and stderr is empty.

See ../nexus-import-require-module-alias/WAVE2_LANDING_REPORT.md and its evidence
for the full package run, exact commands and digest proof. Earlier REPORT.md
records are historical and do not describe this restored landing.
