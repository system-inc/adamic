interface Options { outFile?: string; [name: string]: string | number | undefined; }
type Dict = { [name: string]: string | number | undefined };
function key(): string { return 'outFile'; }
const name: string = key();
function widen<T extends Options>(o: T): Dict { return o; }
const options: Options = {};
widen(options).outFile = 3;
console.log(typeof options.outFile);
