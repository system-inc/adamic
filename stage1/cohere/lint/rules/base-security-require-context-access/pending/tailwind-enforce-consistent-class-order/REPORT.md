Built: portable class order components in .a: UTF-16 lexical ascending/descending ordering, separate slot fixes and message text. This is not a registered full rule.
Commits: components 2cc9128a; claim d316a491 was pushed before implementation.
Commands and outputs: components_validate.py PASS, 336 mixed component cases and 69,771 identical bytes on Node, emitted JavaScript and ASan/UBSan native.
Mutants: class_order_descending_reversed builds and runs with zero exits and empty stderr, then fails only byte comparison on all three backends. This is component credit, not full-rule mutant credit.
Not covered: full-rule findings/fixes parity over compiler, stage1 or upstream fixtures; full-rule throughput is unavailable. project CSS loading, candidate resolution and official/strict ranking, unknown/component groups, class-literal and template surfaces, source/deprecation fix checks and the full options/entry point.

## Blockers and scope

Both Tailwind rule bodies declare ReadsCompilerOptions | ReadsDesignSystem and read DesignSystemForProgram(ctx.Program) before registering listeners. The shared Adamic RuleContext has no Program or CSS input, including on the now-published lint-harness-dot-a branch. No Adamic project design-system/candidate/ranking provider exists in stage1. A hardcoded repository table or a Go-projected rule verdict would not hold independent native behavior byte for byte. Those dependencies cannot be supplied by editing the shared context under this unit's ownership restriction. No shared file was changed.

The unmodified upstream control tests fail, rather than reporting skipped tests as a pass, because they cannot find Tailwind 4.3.3 from /Users/kirkouimet/Projects/ahra/app/_theme/styles. Raw output is in evidence/upstream-blocker.log. The class-order rule also emits multiple separate slot fixes; the current shared Finding/oracle contract cannot serialize them, and the published harness oracle still panics when len(d.Fixes) != 1.

The Go scratch overlay exposes actual private component functions. Canonical tests supply explicit candidate facts to mergeKey/collapseOutputFor/rebuildClass; they do not validate candidate parsing or CSS loading. Ordering tests exercise actual orderClasses in asc/desc mode without a design system and actual slotFixes. The mixed corpus covers every family in both directions, logical on/off, mismatched values/variants/importance, writable source values, Unicode ordering and slot ranges. Source Node, emitted JavaScript and sanitized native run the same owned components. The full rule entry points are neither registered nor claimed complete.

Reproduce using the script in ../tailwind-enforce-canonical-classes/components_validate.py with --scratch /tmp/wave11-tailwind-components-final, redirecting its output to a log. The components are retained separately under this owned unit's pending directory until their provider and listener contracts can be implemented.
