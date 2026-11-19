// An instance of a class declaring name holds its slot, optional or not, so the write through Element's
// checked view is admitted (#w8nanzw), and agrees with Node.
interface Node { readonly kind: number; readonly pos: number; }
interface Declaration extends Node { _declarationBrand: undefined; }
interface Element extends Declaration { _elementBrand: undefined; name?: string; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
function rename(n: Node): void { if (isElement(n)) { n.name = "w"; console.log(String(n.name)); } }
class Box { readonly kind = 171; readonly pos = 1; name?: string; label = "l"; }
const box = new Box();
rename(box);
console.log(String(box.name));
