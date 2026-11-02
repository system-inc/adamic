Binding guard evidence for #he8tysb head 1 revision.

Command, with the pinned stage3/api packages on NODE_PATH:

    node stage3/adapt/93-builtin-imports/proofs/bindings.cjs /workspace/scratch/adapt-93-revision-proof/control /workspace/scratch/adapt-93-revision-proof/after /workspace/scratch/adapt-93/stage3/adapt/93-builtin-imports/proofs/bindings.json

The actual before and after trees both pass: 27 qualified AST accesses on 22
source lines, no bare uses, no competing declarations, and all accesses resolve
to the expected binding (the original local before, the import after).

The same command then plants three virtual-source mutants, each alone:

1. Append `const _path = { join: () => "wrong" };` at module scope.
2. Add `_fs?: typeof import("fs")` as a parameter of readFile, enclosing its
   existing _fs.readFileSync access.
3. Append `void _os;`, a bare namespace value read.

It calls the real adapter's validateBindings on independently bound virtual
programs, bypassing the complete-file hash guard. Thus the hash check cannot
mask a missing semantic collision check. No build or -Werror failure is used.
Exact observed lines:

    CAUGHT module collision: collision or bare namespace use: _path
    CAUGHT enclosing shadow: collision or bare namespace use: _fs
    CAUGHT bare namespace value: collision or bare namespace use: _os

Exit 0 means all three negatives were rejected as required; a survivor makes
this proof fail. bindings.json records the actual exception messages, and
logs/bindings.log.gz retains the complete output. These probes write no source
files. All normal before/after symbol checks also run inside the adapter, with
the planned after program validated before its sole possible source write.
