interface Options { outFile?: string; [name: string]: string | number | undefined; }
const name: string = "outFile";
const options: Options = { [name]: 3 };
console.log(String(options.outFile));
