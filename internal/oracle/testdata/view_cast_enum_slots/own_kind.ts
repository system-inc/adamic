// binder 3369, moduleSpecifiers 915, parser 10804 and 10805, ts.ts 2191: an overlap cast, from a union one
// member of which upcasts to the target, whose own kind is the member slot. No test at the cast; the kind is
// checked at each read through the view, so a Wide that holds the name's fields reads through it.
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
const wide = { kind: Kind.Name, parent: undefined, flags: 0, text: "width" };
console.log(text(identifier));
console.log(text(wide));
