interface Options { target?: number; lib: string[]; [option: string]: string | number | boolean | string[] | (string | number)[] | undefined; }
const keys: string[] = ["lib", "other"];
const key = keys[0] ?? "other";
const alias: (string | number)[] = ["a", 1];
const opts: Options = { target: 9, lib: ["es2020"], [key]: alias };
console.log('before');
const second = opts.lib[1];
console.log('after');
