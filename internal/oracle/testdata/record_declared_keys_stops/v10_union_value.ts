interface Options { outFile?: string; [name: string]: string | number | undefined; }
type Dict = { [name: string]: string | number | undefined };
function key(): string { return 'outFile'; }
const name: string = key();
const options: Options = {};
const v: string | number = name.length > 3 ? 3 : "a";
options[name] = v;
console.log(typeof options.outFile);
