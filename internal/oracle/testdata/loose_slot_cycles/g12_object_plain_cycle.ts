interface Node { name: string; box: object | undefined }
const a: Node = { name: "a", box: undefined };
const holder: { n: Node } = { n: a };
a.box = holder;
console.log(a.name);
