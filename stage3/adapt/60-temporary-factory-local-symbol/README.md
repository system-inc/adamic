Temporary: comes out when factory generic initialization proof lands

The write uses the explicitly undefined-admitting structural field view. It preserves the unproved generic return contract as an explicit temporary limitation.

Source: TypeScript 6.0.3, `src/compiler/factory/nodeFactory.ts`, `createBaseDeclaration`.

Before: `node.localSymbol = undefined`.

After: `(node as { localSymbol: Declaration["localSymbol"] | undefined }).localSymbol = undefined`.

The script requires exactly one owner/site, is idempotent, and compares stock
TypeScript emitted JavaScript byte for byte before writing. Full stage 3
oracle and parser comparison results are recorded in the parser evidence.

Validated: full default stage 3 oracle 106,367 passing, zero failing/pending,
empty baseline diff, 243.492 seconds. Only the existing mechanically sanctioned
API snapshot is accepted; the post-run exact attribution check passes. The
parser dump retains its original hash; removing the view restores its diagnostic
and the planted runtime edit fails the JavaScript equality guard.
