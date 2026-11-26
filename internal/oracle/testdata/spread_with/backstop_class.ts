// An accessor-free class instance through { readonly x: number } copies as Node's spread does: its own fields,
// then k (Adamic's condition on #k7g6f9h, beside backstop_plain.ts).
class Point {
    x: number;
    y: number;
    constructor(x: number, y: number) {
        this.x = x;
        this.y = y;
    }
}
function copy(source: { readonly x: number }): void {
    const made = { ...source, k: 1 };
    console.log(Object.keys(made).join(","));
    console.log(String(JSON.stringify(made)));
}
copy(new Point(1, 2));
console.log("end");
