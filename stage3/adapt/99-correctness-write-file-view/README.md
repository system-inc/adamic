# Readonly writeFile capture

Adamic and Compiler, #b8ffarm. Base: adamic
8e3a784aa5a601f57342c07b68f145fe56252c21; TypeScript v6.0.3,
050880ce59e30b356b686bd3144efe24f875ebc8.
This is JavaScript-changing and value-identical: one local alias statement is
added, and the existing read selects that same alias. There is no public row.

At adapted sys.ts:1385, patchWriteFileEnsuringDirectory now reads:

```ts
const writeFileView: { readonly writeFile: System["writeFile"]; } = sys;
const originalWriteFile = writeFileView.writeFile;
```

Nothing else in this owner changes. The readonly view is the original object,
not a copy or wrapper. The read obtains the exact function once, with the same
getter receiver, before sys.writeFile is reassigned. The existing
originalWriteFile.call(sys, ...) remains; function identity and own .call
behavior are preserved. A live-reader arrow would recurse into the replacement;
bind would change the function and its .call lookup. Neither is used here.
The new local adds no function call or allocation. The measured lane has 2,200
captures and 4,627 callback calls on both sides, with zero added calls.

The initial ownership search covered owners.json, targets.json and every
adapt.cjs under stage3/adapt. The timer-closures pin contains a call in
TestServerHost, not this owner. Its sys.ts function/interface owners, 93 and
crypto's getNodeSystem, and 41's ambient declarations are separate. This is a
new whole-owner adaptation, not an amendment to those pins.

owners.json pins the complete function before and after, including exact text
and SHA-256. The adapter accepts only those states, refuses missing, duplicate,
overloaded, moved, drifted or half-adapted owners and mixed/bare-CR endings
before writing. Uniform LF/CRLF and unrelated source are preserved. Replay
writes nothing and retains bytes and mtimes. Both JavaScript comment modes must
differ only inside this owner by the exact alias and redirected read.

Run contract.cjs MAIN with the pinned API on NODE_PATH. The contract checks
all refusal/no-write controls and both line endings. mutant.cjs TREE restores
the original whole owner; a second restoration also writes nothing.

witness.cjs executes both real pinned function bodies. Its filesystem helper is
a focused stub, and an in-memory hook observes the private capture. It covers
receiver-sensitive calls, one getter read, later reassignment, exact captured
identity, and an own .call override changed between invocations. Neither hook
nor stub is present in adapted source.

The proof and complete unfiltered lane receipts live on
[typescript/write-file-view-1385-evidence](https://github.com/system-inc/adamic/tree/typescript/write-file-view-1385-evidence/stage3/census/write-file-view-1385).
No checker, compiler, public declaration, API sanction, oracle, worker count or
expected count is changed. PROOF.md gives the acceptance contract and measured
source checks.
