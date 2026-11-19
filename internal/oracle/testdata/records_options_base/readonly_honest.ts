// #6c4swvt piece 4: a record seen as Readonly<OptionsBase>, tsc's serializeOptionBaseObject's parameter. A
// readonly string-index type the program first puts a record into is held as that record (records_readonly.go),
// and nothing writes through its index, so the view may widen what the values are read as. Reads by name, by
// key and by for...in agree with Node.
type Value = string | number | boolean | string[] | undefined;
interface Source {
    fileName: string;
    text: string;
}
interface WatchOptions {
    watchFile?: number;
    excludeFiles?: string[];
    [option: string]: Value;
}
interface OptionsBase {
    [option: string]: Value | Source;
}
function serialize(options: Readonly<OptionsBase>): string {
    const rows: string[] = [];
    for (const name in options) {
        const value = options[name];
        rows.push(name + "=" + typeof value);
    }
    return rows.join(",") + ":" + String(options["watchFile"] === 4);
}
const watch: WatchOptions = { watchFile: 4 };
watch.excludeFiles = ["a.ts"];
watch["synchronous"] = true;
console.log(serialize(watch));
watch.watchFile = 2;
console.log(serialize(watch));
