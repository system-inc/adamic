interface Options { outFile?: string; [name: string]: string | number | undefined; }
type Dict = { [name: string]: string | number | undefined };
function key(): string { return 'outFile'; }
const name: string = key();
function maker(): <T extends Dict>(target: T, value: number) => void {
    return <T extends Dict>(target: T, value: number): void => { const wide: Dict = target; wide[name] = value; };
}
const put: (target: Options, value: number) => void = maker();
const options: Options = {};
put(options, 3);
console.log(typeof options.outFile);
