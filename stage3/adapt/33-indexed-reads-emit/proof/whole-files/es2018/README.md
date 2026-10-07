# transformers/es2018.ts at zero

All-code pinned latent census **4 -> 0**, clean files **39 -> 40 of 78**.
The source was unchanged from the previous reviewed wave before this edit.
Four existing ledger declines become original-point assertions: two U-endpoint
reads of objects[0], U-loop objects[i], and U-loop node.elements[i].

The object-literal factory obtains its spread flag from property children; an
empty property list cannot produce that flag. The chunk producer pushes defined
visited spread expressions or factory-created objects for ordinary properties.
Its expression/property visitors preserve those nodes, so this nonempty flagged
literal yields a populated nonempty chunk list. The first read also has a direct
objects.length guard. Inserting an initial empty object preserves nonemptiness.
The merge loop is 1 <= i < objects.length; helper callbacks receive a separate
two-element array, leaving the private chunk array unchanged. The comma-list
loop is 0 <= i < node.elements.length on a populated NodeArray; the required read
precedes its visitor, and replacements are collected in a different result array.
No read moves, defaults or skips.

All stock JavaScript bytes and adapter idempotence pass. The endpoint ! -> ?? 0
mutant fails the JavaScript and site-contract checks. Default oracle:
**106367 passing**, zero failing/pending, **empty baseline diff**,
217.481 seconds. Exact API exceptions are 20's 189 and 40's 28 lines;
every other reference byte and the complete file set match pristine.
Reproduce with before snapshot /tmp/emit33-close-source-before-es2018 and the
same source-only latent census, verify, mutant and default oracle commands as ts.
