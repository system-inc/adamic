interface Options { outFile?: string; [name: string]: string | number | undefined; }
type Dict = { [name: string]: string | number | undefined };
function key(): string { return 'outFile'; }
const name: string = key();
const entries: [string, number][] = [[name, 3]];
const options: Options = Object.fromEntries(entries);
console.log(typeof options.outFile);
