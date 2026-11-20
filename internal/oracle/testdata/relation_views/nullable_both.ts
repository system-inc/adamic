// rulings.md:72's second slice (#0b9xqhh), nullable on both sides: moduleNameResolver.ts:140's shape,
// r && { path: r.path, extension: r.ext }, of type undefined | { ... }, returned as Resolved | undefined. Each
// side is one object type held as one nullable pointer, so the object arms are viewed and undefined crosses
// unchanged. A Holder's Source returned as Target | undefined the same way is nullable_both_lie.ts's honest
// twin. Every read agrees with Node.
interface Found { readonly path: string; readonly ext: string; }
interface Resolved { readonly path: string; readonly extension: string; readonly packageId?: string; }
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }
interface Holder { readonly source: Source; }

function resolved(r: Found | undefined): Resolved | undefined {
    return r && { path: r.path, extension: r.ext };
}
function inner(holder: Holder | undefined): Target | undefined {
    return holder && holder.source;
}
function showResolved(found: Found | undefined): string {
    const result = resolved(found);
    return result === undefined ? "none" : result.path + result.extension + " " + (result.packageId ?? "no package");
}
function show(target: Target | undefined): string {
    return target === undefined ? "none" : target.label ?? "absent";
}
const honest = { kind: 2, label: "two" };
const bare: Holder = { source: { kind: 1 } };
const labelled: Holder = { source: honest };
console.log(showResolved({ path: "index", ext: ".ts" }));
console.log(showResolved(undefined));
console.log(show(inner(undefined)));
console.log(show(inner(bare)));
console.log(show(inner(labelled)));
