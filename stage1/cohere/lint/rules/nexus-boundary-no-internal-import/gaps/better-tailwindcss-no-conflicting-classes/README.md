# better-tailwindcss/no-conflicting-classes

Unregistered partial port, claimed by wave1-12. All Adamic code is `.a`.

Implemented: Deduplication, symmetric per-class pairing, variant and selector agreement, whole-property-set matching, composing-utility exclusion, property ordering and exact Go message.

Blocked: RuleContext has no Program, Tailwind entry point, design-system filesystem or resolveClassFacts bridge. The shared reader and JSX adapter are also absent. Go returns no listeners for nil Program, so empty raw-oracle parity would not certify this rule. Skip/error behavior also needs the project bridge.

No rule.json is supplied: registering an incomplete adapter would silently miss findings. No fix or suggestion is invented; all three pinned Go implementations report diagnostics without edits.

The comparison package lives in ../better-tailwindcss-enforce-shorthand-classes. Its 562 decision cases use actual Go private functions through a test-only overlay. Conflict tests inject already-resolved facts, leaving Go pairing and message code unchanged. These are decision-contract tests, not end-to-end rule tests or corpus certification.
