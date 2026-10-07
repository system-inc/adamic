# Module-variable candidate port

This `.a` directory preserves Go cohere's variable-statement range, first plain
identifier named module, and one report per statement. Destructuring, parameters
and for-loop declarations remain exempt. It has no fixes or suggestions.

All 13 captured upstream cases and the compiler/stage1 corpus agree with Go on
source Node, emitted JavaScript and sanitized native using the scratch
compatibility proposal. Its compiling module_binding_ignored mutant is caught
only by output comparison on all three paths.

Normal integration is blocked by the shared registry's rule.ts requirement.
This is a validated candidate, not a normal-gate pass. Commands and evidence:
[continuation report](../structure-tailwind-no-physical-direction/REPORT.md).
