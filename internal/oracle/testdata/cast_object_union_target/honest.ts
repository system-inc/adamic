// Cause U of #eg3krrp: one plain object cast to a union target no literal tag tells apart is union target (a)'s
// checked view, the source taken as the union's one member. Each value reads through it as Node reads it: tsc's
// objectFlags shape (utilities.ts:8319), and a build info shape whose members hold no array field.
interface Type { readonly flags: number }
interface ObjectType extends Type { readonly objectFlags: number }
interface UnionType extends Type { readonly types: readonly Type[]; readonly objectFlags: number }
function flagsOf(t: Type): number { return (t as ObjectType | UnionType).objectFlags; }
const object: ObjectType = { flags: 1, objectFlags: 4 };
const union: UnionType = { flags: 2, types: [], objectFlags: 8 };
console.log(String(flagsOf(object)));
console.log(String(flagsOf(union)));
interface BuildInfo { readonly version: string }
interface BundleInfo extends BuildInfo { readonly outSignature?: string; readonly errorCount: number }
interface ProjectInfo extends BuildInfo { readonly root: number; readonly errorCount: number }
function errorsOf(info: BuildInfo): number { return (info as BundleInfo | ProjectInfo).errorCount; }
const bundle: BundleInfo = { version: "1", errorCount: 1 };
const project: ProjectInfo = { version: "1", root: 0, errorCount: 2 };
console.log(String(errorsOf(bundle)));
console.log(String(errorsOf(project)));
