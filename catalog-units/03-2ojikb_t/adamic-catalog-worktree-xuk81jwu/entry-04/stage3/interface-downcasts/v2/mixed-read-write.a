interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: string | number | {readonly text: string}; }
const raw = { kind: 'Holder' as const, value: 'first' };
const base: Root = raw;
const viewed = base as View;
const first = viewed.value;
console.log(typeof first === 'string' ? first : 'wrong');
raw.value = 'second';
const second = viewed.value;
console.log(typeof second === 'string' ? second : 'wrong');
