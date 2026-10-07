Temporary: comes out when codex/stage3-optional-declarations lands its upstream-build fix

Slice-only plan, published before implementation. Require slice.json at the input
root. Never apply to the full adapted upstream tree. Add explicit `| undefined`
to DiagnosticMessage's three optional fields that diag always materializes:
reportsUnnecessary, reportsDeprecated, and elidedInCompatabilityPyramid.
This addresses TS2375 and changes no JavaScript behavior. Preserve optional
property presence and absence. Adaptation 20 stays excluded.

Validate idempotence, the unchanged 509,014 Node token bytes, and the stage 3
baseline oracle before calling the adaptation complete. Remove one added union
as a checker mutant and require TS2375 to return. Unvalidated work is reported
as a scratch probe.
