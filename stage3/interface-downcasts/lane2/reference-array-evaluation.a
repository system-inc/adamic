interface Base { readonly kind: 'items' | 'other'; }
interface Entry { readonly value: number; readonly text: string; }
interface Items extends Base { readonly kind: 'items'; readonly values: Entry[]; }
function items(node: Base): Items { return node as Items; }
function entry(value: number): Entry { return {value, text: `item${value}`}; }
let calls = 0;
const raw = {kind: 'items' as const, values: [entry(1)]};
function array(): Entry[] { calls++; return items(raw).values; }
function index(): number { calls++; return 0; }
function incoming(): Entry { calls++; return entry(2); }
array()[index()] = incoming();
console.log(`${calls}:${raw.values[0]!.text}`);
