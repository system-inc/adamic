// #70kdt66: ?.push on a narrowed record value a call removed; Node pushes nothing and prints pushed.
const r: Record<string, number[]> = {};
function clear(): void { delete r["k"]; }
r["k"] = [];
clear();
try { r["k"]?.push(1); console.log("pushed"); } catch (error) { console.log("caught"); }
console.log(String(Object.keys(r).length));
