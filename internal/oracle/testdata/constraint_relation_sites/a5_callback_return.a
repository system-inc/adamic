interface Options { readonly overrideImportMode?: number; readonly other?: string; }
function read<T extends Options>(make: () => T): string { return String(make().overrideImportMode); }
const wide = { other: "o", overrideImportMode: "lie" };
const narrow: { readonly other: string } = wide;
console.log(read(() => narrow));
