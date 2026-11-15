const r: Record<string, number[]> = {};
function clear(): void { delete r["k"]; }
r["k"] = [];
clear();
if (r["k"]?.includes(1)) { console.log("y"); } else { console.log("n"); }
console.log("after");
