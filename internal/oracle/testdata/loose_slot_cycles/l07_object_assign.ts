interface Node { name: string; box: object }
const a: Node = { name: "a", box: {} };
Object.assign(a.box, { n: a });
console.log(a.name);
