// An instance of a class whose base declares name holds its slot, so the write through Element's checked
// view is admitted (#w8nanzw), and agrees with Node.
interface Node { readonly kind: number; readonly pos: number; }
interface Declaration extends Node { _declarationBrand: undefined; }
interface Element extends Declaration { _elementBrand: undefined; name?: string; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
function rename(n: Node): void { if (isElement(n)) { n.name = "w"; console.log(String(n.name)); } }
class Base { readonly kind = 171; readonly pos = 1; name?: string; }
class Derived extends Base { label = "l"; }
const derived = new Derived();
rename(derived);
console.log(String(derived.name));
