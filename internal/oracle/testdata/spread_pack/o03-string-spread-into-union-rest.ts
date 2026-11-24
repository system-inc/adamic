type A = number | string;
const strs: string[] = ["a", "b"];
function show(...rest: A[]): void { for (const value of rest) console.log(typeof value + ":" + String(value)); }
show(...strs);
const more: string[] = ["c"];
function lead(first: string, ...rest: A[]): void {
    let text = first + ":" + String(rest.length);
    for (const value of rest) text += "," + String(value);
    console.log(text);
}
lead("x", ...strs, "d", ...more);
class Shelf {
    put(...rest: A[]): number { return rest.length; }
}
console.log(String(new Shelf().put(...strs, ...more)));
