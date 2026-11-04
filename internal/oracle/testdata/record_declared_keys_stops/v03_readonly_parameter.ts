interface Options { outFile?: string; [name: string]: string | number | undefined; }
type Dict = { [name: string]: string | number | undefined };
function key(): string { return 'outFile'; }
const name: string = key();
function widen(ro: Readonly<Options>): void { const d: Dict = ro; d.outFile = 3; }
const options: Options = {};
widen(options);
console.log(typeof options.outFile);
