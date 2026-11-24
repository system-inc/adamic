interface TextRange { readonly pos: number; readonly end: number; }
interface Node extends TextRange { readonly kind: number; readonly modifiers?: NodeArray<Node>; }
interface NodeArray<T> extends ReadonlyArray<T>, TextRange { }
function isGrammarError(parent: Node, child: TextRange): boolean {
    return child === parent.modifiers;
}
const parent: Node = { kind: 1, pos: 0, end: 1 };
console.log(String(isGrammarError(parent, parent)));
