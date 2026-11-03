interface Options { outFile?: string; [name: string]: string | number | undefined; }
type Dict = { [name: string]: string | number | undefined };
function key(): string { return 'outFile'; }
const name: string = key();
const put = <T extends Dict>(target: T, value: number): void => { const wide: Dict = target; wide[name] = value; };
const options: Options = {};
put(options, 3);
console.log(typeof options.outFile);
