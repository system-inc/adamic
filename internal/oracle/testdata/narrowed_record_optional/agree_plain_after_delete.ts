// No ?.: Node throws a TypeError reading length of undefined too, and the catch takes it on both.
const r: Record<string, number[]> = {};
function clear(): void { delete r["k"]; }
r["k"] = [];
clear();
try { console.log(String(r["k"].length)); } catch (error) { console.log("caught"); }
console.log("after");
