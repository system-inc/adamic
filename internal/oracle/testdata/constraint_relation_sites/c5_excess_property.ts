interface Options { readonly overrideImportMode?: number; readonly other?: string; }
function read<T extends Options>(options: T): string { return String(options.overrideImportMode); }
console.log(read({ overrideImportMode: 8, extra: "x" }));
