// #6c4swvt piece 4: a readonly record may hold a narrower record, so the way back to a writable one is a widening.
// Seen as OptionsBase, the WatchOptions this view holds could be written a Source under any key, which its own
// index reads as Value. tsc allows the assignment; Adamic refuses it, so the view never hands out a writable alias.
type Value = string | number | boolean | undefined;
interface Source {
    fileName: string;
}
interface WatchOptions {
    watchFile?: number;
    [option: string]: Value;
}
interface OptionsBase {
    [option: string]: Value | Source;
}
function writable(options: Readonly<OptionsBase>): OptionsBase {
    return options;
}
const watch: WatchOptions = { watchFile: 4 };
const alias = writable(watch);
alias["k"] = { fileName: "x" };
console.log(typeof watch["k"]);
