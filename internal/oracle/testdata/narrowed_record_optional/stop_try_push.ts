const r: Record<string, number[]> = {};
function clear(): void { delete r["k"]; }
r["k"] = [];
clear();
try { r["k"]?.push(1); console.log("pushed"); } catch (error) { console.log("caught"); }
console.log("after");
