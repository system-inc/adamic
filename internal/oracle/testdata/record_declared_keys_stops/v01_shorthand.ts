interface Options { outFile?: string; [name: string]: string | number | undefined; }
type Dict = { [name: string]: string | number | undefined };
function key(): string { return 'outFile'; }
const name: string = key();
const options: Options = {};
const outer: { readonly options: Dict } = { options };
outer.options.outFile = 3;
console.log(typeof options.outFile);
