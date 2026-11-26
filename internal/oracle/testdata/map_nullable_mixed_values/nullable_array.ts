function read(present: boolean): number[] | null {
    if (present) return [7];
    return null;
}
type Value = number[] | string | null | undefined;
const map = new Map<string, Value>([["constructor", read(false)]]);
map.set("set", read(false));
const present = read(true);
map.set("present", present);
map.set("undefined", undefined);
for (const key of ["constructor", "set", "present", "undefined", "absent"]) {
    const value = map.get(key);
    if (value === null) console.log("null");
    else if (value === undefined) console.log(map.has(key) ? "stored undefined" : "absent");
    else console.log(value === present ? "same reference" : "other");
}
