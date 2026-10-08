const narrow: { readonly count: 1 | 2 | undefined } = { count: 1 };
function store(view: { count: number | undefined }, value: number | undefined): void { view.count = value; }
store(narrow, undefined);
console.log(narrow.count === undefined ? 'true' : 'false');
store(narrow, 2);
if (narrow.count !== undefined) { console.log(narrow.count.toString()); }
