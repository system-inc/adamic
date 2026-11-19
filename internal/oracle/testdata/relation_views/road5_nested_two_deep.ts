// Language's road 5 and Lower's first row on #0b9xqhh: the absent optional field is object-typed, and it's read
// two deep. The view's contract checks inner's own fields lazily, at each read through Inner (viewSchema's
// nested contract), never once at the crossing: the absent and honest reads print, and the lying text stops at
// its read (exit 70) where Node prints 2.
interface Source { readonly kind: number; }
interface Inner { readonly text: string; }
interface Target { readonly kind: number; readonly inner?: Inner; }

function text(target: Target): string {
    return target.inner === undefined ? "absent" : target.inner.text;
}
const bare: Source = { kind: 1 };
const honest = { kind: 2, inner: { text: "two" } };
const lying = { kind: 3, inner: { text: 3 } };
const truthful: Source = honest;
const liar: Source = lying;
console.log(text(bare));
console.log(text(truthful));
console.log(text(liar));
