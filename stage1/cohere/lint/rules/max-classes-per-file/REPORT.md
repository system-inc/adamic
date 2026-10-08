Ported max-classes-per-file from batch4 onto the unified registry harness.
The descriptor has node:true, subscribes to ClassDeclaration/ClassExpression,
counts only handed nodes and emits the one file-level finding in its finish hook.
Declarations inside functions do not count; expressions follow ignoreExpressions.
The reported span excludes leading and trailing trivia, matching pinned Go cohere.
Messages are moved verbatim. No helper or shared file edit was needed.

The Go adapter returns typed MaxClassesPerFileOptions, decodes captured Maximum
and raw max/integer shapes, preserves defaults and IgnoreExpressions, and ignores
unrelated object fields on shared all-rule rows. The native factory reads options
only when enabled, because factories are also constructed for another selected
rule. The options guard is unchanged. Upstream test prefix Test includes all six
actual tests, including TestDecodeMaxClassesPerFileOptions; captured records are
filtered by exact rule name. The sidecar helper is not on this base; all three
owned witnesses fire on default options, and captured upstream cases cover options.

Validation on area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898:
- lint-registry, gofmt and full go vet pass.
- TestRulesAgree: 2,143 captured cases; Go, Node, emitted JavaScript and sanitized
  native match 13,092,595 bytes; PASS 50.88s.
- TestOwnedWitnesses: all four engines match 94,697 bytes; PASS (see final log).
- TestMutants/one_excess_class_is_silently_allowed: compiled mutant permits one
  excess class; caught on Node, emitted JavaScript and sanitized native at owned
  witness case 145, line 2738. No compiler failure is counted as a kill.
- Full final command passes; exact command and durations are in evidence.

The initial agreement runs exposed two owned options bugs, both fixed above;
those failure logs are retained beside the final green log. The pre-existing
method-signature-style malformed corpus cases are explicitly checked for parser
refusal by the shared harness. This rule has no known checker/RegExp/parser/
Tailwind blocker. Other rules' mutant subtests, the full repository gate and the
17 external stage-1 correctness checks were not run in this assigned unit.

Setup succeeded: cache warm/done 46s; Go/clang/Node/submodules ready 0s,
nproc 5. Only identified generated ELF/ar files from prior owned scratch were
removed for space; source, patches and oracle logs were kept.
