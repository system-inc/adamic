const narrow: { readonly count: 1 | 2 } = { count: 1 };
function store(view: { count: number }, value: number): void { view.count = value; }
store(narrow, 3);
console.log('stored');
