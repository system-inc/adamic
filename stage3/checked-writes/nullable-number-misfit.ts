const narrow: { readonly count: 1 | 2 | undefined } = { count: 1 };
function store(view: { count: number | undefined }, value: number | undefined): void { view.count = value; }
store(narrow, 3);
console.log('stored');
