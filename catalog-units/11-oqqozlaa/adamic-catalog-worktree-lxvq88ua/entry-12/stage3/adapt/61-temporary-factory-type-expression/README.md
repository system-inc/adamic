Temporary: comes out when factory generic initialization proof lands

Only the write receiver is widened; no assertion of presence or change to the returned generic type is made.

Source: TypeScript 6.0.3, `src/compiler/factory/nodeFactory.ts`, `createJSDocTypeLikeTagWorker`.

Before: `node.typeExpression = typeExpression`.

After: `(node as { typeExpression: JSDocTypeExpression | undefined }).typeExpression = typeExpression`.

The script requires exactly one owner/site, is idempotent, and compares stock
TypeScript emitted JavaScript byte for byte before writing. Full stage 3
oracle and parser comparison results are recorded in the parser evidence.

Validated cumulatively after 60: full default stage 3 oracle 106,367 passing,
zero failing/pending, empty baseline diff, 251.273 seconds. Only
the existing mechanically sanctioned API snapshot is accepted; its post-run
exact attribution check passes. The parser dump retains its original hash;
removing the view restores its diagnostic and the planted runtime edit fails
the JavaScript equality guard.
