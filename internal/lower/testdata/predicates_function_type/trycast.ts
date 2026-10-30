interface Node { readonly kind: number; }
interface Identifier extends Node { readonly kind: 80; readonly escapedText: string; }
export function tryCast<TOut extends TIn, TIn = any>(value: TIn | undefined, test: (value: TIn) => value is TOut): TOut | undefined { return value !== undefined && test(value) ? value : undefined; }
function isIdentifier(node: Node): node is Identifier { return node.kind === 80; }
const id = tryCast({ kind: 80, escapedText: "x" } as Node, isIdentifier);
console.log(id === undefined ? "none" : id.escapedText);
const no = tryCast({ kind: 11 } as Node, isIdentifier);
console.log(no === undefined ? "none" : no.escapedText);
