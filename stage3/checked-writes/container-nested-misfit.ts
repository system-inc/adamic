const narrow: (1 | 2)[][] = [];
const good: (1 | 2)[] = [1];
const bad: number[] = [3];
function store(values: number[][], value: number[]): void { values.push(value); }
store(narrow, bad);
console.log(narrow.length.toString());
