// emitter 1868: `node as JSDocTag`, whose parent is a union of members, a Doc or a TypeLiteral. Its kind is
// checked at each read through the view against both members' literals.
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
const literal: Node = { kind: Kind.Literal, parent: root };
const documented: Node = { kind: Kind.Tag, parent: doc };
const typed: Node = { kind: Kind.Param, parent: literal };
console.log(owner(documented));
console.log(owner(typed));
