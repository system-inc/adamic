type A = number | string;
const strs: string[] = ["a"];
function show(first: string, ...rest: A[]): void { console.log(first + ":" + String(rest.length)); }
show("x", ...strs, 3);
