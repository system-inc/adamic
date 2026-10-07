# better-tailwindcss/no-concatenated-classes

Unregistered partial port, claimed by wave1-12. All Adamic code is `.a`.

Implemented: Before/after interpolation-seam checks, Go ASCII seam whitespace, Unicode strings.Fields splitting, first/last fragment selection and exact Go message.

Blocked: A production ClassTemplatesIn adapter is absent. It must find only configured attributes/callees/variables, retain traversal order and static segment ranges, and decode escapes; JSX parsing is also absent. Plus concatenation must remain quiet.

No rule.json is supplied: registering an incomplete adapter would silently miss findings. No fix or suggestion is invented; all three pinned Go implementations report diagnostics without edits.

The comparison package lives in ../better-tailwindcss-enforce-shorthand-classes. Its 562 decision cases use actual Go private functions through a test-only overlay. Conflict tests inject already-resolved facts, leaving Go pairing and message code unchanged. These are decision-contract tests, not end-to-end rule tests or corpus certification.
