interface Node { name: string; next: Node | undefined }
function describe(value: unknown): string {
    return typeof value;
}
const a: Node = { name: "a", next: undefined };
console.log(describe(a) + " " + a.name);
