interface Node { name: string; box: { n: Node } | undefined }
const a: Node = { name: "a", box: undefined };
a.box = { n: a };
console.log(a.box.n.name);
