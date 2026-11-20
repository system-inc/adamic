// binder 3078 and emitter 1888's shape, lying: the second tag's parent is no Doc, so the read of its kind
// through the view stops there, after the honest call has printed what Node prints.
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
const other: Node = { kind: Kind.Other, parent: root };
const parameter = { kind: Kind.Param, parent: doc, name: "size" };
const lying = { kind: Kind.Param, parent: other, name: "size" };
console.log(owner(parameter));
console.log(owner(lying));
