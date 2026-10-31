interface Options { outFile?: string; [name: string]: string | number | undefined; }
type Dict = { [name: string]: string | number | undefined };
function key(): string { return 'outFile'; }
const name: string = key();
const options: Options = {};
const d: Dict = options;
const later = (): void => { d.outFile = 3; };
later();
console.log(typeof options.outFile);
