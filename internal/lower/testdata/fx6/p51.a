interface Base { readonly kind: 'box' | 'other' }
interface Box extends Base { readonly kind: 'box'; readonly value: number | string }
const raw = {kind: 'box' as const, value: 5};
const base: Base = raw;
const view = base as Box;
const {value} = view;
console.log(typeof value === 'number' ? `n${value}` : 'other');
