interface Options { outFile?: string; [name: string]: string | number | undefined; }
type Dict = { [name: string]: string | number | undefined };
function key(): string { return 'outFile'; }
const name: string = key();
function read<T extends Options>(options: T): string { return typeof options.outFile; }
function made(): Dict { const dict: Dict = {}; dict[name] = 3; return dict; }
console.log(read(made()));
