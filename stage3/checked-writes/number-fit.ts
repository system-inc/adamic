const narrow: { readonly count: 1 | 2 } = { count: 1 };
const other: { readonly count: 7 } = { count: 7 };
function store(view: { count: number }, value: number): void { view.count = value; }
store(narrow, 2);
store(other, 7);
console.log(narrow.count.toString() + ' ' + other.count.toString());
