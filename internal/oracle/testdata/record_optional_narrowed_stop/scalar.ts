// #70kdt66: a scalar record value narrowed present, read through ?.: stopped until a maybe reaches every ?..
const n: Record<string, number> = {};
function clear(): void { delete n["k"]; }
n["k"] = 1.5;
clear();
console.log(String(n["k"]?.toFixed(1)));
