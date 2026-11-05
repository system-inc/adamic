interface Options { outFile?: string; [name: string]: string | number | undefined; }
function key(): string { return 'outFile'; }
const name: string = key();
const options: Options = {};
const source: { x: string | number | undefined } = { x: 3 };
options[name] = source.x;
console.log(typeof options.outFile);
console.log(String(options.outFile));
