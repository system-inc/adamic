// Language's road 1 on #0b9xqhh: a getter in a subtype answers a string on its first read and a number on its
// second. A checked view that re-read the getter would check one value and use another, so the getter named
// as the view's field stops first, before any read, as "a getter in a checked field contract". Node prints
// first, first, 2.
interface Source { readonly kind: number; }
interface Target { readonly kind: number; readonly label?: string; }

class Shifting implements Source {
    readonly kind = 1;
    private reads = 0;
    get label(): string | number {
        this.reads++;
        return this.reads === 1 ? "first" : 2;
    }
}
function show(target: Target): void {
    const first = target.label;
    console.log(first ?? "absent");
    console.log(first ?? "absent");
    console.log(target.label ?? "absent");
}
const source: Source = new Shifting();
show(source);
