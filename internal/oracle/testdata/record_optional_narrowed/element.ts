// #70kdt66: ?.[0] on a narrowed record value a call removed (Compiler's road 2).
const r: Record<string, string[]> = {};
function clear(): void { delete r["k"]; }
r["k"] = ["a"];
clear();
try { console.log(String(r["k"]?.[0])); } catch (error) { console.log("caught"); }
r["k"] = ["b"];
console.log(String(r["k"]?.[0]));
