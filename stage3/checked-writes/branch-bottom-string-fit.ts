const empty: never[] = [];
const broad: string[] = ['outside'.repeat(2)];
function select(flag: boolean) { return flag ? empty : broad; }
function store(values: string[], insert: boolean): void { if (insert) { values.push('outside'.repeat(2)); } }
store(select(true), false);
console.log(empty.length.toString());
store(select(false), true);
console.log(broad.length.toString());
