const original = new Map<string, 1 | 2>();
original.set('key', 1);
const narrow = [...original.values()];
function store(values: number[], value: number): void { values[0] = value; }
store(narrow, 2);
console.log((narrow[0] ?? 0).toString());
