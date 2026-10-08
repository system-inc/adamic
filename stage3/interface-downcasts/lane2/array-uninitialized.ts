interface Base { readonly kind: 'items' | 'other'; }
interface Items extends Base { readonly kind: 'items'; readonly values: readonly number[]; }
function items(node: Base): Items { return node as Items; }
const raw: Items = {kind: 'items', values: undefined!};
console.log(`${items(raw).values.length}`);
