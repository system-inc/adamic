type Value = string | null | undefined;
const source = new Map<number, Value>([[1, null], [2, undefined], [3, "three"]]);
const copy = new Map<number, Value>(source.entries());
source.clear();
for (const value of copy.values()) {
    if (value === null) console.log("null");
    else if (value === undefined) console.log("undefined");
    else console.log(value);
}
console.log(copy.has(2) && !copy.has(4) ? "presence preserved" : "presence lost");
