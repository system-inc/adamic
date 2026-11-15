const r: Record<string, number[]> = {};
function clear(): void { delete r["k"]; }
r["k"] = [];
clear();
try { console.log(String((r["k"])?.length)); } catch (error) { console.log("caught"); }
console.log("after");
