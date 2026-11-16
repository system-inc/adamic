interface Node { name: string; box: unknown }
function hide(value: unknown): unknown {
    return value;
}
const holder: { n: Node | undefined } = { n: undefined };
const a: Node = { name: "a", box: hide(holder) };
holder.n = a;
console.log(a.name);
