interface Node { name: string; run: (() => boolean) | undefined }
function link(node: Node): void {
    const held: unknown = node;
    node.run = () => held !== undefined;
}
const a: Node = { name: "a", run: undefined };
link(a);
console.log(String(a.run !== undefined));
