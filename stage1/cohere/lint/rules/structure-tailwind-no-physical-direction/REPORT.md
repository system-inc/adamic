# structure/tailwind-no-physical-direction

Maps physical Tailwind utilities to the exact logical-class message, preserving Go token-shape heuristics, Unicode Fields splitting, file gates and rtl/ltr exemptions.

Implementation is `rule.a`; `rule.json` registers through the directory contract. `oracle.go` calls the unchanged Go cohere rule. `mutant.json` owns `physical_direction_exemption_removed`. The mutant compiled and ran successfully and was caught by output comparison on source Node, emitted JavaScript and ASan/UBSan native.

See [the continuation report](../../claims/wave1-10-next-report.md) for exact commands, corpus results, rates, exclusions and raw logs. The shared registry on this branch still requires `rule.ts`; the owned Tailwind directory includes a reviewable scratch compatibility overlay for `.a` registration, independent edit ranges and parser modes. These are measured candidate ports awaiting that shared integration, not a claim that the unmodified repository gate passes.
