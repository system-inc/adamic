// Lower's risk row, lying: the Doc's owner is neither a Statement nor an Identifier, so the read of its kind
// through the view stops there, after the honest call has printed what Node prints.
enum Kind { Root, Doc, Tag, Param, Prop, Name, Computed, Literal, Other }
interface Node { readonly kind: Kind; readonly parent: Node | undefined; }
interface Statement extends Node { readonly kind: Kind.Root; }
interface Identifier extends Node { readonly kind: Kind.Name; }
type HasDoc = Statement | Identifier;
interface Tags extends ReadonlyArray<Node> { readonly hasTrailingComma: boolean; }
interface Doc extends Node { readonly kind: Kind.Doc; readonly parent: HasDoc; readonly tags: Tags | undefined; }
interface Tag extends Node { readonly kind: Kind.Param | Kind.Prop; readonly parent: Doc; }
function owners(node: Node): string {
    const tag = node as Tag;
    return `${tag.parent.kind} ${tag.parent.parent.kind}`;
}
const statement: Node = { kind: Kind.Root, parent: undefined };
const other: Node = { kind: Kind.Other, parent: undefined };
const doc = { kind: Kind.Doc, parent: statement, tags: undefined };
const lyingDoc = { kind: Kind.Doc, parent: other, tags: undefined };
const tag = { kind: Kind.Param, parent: doc };
const lying = { kind: Kind.Param, parent: lyingDoc };
console.log(owners(tag));
console.log(owners(lying));
