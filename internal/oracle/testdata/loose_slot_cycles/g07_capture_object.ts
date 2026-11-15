interface Node { name: string; run: (() => number) | undefined }
const a: Node = { name: "a", run: undefined };
const held: object = { k: a } as Record<string, Node>;
a.run = () => (held === a ? 0 : 1);
console.log(String(a.run()));
