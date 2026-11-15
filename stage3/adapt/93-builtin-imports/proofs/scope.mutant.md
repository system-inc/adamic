Current scope proof (#p97jpsy)

The gate is TestBuiltinImportsP97jpsySingleStepScope in stage3/lane/adaptedtree.
It reads TestProduct_Stage3AdaptedTree's tree at apply.py's
refs/stage3/93-builtin-imports-before and -after, extracting src into its own
temporary directories. It runs check.cjs, both scope-mutant.cjs modes, then the
restored check. It never adapts or writes the product and never runs Go.

The oracle reads targets.json. Each whole-owner before/after SHA-256
authenticates the corresponding AST text from the two snapshots before the
after-text is replaced with the before-text. The remaining source must be byte
identical, including CRLF, indentation and trailing whitespace. All other src
files and the source inventory remain exact. New pinned owners participate
without a separate oracle allowlist. The removed declarations are inside the
pinned getNodeSystem; the three import insertions retain their separate exact
undo. Both comments-retained and comments-removed transpiles of the restored
source must equal the before transpile.

The default mutant adds /* scope-outside-mutant */ before the unrelated
getFileWatcherEventKind declaration in sys.ts, outside the pinned owner. The
check must exit 1 naming that marker, file and line. The inside mode inserts
/* scope-inside-mutant */ into getNodeSystem. The check must exit 1 naming the
pinned After owner getNodeSystem. Both edits are valid TypeScript; only the
scope oracle is run to reject them. Both restore saved bytes in finally.

stage3/lane/adaptedtree already has a compiler-dependencies.json entry covering
the compiler dependencies, so no census change is needed.

Historical evidence below predates these mutants and the owner-derived oracle.

Mutant evidence for #he8tysb head 1 revision's exact-source scope proof.

The with-93 tree gets one extra edit to a private ambient declaration:

    declare function clearTimeout(handle: NodeJS.Timeout | undefined): void;

becomes:

    declare function clearTimeout(handle: NodeJS.Timeout | undefined | null): void;

This extra accepted input type has no emitted JavaScript. The mutant deliberately
passes both comments-retained and comments-removed JavaScript comparisons first.
Only the full source undo check rejects the extra edit in this probe; a green
transpile comparison alone cannot enforce the adaptation's source scope.
No malformed TypeScript or build error is used to kill it.

Command (stage3/api/node_modules on NODE_PATH, Node v24.19.0):

    node stage3/adapt/93-builtin-imports/proofs/scope-mutant.cjs /workspace/scratch/adapt-93-revision-proof/control /workspace/scratch/adapt-93-revision-proof/after

Exit 1. Exact observed lines:

    PASS TestBuiltinTranspileScope: the extra type-only edit is invisible in both JavaScript comparisons
    AssertionError [ERR_ASSERTION]: TestBuiltinSourceScope: edit outside the three imports and removed declarations
        at TestBuiltinSourceScope (/workspace/scratch/adapt-93/stage3/adapt/93-builtin-imports/proofs/check.cjs:26:9)

The script restores the exact saved bytes in finally, including on this failure.
Restored command:

    node stage3/adapt/93-builtin-imports/proofs/check.cjs /workspace/scratch/adapt-93-revision-proof/control /workspace/scratch/adapt-93-revision-proof/after /workspace/scratch/adapt-93/stage3/adapt/93-builtin-imports/proofs/checks.json

Exit 0: 746 files inventoried, only sys.ts changed, exact undo, both JavaScript
comparisons pass. Raw output is in logs/scope-only-mutant.log.gz and
logs/scope-only-restored.log.gz. No mutant remains.

Additional negatives in the bounded preparation: adding a newline to sys.ts
exits 1 with `AssertionError [ERR_ASSERTION]: target source moved: sys.ts`,
without changing its bytes or timestamp. Replacing the unrelated terminal width
fallback 0 with 1 also exits 1 in TestBuiltinSourceScope. Both are restored.
