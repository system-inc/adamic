interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
interface Resolver { isLateBound(node: Node): node is Identifier; }
const resolver: Resolver = { isLateBound: (node: Node): node is Identifier => node.kind === 80 };
function show(r: Resolver, node: Node): void { if (r.isLateBound(node)) { console.log(node.escapedText); } else { console.log("no"); } }
show(resolver, { kind: 11 } as Node);
show(resolver, { kind: 80, escapedText: "x" } as Node);
