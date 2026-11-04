interface Options { readonly overrideImportMode?: number; readonly other?: string; }
function read<T extends Options>(options: T): string { return String(options.overrideImportMode); }
const wide = { other: "o", overrideImportMode: "lie" };
const narrow: { readonly other: string } = wide;
console.log(read<{ readonly other: string }>(narrow));
