const r: Record<string, number[]> = {};
function clear(): void { delete r["k"]; }
r["k"] = [];
clear();
console.log(`${r["k"]?.length}`);
console.log("after");
