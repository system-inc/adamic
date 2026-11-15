// #70kdt66: ?.join on a narrowed record value a call removed.
const r: Record<string, number[]> = {};
function clear(): void { delete r["k"]; }
r["k"] = [];
clear();
try { console.log(String(r["k"]?.join(","))); } catch (error) { console.log("caught"); }
r["k"] = [1, 2];
console.log(String(r["k"]?.join(",")));
