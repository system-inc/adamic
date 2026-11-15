// A comparator that keeps its arguments object past the comparison would see the shared argument
// array (#367nnz2) rewritten under it. Lowering refuses the arguments object (rule 2600), so this
// program never reaches a backend.
const kept: IArguments[] = [];
const order = [3, 1, 2].sort(function (l: number, r: number) { kept.push(arguments); return l - r; });
console.log(`${String(order)} ${kept.length} ${String(kept[0]?.[0])} ${String(kept[0]?.[1])}`);
