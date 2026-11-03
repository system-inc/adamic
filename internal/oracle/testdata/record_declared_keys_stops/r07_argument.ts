interface Options { outFile?: string; [name: string]: string | number | undefined; }
type Dict = { [name: string]: string | number | undefined };
function key(): string { return 'outFile'; }
const name: string = key();
function set(d: Dict): void { d.outFile = 3; }
const options: Options = {};
set(options);
console.log(typeof options.outFile);
