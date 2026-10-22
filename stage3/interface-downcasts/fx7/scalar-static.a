interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly value: boolean; }
class StaticValue { static readonly value: number = 7; }
const raw = {kind: 'Holder' as const, value: StaticValue};
const base: Root = raw;
const viewed = base as View;
console.log(String(viewed.value));
