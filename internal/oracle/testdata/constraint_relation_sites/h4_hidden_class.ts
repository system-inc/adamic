interface Options { readonly overrideImportMode?: number; readonly other?: string; }
class Wide { readonly other = "o"; readonly overrideImportMode = "lie"; }
function read<T extends Options>(options: T): string { return String(options.overrideImportMode); }
const narrow: { readonly other: string } = new Wide();
console.log(read(narrow));
