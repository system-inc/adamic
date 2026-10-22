// A producer identical to the first callable member, and assignable to the second, is carried by the first.
interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: ((value: string | number) => string) | ((value: number) => string); }
const value = (value: string | number): string => typeof value === 'number' ? `n${value + 1}` : `s${value.length}`;
const raw = { kind: 'Holder' as const, value };
const base: Root = raw;
const view = base as View;
console.log(view.value(41));
