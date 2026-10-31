interface Options { outFile?: string; [name: string]: string | number | undefined; }
type Dict = { [name: string]: string | number | undefined };
function key(): string { return 'outFile'; }
const name: string = key();
const dict: Dict = { outFile: 3 };
const options: Options = { outFile: 'a', ...dict };
console.log(typeof options.outFile);
