// A lie (#w8nanzw): lie was built with name, so it holds the slot the write through Element's checked view
// needs, but holds a number there where Element says a string. Its own type says so, so the write is NotYet at
// compile time, by the type each reaching object is built at, never a stop at run time. Node runs on.
interface Node { readonly kind: number; readonly pos: number; }
interface Declaration extends Node { _declarationBrand: undefined; }
interface Element extends Declaration { _elementBrand: undefined; name?: string; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
function rename(n: Node): void { if (isElement(n)) { console.log("writing"); n.name = "w"; console.log("after"); } }
const lie = { kind: 171, pos: 1, name: 5 };
rename(lie);
