type Target = ((value: number) => number) | ((value: string | number) => number);
interface Root { readonly kind: 'Receiver'; }
interface Receiver extends Root { readonly value: Target; readonly count: number | string; }
const raw = { kind: 'Receiver' as const, value: (value: number): number => value * 3, count: 3 };
const base: Root = raw;
const view = base as Receiver;
console.log(`${view.count}`);
console.log(`${view.value(4)}`);
