// The third condition on #6c4swvt: a named read of a field held as a union, kind?: string | number, is the slot's box
// handed back unchecked, so a boolean written under kind through OptionsBase would read as a string | number. The
// view keeps kind's contract, and stops not yet at the existing storage stop.
interface Options {
    kind?: string | number;
    [option: string]: string | number | boolean | undefined;
}
interface OptionsBase {
    [option: string]: string | number | boolean | undefined;
}
function count(options: OptionsBase): number {
    let keys = 0;
    for (const name in options) {
        keys += name.length > 0 ? 1 : 0;
    }
    return keys;
}
const options: Options = { kind: 1 };
console.log(`${count(options)}`);
