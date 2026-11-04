interface Options { readonly overrideImportMode?: number; readonly other?: string; }
function read<T extends { readonly inner: Options }>(holder: T): string { return String(holder.inner.overrideImportMode); }
const wide = { other: "o", overrideImportMode: "lie" };
const narrow: { readonly other: string } = wide;
console.log(read({ inner: narrow }));
