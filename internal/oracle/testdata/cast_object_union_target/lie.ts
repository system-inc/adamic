// The lie: a Type that is neither member cast as ObjectType | UnionType, cause U's checked view with no test at the
// cast (#eg3krrp). It stops at the first read of a field the value lacks, with the view's own message. Node runs on.
interface Type { readonly flags: number }
interface ObjectType extends Type { readonly objectFlags: number }
interface UnionType extends Type { readonly types: readonly Type[]; readonly objectFlags: number }
function flagsOf(t: Type): number { return (t as ObjectType | UnionType).objectFlags; }
const value: Type = { flags: 1 };
console.log("casting");
console.log(String(flagsOf(value)));
