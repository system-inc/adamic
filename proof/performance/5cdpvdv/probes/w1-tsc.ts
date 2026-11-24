interface Node { readonly kind: number; readonly modifiers?: NodeArray<Node>; }
interface NodeArray<T> extends ReadonlyArray<T> { readonly pos: number; }
function isGrammarError(parent: Node, child: Node | NodeArray<Node>): boolean {
    return child === parent.modifiers;
}
const parent: Node = { kind: 1 };
console.log(String(isGrammarError(parent, parent)));
