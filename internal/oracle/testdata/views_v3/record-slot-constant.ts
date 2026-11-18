interface Options { target?: number; lib: string[]; rootDirs?: string[]; [option: string]: string | number | boolean | string[] | undefined; }
function each(options: Options): string {
    let out = "";
    for (const d of options.rootDirs ?? []) out += d.toUpperCase() + ",";
    return out;
}
const key: string = "rootDirs";
const other: Options = { lib: [], rootDirs: ["a"], [key]: "zz" };
console.log(each(other));
