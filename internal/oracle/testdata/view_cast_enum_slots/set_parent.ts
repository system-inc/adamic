// Lower's risk row: a checked parent meeting a setParent-style write elsewhere in the program. The draft's
// parent is written after construction, through the draft's own writable field, never through the view, and
// the read through the view checks what the write stored.
enum Kind { Root, Doc, Tag, Param, Prop, Name, Computed, Literal, Other }
interface Node { readonly kind: Kind; readonly parent: Node | undefined; }
interface Doc extends Node { readonly kind: Kind.Doc; }
interface Draft { readonly kind: Kind; parent: Node | undefined; }
interface PropertyLikeTag extends Node { readonly kind: Kind.Param | Kind.Prop; readonly parent: Doc; }
function setParent(child: Draft, parent: Node): Draft {
    child.parent = parent;
    return child;
}
function owner(node: Node): Kind {
    const tag = node as PropertyLikeTag;
    return tag.parent.kind;
}
const root: Node = { kind: Kind.Root, parent: undefined };
const doc: Node = { kind: Kind.Doc, parent: root };
const draft: Draft = { kind: Kind.Param, parent: undefined };
console.log(`${owner(setParent(draft, doc))}`);
