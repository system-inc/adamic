interface Options { outFile?: string; [name: string]: string | number | undefined; }
type Dict = { [name: string]: string | number | undefined };
function key(): string { return 'outFile'; }
const name: string = key();
const options: Options = { outFile: 'a' };
for (const k in options) options[k] = 3;
console.log(typeof options.outFile);
