interface Options { outFile?: string; [name: string]: string | number | undefined; }
type Dict = { [name: string]: string | number | undefined };
function key(): string { return 'outFile'; }
const name: string = key();
class Holder { readonly d: Dict; constructor(o: Options) { this.d = o; } }
const options: Options = {};
new Holder(options).d.outFile = 3;
console.log(typeof options.outFile);
