const narrow: (1 | 2)[] = [1];
function store(values: number[], value: number): void { values.fill(value); }
store(narrow, 2);
console.log((narrow[0] ?? 0).toString());
