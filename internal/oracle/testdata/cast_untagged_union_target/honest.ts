// Union target (a), #b5w3ycg: an untagged union cast to a union of its members is a checked view of the target union,
// read through the fields every target member declares in one representation. Each member's value reads through it
// as Node reads it, and so does tsc's objectFlags shape (utilities.ts:8280).
interface A { readonly a: number; readonly shared: string }
interface B { readonly b: string; readonly shared: string }
interface C { readonly c: boolean }
function shared(x: A | B | C): string { return (x as A | B).shared; }
console.log(shared({ a: 1, shared: "from a" }));
console.log(shared({ b: "b", shared: "from b" }));
interface Type { readonly flags: number }
interface ObjectType extends Type { readonly objectFlags: number }
interface UnionType extends Type { readonly types: readonly Type[]; readonly objectFlags: number }
interface IntrinsicType extends Type { readonly intrinsicName: string }
function flagsOf(t: ObjectType | UnionType | IntrinsicType): number { return (t as ObjectType | UnionType).objectFlags; }
console.log(String(flagsOf({ flags: 1, objectFlags: 4 })));
console.log(String(flagsOf({ flags: 2, types: [], objectFlags: 8 })));
