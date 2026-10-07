# Frozen rule validation contract

This fixture contains the registration and shared lint contracts from rule branch
99c66e1e34036ef20bde058e43bb8ae79136ccb8. Source text is frozen as raw witnesses,
not installed into the repository shared harness. The generator is an unchanged
historical copy apart from its private package name and explanatory comment.
Install writes these contracts only into a temporary directory; the owned tests
then copy current rule implementations, use the current Adamic compiler/parser,
and compare real pinned Go cohere findings, repairs and suggestions on source Node,
emitted JavaScript and ASan/UBSan native.

This permits the owned rule commits to rebase onto current main without restoring
unowned shared files. It certifies rule semantics under the prior registration
contract. It does not certify the current production monolithic dispatcher or
provide the pending numeric handed-node dispatcher. Integration is parked on
shared harness unification #zmh9v36. Once that contract lands, remove this fixture
and adopt the shared harness in the owned validators.

The original JSX, complete-suggestion and Tailwind provider refusal controls remain
in the owned suites. Go AST projection and Go-fact replay are still conditional
semantic checks, not proof that those missing production adapters exist. No shared
finding, context, main, registry, oracle or comparison file is modified.
