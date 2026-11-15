interface Inner { b: number }
interface Holder { a?: Inner }
const r: Record<string, Holder> = {};
function clear(): void { delete r["k"]; }
r["k"] = { a: { b: 1 } };
clear();
console.log(String(r["k"]?.a?.b));
console.log("after");
