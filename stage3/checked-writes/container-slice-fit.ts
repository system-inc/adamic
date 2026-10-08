const original: (1 | 2)[] = [1];
const narrow = original.slice();
function store(values: number[], value: number): void { values[0] = value; }
store(narrow, 2);
console.log((narrow[0] ?? 0).toString());
