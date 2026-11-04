An extra edit outside the owner is rejected by the edit-span assertion

A temporary broken copy of the actual adapter appends one additional edit: insert a space before invokeCallbacks' `(existing as string[]).push(...fileNames);`. It changes neither pinned region and remains valid TypeScript. The production adapter is never modified.

Command:

```sh
node stage3/adapt/99-wasm-crypto-import/.span-escape-adapter.cjs /workspace/scratch/crypto-scoped-proof/probe
```

Exit 1, exact red:

```text
AssertionError [ERR_ASSERTION]: edit outside getNodeSystem and node: import insertion
```

Before bytes and nanosecond mtime are unchanged. A separate control disables only that assertion in the broken copy: it exits 0, files=1, writing both the legitimate crypto edit and the extra space. Its owner/import after hashes and binding checks still pass. Thus no parse, binding or scoped-hash failure masks the span check. scoped.py restores the source and removes both temporary adapter copies in finally. Full logs: scoped-span-escape.log.gz and scoped-span-disabled-control.log.gz.
