# TestJSDocTagsAreSharedByClones's mutant

Task #7ez1s16. A clone made before the first tag read shares the parsed
function's JSDoc array. After the original fills the cache, the clone must read
the same deprecated tag. Stock TypeScript 6.0.3 and the adapted product each
print `[[332,"hello"]]`.

The mutant changes the adapter's `owners.json` after texts and
`jsDocCache.ts` so the product rebuilds with a node-keyed table.

```diff
-getCachedJSDocTags(node.jsDoc)
+getCachedJSDocTags(node)
-setCachedJSDocTags(node.jsDoc, tags)
+setCachedJSDocTags(node, tags)
-clearCachedJSDocTags(node.jsDoc)
+clearCachedJSDocTags(node)
-const cache = new Map<JSDoc[], readonly JSDocTag[]>();
+const cache = new Map<object, readonly JSDocTag[]>();
```

The three cache functions also accept `object` keys. Both hereby build tasks
type-check the resulting tree. The test reaches Node, which reports that the
clone's tag array is empty.

- Head: `--- PASS: TestJSDocTagsAreSharedByClones (573.31s)`.
- Mutant: `--- FAIL: TestJSDocTagsAreSharedByClones (597.87s)`. The clone probe reports `+ []`.
- Restored: `--- PASS: TestJSDocTagsAreSharedByClones (1.21s)`. Both adapter files compare equal to their pre-mutant copies.

The house's sealed proof follows.
