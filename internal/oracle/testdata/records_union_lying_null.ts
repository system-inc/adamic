interface Options { outFile?: string; [name: string]: string | number | null | undefined; }
const name: string = "outFile";
const options: Options = { [name]: null };
console.log(String(options.outFile));
