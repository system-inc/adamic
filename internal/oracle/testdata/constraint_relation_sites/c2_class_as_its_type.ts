interface Options { readonly overrideImportMode?: number; readonly other?: string; }
class Exact { readonly other = "o"; readonly overrideImportMode = 3; }
function read<T extends Options>(options: T): string { return String(options.overrideImportMode); }
const exact: Exact = new Exact();
console.log(read(exact));
console.log(read(new Exact()));
