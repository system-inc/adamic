const r: Record<string, number[]> = {};
function clear(): void { delete r["k"]; }
r["k"] = [];
clear();
console.log(String(r["k"]?.[0]));
console.log("after");
