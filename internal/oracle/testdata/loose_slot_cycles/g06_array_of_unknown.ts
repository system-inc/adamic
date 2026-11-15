interface Node { name: string; items: unknown[] | undefined }
const a: Node = { name: "a", items: undefined };
const record: Record<string, Node> = { k: a };
a.items = [record];
console.log(String(a.items.length));
