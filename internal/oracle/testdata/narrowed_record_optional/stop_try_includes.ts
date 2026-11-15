const r: Record<string, number[]> = {};
function clear(): void { delete r["k"]; }
r["k"] = [];
clear();
try { console.log(String(r["k"]?.includes(1))); } catch (error) { console.log("caught"); }
console.log("after");
