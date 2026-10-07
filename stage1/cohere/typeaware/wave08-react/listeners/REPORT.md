Built: listener-only rule.json files for all nine owned rules, matching their numeric .a subscriptions.
Commits: parent 45a71d4d, based on current main e8ba3d5d; this report accompanies the manifest commit.
Checks: validate_manifests.py passes nine manifests against sanitized native declarations and the independent Go AST enum oracle.
Mutants: valid NewExpression manifest substitution is caught by numeric comparison; the compiling 214-to-215 native declaration mutant is caught again.
Uncovered: these are subscription metadata, not registered factory/visit adapters; numeric handed-node dispatch, JSX and React HIR/SSA remain blocked.

The harness branch's existing rule.json contract encodes kinds as SyntaxKind
names in configuration. These listener-only manifests use that spelling; the
owned .a modules already export the corresponding numeric listenerKinds arrays.
Rule code does not consume these configuration strings to decide relevance.
The test converts each manifest kind through the pinned TypeScript enum, then
compares with the Go AST numeric oracle and the sanitized native exports.

All nine manifest kind lists pass. Replacing the require rule's CallExpression
with the valid enum kind NewExpression passes JSON/name lookup but fails numeric
subscription comparison. The validator also reruns validate_listeners.py, which
builds a native wrong-kind mutant that exits zero with empty stderr and is caught
only by the independent Go numeric bytes. Raw outputs and logs are in validation/.

These files intentionally contain only name and kinds. They are not advertised
as complete shared registry descriptors: the shared registry also expects
factory, class, visit, ordering and oracle integration, whose contracts for the
existing type-aware Rules/ProcessPool profiles are unavailable. No factory or
handed-node visit export is invented. These manifests live in the owned wave
folder, outside the shared registry's production discovery path. Integration
must wire them to the eventual type-aware driver and callback adapters.

The existing shared ParseNode still exposes kind as string, and the type-aware
driver calls whole-file run methods. Legacy string comparisons and refetches
therefore remain; removing them requires the prohibited shared parser/driver API
change. The JSX parser and native source-to-React-HIR/SSA blockers remain as
recorded in ../REPORT.md. No batch 8 Diagnostic integration SHA was supplied in
this request, and current main remains e8ba3d5d. No shared source was edited.

No executable rule, compiler or bridge production code changed in this update,
so the fresh six-port parity/sanitizer/mutant and released-handle evidence from
../LISTENER_REPORT.md remains applicable. Full corpus and ownership gates were
not repeated for JSON metadata alone. The declaration sanitizer/oracle and
mutant checks were rerun. All new native sources remain .a. No new claims were
taken, and the push targets only codex/typeaware-wave-08.

Reproduce with validate_manifests.py --artifacts, --stage0, --archive and
--asan-archive, using the same compiler/checker paths as validate_listeners.py.
All test output was captured to files.
