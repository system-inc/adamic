// Lower's risk row: a Node subtype's full view contract, a Doc whose parent is a union of owners (tsc's
// HasJSDoc) and whose tags are a readonly array with a field of its own (tsc's NodeArray). The read either
// agrees with Node or the cast stops; it never reads unchecked.
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
const doc = { kind: Kind.Doc, parent: statement, tags: undefined };
const tag = { kind: Kind.Param, parent: doc };
console.log(owners(tag));
