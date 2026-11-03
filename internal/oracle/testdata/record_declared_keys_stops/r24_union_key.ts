interface Options { outFile?: string; [name: string]: string | number | undefined; }
type Dict = { [name: string]: string | number | undefined };
function key(): string { return 'outFile'; }
const name: string = key();
function put(o: Options, k: 'outFile' | string): void { o[k] = 3; }
const options: Options = {};
put(options, 'outFile');
console.log(typeof options.outFile);
