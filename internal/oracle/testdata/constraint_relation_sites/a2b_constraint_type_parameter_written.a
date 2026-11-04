interface Options { readonly overrideImportMode?: number; readonly other?: string; }
type Narrow = { readonly other: string };
function read<U extends Options, T extends U>(options: T): string { return String(options.overrideImportMode); }
const wide = { other: "o", overrideImportMode: "lie" };
const narrow: Narrow = wide;
console.log(read<Narrow, Narrow>(narrow));
