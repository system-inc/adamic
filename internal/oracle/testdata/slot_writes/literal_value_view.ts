// A literal that wrote name, built at Named, holds its slot as Element types it, so the write through Element's
// checked view is admitted (#w8nanzw), and agrees with Node.
interface Node { readonly kind: number; readonly pos: number; }
interface Declaration extends Node { _declarationBrand: undefined; }
interface Element extends Declaration { _elementBrand: undefined; name?: string; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
function rename(n: Node): void { if (isElement(n)) { n.name = "w"; console.log(String(n.name)); } }
interface Named extends Node { name?: string; }
const raw: Named = { kind: 171, pos: 1, name: "a" };
rename(raw);
