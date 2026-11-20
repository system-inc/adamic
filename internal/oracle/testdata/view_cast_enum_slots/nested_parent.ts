// binder 3078 and emitter 1888: `node as JSDocPropertyLikeTag`, whose parent is one member, a Doc. The cast has
// no test, and the parent's kind is checked at each read through the view.
enum Kind { Root, Doc, Tag, Param, Prop, Name, Computed, Literal, Other }
interface Node { readonly kind: Kind; readonly parent: Node | undefined; }
interface Doc extends Node { readonly kind: Kind.Doc; }
interface PropertyLikeTag extends Node { readonly kind: Kind.Param | Kind.Prop; readonly parent: Doc; readonly name: string; }
function owner(node: Node): string {
    const tag = node as PropertyLikeTag;
    return `${tag.name} ${tag.parent.kind}`;
}
const root: Node = { kind: Kind.Root, parent: undefined };
const doc: Node = { kind: Kind.Doc, parent: root };
const parameter = { kind: Kind.Param, parent: doc, name: "size" };
const property = { kind: Kind.Prop, parent: doc, name: "width" };
console.log(owner(parameter));
console.log(owner(property));
