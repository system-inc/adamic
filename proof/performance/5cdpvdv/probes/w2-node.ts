interface Node { readonly kind: number; readonly modifiers?: NodeArray<Node>; }
interface NodeArray<T> extends ReadonlyArray<T> { readonly pos: number; }
function same(parent: Node, child: Node): boolean {
    return (child as Node | NodeArray<Node>) === parent.modifiers;
}
const parent: Node = { kind: 1 };
console.log(String(same(parent, parent)));
