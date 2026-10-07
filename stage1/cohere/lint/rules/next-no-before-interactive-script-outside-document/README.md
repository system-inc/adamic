# no-before-interactive-script-outside-document candidate

Already claimed by wave1-12. decision.a ports the upstream predicates and messages.a preserves the exact Go id and description. This is an unregistered decision candidate, not a working rule.

Blocked: stage1 Parser refuses JSX before visitation; there is no JSX node, attribute or decoded entity adapter contract to attach this decision to. The shared Go oracle also forces ScriptKindTS on JSX. Import-dependent rules additionally require the whole-file bindings supplied by that adapter. No shared parser, generator or harness source is edited.

Attribute names are exact; values are decoded plain strings.

Ten manually extracted decision cases match real Go verdicts and exact messages on source Node, emitted JavaScript and sanitized native. One compiling decision mutant per candidate is caught only by output comparison on all three backends. This is predicate validation, not JSX extraction or finding range parity. No source-corpus parity or findings throughput is claimed. Prior independently run positive JSX blocker witnesses remain in the claim evidence directories. Integration must supply extraction, registration, ranges and original-fixture comparisons before this can be called ported.
