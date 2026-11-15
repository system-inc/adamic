// #70kdt66: ?.length on a narrowed record value a call removed.
const r: Record<string, number[]> = {};
function clear(): void { delete r["k"]; }
r["k"] = [];
clear();
try { console.log(String(r["k"]?.length)); } catch (error) { console.log("caught"); }
console.log("after");
