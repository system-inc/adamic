// The overlap cast's shape, lying: the Wide holds the name's text but not its kind, so the read of its kind
// through the view stops there, after the honest call has printed what Node prints.
enum Kind { Root, Doc, Tag, Param, Prop, Name, Computed, Literal, Other }
interface Node { readonly kind: Kind; readonly parent: Node | undefined; }
interface Identifier extends Node { readonly kind: Kind.Name; readonly text: string; readonly escaped: string; }
interface Wide extends Node { readonly flags: number; }
interface NameLike extends Node { readonly kind: Kind.Name; readonly text: string; }
function text(node: Identifier | Wide): string {
    const named = node as NameLike;
    return `${named.kind} ${named.text}`;
}
const identifier: Identifier = { kind: Kind.Name, parent: undefined, text: "size", escaped: "size" };
const lying = { kind: Kind.Other, parent: undefined, flags: 0, text: "width" };
console.log(text(identifier));
console.log(text(lying));
