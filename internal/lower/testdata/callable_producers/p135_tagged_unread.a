interface Left { readonly kind: 'left'; readonly run: (value: number) => number; }
interface Right { readonly kind: 'right'; readonly count: number; }
interface Holder { readonly kind: 'Holder'; readonly value: Left | Right; }
const inner = { kind: 'left' as const, run(value: number): number { return value + 1; } };
const raw = { kind: 'Holder' as const, value: inner };
const base: { readonly kind: 'Holder' } = raw;
const value = (base as Holder).value;
console.log(`${value === inner}`);
