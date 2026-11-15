interface Node { name: string; extra: { [key: string]: unknown } }
const a: Node = { name: "a", extra: {} };
a.extra["self"] = a;
console.log(a.name);
