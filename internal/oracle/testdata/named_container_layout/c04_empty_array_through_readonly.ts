const emptyArray: never[] = [];
function parse(spec: string): number[][] { return [[spec.length]]; }
const all: readonly (readonly number[])[] = "ab".length > 1 ? parse("ab") : emptyArray;
const none: readonly (number | string)[] = emptyArray;
console.log(String(all.length) + " " + String(none.length));
