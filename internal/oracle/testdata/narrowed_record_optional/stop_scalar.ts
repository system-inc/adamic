const r: Record<string, number> = {};
function clear(): void { delete r["k"]; }
r["k"] = 1;
clear();
try { console.log(String(r["k"]?.toFixed(1))); } catch (error) { console.log("caught"); }
console.log("after");
