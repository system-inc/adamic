// A cast from a union of plain objects no literal tag tells apart, to one member or narrower, is a checked view of
// the target with no test at the cast (#axqgdck, rulings.md:18). Each shape is one of tsc's first stops, held to Node:
// a value of the cast's member reads what it holds, a value of another member reads what it shares, and the value
// keeps its identity, in a comparison and as a Map key.
interface PackageJsonInfo { readonly packageDirectory: string; readonly contents: { readonly name: string } }
interface MissingPackageJsonInfo { readonly packageDirectory: string; readonly directoryExists: boolean }
type PackageJsonInfoCacheEntry = PackageJsonInfo | MissingPackageJsonInfo;

// moduleNameResolver.ts:931
function contentsOf(entry: PackageJsonInfoCacheEntry): string {
    return (entry as PackageJsonInfo).contents.name;
}
// A value of the other member, read through the view only where the two share a field.
function directoryOf(entry: PackageJsonInfoCacheEntry): string {
    return (entry as PackageJsonInfo).packageDirectory;
}
// absentArm's presence check first, then the view.
function presentContents(entry: PackageJsonInfo | MissingPackageJsonInfo | undefined): string {
    return (entry as PackageJsonInfo).contents.name;
}

// utilitiesPublic.ts:2632, a member out of FileReference | Node.
interface FileReference { readonly pos: number; readonly end: number; readonly fileName: string }
interface Node { readonly pos: number; readonly end: number; readonly kind: number }
function kindOf(location: FileReference | Node): number {
    return (location as Node).kind;
}

// utilitiesPublic.ts:960, a target narrower than a member, whose kind is open on both members.
interface Declaration { readonly kind: number; readonly pos: number }
interface Expression { readonly kind: number; readonly end: number }
interface Identifier extends Expression { readonly escapedText: string }
function textOf(declaration: Declaration | Expression): string {
    return (declaration as Identifier).escapedText;
}

// utilities.ts:11174, three Type-ish members that share flags.
interface NumberLiteralType { readonly flags: number; readonly value: number }
interface StringLiteralType { readonly flags: number; readonly value: string }
interface UniqueESSymbolType { readonly flags: number; readonly escapedName: string }
function symbolName(type: NumberLiteralType | StringLiteralType | UniqueESSymbolType): string {
    return (type as UniqueESSymbolType).escapedName;
}

const found: PackageJsonInfoCacheEntry = { packageDirectory: '/found', contents: { name: 'found-package' } };
const missing: PackageJsonInfoCacheEntry = { packageDirectory: '/missing', directoryExists: false };
console.log(contentsOf(found));
console.log(directoryOf(found));
console.log(directoryOf(missing));
console.log(presentContents(found));
const reference: FileReference | Node = { pos: 1, end: 2, kind: 80 };
console.log(String(kindOf(reference)));
const identifier: Identifier = { kind: 80, end: 7, escapedText: 'name' };
console.log(textOf(identifier));
const symbol: UniqueESSymbolType = { flags: 8192, escapedName: '__@iterator' };
console.log(symbolName(symbol));

// The view is the value itself (Lower's pin 2 on #axqgdck, Language's x10). The cast sits in a function whose
// parameter is the union, since a const's own reads are narrowed to what it was given.
function asInfo(entry: PackageJsonInfoCacheEntry): PackageJsonInfo {
    return entry as PackageJsonInfo;
}
const seen = asInfo(found);
console.log(String(seen === found));
const byInfo = new Map<PackageJsonInfo, number>();
byInfo.set(seen, 7);
console.log(String(byInfo.get(asInfo(found))));
