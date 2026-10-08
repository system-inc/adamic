const empty: never[] = [];
const broad: number[] = [1];
function select(flag: boolean) { return flag ? empty : broad; }
function store(values: number[], insert: boolean): void { if (insert) { values.push(1); } }
store(select(true), false);
console.log(empty.length.toString());
store(select(false), true);
console.log(broad.length.toString());
