interface Options { outFile?: string; [name: string]: string | number | undefined; }
type Dict = { [name: string]: string | number | undefined };
function key(): string { return 'outFile'; }
const name: string = key();
const options: Options = {};
const d: Dict = options;
d.outFile = 3;
console.log(typeof options.outFile);
