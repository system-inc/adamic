const narrow: Map<string, 1 | 2> = new Map<string, 1 | 2>();
function store(values: Map<string, number>, value: number): void { values.set('key', value); }
store(narrow, 2);
console.log((narrow.get('key') ?? 0).toString());
