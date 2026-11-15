class Node {
    name = "a";
    box: { tag: number } | undefined = undefined;
}
const a = new Node();
const holder = { tag: 1, n: a };
a.box = holder;
console.log(a.name);
