interface Options { outFile?: string; [name: string]: string | number | undefined; }
type Dict = { [name: string]: string | number | undefined };
function key(): string { return 'outFile'; }
const name: string = key();
const o = { outFile: "a", [name]: 3 };
console.log(typeof o.outFile);
