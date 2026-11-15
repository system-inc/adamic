interface Node { name: string; box: unknown | undefined }
const a: Node = { name: "a", box: undefined };
const record: Record<string, Node> = { k: a };
a.box = record;
console.log(a.name);
