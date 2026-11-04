interface Options { readonly overrideImportMode?: number; readonly other?: string; }
function read<T extends Options>(options: T): string { return String(options.overrideImportMode); }
console.log(read({ overrideImportMode: 1 }));
console.log(read({ other: "o", overrideImportMode: 2 }));
