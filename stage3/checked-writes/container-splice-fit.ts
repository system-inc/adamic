const narrow: (1 | 2)[] = [1];
function store(values: number[]): void { values.splice(0, 1, 2, 2); }
store(narrow);
console.log(narrow.length.toString());
