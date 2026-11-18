# Type the syntax cursor's cached position (#x2cjvy9)

Base commit `46d8f9a44c7d6c7dcbc9f82cab3cf8742f835bc5`. TypeScript v6.0.3 commit `050880ce59e30b356b686bd3144efe24f875ebc8`.

`IncrementalParser.createSyntaxCursor` stores positions from `currentNode(position: number)` in `lastQueriedPosition`. The one-member enum `InvalidPosition` names its initial sentinel. The adaptation adds the explicit `number` annotation to the slot.

The existing parser adaptations pin other owners or individual sites. This adapter pins the entire `createSyntaxCursor` function before and after the annotation. It refuses drift, parse errors and mixed line endings, preserves CRLF, and compares the emitted JavaScript bytes with comments retained and removed before writing. A replay writes nothing. The mutant restores the original declaration.
