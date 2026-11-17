interface Options { target?: number; lib: string[]; [option: string]: string | number | boolean | string[] | (string | undefined)[] | undefined; }
const keys: string[] = ["lib", "other"];
const key = keys[0] ?? "other";
const holey: (string | undefined)[] = ["a", undefined];
const opts: Options = { target: 9, lib: ["es2020"], [key]: holey };
console.log('before');
const second = opts.lib[1];
console.log('after');
