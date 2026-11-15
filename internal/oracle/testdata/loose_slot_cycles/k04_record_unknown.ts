interface Node { name: string; bag: Record<string, unknown> }
const a: Node = { name: "a", bag: {} };
a.bag["n"] = a;
console.log(a.name);
