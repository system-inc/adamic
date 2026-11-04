interface Options { readonly overrideImportMode?: number; readonly other?: string; }
function read<T extends Options>(options: T): string { return String(options.overrideImportMode); }
const options: Options = { overrideImportMode: 4 };
console.log(read(options));
console.log(read<Options>(options));
