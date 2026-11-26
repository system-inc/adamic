const s = Symbol("hidden");
const base = { [s]: 1, shown: 2 };
const spread = { ...base, more: 3 };
console.log(Object.keys(spread).join(",") + " " + String(Object.entries(spread).length));
