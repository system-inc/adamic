// emitter 1868's shape, lying: the second tag's parent is neither a Doc nor a TypeLiteral, so the read of its
// kind through the view stops there, after the honest call has printed what Node prints.
enum Kind { Root, Doc, Tag, Param, Prop, Name, Computed, Literal, Other }
interface Node { readonly kind: Kind; readonly parent: Node | undefined; }
interface Doc extends Node { readonly kind: Kind.Doc; }
interface TypeLiteral extends Node { readonly kind: Kind.Literal; }
interface Tag extends Node { readonly kind: Kind.Tag | Kind.Param | Kind.Prop; readonly parent: Doc | TypeLiteral; }
function owner(node: Node): Kind {
    const tag = node as Tag;
    return tag.parent.kind;
}
const root: Node = { kind: Kind.Root, parent: undefined };
const doc: Node = { kind: Kind.Doc, parent: root };
const other: Node = { kind: Kind.Other, parent: root };
const documented: Node = { kind: Kind.Tag, parent: doc };
const lying: Node = { kind: Kind.Tag, parent: other };
console.log(owner(documented));
console.log(owner(lying));
