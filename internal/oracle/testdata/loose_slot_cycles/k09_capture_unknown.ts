interface Node { name: string; run: (() => boolean) | undefined }
const a: Node = { name: "a", run: undefined };
const held: unknown = a;
a.run = () => held !== undefined;
console.log(String(a.run()));
