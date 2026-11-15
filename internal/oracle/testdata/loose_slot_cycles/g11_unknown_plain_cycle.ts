interface Node { name: string; box: unknown | undefined }
const a: Node = { name: "a", box: undefined };
const holder: { n: Node } = { n: a };
a.box = holder;
console.log(a.name);
