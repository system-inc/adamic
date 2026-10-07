Built: reviewed the named harness and corrected the remaining integration blocker; no rule implementation changed.
Commits: follows pushed 202d30951 on wave-09; current main c01907a7 remains an ancestor; harness ab70f38d4 is not on main.
Commands/output: fetched all origin heads and read the complete harness context, finding model and registry generator; working tree was clean.
Mutants: no new executable check was added; all prior rule/component mutants and sanitizers remain recorded in LANDING_C019_REPORT.md.
Not covered: harness integration or a new oracle sweep; numeric descriptor compatibility and dynamic native RegExp remain blockers; no new claims.

The user named harness revision
ab70f38d47de1d4974082b38f84a56af2368b7af on origin/lint-rules/harness.
It is available after the origin refresh, but is not an ancestor of current
origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. Their merge base is
ef3d907ecdc4c771b016f7d9c52372def057a340. Wave-09 remains based on current
main; the old-base integration branch was inspected read-only. It was not
merged locally ahead of integration, pushed to, or treated as landed main.

The complete RuleContext, Finding and registry generator were read. The
seven-argument report remains supported; reportNode and reportRange add
automatic-edit and suggestion reporting. The optional Node descriptor flag
now produces visit(node,index[,parent]), so an absent handed-node callback is
no longer the correct description of that branch. This supersedes that part
of the earlier blocker reports.

The numeric incompatibility remains concrete. registry.Descriptor.Kinds is
[]string with JSON name kinds. Discovery validates SyntaxKind names as
identifier strings. Render emits switch(node.kind) with string case labels.
ParseNode.kind remains string, and its constructor takes string. This unit's
six requested numeric rule.json arrays are therefore not accepted by that
descriptor decoder. The generator needs the numeric contract before these
descriptors can integrate as requested. No per-rule conversion from string
names, shared generator edit or duplicate numeric parser was introduced.

The supplied RuleContext has no checker-program/bridge state. Existing
type-aware Rules owns that state and its legacy ask path still refetches the
node. Integrating that path also needs the shared type-aware context adapter;
reportRange by itself does not supply checker facts. Existing independent
full-byte finding/fix/suggestion evidence is retained, but there is no claim
that historical execution meets the new speed contract.

Dynamic RegExp compilation remains separately refused by current native
lowering. The shared regex table has been inspected as recorded in
LANDING_C019_REPORT.md; no applicable fixed-pattern row exists for these
claims. No hand-rolled matcher was added. Neither regex claim qualifies for
the React analysis parking exception. They remain incomplete, and no new
batch was claimed. All prior native/Go timings, corpus and released-handle
checks remain in the latest landing report. Toolchain setup is reused:
88 seconds, nproc 5. No test rerun was needed for this read-only review and
documentation correction, which changes no executable behavior.

Only codex/typeaware-wave-09 receives the status update. Integration owns main
and area branches. Shared source excerpts are retained in validation-harness-ab70
to make the inspected contract reviewable without relying on a moving ref.
