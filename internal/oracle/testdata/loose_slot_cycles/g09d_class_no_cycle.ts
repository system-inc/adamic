class Node {
    name = "a";
    box: { tag: number } | undefined = undefined;
}
const a = new Node();
const holder = { tag: 1, r: { k: "x" } as Record<string, string> };
a.box = holder;
console.log(a.name);
