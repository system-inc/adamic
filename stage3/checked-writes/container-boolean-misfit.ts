const narrow: true[] = [true];
function store(values: boolean[], value: boolean): void { values.push(value); }
store(narrow, false);
console.log((narrow[1] ?? false).toString());
