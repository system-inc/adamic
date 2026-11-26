const base = { "2": "two", z: "zed" };
const spread = { a: "a", ...base, 1: "one" };
console.log(Object.keys(spread).join(",") + " " + Object.values(spread).join(","));
