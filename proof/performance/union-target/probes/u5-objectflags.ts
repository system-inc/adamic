interface Type { readonly flags: number }
interface ObjectType extends Type { readonly objectFlags: number }
interface UnionType extends Type { readonly types: readonly Type[]; readonly objectFlags: number }
interface IntrinsicType extends Type { readonly intrinsicName: string }
function flagsOf(t: ObjectType | UnionType | IntrinsicType): number { return (t as ObjectType | UnionType).objectFlags; }
console.log(String(flagsOf({ flags: 1, objectFlags: 4 })));
console.log(String(flagsOf({ flags: 2, types: [], objectFlags: 8 })));
